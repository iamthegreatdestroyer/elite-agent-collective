package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

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

func TestMiddlewareEnabledValidToken(t *testing.T) {
	cfg := &config.OIDCConfig{
		Issuer:   "https://example.com",
		ClientID: "test-client",
	}

	middleware := NewMiddleware(cfg)

	handlerCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer some-token")
	w := httptest.NewRecorder()

	middleware.Authenticate(handler).ServeHTTP(w, req)

	// With the stub implementation, any non-empty token is valid
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if !handlerCalled {
		t.Error("expected handler to be called")
	}
}
