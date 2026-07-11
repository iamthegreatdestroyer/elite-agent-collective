// Package handlers contains individual agent implementations.
package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/auth"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/copilot"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/memory"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/pkg/models"
)

// BaseAgent provides common functionality for all agents.
type BaseAgent struct {
	info     models.Agent
	upstream *copilot.UpstreamClient
	ollama   *copilot.OllamaClient
	mem      *memory.Store
}

// NewBaseAgent creates a new base agent with the given info.
func NewBaseAgent(info models.Agent, upstream *copilot.UpstreamClient) *BaseAgent {
	return &BaseAgent{info: info, upstream: upstream}
}

// SetMemory attaches a persistent memory store to the agent.
func (a *BaseAgent) SetMemory(m *memory.Store) {
	a.mem = m
}

// SetOllama attaches a local Ollama fallback client to the agent. When set,
// Handle uses it as a second-tier fallback (after upstream Copilot, before
// the canned template response).
func (a *BaseAgent) SetOllama(o *copilot.OllamaClient) {
	a.ollama = o
}

// GetInfo returns the agent's metadata.
func (a *BaseAgent) GetInfo() models.Agent {
	return a.info
}

// Handle processes a Copilot request using the base implementation.
// It tries to forward to the upstream Copilot API with a system prompt; on failure
// or when upstream is disabled, it falls back to a template response.
func (a *BaseAgent) Handle(ctx context.Context, req *models.CopilotRequest) (*models.CopilotResponse, error) {
	userID := memory.UserID(auth.GetGitHubToken(ctx))
	userMsg := copilot.GetLastUserMessage(req)

	// Store facts extracted from the user's message before responding.
	if a.mem != nil && userMsg != "" {
		for _, f := range memory.ExtractFacts(a.info.Codename, userMsg) {
			_ = a.mem.Remember(userID, a.info.Codename, f.Key, f.Value)
		}
	}

	systemPrompt := a.buildSystemPrompt(userID, userMsg)

	if a.upstream != nil {
		if resp, _ := a.upstream.Forward(ctx, systemPrompt, req); resp != nil {
			return resp, nil
		}
	}

	// Second-tier fallback: local Ollama model. Only attempt this when the
	// client is configured and its endpoint is actually reachable, so we
	// don't pay the request timeout on every call when Ollama isn't running.
	if a.ollama != nil && a.ollama.Enabled() {
		if resp, err := a.ollama.Forward(ctx, systemPrompt, req); err == nil && resp != nil {
			return resp, nil
		}
	}

	return a.templateResponse(req), nil
}

// buildSystemPrompt constructs the agent's system prompt, prepending any
// remembered context for the current user.
func (a *BaseAgent) buildSystemPrompt(userID, query string) string {
	var sb strings.Builder

	if a.mem != nil {
		if ctx := a.mem.FormatContextRelevant(userID, query, 5); ctx != "" {
			sb.WriteString(ctx)
			sb.WriteString("\n")
		}
	}

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
