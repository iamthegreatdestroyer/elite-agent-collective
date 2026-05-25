package auth

import "fmt"

// StubValidator accepts any non-empty token for testing purposes.
// NEVER use in production.
type StubValidator struct{}

func (s *StubValidator) ValidateToken(tokenString string) (*Claims, error) {
	if tokenString == "" {
		return nil, fmt.Errorf("token required")
	}
	return &Claims{Subject: "test-user", Issuer: "test-issuer"}, nil
}
