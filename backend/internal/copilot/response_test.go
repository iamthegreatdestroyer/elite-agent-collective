package copilot

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/pkg/models"
)

func TestWriteResponse(t *testing.T) {
	w := httptest.NewRecorder()
	resp := &models.CopilotResponse{
		Choices: []models.Choice{
			{
				Message: models.Message{
					Role:    "assistant",
					Content: "Test response",
				},
				FinishReason: "stop",
			},
		},
	}

	err := WriteResponse(w, resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", w.Header().Get("Content-Type"))
	}

	var decoded models.CopilotResponse
	if err := json.NewDecoder(w.Body).Decode(&decoded); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(decoded.Choices) != 1 {
		t.Errorf("expected 1 choice, got %d", len(decoded.Choices))
	}

	if decoded.Choices[0].Message.Content != "Test response" {
		t.Errorf("expected content 'Test response', got %s", decoded.Choices[0].Message.Content)
	}
}

func TestWriteError(t *testing.T) {
	w := httptest.NewRecorder()
	WriteError(w, "Something went wrong", 500)

	if w.Code != 500 {
		t.Errorf("expected status 500, got %d", w.Code)
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", w.Header().Get("Content-Type"))
	}

	var decoded models.CopilotResponse
	if err := json.NewDecoder(w.Body).Decode(&decoded); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(decoded.Choices) != 1 {
		t.Errorf("expected 1 choice, got %d", len(decoded.Choices))
	}

	if decoded.Choices[0].Message.Content != "Something went wrong" {
		t.Errorf("expected error message, got %s", decoded.Choices[0].Message.Content)
	}
}

func TestNewResponse(t *testing.T) {
	resp := NewResponse("Test content")

	if len(resp.Choices) != 1 {
		t.Errorf("expected 1 choice, got %d", len(resp.Choices))
	}

	if resp.Choices[0].Message.Role != "assistant" {
		t.Errorf("expected role 'assistant', got %s", resp.Choices[0].Message.Role)
	}

	if resp.Choices[0].Message.Content != "Test content" {
		t.Errorf("expected content 'Test content', got %s", resp.Choices[0].Message.Content)
	}

	if resp.Choices[0].FinishReason != "stop" {
		t.Errorf("expected finish reason 'stop', got %s", resp.Choices[0].FinishReason)
	}
}
