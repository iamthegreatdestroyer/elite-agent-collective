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

	// OIDC configuration
	OIDC OIDCConfig
}

// OIDCConfig holds OIDC authentication configuration.
type OIDCConfig struct {
	Issuer       string
	ClientID     string
	ClientSecret string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	return &Config{
		Port:     getEnvAsInt("PORT", 8080),
		LogLevel: getEnv("LOG_LEVEL", "info"),
		OIDC: OIDCConfig{
			Issuer:       getEnv("OIDC_ISSUER", "https://token.actions.githubusercontent.com"),
			ClientID:     getEnv("OIDC_CLIENT_ID", ""),
			ClientSecret: getEnv("OIDC_CLIENT_SECRET", ""),
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
