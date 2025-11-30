// Package auth provides authentication middleware and signature validation.
package auth

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
)

// SignatureMiddleware provides GitHub webhook signature verification.
type SignatureMiddleware struct {
	secret  string
	enabled bool
}

// NewSignatureMiddleware creates a new signature verification middleware.
// The secret is the GITHUB_WEBHOOK_SECRET used to verify webhook payloads.
func NewSignatureMiddleware(secret string) *SignatureMiddleware {
	return &SignatureMiddleware{
		secret:  secret,
		enabled: secret != "",
	}
}

// VerifySignature is HTTP middleware that validates GitHub webhook signatures.
// It validates the X-GitHub-Signature-256 header against the request body.
// If signature verification is not enabled (no secret configured), requests pass through.
func (m *SignatureMiddleware) VerifySignature(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip signature verification if not enabled
		if !m.enabled {
			next.ServeHTTP(w, r)
			return
		}

		// Get the signature header
		signature := r.Header.Get("X-GitHub-Signature-256")
		if signature == "" {
			log.Printf("Missing X-GitHub-Signature-256 header")
			http.Error(w, "Missing signature header", http.StatusUnauthorized)
			return
		}

		// Read the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Failed to read request body: %v", err)
			http.Error(w, "Failed to read request body", http.StatusInternalServerError)
			return
		}

		// Replace the body for the next handler
		r.Body = io.NopCloser(bytes.NewReader(body))

		// Validate the signature
		if err := m.validateSignature(signature, body); err != nil {
			log.Printf("Signature validation failed: %v", err)
			http.Error(w, "Invalid signature", http.StatusUnauthorized)
			return
		}

		log.Printf("Webhook signature verified successfully")
		next.ServeHTTP(w, r)
	})
}

// validateSignature validates a GitHub webhook signature.
// The signature format is "sha256=<hex-encoded-signature>".
func (m *SignatureMiddleware) validateSignature(signature string, body []byte) error {
	// The signature should be in format "sha256=<hex>"
	if !strings.HasPrefix(signature, "sha256=") {
		return errors.New("signature must have sha256= prefix")
	}

	// Extract the hex-encoded signature
	signatureHex := strings.TrimPrefix(signature, "sha256=")
	expectedSignature, err := hex.DecodeString(signatureHex)
	if err != nil {
		return errors.New("invalid signature format")
	}

	// Calculate the expected signature
	mac := hmac.New(sha256.New, []byte(m.secret))
	mac.Write(body)
	actualSignature := mac.Sum(nil)

	// Compare using constant-time comparison to prevent timing attacks
	if !hmac.Equal(expectedSignature, actualSignature) {
		return errors.New("signature mismatch")
	}

	return nil
}

// ValidateSignature validates a GitHub webhook signature directly.
// This can be used outside of the middleware context.
func ValidateSignature(secret, signature string, body []byte) error {
	m := &SignatureMiddleware{secret: secret, enabled: true}
	return m.validateSignature(signature, body)
}

// ComputeSignature computes the HMAC SHA-256 signature for a given payload.
// This is useful for testing and generating signatures.
func ComputeSignature(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}
