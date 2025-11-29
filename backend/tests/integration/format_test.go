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

// TestCopilotRequestParsing tests that various Copilot request formats are parsed correctly.
func TestCopilotRequestParsing(t *testing.T) {
	testCases := []struct {
		name     string
		request  models.CopilotRequest
		expected int
	}{
		{
			name: "standard request",
			request: models.CopilotRequest{
				Messages: []models.Message{
					{Role: "user", Content: "@APEX help me"},
				},
				Model:  "gpt-4",
				Stream: false,
			},
			expected: http.StatusOK,
		},
		{
			name: "request with system message",
			request: models.CopilotRequest{
				Messages: []models.Message{
					{Role: "system", Content: "You are helpful."},
					{Role: "user", Content: "@APEX help me"},
				},
				Model:  "gpt-4",
				Stream: false,
			},
			expected: http.StatusOK,
		},
		{
			name: "request with conversation history",
			request: models.CopilotRequest{
				Messages: []models.Message{
					{Role: "system", Content: "You are helpful."},
					{Role: "user", Content: "What is sorting?"},
					{Role: "assistant", Content: "Sorting is organizing items."},
					{Role: "user", Content: "@APEX explain merge sort"},
				},
				Model:  "gpt-4",
				Stream: false,
			},
			expected: http.StatusOK,
		},
		{
			name: "request with different model",
			request: models.CopilotRequest{
				Messages: []models.Message{
					{Role: "user", Content: "@APEX help me"},
				},
				Model:  "gpt-3.5-turbo",
				Stream: false,
			},
			expected: http.StatusOK,
		},
		{
			name: "request with streaming enabled",
			request: models.CopilotRequest{
				Messages: []models.Message{
					{Role: "user", Content: "@APEX help me"},
				},
				Model:  "gpt-4",
				Stream: true,
			},
			expected: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body, err := json.Marshal(tc.request)
			if err != nil {
				t.Fatalf("failed to marshal request: %v", err)
			}

			resp, err := http.Post(getTestServerURL()+"/copilot", "application/json", bytes.NewReader(body))
			if err != nil {
				t.Fatalf("failed to make request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tc.expected {
				t.Fatalf("expected status %d, got %d", tc.expected, resp.StatusCode)
			}
		})
	}
}

// TestCopilotResponseFormat validates the response format matches Copilot expectations.
func TestCopilotResponseFormat(t *testing.T) {
	reqBody := models.CopilotRequest{
		Messages: []models.Message{
			{Role: "user", Content: "@APEX analyze this code"},
		},
		Model:  "gpt-4",
		Stream: false,
	}
	body, _ := json.Marshal(reqBody)

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

	// Parse and validate response structure
	var copilotResp models.CopilotResponse
	if err := json.NewDecoder(resp.Body).Decode(&copilotResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Validate choices array
	if len(copilotResp.Choices) == 0 {
		t.Fatal("expected at least one choice in response")
	}

	// Validate choice structure
	choice := copilotResp.Choices[0]

	// Message should have role = "assistant"
	if choice.Message.Role != "assistant" {
		t.Errorf("expected message role 'assistant', got: %s", choice.Message.Role)
	}

	// Message should have non-empty content
	if choice.Message.Content == "" {
		t.Error("expected non-empty message content")
	}

	// Finish reason should be "stop"
	if choice.FinishReason != "stop" {
		t.Errorf("expected finish_reason 'stop', got: %s", choice.FinishReason)
	}
}

// TestStreamingResponse tests that streaming requests are handled.
// Note: Current implementation may not support actual streaming.
func TestStreamingResponse(t *testing.T) {
	reqBody := models.CopilotRequest{
		Messages: []models.Message{
			{Role: "user", Content: "@APEX help me"},
		},
		Model:  "gpt-4",
		Stream: true, // Request streaming
	}
	body, _ := json.Marshal(reqBody)

	resp, err := http.Post(getTestServerURL()+"/copilot", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Current implementation returns standard JSON response even for streaming requests
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

// TestErrorResponseFormat validates error responses have correct format.
func TestErrorResponseFormat(t *testing.T) {
	testCases := []struct {
		name         string
		body         string
		expectedCode int
	}{
		{
			name:         "malformed JSON",
			body:         `{"messages": not valid json}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "empty messages",
			body:         `{"messages": [], "model": "gpt-4"}`,
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := http.Post(getTestServerURL()+"/copilot", "application/json", strings.NewReader(tc.body))
			if err != nil {
				t.Fatalf("failed to make request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tc.expectedCode {
				t.Fatalf("expected status %d, got %d", tc.expectedCode, resp.StatusCode)
			}

			// For 400 errors, response should still be JSON with error message
			contentType := resp.Header.Get("Content-Type")
			if !strings.Contains(contentType, "application/json") {
				t.Errorf("expected content-type application/json for error, got: %s", contentType)
			}

			var errorResp models.CopilotResponse
			if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
				t.Fatalf("failed to decode error response: %v", err)
			}

			// Error response should have a message
			if len(errorResp.Choices) == 0 {
				t.Fatal("expected error response to have choices")
			}

			if errorResp.Choices[0].Message.Content == "" {
				t.Error("expected error response to have content")
			}
		})
	}
}

// TestMessageRoleParsing tests that different message roles are handled correctly.
func TestMessageRoleParsing(t *testing.T) {
	testCases := []struct {
		name     string
		messages []models.Message
		expected int
	}{
		{
			name: "only user message",
			messages: []models.Message{
				{Role: "user", Content: "@APEX help"},
			},
			expected: http.StatusOK,
		},
		{
			name: "system and user",
			messages: []models.Message{
				{Role: "system", Content: "Be helpful"},
				{Role: "user", Content: "@APEX help"},
			},
			expected: http.StatusOK,
		},
		{
			name: "full conversation",
			messages: []models.Message{
				{Role: "system", Content: "Be helpful"},
				{Role: "user", Content: "Question 1"},
				{Role: "assistant", Content: "Answer 1"},
				{Role: "user", Content: "@APEX Question 2"},
			},
			expected: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reqBody := models.CopilotRequest{
				Messages: tc.messages,
				Model:    "gpt-4",
				Stream:   false,
			}
			body, _ := json.Marshal(reqBody)

			resp, err := http.Post(getTestServerURL()+"/copilot", "application/json", bytes.NewReader(body))
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

// TestAgentMentionFormats tests different ways to mention agents.
func TestAgentMentionFormats(t *testing.T) {
	testCases := []struct {
		name    string
		content string
		agent   string
	}{
		{"at start", "@APEX help me", "APEX"},
		{"at start with newline", "@APEX\nhelp me", "APEX"},
		{"at start with tab", "@APEX\thelp me", "APEX"},
		{"lowercase", "@apex help me", "apex"}, // Note: currently case-sensitive
		{"no space after", "@APEX", "APEX"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reqBody := models.CopilotRequest{
				Messages: []models.Message{
					{Role: "user", Content: tc.content},
				},
				Model:  "gpt-4",
				Stream: false,
			}
			body, _ := json.Marshal(reqBody)

			resp, err := http.Post(getTestServerURL()+"/copilot", "application/json", bytes.NewReader(body))
			if err != nil {
				t.Fatalf("failed to make request: %v", err)
			}
			defer resp.Body.Close()

			// All should succeed (either routes to agent or falls back to APEX)
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("expected status 200 for %s, got %d", tc.name, resp.StatusCode)
			}
		})
	}
}

// TestLargeRequest tests handling of large conversation histories.
func TestLargeRequest(t *testing.T) {
	// Create a conversation with many messages
	messages := []models.Message{
		{Role: "system", Content: "You are a helpful assistant."},
	}

	// Add 50 back-and-forth messages
	for i := 0; i < 50; i++ {
		messages = append(messages, models.Message{
			Role:    "user",
			Content: "This is message number " + string(rune('0'+i%10)),
		})
		messages = append(messages, models.Message{
			Role:    "assistant",
			Content: "Response to message " + string(rune('0'+i%10)),
		})
	}

	// Final user message
	messages = append(messages, models.Message{
		Role:    "user",
		Content: "@APEX summarize our conversation",
	})

	reqBody := models.CopilotRequest{
		Messages: messages,
		Model:    "gpt-4",
		Stream:   false,
	}
	body, _ := json.Marshal(reqBody)

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
