// Package auth provides authentication middleware and OIDC validation.
package auth

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/config"
)

// Claims represents the claims from a validated OIDC token.
type Claims struct {
	Subject   string
	Issuer    string
	Audience  string
	ExpiresAt int64
}

// JWKS represents a JSON Web Key Set.
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// JWK represents a JSON Web Key.
type JWK struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
	Alg string `json:"alg"`
}

// OIDCDiscovery represents the OIDC discovery document.
type OIDCDiscovery struct {
	Issuer  string `json:"issuer"`
	JWKSURI string `json:"jwks_uri"`
}

// cachedJWKS holds cached JWKS data with expiration.
type cachedJWKS struct {
	keys      map[string]*rsa.PublicKey
	expiresAt time.Time
}

// OIDCValidator validates OIDC tokens.
type OIDCValidator struct {
	config     *config.OIDCConfig
	httpClient *http.Client
	cache      *cachedJWKS
	cacheMu    sync.RWMutex
	cacheTTL   time.Duration
}

// NewOIDCValidator creates a new OIDC validator with the given configuration.
func NewOIDCValidator(cfg *config.OIDCConfig) *OIDCValidator {
	return &OIDCValidator{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		cacheTTL: 1 * time.Hour,
	}
}

// ValidateToken validates an OIDC token and returns the claims.
// It performs the following validation steps:
//   - Parses the JWT token structure
//   - Verifies the RS256 signature using the provider's JWKS
//   - Validates the issuer matches the configured OIDC_ISSUER
//   - Validates the audience matches the configured OIDC_CLIENT_ID
//   - Validates the token has not expired
//
// Returns an error if any validation step fails.
func (v *OIDCValidator) ValidateToken(tokenString string) (*Claims, error) {
	if tokenString == "" {
		return nil, errors.New("token is required")
	}

	// Parse and validate the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Ensure the token uses RSA signing
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Get the key ID from the token header
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, errors.New("token missing key ID (kid)")
		}

		// Fetch the public key
		publicKey, err := v.getPublicKey(kid)
		if err != nil {
			return nil, fmt.Errorf("failed to get public key: %w", err)
		}

		return publicKey, nil
	}, jwt.WithValidMethods([]string{"RS256"}),
		jwt.WithIssuer(v.config.Issuer),
		jwt.WithAudience(v.config.ClientID))

	if err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	// Extract claims
	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("failed to parse claims")
	}

	claims := &Claims{}

	if sub, ok := mapClaims["sub"].(string); ok {
		claims.Subject = sub
	}

	if iss, ok := mapClaims["iss"].(string); ok {
		claims.Issuer = iss
	}

	// Handle audience - can be string or []string
	switch aud := mapClaims["aud"].(type) {
	case string:
		claims.Audience = aud
	case []interface{}:
		if len(aud) > 0 {
			if audStr, ok := aud[0].(string); ok {
				claims.Audience = audStr
			}
		}
	}

	if exp, ok := mapClaims["exp"].(float64); ok {
		claims.ExpiresAt = int64(exp)
	}

	return claims, nil
}

// getPublicKey retrieves an RSA public key by key ID (kid) from the cached JWKS.
// If the cache is expired or the key is not found, it triggers a refresh from the OIDC provider.
// Thread-safe for concurrent access.
func (v *OIDCValidator) getPublicKey(kid string) (*rsa.PublicKey, error) {
	// Check cache first
	v.cacheMu.RLock()
	if v.cache != nil && time.Now().Before(v.cache.expiresAt) {
		if key, ok := v.cache.keys[kid]; ok {
			v.cacheMu.RUnlock()
			return key, nil
		}
	}
	v.cacheMu.RUnlock()

	// Fetch fresh JWKS
	if err := v.refreshJWKS(); err != nil {
		return nil, err
	}

	// Try again after refresh
	v.cacheMu.RLock()
	defer v.cacheMu.RUnlock()

	if v.cache == nil {
		return nil, errors.New("JWKS cache is empty after refresh")
	}

	key, ok := v.cache.keys[kid]
	if !ok {
		return nil, fmt.Errorf("key not found: %s", kid)
	}

	return key, nil
}

// refreshJWKS fetches and caches the JWKS from the OIDC provider.
// It first retrieves the OIDC discovery document to obtain the JWKS URI,
// then fetches the JWKS and parses all RSA public keys.
// Keys are cached with a 1-hour TTL. Thread-safe for concurrent access.
func (v *OIDCValidator) refreshJWKS() error {
	v.cacheMu.Lock()
	defer v.cacheMu.Unlock()

	// Double-check if cache is still valid (another goroutine might have refreshed)
	if v.cache != nil && time.Now().Before(v.cache.expiresAt) {
		return nil
	}

	// Fetch OIDC discovery document
	discoveryURL := strings.TrimSuffix(v.config.Issuer, "/") + "/.well-known/openid-configuration"
	discovery, err := v.fetchDiscovery(discoveryURL)
	if err != nil {
		return fmt.Errorf("failed to fetch discovery document: %w", err)
	}

	// Fetch JWKS
	jwks, err := v.fetchJWKS(discovery.JWKSURI)
	if err != nil {
		return fmt.Errorf("failed to fetch JWKS: %w", err)
	}

	// Parse keys
	keys := make(map[string]*rsa.PublicKey)
	for _, jwk := range jwks.Keys {
		if jwk.Kty != "RSA" {
			continue
		}

		publicKey, err := parseRSAPublicKey(jwk)
		if err != nil {
			continue // Skip invalid keys
		}

		keys[jwk.Kid] = publicKey
	}

	if len(keys) == 0 {
		return errors.New("no valid RSA keys found in JWKS")
	}

	v.cache = &cachedJWKS{
		keys:      keys,
		expiresAt: time.Now().Add(v.cacheTTL),
	}

	return nil
}

// fetchDiscovery fetches the OIDC discovery document from the provider's
// well-known endpoint (/.well-known/openid-configuration) to obtain the JWKS URI.
func (v *OIDCValidator) fetchDiscovery(url string) (*OIDCDiscovery, error) {
	resp, err := v.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("discovery endpoint returned status %d", resp.StatusCode)
	}

	var discovery OIDCDiscovery
	if err := json.NewDecoder(resp.Body).Decode(&discovery); err != nil {
		return nil, fmt.Errorf("failed to decode discovery document: %w", err)
	}

	return &discovery, nil
}

// fetchJWKS fetches the JSON Web Key Set from the provider's JWKS endpoint.
func (v *OIDCValidator) fetchJWKS(url string) (*JWKS, error) {
	resp, err := v.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("JWKS endpoint returned status %d", resp.StatusCode)
	}

	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, fmt.Errorf("failed to decode JWKS: %w", err)
	}

	return &jwks, nil
}

// parseRSAPublicKey converts a JWK (JSON Web Key) with RSA parameters
// into an *rsa.PublicKey for JWT signature verification.
// The JWK must contain base64url-encoded modulus (n) and exponent (e).
func parseRSAPublicKey(jwk JWK) (*rsa.PublicKey, error) {
	// Decode the modulus (n)
	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode modulus: %w", err)
	}
	n := new(big.Int).SetBytes(nBytes)

	// Decode the exponent (e)
	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode exponent: %w", err)
	}

	// Convert exponent bytes to int
	var e int
	for _, b := range eBytes {
		e = e<<8 + int(b)
	}

	return &rsa.PublicKey{
		N: n,
		E: e,
	}, nil
}
