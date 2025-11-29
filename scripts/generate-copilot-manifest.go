// Package main provides a manifest generator that reads agents-manifest.yaml
// and generates the copilot-extension.json file.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ManifestVersion is the version of the manifest schema.
const ManifestVersion = "2.0.0"

// AgentManifest represents the complete agents manifest YAML structure.
type AgentManifest struct {
	Version     string         `yaml:"version"`
	Name        string         `yaml:"name"`
	Description string         `yaml:"description"`
	Tiers       []TierInfo     `yaml:"tiers"`
	Agents      []AgentConfig  `yaml:"agents"`
}

// TierInfo describes a tier in the agent hierarchy.
type TierInfo struct {
	ID          int    `yaml:"id"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// AgentConfig represents a single agent's configuration.
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

// CopilotExtension represents the copilot-extension.json structure.
type CopilotExtension struct {
	Schema             string              `json:"$schema"`
	Name               string              `json:"name"`
	DisplayName        string              `json:"display_name"`
	Version            string              `json:"version"`
	Description        string              `json:"description"`
	HomepageURL        string              `json:"homepage_url"`
	PrivacyPolicyURL   string              `json:"privacy_policy_url"`
	SupportURL         string              `json:"support_url"`
	DefaultModel       string              `json:"default_model"`
	Capabilities       Capabilities        `json:"capabilities"`
	Tools              []Tool              `json:"tools"`
	AgentCollaboration AgentCollaboration  `json:"agent_collaboration"`
	Tiers              map[string]TierDef  `json:"tiers"`
}

// Capabilities represents extension capabilities.
type Capabilities struct {
	Conversation bool `json:"conversation"`
	Tools        bool `json:"tools"`
}

// Tool represents a single tool/agent in the Copilot extension.
type Tool struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Parameters  Parameters `json:"parameters"`
}

// Parameters represents tool parameters.
type Parameters struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required"`
}

// Property represents a single parameter property.
type Property struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

// AgentCollaboration represents multi-agent collaboration settings.
type AgentCollaboration struct {
	Enabled     bool   `json:"enabled"`
	Description string `json:"description"`
}

// TierDef represents a tier definition in the extension.
type TierDef struct {
	Name   string   `json:"name"`
	Agents []string `json:"agents"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Determine paths
	scriptDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Try to find the repository root
	repoRoot := findRepoRoot(scriptDir)
	if repoRoot == "" {
		return fmt.Errorf("could not find repository root (looking for config/agents-manifest.yaml)")
	}

	manifestPath := filepath.Join(repoRoot, "config", "agents-manifest.yaml")
	outputPath := filepath.Join(repoRoot, "copilot-extension.json")

	// Read and parse the YAML manifest
	manifest, err := readManifest(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to read manifest: %w", err)
	}

	// Validate the manifest
	if err := validateManifest(manifest); err != nil {
		return fmt.Errorf("manifest validation failed: %w", err)
	}

	// Generate the Copilot extension JSON
	extension := generateExtension(manifest)

	// Write the output
	if err := writeExtension(outputPath, extension); err != nil {
		return fmt.Errorf("failed to write extension: %w", err)
	}

	fmt.Printf("Successfully generated %s with %d agents\n", outputPath, len(manifest.Agents))
	return nil
}

func findRepoRoot(startDir string) string {
	dir := startDir
	for {
		// Check if config/agents-manifest.yaml exists
		manifestPath := filepath.Join(dir, "config", "agents-manifest.yaml")
		if _, err := os.Stat(manifestPath); err == nil {
			return dir
		}

		// Go up one directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root
			break
		}
		dir = parent
	}
	return ""
}

func readManifest(path string) (*AgentManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}

	var manifest AgentManifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &manifest, nil
}

func validateManifest(manifest *AgentManifest) error {
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
		if agent.Description == "" {
			return fmt.Errorf("agent %s missing description", agent.Codename)
		}
		if agent.Philosophy == "" {
			return fmt.Errorf("agent %s missing philosophy", agent.Codename)
		}
		if len(agent.Directives) == 0 {
			return fmt.Errorf("agent %s missing directives", agent.Codename)
		}
	}

	// Verify all 8 tiers are present
	tierCounts := make(map[int]int)
	for _, agent := range manifest.Agents {
		tierCounts[agent.Tier]++
	}

	expectedTiers := map[int]int{
		1: 5,  // Foundational
		2: 12, // Specialists
		3: 2,  // Innovators
		4: 1,  // Meta
		5: 5,  // Domain Specialists
		6: 5,  // Emerging Tech
		7: 5,  // Human-Centric
		8: 5,  // Enterprise
	}

	for tier, expected := range expectedTiers {
		if tierCounts[tier] != expected {
			return fmt.Errorf("tier %d: expected %d agents, found %d", tier, expected, tierCounts[tier])
		}
	}

	return nil
}

func generateExtension(manifest *AgentManifest) *CopilotExtension {
	extension := &CopilotExtension{
		Schema:           "https://json.schemastore.org/github-copilot-extension.json",
		Name:             "elite-agent-collective",
		DisplayName:      "Elite Agent Collective",
		Version:          ManifestVersion,
		Description:      manifest.Description,
		HomepageURL:      "https://github.com/iamthegreatdestroyer/elite-agent-collective",
		PrivacyPolicyURL: "https://github.com/iamthegreatdestroyer/elite-agent-collective/blob/main/PRIVACY.md",
		SupportURL:       "https://github.com/iamthegreatdestroyer/elite-agent-collective/issues",
		DefaultModel:     "gpt-4",
		Capabilities: Capabilities{
			Conversation: true,
			Tools:        true,
		},
		AgentCollaboration: AgentCollaboration{
			Enabled:     true,
			Description: "Agents can collaborate by invoking multiple agents together using @AGENT1 @AGENT2 syntax.",
		},
		Tiers: make(map[string]TierDef),
	}

	// Generate tools from agents
	extension.Tools = make([]Tool, 0, len(manifest.Agents))
	for _, agent := range manifest.Agents {
		tool := Tool{
			Name:        agent.Codename,
			Description: fmt.Sprintf("%s - %s. Philosophy: %s", agent.Name, agent.Description, agent.Philosophy),
			Parameters: Parameters{
				Type: "object",
				Properties: map[string]Property{
					"task": {
						Type:        "string",
						Description: generateTaskDescription(agent),
					},
				},
				Required: []string{"task"},
			},
		}
		extension.Tools = append(extension.Tools, tool)
	}

	// Generate tier definitions
	tierAgents := make(map[int][]string)
	for _, agent := range manifest.Agents {
		tierAgents[agent.Tier] = append(tierAgents[agent.Tier], agent.Codename)
	}

	tierNames := map[int]string{
		1: "Foundational",
		2: "Specialists",
		3: "Innovators",
		4: "Meta",
		5: "Domain Specialists",
		6: "Emerging Tech",
		7: "Human-Centric",
		8: "Enterprise",
	}

	for tierID, agents := range tierAgents {
		extension.Tiers[fmt.Sprintf("%d", tierID)] = TierDef{
			Name:   tierNames[tierID],
			Agents: agents,
		}
	}

	return extension
}

func generateTaskDescription(agent AgentConfig) string {
	// Extract action verbs from examples
	if len(agent.Examples) > 0 {
		return fmt.Sprintf("The task for %s to perform. Examples from: %s", agent.Codename, agent.Examples[0])
	}
	return fmt.Sprintf("The task for %s to perform based on their specialty: %s", agent.Codename, agent.Description)
}

func writeExtension(path string, extension *CopilotExtension) error {
	data, err := json.MarshalIndent(extension, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
