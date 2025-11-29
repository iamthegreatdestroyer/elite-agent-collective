package agents

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/pkg/models"
)

func setupTestHandler() (*Handler, *chi.Mux) {
	registry := DefaultRegistry()
	handler := NewHandler(registry)
	
	r := chi.NewRouter()
	r.Get("/agents", handler.ListAgents)
	r.Get("/agents/{codename}", handler.GetAgent)
	r.Post("/agents/{codename}/invoke", handler.InvokeAgent)
	r.Post("/copilot", handler.CopilotWebhook)
	
	return handler, r
}

func TestListAgents(t *testing.T) {
	_, r := setupTestHandler()

	req := httptest.NewRequest("GET", "/agents", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var agents []models.Agent
	if err := json.NewDecoder(w.Body).Decode(&agents); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(agents) != 40 {
		t.Errorf("expected 40 agents, got %d", len(agents))
	}
}

func TestGetAgent(t *testing.T) {
	_, r := setupTestHandler()

	// Test getting APEX
	req := httptest.NewRequest("GET", "/agents/APEX", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var agent models.Agent
	if err := json.NewDecoder(w.Body).Decode(&agent); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if agent.Codename != "APEX" {
		t.Errorf("expected codename 'APEX', got %s", agent.Codename)
	}
}

func TestGetAgentNotFound(t *testing.T) {
	_, r := setupTestHandler()

	req := httptest.NewRequest("GET", "/agents/NONEXISTENT", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestInvokeAgent(t *testing.T) {
	_, r := setupTestHandler()

	reqBody := models.CopilotRequest{
		Messages: []models.Message{
			{Role: "user", Content: "Help me with an algorithm"},
		},
		Model:  "gpt-4",
		Stream: false,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/agents/APEX/invoke", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp models.CopilotResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Choices) == 0 {
		t.Fatal("expected at least one choice")
	}

	if resp.Choices[0].Message.Role != "assistant" {
		t.Errorf("expected role 'assistant', got %s", resp.Choices[0].Message.Role)
	}

	if resp.Choices[0].Message.Content == "" {
		t.Error("expected non-empty content")
	}
}

func TestCopilotWebhook(t *testing.T) {
	_, r := setupTestHandler()

	reqBody := models.CopilotRequest{
		Messages: []models.Message{
			{Role: "user", Content: "@APEX help me design an algorithm"},
		},
		Model:  "gpt-4",
		Stream: false,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/copilot", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp models.CopilotResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Choices) == 0 {
		t.Fatal("expected at least one choice")
	}

	// Response should be from APEX
	content := resp.Choices[0].Message.Content
	if content == "" {
		t.Error("expected non-empty content")
	}
}

func TestCopilotWebhookDefaultsToAPEX(t *testing.T) {
	_, r := setupTestHandler()

	reqBody := models.CopilotRequest{
		Messages: []models.Message{
			{Role: "user", Content: "help me with something"},
		},
		Model:  "gpt-4",
		Stream: false,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/copilot", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestExtractAgentCodename(t *testing.T) {
	tests := []struct {
		message  string
		expected string
	}{
		{"@APEX help me", "APEX"},
		{"@CIPHER analyze this", "CIPHER"},
		{"@ARCHITECT design system", "ARCHITECT"},
		{"help me with something", ""},
		{"", ""},
		{"@", ""},
		{"@APEX", "APEX"},
		{"@APEX\nwith newline", "APEX"},
		{"@APEX\twith tab", "APEX"},
	}

	for _, tt := range tests {
		result := extractAgentCodename(tt.message)
		if result != tt.expected {
			t.Errorf("extractAgentCodename(%q) = %q, expected %q", tt.message, result, tt.expected)
		}
	}
}
