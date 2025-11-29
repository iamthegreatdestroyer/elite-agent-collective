// Package agents provides the agent registry and HTTP handlers.
package agents

import (
	"fmt"
	"sync"

	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/pkg/models"
)

// Registry maintains a registry of all available agents.
type Registry struct {
	agents map[string]models.AgentHandler
	mu     sync.RWMutex
}

// NewRegistry creates a new agent registry.
func NewRegistry() *Registry {
	return &Registry{
		agents: make(map[string]models.AgentHandler),
	}
}

// Register adds an agent to the registry.
func (r *Registry) Register(handler models.AgentHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	info := handler.GetInfo()
	r.agents[info.Codename] = handler
}

// Get retrieves an agent by codename.
func (r *Registry) Get(codename string) (models.AgentHandler, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	handler, ok := r.agents[codename]
	if !ok {
		return nil, fmt.Errorf("agent not found: %s", codename)
	}
	return handler, nil
}

// List returns all registered agents.
func (r *Registry) List() []models.Agent {
	r.mu.RLock()
	defer r.mu.RUnlock()
	agents := make([]models.Agent, 0, len(r.agents))
	for _, handler := range r.agents {
		agents = append(agents, handler.GetInfo())
	}
	return agents
}

// Count returns the number of registered agents.
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.agents)
}

// DefaultRegistry creates a registry with all 40 agents registered.
func DefaultRegistry() *Registry {
	registry := NewRegistry()
	RegisterAllAgents(registry)
	return registry
}
