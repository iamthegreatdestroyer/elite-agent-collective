package copilot

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/pkg/models"
)

func TestParseRequest(t *testing.T) {
	body := `{"messages":[{"role":"user","content":"Hello"}],"model":"gpt-4","stream":false}`
	req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(body))

	copilotReq, err := ParseRequest(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(copilotReq.Messages) != 1 {
		t.Errorf("expected 1 message, got %d", len(copilotReq.Messages))
	}

	if copilotReq.Messages[0].Role != "user" {
		t.Errorf("expected role 'user', got %s", copilotReq.Messages[0].Role)
	}

	if copilotReq.Messages[0].Content != "Hello" {
		t.Errorf("expected content 'Hello', got %s", copilotReq.Messages[0].Content)
	}

	if copilotReq.Model != "gpt-4" {
		t.Errorf("expected model 'gpt-4', got %s", copilotReq.Model)
	}
}

func TestParseRequestInvalidJSON(t *testing.T) {
	body := `invalid json`
	req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(body))

	_, err := ParseRequest(req)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestGetLastUserMessage(t *testing.T) {
	tests := []struct {
		name     string
		messages []models.Message
		expected string
	}{
		{
			name: "single user message",
			messages: []models.Message{
				{Role: "user", Content: "Hello"},
			},
			expected: "Hello",
		},
		{
			name: "multiple messages, user last",
			messages: []models.Message{
				{Role: "system", Content: "You are a helpful assistant"},
				{Role: "user", Content: "First question"},
				{Role: "assistant", Content: "First answer"},
				{Role: "user", Content: "Second question"},
			},
			expected: "Second question",
		},
		{
			name: "multiple messages, assistant last",
			messages: []models.Message{
				{Role: "user", Content: "Question"},
				{Role: "assistant", Content: "Answer"},
			},
			expected: "Question",
		},
		{
			name:     "no user messages",
			messages: []models.Message{},
			expected: "",
		},
		{
			name: "only system and assistant",
			messages: []models.Message{
				{Role: "system", Content: "System message"},
				{Role: "assistant", Content: "Assistant message"},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &models.CopilotRequest{Messages: tt.messages}
			result := GetLastUserMessage(req)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
