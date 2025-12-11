// Package agents provides the agent registry and HTTP handlers.
package agents

import (
	"fmt"
	"os"
	"sync"

	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/agents/handlers"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/pkg/models"
	"gopkg.in/yaml.v3"
)

// ManifestConfig represents the structure of agents-manifest.yaml.
type ManifestConfig struct {
	Version     string        `yaml:"version"`
	Name        string        `yaml:"name"`
	Description string        `yaml:"description"`
	Tiers       []TierConfig  `yaml:"tiers"`
	Agents      []AgentConfig `yaml:"agents"`
}

// TierConfig represents a tier definition in the manifest.
type TierConfig struct {
	ID          int    `yaml:"id"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// AgentConfig represents an agent definition in the manifest.
type AgentConfig struct {
	ID            string   `yaml:"id"`
	Codename      string   `yaml:"codename"`
	Tier          int      `yaml:"tier"`
	Name          string   `yaml:"name"`
	Description   string   `yaml:"description"`
	Philosophy    string   `yaml:"philosophy"`
	Keywords      []string `yaml:"keywords"`
	Directives    []string `yaml:"directives"`
	Examples      []string `yaml:"examples"`
	Collaborators []string `yaml:"collaborators"`
}

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
// It attempts to load from .github/agents/ first, falling back to hardcoded definitions.
func DefaultRegistry() *Registry {
	registry := NewRegistry()
	if err := RegisterAllAgents(registry); err != nil {
		// Log error but don't panic - we may have loaded agents via fallback
		fmt.Fprintf(os.Stderr, "Warning: RegisterAllAgents returned error: %v\n", err)
	}
	return registry
}

// LoadManifest reads and parses the agents manifest YAML file.
func LoadManifest(path string) (*ManifestConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest file: %w", err)
	}

	var manifest ManifestConfig
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest YAML: %w", err)
	}

	return &manifest, nil
}

// RegistryFromManifest creates a registry by loading agents from a manifest file.
// It registers handlers for all agents defined in the manifest.
func RegistryFromManifest(manifestPath string) (*Registry, error) {
	manifest, err := LoadManifest(manifestPath)
	if err != nil {
		return nil, err
	}

	registry := NewRegistry()

	for _, agentConfig := range manifest.Agents {
		agent := models.Agent{
			ID:         agentConfig.ID,
			Codename:   agentConfig.Codename,
			Tier:       agentConfig.Tier,
			Specialty:  agentConfig.Name,
			Philosophy: agentConfig.Philosophy,
			Directives: agentConfig.Directives,
		}

		// Use custom handler for APEX, base handler for others
		if agentConfig.Codename == "APEX" {
			registry.Register(handlers.NewApexAgent())
		} else {
			registry.Register(handlers.NewBaseAgent(agent))
		}
	}

	return registry, nil
}

// ValidateManifest checks that a manifest contains all required agents and fields.
func ValidateManifest(manifest *ManifestConfig) error {
	if len(manifest.Agents) != 40 {
		return fmt.Errorf("expected 40 agents, found %d", len(manifest.Agents))
	}

	// Check for duplicate codenames
	seen := make(map[string]bool)
	for _, agent := range manifest.Agents {
		if seen[agent.Codename] {
			return fmt.Errorf("duplicate agent codename: %s", agent.Codename)
		}
		seen[agent.Codename] = true

		// Validate required fields
		if agent.ID == "" {
			return fmt.Errorf("agent missing ID")
		}
		if agent.Codename == "" {
			return fmt.Errorf("agent %s missing codename", agent.ID)
		}
		if agent.Name == "" {
			return fmt.Errorf("agent %s missing name", agent.Codename)
		}
		if agent.Philosophy == "" {
			return fmt.Errorf("agent %s missing philosophy", agent.Codename)
		}
		if len(agent.Directives) == 0 {
			return fmt.Errorf("agent %s missing directives", agent.Codename)
		}
	}

	// Verify all 8 tiers are present with correct counts
	tierCounts := make(map[int]int)
	for _, agent := range manifest.Agents {
		tierCounts[agent.Tier]++
	}

	expectedTiers := map[int]int{
		1: 5, 2: 12, 3: 2, 4: 1, 5: 5, 6: 5, 7: 5, 8: 5,
	}

	for tier, expected := range expectedTiers {
		if tierCounts[tier] != expected {
			return fmt.Errorf("tier %d: expected %d agents, found %d", tier, expected, tierCounts[tier])
		}
	}

	return nil
}
