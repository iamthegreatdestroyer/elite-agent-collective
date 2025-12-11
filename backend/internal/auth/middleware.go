// Package auth provides authentication middleware and OIDC validation.
package auth

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/config"
)

// contextKey is a type for context keys to avoid collisions.
type contextKey string

// ClaimsContextKey is the context key for storing claims.
const ClaimsContextKey contextKey = "claims"

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
// It returns 401 for missing or invalid tokens when authentication is enabled.
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

		// Add claims to request context
		ctx := context.WithValue(r.Context(), ClaimsContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuth is HTTP middleware that validates tokens if present but allows unauthenticated requests.
// If a valid token is provided, claims are added to the request context.
// If no token is provided, the request proceeds without claims.
// If an invalid token is provided, the request is rejected with 401.
func (m *Middleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If auth is not enabled, just proceed
		if !m.enabled {
			next.ServeHTTP(w, r)
			return
		}

		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// No token provided, allow request to proceed without claims
			next.ServeHTTP(w, r)
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

		// Add claims to request context
		ctx := context.WithValue(r.Context(), ClaimsContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetClaims retrieves claims from the request context.
// Returns nil if no claims are present (unauthenticated request with optional auth).
func GetClaims(ctx context.Context) *Claims {
	claims, ok := ctx.Value(ClaimsContextKey).(*Claims)
	if !ok {
		return nil
	}
	return claims
}
