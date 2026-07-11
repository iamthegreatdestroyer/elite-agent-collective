package handlers

import (
	"strings"
	"testing"

	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/pkg/models"
)

func TestWithRetrievedContext(t *testing.T) {
	req := &models.CopilotRequest{Messages: []models.Message{
		{Role: "system", Content: "sys"},
		{Role: "user", Content: "first"},
		{Role: "assistant", Content: "ok"},
		{Role: "user", Content: "the question"},
	}}
	out := withRetrievedContext(req, "SOME NOTES")

	if len(req.Messages) != 4 {
		t.Fatalf("original request must not be mutated, got %d messages", len(req.Messages))
	}
	if len(out.Messages) != 5 {
		t.Fatalf("expected 5 messages after injection, got %d", len(out.Messages))
	}
	if out.Messages[3].Role != "user" || !strings.Contains(out.Messages[3].Content, "SOME NOTES") {
		t.Fatalf("reference must be inserted right before the last user turn: %+v", out.Messages)
	}
	if out.Messages[4].Content != "the question" {
		t.Fatalf("the user's question must remain the last message: %+v", out.Messages[4])
	}
	if withRetrievedContext(req, "") != req {
		t.Fatal("empty block should return the request unchanged")
	}
}

func TestRetrievalInjectMode(t *testing.T) {
	t.Setenv("EAC_RETRIEVER_INJECT", "")
	if retrievalInjectMode() != "message" {
		t.Fatal("default must be point-of-need message mode")
	}
	t.Setenv("EAC_RETRIEVER_INJECT", "system")
	if retrievalInjectMode() != "system" {
		t.Fatal("system override must be honored")
	}
}
