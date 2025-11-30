//go:build !integration
// +build !integration

package auth

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewSignatureMiddleware(t *testing.T) {
	t.Run("enabled with secret", func(t *testing.T) {
		m := NewSignatureMiddleware("test-secret")
		if !m.enabled {
			t.Error("expected middleware to be enabled with secret")
		}
	})

	t.Run("disabled without secret", func(t *testing.T) {
		m := NewSignatureMiddleware("")
		if m.enabled {
			t.Error("expected middleware to be disabled without secret")
		}
	})
}

func TestSignatureMiddleware_VerifySignature(t *testing.T) {
	secret := "test-webhook-secret"
	body := []byte(`{"test": "payload"}`)

	t.Run("valid signature", func(t *testing.T) {
		m := NewSignatureMiddleware(secret)
		signature := ComputeSignature(secret, body)

		handler := m.VerifySignature(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(body))
		req.Header.Set("X-GitHub-Signature-256", signature)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
	})

	t.Run("missing signature header", func(t *testing.T) {
		m := NewSignatureMiddleware(secret)

		handler := m.VerifySignature(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(body))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", rec.Code)
		}
	})

	t.Run("invalid signature", func(t *testing.T) {
		m := NewSignatureMiddleware(secret)

		handler := m.VerifySignature(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(body))
		req.Header.Set("X-GitHub-Signature-256", "sha256=invalid")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", rec.Code)
		}
	})

	t.Run("wrong secret", func(t *testing.T) {
		m := NewSignatureMiddleware(secret)
		wrongSignature := ComputeSignature("wrong-secret", body)

		handler := m.VerifySignature(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(body))
		req.Header.Set("X-GitHub-Signature-256", wrongSignature)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", rec.Code)
		}
	})

	t.Run("disabled middleware passes through", func(t *testing.T) {
		m := NewSignatureMiddleware("") // Empty secret = disabled

		handler := m.VerifySignature(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(body))
		// No signature header
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200 when disabled, got %d", rec.Code)
		}
	})
}

func TestValidateSignature(t *testing.T) {
	secret := "test-secret"
	body := []byte(`{"test": "data"}`)

	t.Run("valid signature", func(t *testing.T) {
		signature := ComputeSignature(secret, body)
		err := ValidateSignature(secret, signature, body)
		if err != nil {
			t.Errorf("expected no error for valid signature, got: %v", err)
		}
	})

	t.Run("invalid prefix", func(t *testing.T) {
		err := ValidateSignature(secret, "md5=abc123", body)
		if err == nil {
			t.Error("expected error for invalid prefix")
		}
	})

	t.Run("invalid hex", func(t *testing.T) {
		err := ValidateSignature(secret, "sha256=notvalidhex!", body)
		if err == nil {
			t.Error("expected error for invalid hex")
		}
	})

	t.Run("mismatched signature", func(t *testing.T) {
		wrongSignature := ComputeSignature("different-secret", body)
		err := ValidateSignature(secret, wrongSignature, body)
		if err == nil {
			t.Error("expected error for mismatched signature")
		}
	})
}

func TestComputeSignature(t *testing.T) {
	secret := "test-secret"
	body := []byte(`{"test": "data"}`)

	signature := ComputeSignature(secret, body)

	// Should start with sha256= prefix
	if len(signature) < 7 || signature[:7] != "sha256=" {
		t.Error("signature should start with sha256=")
	}

	// Should be consistent
	signature2 := ComputeSignature(secret, body)
	if signature != signature2 {
		t.Error("same input should produce same signature")
	}

	// Different input should produce different signature
	differentSignature := ComputeSignature(secret, []byte(`{"different": "data"}`))
	if signature == differentSignature {
		t.Error("different input should produce different signature")
	}
}
