// Package handlers contains individual agent implementations.
package handlers

import (
	"context"
	"fmt"

	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/copilot"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/pkg/models"
)

// BaseAgent provides common functionality for all agents.
type BaseAgent struct {
	info     models.Agent
	greeting string
}

// NewBaseAgent creates a new base agent with the given info.
func NewBaseAgent(info models.Agent) *BaseAgent {
	return &BaseAgent{
		info: info,
	}
}

// GetInfo returns the agent's metadata.
func (a *BaseAgent) GetInfo() models.Agent {
	return a.info
}

// Handle processes a Copilot request using the base implementation.
func (a *BaseAgent) Handle(ctx context.Context, req *models.CopilotRequest) (*models.CopilotResponse, error) {
	userMessage := copilot.GetLastUserMessage(req)
	
	response := fmt.Sprintf(`As %s, the %s Specialist, I'll help you with: %s

My philosophy: %s

I'm ready to assist you with my expertise. Here are my core directives:
%s

How can I help you today?`, 
		a.info.Codename,
		a.info.Specialty,
		userMessage,
		a.info.Philosophy,
		formatDirectives(a.info.Directives))
	
	return copilot.NewResponse(response), nil
}

// formatDirectives formats the directives as a numbered list.
func formatDirectives(directives []string) string {
	result := ""
	for i, d := range directives {
		result += fmt.Sprintf("%d. %s\n", i+1, d)
	}
	return result
}
