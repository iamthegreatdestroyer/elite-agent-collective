// Package config provides configuration management for the backend server.
package config

import (
	"os"
	"strconv"
)

// Config holds all configuration for the server.
type Config struct {
	// Server configuration
	Port     int
	LogLevel string

	// CORS configuration
	CORSAllowedOrigins string

	// OIDC configuration
	OIDC OIDCConfig

	// GitHub App configuration for Copilot Extensions
	GitHub GitHubConfig
}

// OIDCConfig holds OIDC authentication configuration.
type OIDCConfig struct {
	Issuer       string
	ClientID     string
	ClientSecret string
}

// GitHubConfig holds GitHub App configuration for Copilot Extensions.
type GitHubConfig struct {
	// AppID is the GitHub App ID
	AppID string
	// PrivateKey is the GitHub App private key (PEM format)
	PrivateKey string
	// WebhookSecret is the secret used to verify webhook payloads
	WebhookSecret string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	return &Config{
		Port:               getEnvAsInt("PORT", 8080),
		LogLevel:           getEnv("LOG_LEVEL", "info"),
		CORSAllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", ""),
		OIDC: OIDCConfig{
			Issuer:       getEnv("OIDC_ISSUER", "https://token.actions.githubusercontent.com"),
			ClientID:     getEnv("OIDC_CLIENT_ID", ""),
			ClientSecret: getEnv("OIDC_CLIENT_SECRET", ""),
		},
		GitHub: GitHubConfig{
			AppID:         getEnv("GITHUB_APP_ID", ""),
			PrivateKey:    getEnv("GITHUB_APP_PRIVATE_KEY", ""),
			WebhookSecret: getEnv("GITHUB_WEBHOOK_SECRET", ""),
		},
	}
}

// getEnv gets an environment variable or returns a default value.
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as an integer or returns a default value.
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}
