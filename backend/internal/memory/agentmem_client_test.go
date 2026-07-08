package memory

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestAgentMemClient_Disabled(t *testing.T) {
	var c *AgentMemClient
	if c.Enabled() {
		t.Error("nil client should report Enabled() == false")
	}

	c2 := NewAgentMemClient("")
	if c2.Enabled() {
		t.Error("empty baseURL should report Enabled() == false")
	}
}

func TestAgentMemClient_Remember(t *testing.T) {
	var gotBody toolCallRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/call" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		json.NewEncoder(w).Encode(map[string]any{"status": "asserted", "fact_id": "abc"})
	}))
	defer srv.Close()

	c := NewAgentMemClient(srv.URL)
	if !c.Enabled() {
		t.Fatal("expected client to be enabled")
	}

	if err := c.Remember("user1", "APEX", "language", "Rust"); err != nil {
		t.Fatalf("Remember: %v", err)
	}

	if gotBody.Tool != "memory_assert_fact" {
		t.Errorf("expected tool=memory_assert_fact, got %q", gotBody.Tool)
	}
	if gotBody.Args["agent_id"] != "user1" {
		t.Errorf("expected agent_id=user1 (the userID), got %v", gotBody.Args["agent_id"])
	}
	if gotBody.Args["subject"] != agentMemFactSubject {
		t.Errorf("expected subject=%q, got %v", agentMemFactSubject, gotBody.Args["subject"])
	}
	if gotBody.Args["predicate"] != "language" {
		t.Errorf("expected predicate=language, got %v", gotBody.Args["predicate"])
	}
	if gotBody.Args["value"] != "Rust" {
		t.Errorf("expected value=Rust, got %v", gotBody.Args["value"])
	}
}

func TestAgentMemClient_Recall(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"facts": []map[string]any{
				{
					"id":         "f1",
					"similarity": 0.9,
					"content": map[string]any{
						"subject":   agentMemFactSubject,
						"predicate": "language",
						"value":     "Rust",
						"type":      "fact",
					},
				},
				{
					"id":         "f2",
					"similarity": 0.8,
					"content": map[string]any{
						"subject":   agentMemFactSubject,
						"predicate": "project",
						"value":     "vault encryption",
						"type":      "fact",
					},
				},
			},
		})
	}))
	defer srv.Close()

	c := NewAgentMemClient(srv.URL)
	facts := c.Recall("user1")

	if len(facts) != 2 {
		t.Fatalf("expected 2 facts, got %d: %+v", len(facts), facts)
	}
	if facts[0].Key != "language" || facts[0].Value != "Rust" {
		t.Errorf("unexpected first fact: %+v", facts[0])
	}
	if facts[1].Key != "project" || facts[1].Value != "vault encryption" {
		t.Errorf("unexpected second fact: %+v", facts[1])
	}
}

func TestAgentMemClient_RecallReturnsNilOnServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"error": "boom"})
	}))
	defer srv.Close()

	c := NewAgentMemClient(srv.URL)
	if facts := c.Recall("user1"); facts != nil {
		t.Errorf("expected nil facts on server error, got %+v", facts)
	}
}

func TestAgentMemClient_RememberPropagatesServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"error": "boom"})
	}))
	defer srv.Close()

	c := NewAgentMemClient(srv.URL)
	if err := c.Remember("u", "a", "k", "v"); err == nil {
		t.Error("expected an error when the server reports one")
	}
}

func TestAgentMemClient_Healthy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := NewAgentMemClient(srv.URL)
	if !c.Healthy() {
		t.Error("expected Healthy() == true")
	}

	disabled := NewAgentMemClient("")
	if disabled.Healthy() {
		t.Error("disabled client should never report healthy")
	}
}

// Store-level integration: SetAgentMem must redirect Remember/Recall, and
// leaving it unset must be identical to the pre-existing local-JSON path.
func TestStore_DelegatesToAgentMemWhenConfigured(t *testing.T) {
	var recorded []toolCallRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body toolCallRequest
		_ = json.NewDecoder(r.Body).Decode(&body)
		recorded = append(recorded, body)

		switch body.Tool {
		case "memory_assert_fact":
			json.NewEncoder(w).Encode(map[string]any{"status": "asserted", "fact_id": "1"})
		case "memory_recall_facts":
			json.NewEncoder(w).Encode(map[string]any{
				"facts": []map[string]any{
					{"content": map[string]any{"predicate": "language", "value": "Rust"}},
				},
			})
		}
	}))
	defer srv.Close()

	dir := t.TempDir()
	s, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	s.SetAgentMem(NewAgentMemClient(srv.URL))

	if err := s.Remember("user1", "APEX", "language", "Rust"); err != nil {
		t.Fatalf("Remember: %v", err)
	}
	facts := s.Recall("user1")

	if len(recorded) != 2 {
		t.Fatalf("expected 2 agentmem calls (record + recall), got %d", len(recorded))
	}
	if len(facts) != 1 || facts[0].Key != "language" {
		t.Errorf("expected facts to come from agentmem, got %+v", facts)
	}

	// Nothing should have been written to the local JSON file.
	if _, err := os.Stat(filepath.Join(dir, "user1.json")); err == nil {
		t.Error("expected no local JSON file to be written when agentmem is configured")
	}
}

func TestStore_LocalJSONUnaffectedWhenAgentMemNotConfigured(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewStore(dir)
	// SetAgentMem deliberately not called - this must behave exactly like
	// the original store_test.go suite already verifies.
	_ = s.Remember("user1", "APEX", "language", "Rust")
	facts := s.Recall("user1")
	if len(facts) != 1 || facts[0].Value != "Rust" {
		t.Errorf("expected local JSON path to still work: %+v", facts)
	}
}
