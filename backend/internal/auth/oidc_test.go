package auth

import (
	"testing"

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

func TestValidateTokenStub(t *testing.T) {
	cfg := &config.OIDCConfig{
		Issuer:   "https://example.com",
		ClientID: "test-client",
	}
	validator := NewOIDCValidator(cfg)

	// The stub implementation returns a placeholder claims object
	claims, err := validator.ValidateToken("some-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if claims == nil {
		t.Fatal("expected non-nil claims")
	}

	if claims.Subject != "development-user" {
		t.Errorf("expected subject 'development-user', got %s", claims.Subject)
	}
}
