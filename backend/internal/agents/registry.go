// Package agents provides the agent registry and HTTP handlers.
package agents

import (
	"fmt"
	"os"
	"sync"

	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/agents/handlers"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/copilot"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/memory"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/retrieval"
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

// DefaultRegistry creates a registry with all 40 agents registered (no upstream client).
// It attempts to load from .github/agents/ first, falling back to hardcoded definitions.
func DefaultRegistry() *Registry {
	registry := NewRegistry()
	if err := RegisterAllAgents(registry); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: RegisterAllAgents returned error: %v\n", err)
	}
	return registry
}

// DefaultRegistryWithUpstream creates a registry with all 40 agents registered,
// wired to the given upstream client.
func DefaultRegistryWithUpstream(uc *copilot.UpstreamClient) *Registry {
	registry := NewRegistry()
	if err := RegisterAllAgentsWithUpstream(registry, uc); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: RegisterAllAgentsWithUpstream returned error: %v\n", err)
	}
	return registry
}

// InjectMemory attaches mem to every agent handler that supports persistent memory.
// Call this after creating the registry and before serving requests.
func (r *Registry) InjectMemory(mem *memory.Store) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, handler := range r.agents {
		if ms, ok := handler.(interface{ SetMemory(*memory.Store) }); ok {
			ms.SetMemory(mem)
		}
	}
}

// InjectOllama attaches oc to every agent handler that supports a local
// Ollama fallback. Call this after creating the registry and before serving
// requests. Handlers use it as a second-tier fallback: upstream Copilot
// first, then Ollama, then the canned template response.
func (r *Registry) InjectOllama(oc *copilot.OllamaClient) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, handler := range r.agents {
		if oi, ok := handler.(interface{ SetOllama(*copilot.OllamaClient) }); ok {
			oi.SetOllama(oc)
		}
	}
}

// InjectCrews attaches the crew registry to every agent handler that supports
// it (currently OMNISCIENT), so the meta-orchestrator can fall back to a
// configured crew. Call after creating the registry and loading crews.
func (r *Registry) InjectCrews(c *CrewRegistry) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, handler := range r.agents {
		if ci, ok := handler.(interface{ SetCrews(*CrewRegistry) }); ok {
			ci.SetCrews(c)
		}
	}
}

// InjectRetriever attaches rt to every agent handler that supports prompt
// augmentation (BaseAgent + ApexAgent). Call after creating the registry.
func (r *Registry) InjectRetriever(rt retrieval.Retriever) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, handler := range r.agents {
		if ri, ok := handler.(interface{ SetRetriever(retrieval.Retriever) }); ok {
			ri.SetRetriever(rt)
		}
	}
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
func RegistryFromManifest(manifestPath string, uc *copilot.UpstreamClient) (*Registry, error) {
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

		// Custom handler for APEX, dedicated orchestrator for OMNISCIENT,
		// base handler for others.
		if agentConfig.Codename == "APEX" {
			registry.Register(handlers.NewApexAgent(uc))
		} else if agentConfig.Codename == "OMNISCIENT" {
			registry.Register(NewOmniscientAgent(agent, uc, registry))
		} else {
			registry.Register(handlers.NewBaseAgent(agent, uc))
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
