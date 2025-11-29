package agents

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()
	if registry == nil {
		t.Fatal("expected non-nil registry")
	}
	if registry.Count() != 0 {
		t.Errorf("expected 0 agents, got %d", registry.Count())
	}
}

func TestDefaultRegistry(t *testing.T) {
	registry := DefaultRegistry()
	if registry == nil {
		t.Fatal("expected non-nil registry")
	}
	// Should have all 40 agents
	if registry.Count() != 40 {
		t.Errorf("expected 40 agents, got %d", registry.Count())
	}
}

func TestRegistryGet(t *testing.T) {
	registry := DefaultRegistry()

	// Test getting existing agent
	agent, err := registry.Get("APEX")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if agent == nil {
		t.Fatal("expected non-nil agent")
	}
	info := agent.GetInfo()
	if info.Codename != "APEX" {
		t.Errorf("expected codename 'APEX', got %s", info.Codename)
	}

	// Test getting non-existing agent
	_, err = registry.Get("NONEXISTENT")
	if err == nil {
		t.Error("expected error for non-existing agent")
	}
}

func TestRegistryList(t *testing.T) {
	registry := DefaultRegistry()
	agents := registry.List()

	if len(agents) != 40 {
		t.Errorf("expected 40 agents, got %d", len(agents))
	}

	// Check that all tiers are represented
	tierCounts := make(map[int]int)
	for _, agent := range agents {
		tierCounts[agent.Tier]++
	}

	expectedTierCounts := map[int]int{
		1: 5,  // Foundational
		2: 12, // Specialists
		3: 2,  // Innovators
		4: 1,  // Meta
		5: 5,  // Domain Specialists
		6: 5,  // Emerging Tech
		7: 5,  // Human-Centric
		8: 5,  // Enterprise
	}

	for tier, expected := range expectedTierCounts {
		if tierCounts[tier] != expected {
			t.Errorf("expected %d agents in tier %d, got %d", expected, tier, tierCounts[tier])
		}
	}
}

func TestAllAgentsHaveRequiredFields(t *testing.T) {
	registry := DefaultRegistry()
	agents := registry.List()

	for _, agent := range agents {
		if agent.ID == "" {
			t.Errorf("agent %s has empty ID", agent.Codename)
		}
		if agent.Codename == "" {
			t.Error("found agent with empty codename")
		}
		if agent.Tier < 1 || agent.Tier > 8 {
			t.Errorf("agent %s has invalid tier: %d", agent.Codename, agent.Tier)
		}
		if agent.Specialty == "" {
			t.Errorf("agent %s has empty specialty", agent.Codename)
		}
		if agent.Philosophy == "" {
			t.Errorf("agent %s has empty philosophy", agent.Codename)
		}
		if len(agent.Directives) == 0 {
			t.Errorf("agent %s has no directives", agent.Codename)
		}
	}
}

func TestLoadManifest(t *testing.T) {
	// Find the manifest file relative to the repo root
	manifestPath := findManifestPath(t)
	if manifestPath == "" {
		t.Skip("manifest file not found, skipping test")
	}

	manifest, err := LoadManifest(manifestPath)
	if err != nil {
		t.Fatalf("failed to load manifest: %v", err)
	}

	if manifest.Version == "" {
		t.Error("manifest version is empty")
	}
	if manifest.Name == "" {
		t.Error("manifest name is empty")
	}
	if len(manifest.Agents) != 40 {
		t.Errorf("expected 40 agents, got %d", len(manifest.Agents))
	}
}

func TestValidateManifest(t *testing.T) {
	manifestPath := findManifestPath(t)
	if manifestPath == "" {
		t.Skip("manifest file not found, skipping test")
	}

	manifest, err := LoadManifest(manifestPath)
	if err != nil {
		t.Fatalf("failed to load manifest: %v", err)
	}

	if err := ValidateManifest(manifest); err != nil {
		t.Errorf("manifest validation failed: %v", err)
	}
}

func TestRegistryFromManifest(t *testing.T) {
	manifestPath := findManifestPath(t)
	if manifestPath == "" {
		t.Skip("manifest file not found, skipping test")
	}

	registry, err := RegistryFromManifest(manifestPath)
	if err != nil {
		t.Fatalf("failed to create registry from manifest: %v", err)
	}

	if registry.Count() != 40 {
		t.Errorf("expected 40 agents, got %d", registry.Count())
	}

	// Verify APEX is available
	apex, err := registry.Get("APEX")
	if err != nil {
		t.Errorf("failed to get APEX: %v", err)
	}
	if apex == nil {
		t.Error("APEX handler is nil")
	}
}

// findManifestPath looks for the agents-manifest.yaml file.
func findManifestPath(t *testing.T) string {
	t.Helper()

	// Try common locations
	locations := []string{
		"../../../config/agents-manifest.yaml",
		"../../../../config/agents-manifest.yaml",
	}

	cwd, err := os.Getwd()
	if err == nil {
		// Also try from current directory upwards
		dir := cwd
		for i := 0; i < 5; i++ {
			candidate := filepath.Join(dir, "config", "agents-manifest.yaml")
			if _, err := os.Stat(candidate); err == nil {
				return candidate
			}
			dir = filepath.Dir(dir)
		}
	}

	for _, loc := range locations {
		if _, err := os.Stat(loc); err == nil {
			return loc
		}
	}

	return ""
}
