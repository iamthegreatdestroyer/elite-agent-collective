package handlers

import (
	"context"
	"testing"

	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/pkg/models"
)

func TestApexAgentGetInfo(t *testing.T) {
	agent := NewApexAgent()
	info := agent.GetInfo()

	if info.Codename != "APEX" {
		t.Errorf("expected codename 'APEX', got %s", info.Codename)
	}

	if info.Tier != 1 {
		t.Errorf("expected tier 1, got %d", info.Tier)
	}

	if info.ID != "01" {
		t.Errorf("expected ID '01', got %s", info.ID)
	}

	if len(info.Directives) == 0 {
		t.Error("expected non-empty directives")
	}
}

func TestApexAgentHandle(t *testing.T) {
	agent := NewApexAgent()
	req := &models.CopilotRequest{
		Messages: []models.Message{
			{Role: "user", Content: "Help me design an algorithm"},
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
	if content == "" {
		t.Error("expected non-empty response content")
	}

	// Check that the response mentions APEX's methodology
	if !containsString(content, "APEX") {
		t.Error("expected response to mention APEX")
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStringHelper(s, substr))
}

func containsStringHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
