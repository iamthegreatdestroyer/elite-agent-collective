// Package agents provides the agent registry and HTTP handlers.
package agents

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/copilot"
)

// Handler provides HTTP handlers for agent endpoints.
type Handler struct {
	registry *Registry
}

// NewHandler creates a new agent handler.
func NewHandler(registry *Registry) *Handler {
	return &Handler{
		registry: registry,
	}
}

// ListAgents handles GET /agents - returns all registered agents.
func (h *Handler) ListAgents(w http.ResponseWriter, r *http.Request) {
	agents := h.registry.List()
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(agents); err != nil {
		log.Printf("Error encoding agents list: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// GetAgent handles GET /agents/{codename} - returns a specific agent's info.
func (h *Handler) GetAgent(w http.ResponseWriter, r *http.Request) {
	codename := chi.URLParam(r, "codename")
	
	agent, err := h.registry.Get(codename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(agent.GetInfo()); err != nil {
		log.Printf("Error encoding agent info: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// InvokeAgent handles POST /agents/{codename}/invoke - invokes a specific agent.
func (h *Handler) InvokeAgent(w http.ResponseWriter, r *http.Request) {
	codename := chi.URLParam(r, "codename")
	
	agent, err := h.registry.Get(codename)
	if err != nil {
		copilot.WriteError(w, err.Error(), http.StatusNotFound)
		return
	}
	
	req, err := copilot.ParseRequest(r)
	if err != nil {
		log.Printf("Error parsing request: %v", err)
		copilot.WriteError(w, "Invalid request format", http.StatusBadRequest)
		return
	}
	
	log.Printf("Invoking agent %s with %d messages", codename, len(req.Messages))
	
	resp, err := agent.Handle(r.Context(), req)
	if err != nil {
		log.Printf("Error handling request: %v", err)
		copilot.WriteError(w, "Error processing request", http.StatusInternalServerError)
		return
	}
	
	if err := copilot.WriteResponse(w, resp); err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

// CopilotWebhook handles POST /copilot - main Copilot webhook endpoint.
// This endpoint parses the agent codename from the message content.
func (h *Handler) CopilotWebhook(w http.ResponseWriter, r *http.Request) {
	req, err := copilot.ParseRequest(r)
	if err != nil {
		log.Printf("Error parsing Copilot request: %v", err)
		copilot.WriteError(w, "Invalid request format", http.StatusBadRequest)
		return
	}
	
	// Get the last user message
	userMessage := copilot.GetLastUserMessage(req)
	if userMessage == "" {
		copilot.WriteError(w, "No user message found", http.StatusBadRequest)
		return
	}
	
	// Try to extract agent codename from the message (e.g., "@APEX help me")
	codename := extractAgentCodename(userMessage)
	if codename == "" {
		// Default to APEX if no agent is specified
		codename = "APEX"
	}
	
	agent, err := h.registry.Get(codename)
	if err != nil {
		// Fall back to APEX if agent not found
		agent, _ = h.registry.Get("APEX")
	}
	
	log.Printf("Copilot webhook: routing to agent %s", codename)
	
	resp, err := agent.Handle(r.Context(), req)
	if err != nil {
		log.Printf("Error handling Copilot request: %v", err)
		copilot.WriteError(w, "Error processing request", http.StatusInternalServerError)
		return
	}
	
	if err := copilot.WriteResponse(w, resp); err != nil {
		log.Printf("Error writing Copilot response: %v", err)
	}
}

// extractAgentCodename extracts an agent codename from a message.
// It looks for @CODENAME patterns at the start of the message.
func extractAgentCodename(message string) string {
	if len(message) < 2 || message[0] != '@' {
		return ""
	}
	
	// Find the end of the codename (first space or end of string)
	end := 1
	for end < len(message) && message[end] != ' ' && message[end] != '\n' && message[end] != '\t' {
		end++
	}
	
	return message[1:end]
}
