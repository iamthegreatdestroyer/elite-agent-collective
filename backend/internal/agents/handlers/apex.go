// Package handlers contains individual agent implementations.
package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/copilot"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/pkg/models"
)

// ApexAgent is the Elite Computer Science Engineering Specialist.
type ApexAgent struct {
	upstream *copilot.UpstreamClient
}

// NewApexAgent creates a new APEX agent.
func NewApexAgent(upstream *copilot.UpstreamClient) *ApexAgent {
	return &ApexAgent{upstream: upstream}
}

// GetInfo returns APEX agent metadata.
func (a *ApexAgent) GetInfo() models.Agent {
	return models.Agent{
		ID:         "01",
		Codename:   "APEX",
		Tier:       1,
		Specialty:  "Elite Computer Science Engineering",
		Philosophy: "Every problem has an elegant solution waiting to be discovered.",
		Directives: []string{
			"Deliver production-grade, enterprise-quality code",
			"Apply computer science fundamentals at the deepest level",
			"Anticipate edge cases before they manifest",
			"Optimize for both performance and maintainability",
			"Evolve continuously through pattern recognition",
		},
	}
}

// Handle processes a Copilot request using APEX's methodology.
// Forwards to upstream with APEX's system prompt; falls back to template response.
func (a *ApexAgent) Handle(ctx context.Context, req *models.CopilotRequest) (*models.CopilotResponse, error) {
	if a.upstream != nil {
		if resp, _ := a.upstream.Forward(ctx, apexSystemPrompt(), req); resp != nil {
			return resp, nil
		}
	}
	return a.templateResponse(req), nil
}

func apexSystemPrompt() string {
	info := (&ApexAgent{}).GetInfo()
	var sb strings.Builder
	fmt.Fprintf(&sb, "You are %s, the %s Specialist in the Elite Agent Collective.\n", info.Codename, info.Specialty)
	fmt.Fprintf(&sb, "Philosophy: %s\n", info.Philosophy)
	sb.WriteString("Core Directives:\n")
	for i, d := range info.Directives {
		fmt.Fprintf(&sb, "%d. %s\n", i+1, d)
	}
	sb.WriteString(`
Your methodology — apply it to every request:
1. DECOMPOSE → Break problem into atomic components
2. CLASSIFY  → Map to known patterns & paradigms
3. THEORIZE  → Generate multiple solution hypotheses
4. ANALYZE   → Evaluate time/space complexity, edge cases
5. SYNTHESIZE → Construct optimal solution with patterns
6. VALIDATE  → Mental execution, trace through
7. DOCUMENT  → Clear explanation with trade-offs

Respond with the expertise of APEX. Be precise, actionable, and production-grade.`)
	return sb.String()
}

func (a *ApexAgent) templateResponse(req *models.CopilotRequest) *models.CopilotResponse {
	userMessage := copilot.GetLastUserMessage(req)
	content := fmt.Sprintf(`As APEX, the Elite Computer Science Engineering Specialist, I'll help you with: %s

My approach follows these principles:
1. DECOMPOSE → Break problem into atomic components
2. CLASSIFY → Map to known patterns & paradigms
3. THEORIZE → Generate multiple solution hypotheses
4. ANALYZE → Evaluate time/space complexity, edge cases
5. SYNTHESIZE → Construct optimal solution with patterns
6. VALIDATE → Mental execution, trace through
7. DOCUMENT → Clear explanation with trade-offs

Let me analyze your request...`, userMessage)
	return copilot.NewResponse(content)
}
