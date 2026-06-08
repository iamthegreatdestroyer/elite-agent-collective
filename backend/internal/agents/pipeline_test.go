package agents

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/pkg/models"
)

// stubAgent is a minimal AgentHandler for pipeline tests.
type stubAgent struct {
	codename string
	output   string
}

func (s *stubAgent) GetInfo() models.Agent {
	return models.Agent{Codename: s.codename}
}

func (s *stubAgent) Handle(_ context.Context, _ *models.CopilotRequest) (*models.CopilotResponse, error) {
	return &models.CopilotResponse{
		Choices: []models.Choice{{Message: models.Message{Role: "assistant", Content: s.output}}},
	}, nil
}

// errorAgent always fails.
type errorAgent struct{ codename string }

func (e *errorAgent) GetInfo() models.Agent { return models.Agent{Codename: e.codename} }
func (e *errorAgent) Handle(_ context.Context, _ *models.CopilotRequest) (*models.CopilotResponse, error) {
	return nil, fmt.Errorf("agent unavailable")
}

func makeRegistry(agents ...models.AgentHandler) *Registry {
	r := NewRegistry()
	for _, a := range agents {
		r.Register(a)
	}
	return r
}

func TestPipeline_SingleAgent(t *testing.T) {
	r := makeRegistry(&stubAgent{"APEX", "apex response"})
	p := NewAgentPipeline(r)
	req := &models.CopilotRequest{Messages: []models.Message{{Role: "user", Content: "hello"}}}
	resp, err := p.Execute(context.Background(), req, []string{"APEX"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if len(resp.Choices) == 0 {
		t.Fatal("no choices in response")
	}
	body := resp.Choices[0].Message.Content
	if !strings.Contains(body, "APEX") {
		t.Errorf("expected APEX in response, got: %q", body)
	}
	if !strings.Contains(body, "apex response") {
		t.Errorf("expected agent output in response, got: %q", body)
	}
}

func TestPipeline_MultiAgent_Sequential(t *testing.T) {
	r := makeRegistry(
		&stubAgent{"APEX", "apex output"},
		&stubAgent{"ECLIPSE", "eclipse output"},
	)
	p := NewAgentPipeline(r)
	req := &models.CopilotRequest{Messages: []models.Message{{Role: "user", Content: "review this"}}}
	resp, err := p.Execute(context.Background(), req, []string{"APEX", "ECLIPSE"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	body := resp.Choices[0].Message.Content
	if !strings.Contains(body, "APEX") || !strings.Contains(body, "ECLIPSE") {
		t.Errorf("expected both agents in response, got: %q", body)
	}
	// Pipeline header should show multi-agent with arrow.
	if !strings.Contains(body, "Multi-Agent") {
		t.Errorf("expected 'Multi-Agent' in response, got: %q", body)
	}
	if !strings.Contains(body, "APEX → ECLIPSE") {
		t.Errorf("expected pipeline header APEX → ECLIPSE, got: %q", body)
	}
}

func TestPipeline_SkipsUnavailableAgent(t *testing.T) {
	r := makeRegistry(
		&stubAgent{"APEX", "apex output"},
		&errorAgent{"BROKEN"},
		&stubAgent{"ECLIPSE", "eclipse output"},
	)
	p := NewAgentPipeline(r)
	req := &models.CopilotRequest{Messages: []models.Message{{Role: "user", Content: "test"}}}
	resp, err := p.Execute(context.Background(), req, []string{"APEX", "BROKEN", "ECLIPSE"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	body := resp.Choices[0].Message.Content
	if !strings.Contains(body, "apex output") {
		t.Errorf("expected apex output, got: %q", body)
	}
	if !strings.Contains(body, "eclipse output") {
		t.Errorf("expected eclipse output, got: %q", body)
	}
	if !strings.Contains(body, "unavailable") {
		t.Errorf("expected 'unavailable' for BROKEN agent, got: %q", body)
	}
}

func TestPipeline_UnknownAgent(t *testing.T) {
	r := makeRegistry(&stubAgent{"APEX", "apex output"})
	p := NewAgentPipeline(r)
	req := &models.CopilotRequest{Messages: []models.Message{{Role: "user", Content: "test"}}}
	resp, err := p.Execute(context.Background(), req, []string{"APEX", "NOEXIST"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	body := resp.Choices[0].Message.Content
	if !strings.Contains(body, "apex output") {
		t.Errorf("expected apex output in response")
	}
}

func TestCloneRequest(t *testing.T) {
	orig := &models.CopilotRequest{
		Messages: []models.Message{{Role: "user", Content: "hello"}},
		Model:    "gpt-4",
		Stream:   true,
	}
	clone := cloneRequest(orig)
	if clone == orig {
		t.Error("clone should not be the same pointer as original")
	}
	if clone.Model != orig.Model {
		t.Errorf("model mismatch: %q vs %q", clone.Model, orig.Model)
	}
	// Mutating clone should not affect original.
	clone.Messages[0].Content = "modified"
	if orig.Messages[0].Content == "modified" {
		t.Error("clone shares slice with original")
	}
}

func TestAppendAssistant(t *testing.T) {
	req := &models.CopilotRequest{Messages: []models.Message{{Role: "user", Content: "hi"}}}
	next := appendAssistant(req, "assistant reply")
	if len(next.Messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(next.Messages))
	}
	if next.Messages[1].Role != "assistant" || next.Messages[1].Content != "assistant reply" {
		t.Errorf("appended message incorrect: %+v", next.Messages[1])
	}
	// Original must be unchanged.
	if len(req.Messages) != 1 {
		t.Error("original request was mutated")
	}
}
