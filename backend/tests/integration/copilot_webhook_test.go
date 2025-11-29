//go:build integration
// +build integration

package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/pkg/models"
)

// TestCopilotWebhook_SingleAgentInvocation tests invoking a single agent via the /copilot endpoint.
func TestCopilotWebhook_SingleAgentInvocation(t *testing.T) {
	reqBody := models.CopilotRequest{
		Messages: []models.Message{
			{Role: "user", Content: "@APEX help me design a sorting algorithm"},
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
		t.Fatalf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status 200, got %d: %s", resp.StatusCode, string(respBody))
	}

	var copilotResp models.CopilotResponse
	if err := json.NewDecoder(resp.Body).Decode(&copilotResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(copilotResp.Choices) == 0 {
		t.Fatal("expected at least one choice in response")
	}

	content := copilotResp.Choices[0].Message.Content
	if !strings.Contains(content, "APEX") {
		t.Errorf("expected response to contain 'APEX', got: %s", content)
	}
}

// TestCopilotWebhook_MultiAgentInvocation tests invoking multiple agents via the /copilot endpoint.
func TestCopilotWebhook_MultiAgentInvocation(t *testing.T) {
	reqBody := models.CopilotRequest{
		Messages: []models.Message{
			{Role: "user", Content: "@APEX @ARCHITECT design a distributed system"},
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
		t.Fatalf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var copilotResp models.CopilotResponse
	if err := json.NewDecoder(resp.Body).Decode(&copilotResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Currently, the system routes to the first agent found
	if len(copilotResp.Choices) == 0 {
		t.Fatal("expected at least one choice in response")
	}

	content := copilotResp.Choices[0].Message.Content
	if !strings.Contains(content, "APEX") {
		t.Errorf("expected response to contain 'APEX', got: %s", content)
	}
}

// TestCopilotWebhook_UnknownAgent tests handling of an unknown agent request.
func TestCopilotWebhook_UnknownAgent(t *testing.T) {
	reqBody := models.CopilotRequest{
		Messages: []models.Message{
			{Role: "user", Content: "@UNKNOWN_AGENT help me"},
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
		t.Fatalf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Should fall back to APEX and return 200
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 (fallback to APEX), got %d", resp.StatusCode)
	}

	var copilotResp models.CopilotResponse
	if err := json.NewDecoder(resp.Body).Decode(&copilotResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Should have gotten a response from APEX (the fallback)
	if len(copilotResp.Choices) == 0 {
		t.Fatal("expected at least one choice in response")
	}
}

// TestCopilotWebhook_MalformedRequest tests handling of malformed JSON requests.
func TestCopilotWebhook_MalformedRequest(t *testing.T) {
	malformedJSON := `{"messages": "not an array", "model": 123}`

	resp, err := http.Post(getTestServerURL()+"/copilot", "application/json", strings.NewReader(malformedJSON))
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}
}

// TestCopilotWebhook_EmptyMessage tests handling of requests with empty messages.
func TestCopilotWebhook_EmptyMessage(t *testing.T) {
	reqBody := models.CopilotRequest{
		Messages: []models.Message{},
		Model:    "gpt-4",
		Stream:   false,
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	resp, err := http.Post(getTestServerURL()+"/copilot", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}
}

// TestCopilotWebhook_ResponseFormat validates the response format matches Copilot expectations.
func TestCopilotWebhook_ResponseFormat(t *testing.T) {
	reqBody := models.CopilotRequest{
		Messages: []models.Message{
			{Role: "user", Content: "@APEX analyze this algorithm"},
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
		t.Fatalf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	// Verify content type
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		t.Errorf("expected content-type application/json, got: %s", contentType)
	}

	var copilotResp models.CopilotResponse
	if err := json.NewDecoder(resp.Body).Decode(&copilotResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Validate response structure
	if len(copilotResp.Choices) == 0 {
		t.Fatal("expected at least one choice in response")
	}

	choice := copilotResp.Choices[0]
	if choice.Message.Role != "assistant" {
		t.Errorf("expected role 'assistant', got: %s", choice.Message.Role)
	}

	if choice.FinishReason != "stop" {
		t.Errorf("expected finish_reason 'stop', got: %s", choice.FinishReason)
	}

	if choice.Message.Content == "" {
		t.Error("expected non-empty content in response")
	}
}

// TestCopilotWebhook_DefaultsToAPEX tests that requests without an agent default to APEX.
func TestCopilotWebhook_DefaultsToAPEX(t *testing.T) {
	reqBody := models.CopilotRequest{
		Messages: []models.Message{
			{Role: "user", Content: "help me with an algorithm"},
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
		t.Fatalf("failed to make request: %v", err)
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
		t.Fatal("expected at least one choice in response")
	}

	// Response should be from APEX
	content := copilotResp.Choices[0].Message.Content
	if !strings.Contains(content, "APEX") {
		t.Errorf("expected response from APEX, got: %s", content)
	}
}

// TestCopilotWebhook_WithConversationContext tests handling of multi-turn conversations.
func TestCopilotWebhook_WithConversationContext(t *testing.T) {
	reqBody := models.CopilotRequest{
		Messages: []models.Message{
			{Role: "system", Content: "You are a helpful assistant."},
			{Role: "user", Content: "What is sorting?"},
			{Role: "assistant", Content: "Sorting is the process of arranging items in order."},
			{Role: "user", Content: "@APEX explain merge sort"},
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
		t.Fatalf("failed to make request: %v", err)
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
		t.Fatal("expected at least one choice in response")
	}
}
