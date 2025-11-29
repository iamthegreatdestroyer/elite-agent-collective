package auth

import (
	"context"
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

func TestMiddlewareDisabled(t *testing.T) {
	// When ClientID is empty, auth is disabled
	cfg := &config.OIDCConfig{
		Issuer:   "https://example.com",
		ClientID: "", // Empty = disabled
	}

	middleware := NewMiddleware(cfg)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	middleware.Authenticate(handler).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestMiddlewareEnabledNoHeader(t *testing.T) {
	cfg := &config.OIDCConfig{
		Issuer:   "https://example.com",
		ClientID: "test-client", // Non-empty = enabled
	}

	middleware := NewMiddleware(cfg)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	middleware.Authenticate(handler).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestMiddlewareEnabledInvalidFormat(t *testing.T) {
	cfg := &config.OIDCConfig{
		Issuer:   "https://example.com",
		ClientID: "test-client",
	}

	middleware := NewMiddleware(cfg)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "InvalidFormat")
	w := httptest.NewRecorder()

	middleware.Authenticate(handler).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestMiddlewareEnabledInvalidToken(t *testing.T) {
	cfg := &config.OIDCConfig{
		Issuer:   "https://example.com",
		ClientID: "test-client",
	}

	middleware := NewMiddleware(cfg)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()

	middleware.Authenticate(handler).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

// Helper function to create a mock OIDC server and valid token
func setupMockOIDC(t *testing.T) (*config.OIDCConfig, string, func()) {
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

	cfg := &config.OIDCConfig{
		Issuer:   discoveryServer.URL,
		ClientID: "test-client",
	}

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

	cleanup := func() {
		jwksServer.Close()
		discoveryServer.Close()
	}

	return cfg, tokenString, cleanup
}

func TestMiddlewareEnabledValidToken(t *testing.T) {
	cfg, tokenString, cleanup := setupMockOIDC(t)
	defer cleanup()

	middleware := NewMiddleware(cfg)

	handlerCalled := false
	var capturedClaims *Claims
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		capturedClaims = GetClaims(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	w := httptest.NewRecorder()

	middleware.Authenticate(handler).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if !handlerCalled {
		t.Error("expected handler to be called")
	}

	if capturedClaims == nil {
		t.Error("expected claims to be set in context")
	} else if capturedClaims.Subject != "test-user" {
		t.Errorf("expected subject 'test-user', got %s", capturedClaims.Subject)
	}
}

func TestOptionalAuthNoHeader(t *testing.T) {
	cfg := &config.OIDCConfig{
		Issuer:   "https://example.com",
		ClientID: "test-client",
	}

	middleware := NewMiddleware(cfg)

	handlerCalled := false
	var capturedClaims *Claims
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		capturedClaims = GetClaims(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	middleware.OptionalAuth(handler).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if !handlerCalled {
		t.Error("expected handler to be called")
	}

	if capturedClaims != nil {
		t.Error("expected no claims for unauthenticated request")
	}
}

func TestOptionalAuthInvalidFormat(t *testing.T) {
	cfg := &config.OIDCConfig{
		Issuer:   "https://example.com",
		ClientID: "test-client",
	}

	middleware := NewMiddleware(cfg)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "InvalidFormat")
	w := httptest.NewRecorder()

	middleware.OptionalAuth(handler).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestOptionalAuthInvalidToken(t *testing.T) {
	cfg := &config.OIDCConfig{
		Issuer:   "https://example.com",
		ClientID: "test-client",
	}

	middleware := NewMiddleware(cfg)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()

	middleware.OptionalAuth(handler).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestOptionalAuthValidToken(t *testing.T) {
	cfg, tokenString, cleanup := setupMockOIDC(t)
	defer cleanup()

	middleware := NewMiddleware(cfg)

	handlerCalled := false
	var capturedClaims *Claims
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		capturedClaims = GetClaims(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	w := httptest.NewRecorder()

	middleware.OptionalAuth(handler).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if !handlerCalled {
		t.Error("expected handler to be called")
	}

	if capturedClaims == nil {
		t.Error("expected claims to be set in context")
	} else if capturedClaims.Subject != "test-user" {
		t.Errorf("expected subject 'test-user', got %s", capturedClaims.Subject)
	}
}

func TestOptionalAuthDisabled(t *testing.T) {
	cfg := &config.OIDCConfig{
		Issuer:   "https://example.com",
		ClientID: "", // Disabled
	}

	middleware := NewMiddleware(cfg)

	handlerCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	middleware.OptionalAuth(handler).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if !handlerCalled {
		t.Error("expected handler to be called")
	}
}

func TestGetClaimsNoClaims(t *testing.T) {
	ctx := context.Background()
	claims := GetClaims(ctx)

	if claims != nil {
		t.Error("expected nil claims for empty context")
	}
}

func TestGetClaimsWithClaims(t *testing.T) {
	expectedClaims := &Claims{
		Subject:   "test-user",
		Issuer:    "https://example.com",
		Audience:  "test-client",
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
	}

	ctx := context.WithValue(context.Background(), ClaimsContextKey, expectedClaims)
	claims := GetClaims(ctx)

	if claims == nil {
		t.Fatal("expected non-nil claims")
	}

	if claims.Subject != expectedClaims.Subject {
		t.Errorf("expected subject '%s', got '%s'", expectedClaims.Subject, claims.Subject)
	}
}
