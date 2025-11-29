//go:build integration
// +build integration

package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/pkg/models"
)

// TestHealthEndpoint tests the /health endpoint.
func TestHealthEndpoint(t *testing.T) {
	resp, err := http.Get(getTestServerURL() + "/health")
	if err != nil {
		t.Fatalf("failed to get health endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		t.Errorf("expected content-type application/json, got: %s", contentType)
	}

	var healthResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&healthResp); err != nil {
		t.Fatalf("failed to decode health response: %v", err)
	}

	if status, ok := healthResp["status"].(string); !ok || status != "healthy" {
		t.Errorf("expected status 'healthy', got: %v", healthResp["status"])
	}
}

// TestListAgentsEndpoint tests the GET /agents endpoint.
func TestListAgentsEndpoint(t *testing.T) {
	resp, err := http.Get(getTestServerURL() + "/agents")
	if err != nil {
		t.Fatalf("failed to get agents: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		t.Errorf("expected content-type application/json, got: %s", contentType)
	}

	var agents []models.Agent
	if err := json.NewDecoder(resp.Body).Decode(&agents); err != nil {
		t.Fatalf("failed to decode agents: %v", err)
	}

	if len(agents) != 40 {
		t.Errorf("expected 40 agents, got %d", len(agents))
	}

	// Verify agents have required fields
	for _, agent := range agents {
		if agent.ID == "" {
			t.Errorf("agent has empty ID: %v", agent)
		}
		if agent.Codename == "" {
			t.Errorf("agent has empty Codename: %v", agent)
		}
		if agent.Tier < 1 || agent.Tier > 8 {
			t.Errorf("agent has invalid Tier: %d for %s", agent.Tier, agent.Codename)
		}
		if agent.Specialty == "" {
			t.Errorf("agent has empty Specialty: %s", agent.Codename)
		}
		if agent.Philosophy == "" {
			t.Errorf("agent has empty Philosophy: %s", agent.Codename)
		}
		if len(agent.Directives) == 0 {
			t.Errorf("agent has no Directives: %s", agent.Codename)
		}
	}
}

// TestGetAgentEndpoint tests the GET /agents/{codename} endpoint.
func TestGetAgentEndpoint(t *testing.T) {
	testCases := []struct {
		codename string
		tier     int
	}{
		{"APEX", 1},
		{"CIPHER", 1},
		{"QUANTUM", 2},
		{"NEXUS", 3},
		{"OMNISCIENT", 4},
		{"ATLAS", 5},
		{"PHOTON", 6},
		{"CANVAS", 7},
		{"AEGIS", 8},
	}

	for _, tc := range testCases {
		t.Run(tc.codename, func(t *testing.T) {
			resp, err := http.Get(getTestServerURL() + "/agents/" + tc.codename)
			if err != nil {
				t.Fatalf("failed to get agent %s: %v", tc.codename, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("expected status 200 for %s, got %d", tc.codename, resp.StatusCode)
			}

			var agent models.Agent
			if err := json.NewDecoder(resp.Body).Decode(&agent); err != nil {
				t.Fatalf("failed to decode agent %s: %v", tc.codename, err)
			}

			if agent.Codename != tc.codename {
				t.Errorf("expected codename %s, got %s", tc.codename, agent.Codename)
			}

			if agent.Tier != tc.tier {
				t.Errorf("expected tier %d for %s, got %d", tc.tier, tc.codename, agent.Tier)
			}
		})
	}
}

// TestGetAgentEndpoint_NotFound tests the GET /agents/{codename} endpoint for non-existent agents.
func TestGetAgentEndpoint_NotFound(t *testing.T) {
	resp, err := http.Get(getTestServerURL() + "/agents/NONEXISTENT")
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", resp.StatusCode)
	}
}

// TestInvokeAgentEndpoint tests the POST /agents/{codename}/invoke endpoint.
func TestInvokeAgentEndpoint(t *testing.T) {
	reqBody := models.CopilotRequest{
		Messages: []models.Message{
			{Role: "user", Content: "Help me with an algorithm"},
		},
		Model:  "gpt-4",
		Stream: false,
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	resp, err := http.Post(getTestServerURL()+"/agents/APEX/invoke", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("failed to invoke APEX: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var copilotResp models.CopilotResponse
	if err := json.NewDecoder(resp.Body).Decode(&copilotResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(copilotResp.Choices) == 0 {
		t.Fatal("expected at least one choice")
	}

	if copilotResp.Choices[0].Message.Role != "assistant" {
		t.Errorf("expected role 'assistant', got %s", copilotResp.Choices[0].Message.Role)
	}

	if copilotResp.Choices[0].Message.Content == "" {
		t.Error("expected non-empty content")
	}
}

// TestInvokeAgentEndpoint_NotFound tests invoking a non-existent agent.
func TestInvokeAgentEndpoint_NotFound(t *testing.T) {
	reqBody := models.CopilotRequest{
		Messages: []models.Message{
			{Role: "user", Content: "Help me"},
		},
		Model:  "gpt-4",
		Stream: false,
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	resp, err := http.Post(getTestServerURL()+"/agents/NONEXISTENT/invoke", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", resp.StatusCode)
	}
}

// TestCopilotEndpoint tests the POST /copilot endpoint.
func TestCopilotEndpoint(t *testing.T) {
	reqBody := models.CopilotRequest{
		Messages: []models.Message{
			{Role: "user", Content: "@APEX help me design an algorithm"},
		},
		Model:  "gpt-4",
		Stream: false,
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	resp, err := http.Post(getTestServerURL()+"/copilot", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("failed to post to /copilot: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var copilotResp models.CopilotResponse
	if err := json.NewDecoder(resp.Body).Decode(&copilotResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(copilotResp.Choices) == 0 {
		t.Fatal("expected at least one choice")
	}
}

// TestCopilotEndpoint_InvalidMethod tests that only POST is allowed for /copilot.
func TestCopilotEndpoint_InvalidMethod(t *testing.T) {
	resp, err := http.Get(getTestServerURL() + "/copilot")
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Should return 405 Method Not Allowed
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected status 405, got %d", resp.StatusCode)
	}
}

// TestAPIContentNegotiation tests that endpoints return proper content types.
func TestAPIContentNegotiation(t *testing.T) {
	endpoints := []string{
		"/health",
		"/agents",
		"/agents/APEX",
	}

	for _, endpoint := range endpoints {
		t.Run(endpoint, func(t *testing.T) {
			resp, err := http.Get(getTestServerURL() + endpoint)
			if err != nil {
				t.Fatalf("failed to get %s: %v", endpoint, err)
			}
			defer resp.Body.Close()

			contentType := resp.Header.Get("Content-Type")
			if !strings.Contains(contentType, "application/json") {
				t.Errorf("expected content-type application/json for %s, got: %s", endpoint, contentType)
			}
		})
	}
}

// TestAPIValidationErrors tests that invalid requests return appropriate errors.
func TestAPIValidationErrors(t *testing.T) {
	testCases := []struct {
		name     string
		body     string
		expected int
	}{
		{"malformed JSON", `{"messages": not json}`, http.StatusBadRequest},
		{"invalid message format", `{"messages": "not an array"}`, http.StatusBadRequest},
		{"empty body", ``, http.StatusBadRequest},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := http.Post(getTestServerURL()+"/copilot", "application/json", strings.NewReader(tc.body))
			if err != nil {
				t.Fatalf("failed to make request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tc.expected {
				t.Fatalf("expected status %d for %s, got %d", tc.expected, tc.name, resp.StatusCode)
			}
		})
	}
}
