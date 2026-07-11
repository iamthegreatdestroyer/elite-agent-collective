package memory

import (
	"context"
	"strings"
	"testing"
)

func TestCreatePlanForSubtasks(t *testing.T) {
	sp := NewStrategicPlanner(PlanningConfig{PlanCachingEnabled: true})
	goal := &Goal{ID: "g1", Name: "test goal", Description: "d", Priority: PriorityHigh, Status: GoalActive}
	subs := []string{"design the auth layer", "write the crypto module", "add integration tests"}

	agentFor := func(s string) string {
		switch {
		case strings.Contains(s, "crypto"):
			return "CIPHER"
		case strings.Contains(s, "test"):
			return "ECLIPSE"
		default:
			return "ARCHITECT"
		}
	}

	plan, err := sp.CreatePlanForSubtasks(context.Background(), goal, subs, agentFor)
	if err != nil {
		t.Fatalf("CreatePlanForSubtasks error: %v", err)
	}
	if len(plan.Actions) != 3 {
		t.Fatalf("want 3 actions, got %d", len(plan.Actions))
	}
	if plan.Actions[0].AgentRequired != "ARCHITECT" ||
		plan.Actions[1].AgentRequired != "CIPHER" ||
		plan.Actions[2].AgentRequired != "ECLIPSE" {
		t.Fatalf("agent assignments wrong: %q %q %q",
			plan.Actions[0].AgentRequired, plan.Actions[1].AgentRequired, plan.Actions[2].AgentRequired)
	}
	if !plan.Feasible {
		t.Fatal("plan should be feasible")
	}
	if sp.GetLookaheadTree() == nil {
		t.Fatal("lookahead tree should have been built")
	}

	// Second call for the same goal ID should hit the plan cache.
	plan2, err := sp.CreatePlanForSubtasks(context.Background(), goal, subs, agentFor)
	if err != nil {
		t.Fatal(err)
	}
	if plan2.ID != plan.ID {
		t.Fatalf("expected cached plan (same ID), got %q vs %q", plan2.ID, plan.ID)
	}
}
