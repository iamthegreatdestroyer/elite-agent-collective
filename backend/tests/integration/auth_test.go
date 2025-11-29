//go:build integration
// +build integration

package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/agents"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/auth"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/config"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/pkg/models"
)

// setupAuthEnabledServer creates a test server with authentication enabled.
func setupAuthEnabledServer() *httptest.Server {
	registry := agents.DefaultRegistry()
	agentHandler := agents.NewHandler(registry)

	// Enable authentication
	cfg := &config.OIDCConfig{
		Issuer:   "https://token.actions.githubusercontent.com",
		ClientID: "test-client-id", // Non-empty = auth enabled
	}
	authMiddleware := auth.NewMiddleware(cfg)

	r := chi.NewRouter()
	r.Route("/agents", func(r chi.Router) {
		r.Get("/", agentHandler.ListAgents)
		r.Get("/{codename}", agentHandler.GetAgent)
		r.With(authMiddleware.Authenticate).Post("/{codename}/invoke", agentHandler.InvokeAgent)
	})
	r.With(authMiddleware.Authenticate).Post("/copilot", agentHandler.CopilotWebhook)

	return httptest.NewServer(r)
}

// TestOIDCValidation_ValidToken tests that a valid token allows access.
func TestOIDCValidation_ValidToken(t *testing.T) {
	server := setupAuthEnabledServer()
	defer server.Close()

	reqBody := models.CopilotRequest{
		Messages: []models.Message{
			{Role: "user", Content: "@APEX help me"},
		},
		Model:  "gpt-4",
		Stream: false,
	}
	body, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", server.URL+"/copilot", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer valid-test-token")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// With stub validator, any non-empty token is valid
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
}

// TestOIDCValidation_NoToken tests that missing token returns 401.
func TestOIDCValidation_NoToken(t *testing.T) {
	server := setupAuthEnabledServer()
	defer server.Close()

	reqBody := models.CopilotRequest{
		Messages: []models.Message{
			{Role: "user", Content: "@APEX help me"},
		},
		Model:  "gpt-4",
		Stream: false,
	}
	body, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", server.URL+"/copilot", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	// No Authorization header

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", resp.StatusCode)
	}
}

// TestOIDCValidation_InvalidFormat tests that invalid token format returns 401.
func TestOIDCValidation_InvalidFormat(t *testing.T) {
	server := setupAuthEnabledServer()
	defer server.Close()

	reqBody := models.CopilotRequest{
		Messages: []models.Message{
			{Role: "user", Content: "@APEX help me"},
		},
		Model:  "gpt-4",
		Stream: false,
	}
	body, _ := json.Marshal(reqBody)

	testCases := []struct {
		name   string
		header string
	}{
		{"no bearer prefix", "some-token"},
		{"basic auth", "Basic dXNlcjpwYXNz"},
		{"empty bearer", "Bearer "},
		{"bearer only", "Bearer"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", server.URL+"/copilot", bytes.NewReader(body))
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", tc.header)

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("failed to make request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusUnauthorized {
				t.Fatalf("expected status 401 for %s, got %d", tc.name, resp.StatusCode)
			}
		})
	}
}

// TestWebhookSignatureValidation tests webhook signature validation behavior.
func TestWebhookSignatureValidation(t *testing.T) {
	server := setupAuthEnabledServer()
	defer server.Close()

	reqBody := models.CopilotRequest{
		Messages: []models.Message{
			{Role: "user", Content: "@APEX help me"},
		},
		Model:  "gpt-4",
		Stream: false,
	}
	body, _ := json.Marshal(reqBody)

	// Test with valid bearer token
	req, err := http.NewRequest("POST", server.URL+"/copilot", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer webhook-signature-token")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Stub validator accepts any non-empty token
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
}

// TestAuthMiddlewareBypassForPublicEndpoints tests that public endpoints don't require auth.
func TestAuthMiddlewareBypassForPublicEndpoints(t *testing.T) {
	server := setupAuthEnabledServer()
	defer server.Close()

	// Test that GET /agents is accessible without auth
	resp, err := http.Get(server.URL + "/agents")
	if err != nil {
		t.Fatalf("failed to get agents: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 for /agents, got %d", resp.StatusCode)
	}

	// Test that GET /agents/{codename} is accessible without auth
	resp2, err := http.Get(server.URL + "/agents/APEX")
	if err != nil {
		t.Fatalf("failed to get APEX: %v", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 for /agents/APEX, got %d", resp2.StatusCode)
	}
}

// TestAuthRequiredForProtectedEndpoints tests that protected endpoints require auth.
func TestAuthRequiredForProtectedEndpoints(t *testing.T) {
	server := setupAuthEnabledServer()
	defer server.Close()

	reqBody := models.CopilotRequest{
		Messages: []models.Message{
			{Role: "user", Content: "help me"},
		},
		Model:  "gpt-4",
		Stream: false,
	}
	body, _ := json.Marshal(reqBody)

	// Test that POST /agents/{codename}/invoke requires auth
	resp, err := http.Post(server.URL+"/agents/APEX/invoke", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("failed to invoke APEX: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected status 401 for /agents/APEX/invoke, got %d", resp.StatusCode)
	}

	// Test that POST /copilot requires auth
	body2, _ := json.Marshal(reqBody)
	resp2, err := http.Post(server.URL+"/copilot", "application/json", bytes.NewReader(body2))
	if err != nil {
		t.Fatalf("failed to post to /copilot: %v", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected status 401 for /copilot, got %d", resp2.StatusCode)
	}
}

// TestOIDCValidation_BearerCaseInsensitive tests that "bearer" is case-insensitive.
func TestOIDCValidation_BearerCaseInsensitive(t *testing.T) {
	server := setupAuthEnabledServer()
	defer server.Close()

	reqBody := models.CopilotRequest{
		Messages: []models.Message{
			{Role: "user", Content: "@APEX help me"},
		},
		Model:  "gpt-4",
		Stream: false,
	}
	body, _ := json.Marshal(reqBody)

	testCases := []struct {
		name   string
		header string
	}{
		{"lowercase bearer", "bearer valid-token"},
		{"uppercase bearer", "Bearer valid-token"},
		{"mixed case bearer", "BEARER valid-token"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", server.URL+"/copilot", bytes.NewReader(body))
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", tc.header)

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("failed to make request: %v", err)
			}
			defer resp.Body.Close()

			// With stub validator, any properly formatted token should work
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("expected status 200 for %s, got %d", tc.name, resp.StatusCode)
			}
		})
	}
}
