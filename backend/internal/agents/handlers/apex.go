// Package handlers contains individual agent implementations.
package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/auth"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/copilot"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/memory"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/metrics"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/pkg/models"
)

// ApexAgent is the Elite Computer Science Engineering Specialist.
type ApexAgent struct {
	upstream *copilot.UpstreamClient
	ollama   *copilot.OllamaClient
	mem      *memory.Store
}

// NewApexAgent creates a new APEX agent.
func NewApexAgent(upstream *copilot.UpstreamClient) *ApexAgent {
	return &ApexAgent{upstream: upstream}
}

// SetMemory attaches a persistent memory store to APEX.
func (a *ApexAgent) SetMemory(m *memory.Store) {
	a.mem = m
}

// SetOllama attaches a local Ollama fallback client to APEX. When set,
// Handle uses it as a second-tier fallback (after upstream Copilot, before
// the canned template response).
func (a *ApexAgent) SetOllama(o *copilot.OllamaClient) {
	a.ollama = o
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
	metrics.IncAgent("APEX")
	userID := memory.UserID(auth.GetGitHubToken(ctx))
	userMsg := copilot.GetLastUserMessage(req)

	if a.mem != nil && userMsg != "" {
		for _, f := range memory.ExtractFacts("APEX", userMsg) {
			_ = a.mem.Remember(userID, "APEX", f.Key, f.Value)
		}
	}

	systemPrompt := a.buildSystemPrompt(userID, userMsg)

	if a.upstream != nil {
		if resp, _ := a.upstream.Forward(ctx, systemPrompt, req); resp != nil {
			metrics.UpstreamTotal.Add(1)
			return resp, nil
		}
	}

	// Second-tier fallback: local Ollama model. Only attempt this when the
	// client is configured and its endpoint is actually reachable, so we
	// don't pay the request timeout on every call when Ollama isn't running.
	if a.ollama != nil && a.ollama.Enabled() {
		if resp, err := a.ollama.Forward(ctx, systemPrompt, req); err == nil && resp != nil {
			metrics.OllamaFallbackTotal.Add(1)
			return resp, nil
		}
	}

	return a.templateResponse(req), nil
}

// buildSystemPrompt constructs APEX's system prompt, prepending remembered context.
func (a *ApexAgent) buildSystemPrompt(userID, query string) string {
	info := a.GetInfo()
	var sb strings.Builder

	if a.mem != nil {
		if ctx := a.mem.FormatContextRelevant(userID, query, 5); ctx != "" {
			sb.WriteString(ctx)
			sb.WriteString("\n")
		}
	}

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
