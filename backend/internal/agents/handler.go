// Package agents provides the agent registry and HTTP handlers.
package agents

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/copilot"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/metrics"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/pkg/models"
)

// agentMentionPattern matches @AGENT_NAME patterns in messages.
var agentMentionPattern = regexp.MustCompile(`@([A-Za-z_]+)`)

// crewMentionPattern matches @CREW:name or @CREW name patterns.
var crewMentionPattern = regexp.MustCompile(`(?i)@CREW[:\s]+(\w+)`)

// Handler provides HTTP handlers for agent endpoints.
type Handler struct {
	registry *Registry
	crews    *CrewRegistry
	pipeline *AgentPipeline
	router   *SemanticRouter
}

// NewHandler creates a new agent handler.
func NewHandler(registry *Registry) *Handler {
	return &Handler{
		registry: registry,
		pipeline: NewAgentPipeline(registry),
		router:   NewSemanticRouter(registry),
	}
}

// WithCrews attaches a crew registry so @CREW:name mentions are expanded.
func (h *Handler) WithCrews(crews *CrewRegistry) *Handler {
	h.crews = crews
	return h
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
// Supports single agent, crew invocation (@CREW:name), and multi-agent pipelines.
func (h *Handler) CopilotWebhook(w http.ResponseWriter, r *http.Request) {
	req, err := copilot.ParseRequest(r)
	if err != nil {
		log.Printf("Error parsing Copilot request: %v", err)
		copilot.WriteError(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	userMessage := copilot.GetLastUserMessage(req)
	if userMessage == "" {
		copilot.WriteError(w, "No user message found", http.StatusBadRequest)
		return
	}

	// Check for @CREW:name mentions first.
	codenames := h.resolveCodenames(userMessage)
	if len(codenames) == 0 {
		metrics.IncRoutingMode("default")
		codenames = []string{"APEX"}
	}

	var resp *models.CopilotResponse
	if len(codenames) > 1 {
		// Sequential pipeline — each agent sees prior agents' reasoning.
		log.Printf("Copilot webhook: pipeline %v", codenames)
		resp, err = h.pipeline.Execute(r.Context(), req, codenames)
	} else {
		codename := codenames[0]
		agent, agentErr := h.registry.Get(codename)
		if agentErr != nil {
			agent, _ = h.registry.Get("APEX")
			codename = "APEX"
		}
		log.Printf("Copilot webhook: routing to agent %s", codename)
		resp, err = agent.Handle(r.Context(), req)
	}

	if err != nil {
		log.Printf("Error handling Copilot request: %v", err)
		copilot.WriteError(w, "Error processing request", http.StatusInternalServerError)
		return
	}

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

// resolveCodenames extracts agent codenames from a message, expanding @CREW:name
// references into their agent lists. Returns unique codenames in order.
func (h *Handler) resolveCodenames(message string) []string {
	// Check for @CREW:name or @CREW name pattern.
	if h.crews != nil {
		if m := crewMentionPattern.FindStringSubmatch(message); len(m) == 2 {
			if crew, ok := h.crews.Get(m[1]); ok {
				log.Printf("Copilot webhook: expanding crew %q → %v", m[1], crew.Agents)
				metrics.IncRoutingMode("crew")
				return crew.Agents
			}
		}
	}

	// Explicit @AGENT mentions always win.
	if names := extractAllAgentCodenames(message); len(names) > 0 {
		metrics.IncRoutingMode("explicit")
		return names
	}

	// No explicit mention: fall back to semantic auto-routing (the woken
	// Bloom skill-cascade) so a bare prompt reaches a relevant specialist
	// instead of the blind APEX default. Returns nil when nothing resonates.
	if h.router != nil {
		if routed := h.router.Route(message); len(routed) > 0 {
			log.Printf("Copilot webhook: semantic routing \u2192 %v", routed)
			metrics.IncRoutingMode("semantic")
			return routed
		}
	}
	return nil
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
