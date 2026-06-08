package agents

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCrewRegistry_LoadAndGet(t *testing.T) {
	yaml := `
crews:
  code_review:
    description: "Test crew"
    agents: [APEX, ECLIPSE, VELOCITY]
  security_audit:
    description: "Security crew"
    agents: [CIPHER, FORTRESS]
`
	f := filepath.Join(t.TempDir(), "crews.yaml")
	if err := os.WriteFile(f, []byte(yaml), 0600); err != nil {
		t.Fatalf("write: %v", err)
	}

	r, err := NewCrewRegistry(f)
	if err != nil {
		t.Fatalf("NewCrewRegistry: %v", err)
	}

	if r.Count() != 2 {
		t.Errorf("expected 2 crews, got %d", r.Count())
	}

	crew, ok := r.Get("code_review")
	if !ok {
		t.Fatal("code_review crew not found")
	}
	if len(crew.Agents) != 3 {
		t.Errorf("expected 3 agents in code_review, got %d", len(crew.Agents))
	}

	// Case-insensitive lookup.
	_, ok = r.Get("CODE_REVIEW")
	if !ok {
		t.Error("case-insensitive lookup failed")
	}
}

func TestCrewRegistry_MissingFile(t *testing.T) {
	r, err := NewCrewRegistry("/nonexistent/crews.yaml")
	if err != nil {
		t.Fatalf("expected empty registry for missing file, got error: %v", err)
	}
	if r.Count() != 0 {
		t.Errorf("expected 0 crews, got %d", r.Count())
	}
}

func TestCrewRegistry_GetUnknown(t *testing.T) {
	r, _ := NewCrewRegistry("")
	if _, ok := r.Get("unknown"); ok {
		t.Error("unexpected crew found for unknown name")
	}
}

func TestCrewRegistry_List(t *testing.T) {
	yaml := "crews:\n  a:\n    agents: [APEX]\n  b:\n    agents: [CIPHER]\n"
	f := filepath.Join(t.TempDir(), "c.yaml")
	os.WriteFile(f, []byte(yaml), 0600)
	r, _ := NewCrewRegistry(f)
	names := r.List()
	if len(names) != 2 {
		t.Errorf("expected 2 crew names, got %d", len(names))
	}
}
