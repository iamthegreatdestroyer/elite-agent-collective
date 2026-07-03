package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/copilot"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/pkg/models"
)

// newMockOllamaClient points a real *copilot.OllamaClient at a local
// httptest server that mimics Ollama's /api/chat response shape, so the
// wiring in Handle() (upstream nil/fails -> ollama -> template) can be
// exercised fast and deterministically, without a live Ollama server.
// The real end-to-end proof against an actual Ollama instance lives in
// backend/tests/integration/ollama_fallback_test.go (TestOllamaFallback_APEX).
func newMockOllamaClient(t *testing.T, replyContent string) *copilot.OllamaClient {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/tags":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"models":[]}`))
		case "/api/chat":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"message": map[string]string{
					"role":    "assistant",
					"content": replyContent,
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	// NewOllamaClient reads OLLAMA_URL/OLLAMA_MODEL from the environment;
	// point it at the mock server for the duration of this test.
	oldURL, hadURL := os.LookupEnv("OLLAMA_URL")
	os.Setenv("OLLAMA_URL", server.URL)
	t.Cleanup(func() {
		if hadURL {
			os.Setenv("OLLAMA_URL", oldURL)
		} else {
			os.Unsetenv("OLLAMA_URL")
		}
	})

	return copilot.NewOllamaClient()
}

// TestBaseAgentHandle_OllamaFallback verifies that when upstream is nil,
// BaseAgent.Handle falls through to the Ollama client (second-tier fallback)
// and returns its content, rather than the canned template response.
func TestBaseAgentHandle_OllamaFallback(t *testing.T) {
	info := models.Agent{
		ID:         "99",
		Codename:   "TESTAGENT",
		Specialty:  "Test Specialty",
		Philosophy: "Test philosophy.",
		Directives: []string{"Directive one", "Directive two"},
	}
	agent := NewBaseAgent(info, nil) // upstream nil -> must fall through
	agent.SetOllama(newMockOllamaClient(t, "This is a real Ollama-generated reply about testing."))

	req := &models.CopilotRequest{
		Messages: []models.Message{
			{Role: "user", Content: "Help me test something"},
		},
	}

	resp, err := agent.Handle(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Choices) == 0 {
		t.Fatal("expected at least one choice")
	}

	content := resp.Choices[0].Message.Content
	if content != "This is a real Ollama-generated reply about testing." {
		t.Errorf("expected Ollama-generated content, got template/other content: %q", content)
	}
}

// TestBaseAgentHandle_FallsBackToTemplateWhenOllamaDisabled verifies that
// when both upstream and Ollama are nil/unavailable, Handle still falls back
// to the template response (i.e. the new tier doesn't break the existing
// last-resort fallback).
func TestBaseAgentHandle_FallsBackToTemplateWhenOllamaDisabled(t *testing.T) {
	info := models.Agent{
		ID:         "99",
		Codename:   "TESTAGENT",
		Specialty:  "Test Specialty",
		Philosophy: "Test philosophy.",
		Directives: []string{"Directive one"},
	}
	agent := NewBaseAgent(info, nil) // upstream nil, ollama never set (nil)

	req := &models.CopilotRequest{
		Messages: []models.Message{
			{Role: "user", Content: "Help me test something"},
		},
	}

	resp, err := agent.Handle(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Choices) == 0 {
		t.Fatal("expected at least one choice")
	}

	content := resp.Choices[0].Message.Content
	if !containsString(content, "TESTAGENT") {
		t.Errorf("expected template response mentioning TESTAGENT, got: %q", content)
	}
}
