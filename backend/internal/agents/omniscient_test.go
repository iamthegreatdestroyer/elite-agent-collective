package agents

import (
	"context"
	"strings"
	"testing"

	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/pkg/models"
)

func TestOmniscientAgent_Registered(t *testing.T) {
	reg := DefaultRegistry()
	h, err := reg.Get("OMNISCIENT")
	if err != nil {
		t.Fatalf("OMNISCIENT not registered: %v", err)
	}
	if _, ok := h.(*OmniscientAgent); !ok {
		t.Fatalf("OMNISCIENT should resolve to *OmniscientAgent, got %T", h)
	}
	if info := h.GetInfo(); info.Codename != "OMNISCIENT" || info.Tier != 4 {
		t.Fatalf("unexpected OMNISCIENT info: %+v", info)
	}
}

func TestOmniscientAgent_Orchestrates(t *testing.T) {
	reg := DefaultRegistry()
	h, _ := reg.Get("OMNISCIENT")

	// Multi-clause prompt -> multi-agent orchestration. With upstream + ollama
	// both unavailable in tests, specialists return their template response;
	// OMNISCIENT must still synthesize a non-empty, headed answer and never err.
	req := &models.CopilotRequest{Messages: []models.Message{
		{Role: "user", Content: "design the encryption architecture and write documentation for the module"},
	}}
	resp, err := h.Handle(context.Background(), req)
	if err != nil {
		t.Fatalf("Handle error: %v", err)
	}
	if resp == nil || len(resp.Choices) == 0 || resp.Choices[0].Message.Content == "" {
		t.Fatal("empty response")
	}
	body := resp.Choices[0].Message.Content
	if !strings.Contains(body, "[OMNISCIENT]") {
		t.Fatalf("expected an [OMNISCIENT] header, got: %q", body[:min(120, len(body))])
	}
	if !strings.Contains(body, "feasible=") {
		t.Fatalf("expected the planner summary in the header, got: %q", body[:min(120, len(body))])
	}
}

func TestOmniscientAgent_EmptyPromptFailOpen(t *testing.T) {
	reg := DefaultRegistry()
	h, _ := reg.Get("OMNISCIENT")
	resp, err := h.Handle(context.Background(), &models.CopilotRequest{})
	if err != nil {
		t.Fatalf("Handle error: %v", err)
	}
	if resp == nil || len(resp.Choices) == 0 {
		t.Fatal("expected a response even for an empty prompt")
	}
}

func TestOmniscientAgent_NoSelfRecursion(t *testing.T) {
	// A prompt loaded with orchestration keywords must NOT route back to
	// OMNISCIENT (which would recurse). It should still return a real answer.
	reg := DefaultRegistry()
	h, _ := reg.Get("OMNISCIENT")
	req := &models.CopilotRequest{Messages: []models.Message{
		{Role: "user", Content: "orchestrate the collective and coordinate meta-learning evolution"},
	}}
	resp, err := h.Handle(context.Background(), req)
	if err != nil {
		t.Fatalf("Handle error: %v", err)
	}
	if resp == nil || len(resp.Choices) == 0 || resp.Choices[0].Message.Content == "" {
		t.Fatal("empty response")
	}
}
