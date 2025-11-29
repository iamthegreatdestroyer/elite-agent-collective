// Package auth provides authentication middleware and OIDC validation.
package auth

import (
	"log"
	"net/http"
	"strings"

	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/config"
)

// Middleware creates authentication middleware for protecting routes.
type Middleware struct {
	validator *OIDCValidator
	enabled   bool
}

// NewMiddleware creates a new authentication middleware.
func NewMiddleware(cfg *config.OIDCConfig) *Middleware {
	// Enable auth only if OIDC client ID is configured
	enabled := cfg.ClientID != ""
	
	return &Middleware{
		validator: NewOIDCValidator(cfg),
		enabled:   enabled,
	}
}

// Authenticate is HTTP middleware that validates authentication tokens.
func (m *Middleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip authentication if not enabled
		if !m.enabled {
			next.ServeHTTP(w, r)
			return
		}

		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Expect "Bearer <token>" format
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		token := parts[1]
		claims, err := m.validator.ValidateToken(token)
		if err != nil {
			log.Printf("Token validation failed: %v", err)
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Log successful authentication
		log.Printf("Authenticated user: %s", claims.Subject)

		next.ServeHTTP(w, r)
	})
}
