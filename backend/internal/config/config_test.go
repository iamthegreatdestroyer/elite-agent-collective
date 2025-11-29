package config

import (
	"os"
	"testing"
)

func TestLoadWithDefaults(t *testing.T) {
	// Clear environment variables
	os.Unsetenv("PORT")
	os.Unsetenv("LOG_LEVEL")
	os.Unsetenv("OIDC_ISSUER")
	os.Unsetenv("OIDC_CLIENT_ID")
	os.Unsetenv("OIDC_CLIENT_SECRET")

	cfg := Load()

	if cfg.Port != 8080 {
		t.Errorf("expected default port 8080, got %d", cfg.Port)
	}

	if cfg.LogLevel != "info" {
		t.Errorf("expected default log level 'info', got %s", cfg.LogLevel)
	}

	if cfg.OIDC.Issuer != "https://token.actions.githubusercontent.com" {
		t.Errorf("expected default OIDC issuer, got %s", cfg.OIDC.Issuer)
	}

	if cfg.OIDC.ClientID != "" {
		t.Errorf("expected empty OIDC client ID, got %s", cfg.OIDC.ClientID)
	}
}

func TestLoadWithEnvironmentVariables(t *testing.T) {
	os.Setenv("PORT", "3000")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("OIDC_ISSUER", "https://example.com")
	os.Setenv("OIDC_CLIENT_ID", "test-client")
	os.Setenv("OIDC_CLIENT_SECRET", "test-secret")
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("OIDC_ISSUER")
		os.Unsetenv("OIDC_CLIENT_ID")
		os.Unsetenv("OIDC_CLIENT_SECRET")
	}()

	cfg := Load()

	if cfg.Port != 3000 {
		t.Errorf("expected port 3000, got %d", cfg.Port)
	}

	if cfg.LogLevel != "debug" {
		t.Errorf("expected log level 'debug', got %s", cfg.LogLevel)
	}

	if cfg.OIDC.Issuer != "https://example.com" {
		t.Errorf("expected OIDC issuer 'https://example.com', got %s", cfg.OIDC.Issuer)
	}

	if cfg.OIDC.ClientID != "test-client" {
		t.Errorf("expected OIDC client ID 'test-client', got %s", cfg.OIDC.ClientID)
	}

	if cfg.OIDC.ClientSecret != "test-secret" {
		t.Errorf("expected OIDC client secret 'test-secret', got %s", cfg.OIDC.ClientSecret)
	}
}

func TestLoadWithInvalidPort(t *testing.T) {
	os.Setenv("PORT", "invalid")
	defer os.Unsetenv("PORT")

	cfg := Load()

	// Should fall back to default
	if cfg.Port != 8080 {
		t.Errorf("expected default port 8080 for invalid value, got %d", cfg.Port)
	}
}
