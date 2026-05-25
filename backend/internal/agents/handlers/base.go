// Package handlers contains individual agent implementations.
package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/copilot"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/pkg/models"
)

// BaseAgent provides common functionality for all agents.
type BaseAgent struct {
	info     models.Agent
	upstream *copilot.UpstreamClient
}

// NewBaseAgent creates a new base agent with the given info.
func NewBaseAgent(info models.Agent, upstream *copilot.UpstreamClient) *BaseAgent {
	return &BaseAgent{
		info:     info,
		upstream: upstream,
	}
}

// GetInfo returns the agent's metadata.
func (a *BaseAgent) GetInfo() models.Agent {
	return a.info
}

// Handle processes a Copilot request using the base implementation.
// It tries to forward to the upstream Copilot API with a system prompt; on failure
// or when upstream is disabled, it falls back to a template response.
func (a *BaseAgent) Handle(ctx context.Context, req *models.CopilotRequest) (*models.CopilotResponse, error) {
	if a.upstream != nil {
		systemPrompt := a.buildSystemPrompt()
		if resp, _ := a.upstream.Forward(ctx, systemPrompt, req); resp != nil {
			return resp, nil
		}
	}
	return a.templateResponse(req), nil
}

// buildSystemPrompt constructs the agent's system prompt from its metadata.
func (a *BaseAgent) buildSystemPrompt() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "You are %s, the %s Specialist in the Elite Agent Collective.\n", a.info.Codename, a.info.Specialty)
	fmt.Fprintf(&sb, "Philosophy: %s\n", a.info.Philosophy)
	sb.WriteString("Core Directives:\n")
	for i, d := range a.info.Directives {
		fmt.Fprintf(&sb, "%d. %s\n", i+1, d)
	}
	fmt.Fprintf(&sb, "Respond with the expertise of %s. Be precise, actionable, and production-grade.", a.info.Codename)
	return sb.String()
}

// templateResponse returns a canned response when upstream is unavailable.
func (a *BaseAgent) templateResponse(req *models.CopilotRequest) *models.CopilotResponse {
	userMessage := copilot.GetLastUserMessage(req)
	content := fmt.Sprintf(`As %s, the %s Specialist, I'll help you with: %s

My philosophy: %s

I'm ready to assist you with my expertise. Here are my core directives:
%s
How can I help you today?`,
		a.info.Codename,
		a.info.Specialty,
		userMessage,
		a.info.Philosophy,
		formatDirectives(a.info.Directives))
	return copilot.NewResponse(content)
}

// formatDirectives formats the directives as a numbered list.
func formatDirectives(directives []string) string {
	result := ""
	for i, d := range directives {
		result += fmt.Sprintf("%d. %s\n", i+1, d)
	}
	return result
}
