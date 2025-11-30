// Package agents provides the agent registry and HTTP handlers.
package agents

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/copilot"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/pkg/models"
)

// agentMentionPattern matches @AGENT_NAME patterns in messages.
var agentMentionPattern = regexp.MustCompile(`@([A-Za-z]+)`)

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

	// Support streaming responses if requested
	if req.Stream {
		if err := copilot.WriteStreamingResponse(w, resp.Choices[0].Message.Content); err != nil {
			log.Printf("Error writing streaming response: %v", err)
		}
		return
	}
	
	if err := copilot.WriteResponse(w, resp); err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

// CopilotWebhook handles POST /copilot - main Copilot webhook endpoint.
// This endpoint parses the agent codename from the message content.
// Supports multi-agent collaboration when multiple @AGENT_NAME mentions are found.
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
	
	// Extract all agent codenames from the message (supports multi-agent collaboration)
	codenames := extractAllAgentCodenames(userMessage)
	
	// If no agents specified, default to APEX
	if len(codenames) == 0 {
		codenames = []string{"APEX"}
	}
	
	// Handle multi-agent collaboration
	if len(codenames) > 1 {
		h.handleMultiAgentRequest(w, r, req, codenames)
		return
	}
	
	// Single agent invocation
	codename := codenames[0]
	agent, err := h.registry.Get(codename)
	if err != nil {
		// Fall back to APEX if agent not found
		agent, _ = h.registry.Get("APEX")
		codename = "APEX"
	}
	
	log.Printf("Copilot webhook: routing to agent %s", codename)
	
	resp, err := agent.Handle(r.Context(), req)
	if err != nil {
		log.Printf("Error handling Copilot request: %v", err)
		copilot.WriteError(w, "Error processing request", http.StatusInternalServerError)
		return
	}

	// Support streaming responses if requested
	if req.Stream {
		if err := copilot.WriteStreamingResponse(w, resp.Choices[0].Message.Content); err != nil {
			log.Printf("Error writing streaming response: %v", err)
		}
		return
	}
	
	if err := copilot.WriteResponse(w, resp); err != nil {
		log.Printf("Error writing Copilot response: %v", err)
	}
}

// handleMultiAgentRequest handles requests that invoke multiple agents.
// It combines responses from all specified agents into a single response.
func (h *Handler) handleMultiAgentRequest(w http.ResponseWriter, r *http.Request, req *models.CopilotRequest, codenames []string) {
	log.Printf("Copilot webhook: multi-agent collaboration with agents: %v", codenames)
	
	var responses []string
	var validAgents []string
	
	for _, codename := range codenames {
		agent, err := h.registry.Get(codename)
		if err != nil {
			log.Printf("Agent %s not found, skipping", codename)
			continue
		}
		
		resp, err := agent.Handle(r.Context(), req)
		if err != nil {
			log.Printf("Error from agent %s: %v", codename, err)
			continue
		}
		
		if len(resp.Choices) > 0 {
			responses = append(responses, resp.Choices[0].Message.Content)
			validAgents = append(validAgents, codename)
		}
	}
	
	if len(responses) == 0 {
		copilot.WriteError(w, "No valid agents could process the request", http.StatusInternalServerError)
		return
	}
	
	// Combine responses with clear separation
	var combinedContent strings.Builder
	combinedContent.WriteString(fmt.Sprintf("## Multi-Agent Collaboration: %s\n\n", strings.Join(validAgents, " + ")))
	
	for i, content := range responses {
		if i > 0 {
			combinedContent.WriteString("\n---\n\n")
		}
		combinedContent.WriteString(content)
	}
	
	combinedResp := copilot.NewResponse(combinedContent.String())
	
	// Support streaming responses if requested
	if req.Stream {
		if err := copilot.WriteStreamingResponse(w, combinedContent.String()); err != nil {
			log.Printf("Error writing streaming response: %v", err)
		}
		return
	}
	
	if err := copilot.WriteResponse(w, combinedResp); err != nil {
		log.Printf("Error writing multi-agent response: %v", err)
	}
}

// extractAgentCodename extracts the first agent codename from a message.
// It looks for @CODENAME patterns at the start of the message.
func extractAgentCodename(message string) string {
	codenames := extractAllAgentCodenames(message)
	if len(codenames) == 0 {
		return ""
	}
	return codenames[0]
}

// extractAllAgentCodenames extracts all agent codenames from a message.
// It looks for @CODENAME patterns anywhere in the message.
// Returns unique codenames in the order they appear.
func extractAllAgentCodenames(message string) []string {
	matches := agentMentionPattern.FindAllStringSubmatch(message, -1)
	if len(matches) == 0 {
		return nil
	}
	
	seen := make(map[string]bool)
	var codenames []string
	
	for _, match := range matches {
		if len(match) >= 2 {
			codename := strings.ToUpper(match[1])
			if !seen[codename] {
				seen[codename] = true
				codenames = append(codenames, codename)
			}
		}
	}
	
	return codenames
}
