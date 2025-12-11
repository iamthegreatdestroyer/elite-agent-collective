// Package models contains data models for the Elite Agent Collective backend.
package models

import "context"

// Agent represents a single agent in the Elite Agent Collective.
type Agent struct {
	ID            string   `json:"id"`
	Codename      string   `json:"codename"`
	Tier          int      `json:"tier"`
	Name          string   `json:"name"`
	Specialty     string   `json:"specialty"`
	Philosophy    string   `json:"philosophy"`
	Directives    []string `json:"directives"`
	Keywords      []string `json:"keywords"`
	Examples      []string `json:"examples"`
	Collaborators []string `json:"collaborators"`
	Category      string   `json:"category"`
	MarkdownPath  string   `json:"-"` // Internal: path to .agent.md file
}

// CopilotRequest represents a request from GitHub Copilot.
type CopilotRequest struct {
	Messages []Message `json:"messages"`
	Model    string    `json:"model"`
	Stream   bool      `json:"stream"`
}

// Message represents a single message in a conversation.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// CopilotResponse represents a response to GitHub Copilot.
type CopilotResponse struct {
	Choices []Choice `json:"choices"`
}

// Choice represents a single response choice.
type Choice struct {
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// AgentHandler defines the interface for agent handlers.
type AgentHandler interface {
	// Handle processes a Copilot request and returns a response.
	Handle(ctx context.Context, request *CopilotRequest) (*CopilotResponse, error)
	// GetInfo returns the agent's metadata.
	GetInfo() Agent
}
