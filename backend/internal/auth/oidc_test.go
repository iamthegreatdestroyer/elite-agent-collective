package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/config"
)

func TestNewOIDCValidator(t *testing.T) {
	cfg := &config.OIDCConfig{
		Issuer:   "https://example.com",
		ClientID: "test-client",
	}

	validator := NewOIDCValidator(cfg)
	if validator == nil {
		t.Fatal("expected non-nil validator")
	}
}

func TestValidateTokenEmpty(t *testing.T) {
	cfg := &config.OIDCConfig{}
	validator := NewOIDCValidator(cfg)

	_, err := validator.ValidateToken("")
	if err == nil {
		t.Error("expected error for empty token")
	}
}

func TestValidateTokenInvalidFormat(t *testing.T) {
	cfg := &config.OIDCConfig{
		Issuer:   "https://example.com",
		ClientID: "test-client",
	}
	validator := NewOIDCValidator(cfg)

	// Test with non-JWT format
	_, err := validator.ValidateToken("not-a-valid-jwt")
	if err == nil {
		t.Error("expected error for invalid token format")
	}
}

func TestValidateTokenMissingKid(t *testing.T) {
	cfg := &config.OIDCConfig{
		Issuer:   "https://example.com",
		ClientID: "test-client",
	}
	validator := NewOIDCValidator(cfg)

	// Create a token without kid in header
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"sub": "test-user",
		"iss": "https://example.com",
		"aud": "test-client",
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	// Don't set kid header

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	_, err = validator.ValidateToken(tokenString)
	if err == nil {
		t.Error("expected error for token without kid")
	}
}

func TestValidateTokenWithMockedJWKS(t *testing.T) {
	// Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	kid := "test-key-id"

	// Create mock JWKS endpoint
	jwksHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jwks := JWKS{
			Keys: []JWK{
				{
					Kty: "RSA",
					Kid: kid,
					Use: "sig",
					Alg: "RS256",
					N:   base64.RawURLEncoding.EncodeToString(privateKey.N.Bytes()),
					E:   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(privateKey.E)).Bytes()),
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(jwks)
	})

	jwksServer := httptest.NewServer(jwksHandler)
	defer jwksServer.Close()

	// Create mock discovery endpoint
	discoveryHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		discovery := OIDCDiscovery{
			Issuer:  jwksServer.URL,
			JWKSURI: jwksServer.URL + "/jwks",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(discovery)
	})

	discoveryServer := httptest.NewServer(discoveryHandler)
	defer discoveryServer.Close()

	cfg := &config.OIDCConfig{
		Issuer:   discoveryServer.URL,
		ClientID: "test-client",
	}
	validator := NewOIDCValidator(cfg)

	// Create a valid token
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"sub": "test-user",
		"iss": discoveryServer.URL,
		"aud": "test-client",
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	token.Header["kid"] = kid

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	claims, err := validator.ValidateToken(tokenString)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if claims == nil {
		t.Fatal("expected non-nil claims")
	}

	if claims.Subject != "test-user" {
		t.Errorf("expected subject 'test-user', got %s", claims.Subject)
	}

	if claims.Issuer != discoveryServer.URL {
		t.Errorf("expected issuer '%s', got %s", discoveryServer.URL, claims.Issuer)
	}
}

func TestValidateTokenExpired(t *testing.T) {
	// Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	kid := "test-key-id"

	// Create mock JWKS endpoint
	jwksHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jwks := JWKS{
			Keys: []JWK{
				{
					Kty: "RSA",
					Kid: kid,
					Use: "sig",
					Alg: "RS256",
					N:   base64.RawURLEncoding.EncodeToString(privateKey.N.Bytes()),
					E:   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(privateKey.E)).Bytes()),
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(jwks)
	})

	jwksServer := httptest.NewServer(jwksHandler)
	defer jwksServer.Close()

	// Create mock discovery endpoint
	discoveryHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		discovery := OIDCDiscovery{
			Issuer:  jwksServer.URL,
			JWKSURI: jwksServer.URL + "/jwks",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(discovery)
	})

	discoveryServer := httptest.NewServer(discoveryHandler)
	defer discoveryServer.Close()

	cfg := &config.OIDCConfig{
		Issuer:   discoveryServer.URL,
		ClientID: "test-client",
	}
	validator := NewOIDCValidator(cfg)

	// Create an expired token
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"sub": "test-user",
		"iss": discoveryServer.URL,
		"aud": "test-client",
		"exp": time.Now().Add(-time.Hour).Unix(), // Expired 1 hour ago
	})
	token.Header["kid"] = kid

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	_, err = validator.ValidateToken(tokenString)
	if err == nil {
		t.Error("expected error for expired token")
	}
}

func TestValidateTokenWrongAudience(t *testing.T) {
	// Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	kid := "test-key-id"

	// Create mock JWKS endpoint
	jwksHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jwks := JWKS{
			Keys: []JWK{
				{
					Kty: "RSA",
					Kid: kid,
					Use: "sig",
					Alg: "RS256",
					N:   base64.RawURLEncoding.EncodeToString(privateKey.N.Bytes()),
					E:   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(privateKey.E)).Bytes()),
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(jwks)
	})

	jwksServer := httptest.NewServer(jwksHandler)
	defer jwksServer.Close()

	// Create mock discovery endpoint
	discoveryHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		discovery := OIDCDiscovery{
			Issuer:  jwksServer.URL,
			JWKSURI: jwksServer.URL + "/jwks",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(discovery)
	})

	discoveryServer := httptest.NewServer(discoveryHandler)
	defer discoveryServer.Close()

	cfg := &config.OIDCConfig{
		Issuer:   discoveryServer.URL,
		ClientID: "test-client",
	}
	validator := NewOIDCValidator(cfg)

	// Create a token with wrong audience
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"sub": "test-user",
		"iss": discoveryServer.URL,
		"aud": "wrong-client", // Wrong audience
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	token.Header["kid"] = kid

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	_, err = validator.ValidateToken(tokenString)
	if err == nil {
		t.Error("expected error for wrong audience")
	}
}

func TestValidateTokenWrongIssuer(t *testing.T) {
	// Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	kid := "test-key-id"

	// Create mock JWKS endpoint
	jwksHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jwks := JWKS{
			Keys: []JWK{
				{
					Kty: "RSA",
					Kid: kid,
					Use: "sig",
					Alg: "RS256",
					N:   base64.RawURLEncoding.EncodeToString(privateKey.N.Bytes()),
					E:   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(privateKey.E)).Bytes()),
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(jwks)
	})

	jwksServer := httptest.NewServer(jwksHandler)
	defer jwksServer.Close()

	// Create mock discovery endpoint
	discoveryHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		discovery := OIDCDiscovery{
			Issuer:  jwksServer.URL,
			JWKSURI: jwksServer.URL + "/jwks",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(discovery)
	})

	discoveryServer := httptest.NewServer(discoveryHandler)
	defer discoveryServer.Close()

	cfg := &config.OIDCConfig{
		Issuer:   discoveryServer.URL,
		ClientID: "test-client",
	}
	validator := NewOIDCValidator(cfg)

	// Create a token with wrong issuer
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"sub": "test-user",
		"iss": "https://wrong-issuer.com", // Wrong issuer
		"aud": "test-client",
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	token.Header["kid"] = kid

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	_, err = validator.ValidateToken(tokenString)
	if err == nil {
		t.Error("expected error for wrong issuer")
	}
}

func TestJWKSCaching(t *testing.T) {
	// Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	kid := "test-key-id"
	jwksCallCount := 0

	// Create mock JWKS endpoint that counts calls
	jwksHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jwksCallCount++
		jwks := JWKS{
			Keys: []JWK{
				{
					Kty: "RSA",
					Kid: kid,
					Use: "sig",
					Alg: "RS256",
					N:   base64.RawURLEncoding.EncodeToString(privateKey.N.Bytes()),
					E:   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(privateKey.E)).Bytes()),
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(jwks)
	})

	jwksServer := httptest.NewServer(jwksHandler)
	defer jwksServer.Close()

	// Create mock discovery endpoint
	discoveryHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		discovery := OIDCDiscovery{
			Issuer:  jwksServer.URL,
			JWKSURI: jwksServer.URL + "/jwks",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(discovery)
	})

	discoveryServer := httptest.NewServer(discoveryHandler)
	defer discoveryServer.Close()

	cfg := &config.OIDCConfig{
		Issuer:   discoveryServer.URL,
		ClientID: "test-client",
	}
	validator := NewOIDCValidator(cfg)

	// Create valid tokens
	createToken := func() string {
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
			"sub": "test-user",
			"iss": discoveryServer.URL,
			"aud": "test-client",
			"exp": time.Now().Add(time.Hour).Unix(),
		})
		token.Header["kid"] = kid

		tokenString, err := token.SignedString(privateKey)
		if err != nil {
			t.Fatalf("failed to sign token: %v", err)
		}
		return tokenString
	}

	// Validate multiple tokens - JWKS should only be fetched once due to caching
	for i := 0; i < 5; i++ {
		_, err := validator.ValidateToken(createToken())
		if err != nil {
			t.Fatalf("unexpected error on validation %d: %v", i, err)
		}
	}

	// JWKS should only have been fetched once
	if jwksCallCount != 1 {
		t.Errorf("expected JWKS to be fetched 1 time, but was fetched %d times", jwksCallCount)
	}
}

func TestParseRSAPublicKey(t *testing.T) {
	// Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	jwk := JWK{
		Kty: "RSA",
		Kid: "test-key",
		Use: "sig",
		Alg: "RS256",
		N:   base64.RawURLEncoding.EncodeToString(privateKey.N.Bytes()),
		E:   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(privateKey.E)).Bytes()),
	}

	publicKey, err := parseRSAPublicKey(jwk)
	if err != nil {
		t.Fatalf("failed to parse public key: %v", err)
	}

	if publicKey.N.Cmp(privateKey.N) != 0 {
		t.Error("modulus mismatch")
	}

	if publicKey.E != privateKey.E {
		t.Error("exponent mismatch")
	}
}

func TestParseRSAPublicKeyInvalidN(t *testing.T) {
	jwk := JWK{
		Kty: "RSA",
		Kid: "test-key",
		Use: "sig",
		Alg: "RS256",
		N:   "not-valid-base64!!!",
		E:   "AQAB",
	}

	_, err := parseRSAPublicKey(jwk)
	if err == nil {
		t.Error("expected error for invalid modulus")
	}
}

func TestParseRSAPublicKeyInvalidE(t *testing.T) {
	jwk := JWK{
		Kty: "RSA",
		Kid: "test-key",
		Use: "sig",
		Alg: "RS256",
		N:   "AQAB",
		E:   "not-valid-base64!!!",
	}

	_, err := parseRSAPublicKey(jwk)
	if err == nil {
		t.Error("expected error for invalid exponent")
	}
}
