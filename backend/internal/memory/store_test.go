package memory

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStore_RememberAndRecall(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	if err := s.Remember("user1", "APEX", "language", "Rust"); err != nil {
		t.Fatalf("Remember: %v", err)
	}
	if err := s.Remember("user1", "CIPHER", "project", "vault encryption"); err != nil {
		t.Fatalf("Remember: %v", err)
	}

	facts := s.Recall("user1")
	if len(facts) != 2 {
		t.Fatalf("expected 2 facts, got %d", len(facts))
	}
	// Most recent first.
	if facts[0].Key != "project" {
		t.Errorf("expected most recent fact key=project, got %q", facts[0].Key)
	}
}

func TestStore_Persistence(t *testing.T) {
	dir := t.TempDir()

	s, _ := NewStore(dir)
	_ = s.Remember("u1", "APEX", "k", "v")

	// Reload from disk.
	s2, _ := NewStore(dir)
	facts := s2.Recall("u1")
	if len(facts) != 1 || facts[0].Key != "k" {
		t.Errorf("facts not persisted: %v", facts)
	}
}

func TestStore_FormatContext(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewStore(dir)
	_ = s.Remember("u1", "APEX", "language", "Go")
	_ = s.Remember("u1", "APEX", "project", "NAS OS")

	ctx := s.FormatContext("u1", 5)
	if ctx == "" {
		t.Fatal("expected non-empty context")
	}
	if ctx[:32] != "Previously established context:\n" {
		t.Errorf("unexpected context prefix: %q", ctx[:32])
	}
}

func TestStore_EmptyUserReturnsEmpty(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewStore(dir)
	if facts := s.Recall("nobody"); len(facts) != 0 {
		t.Errorf("expected 0 facts for unknown user, got %d", len(facts))
	}
	if ctx := s.FormatContext("nobody", 5); ctx != "" {
		t.Errorf("expected empty context for unknown user, got %q", ctx)
	}
}

func TestStore_CapAt100(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewStore(dir)
	for i := 0; i < 110; i++ {
		_ = s.Remember("u1", "APEX", "k", "v")
	}
	if len(s.Recall("u1")) != 100 {
		t.Errorf("expected 100 facts after capping")
	}
}

func TestStore_FileIsolation(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewStore(dir)
	_ = s.Remember("alice", "APEX", "lang", "Go")
	_ = s.Remember("bob", "CIPHER", "lang", "Rust")

	aliceFacts := s.Recall("alice")
	if len(aliceFacts) != 1 || aliceFacts[0].Value != "Go" {
		t.Errorf("alice facts wrong: %v", aliceFacts)
	}

	// Verify separate files on disk.
	aliceFile := filepath.Join(dir, "alice.json")
	bobFile := filepath.Join(dir, "bob.json")
	if _, err := os.Stat(aliceFile); err != nil {
		t.Errorf("alice file missing: %v", err)
	}
	if _, err := os.Stat(bobFile); err != nil {
		t.Errorf("bob file missing: %v", err)
	}
}

func TestUserID(t *testing.T) {
	id1 := UserID("token-abc")
	id2 := UserID("token-abc")
	if id1 != id2 {
		t.Errorf("UserID not stable: %q vs %q", id1, id2)
	}
	if id1 == UserID("token-xyz") {
		t.Errorf("different tokens produced same UserID")
	}
	if UserID("") != "anonymous" {
		t.Errorf("empty token should return anonymous")
	}
}

func TestExtractFacts(t *testing.T) {
	tests := []struct {
		msg      string
		wantKey  string
		wantNone bool
	}{
		{"I'm using Rust for this project", "language", false},
		{"we're building a distributed file system", "project", false},
		{"using react for the frontend", "framework", false},
		{"how do I sort a slice?", "", true},
	}
	for _, tt := range tests {
		facts := ExtractFacts("APEX", tt.msg)
		if tt.wantNone {
			if len(facts) != 0 {
				t.Errorf("msg %q: expected no facts, got %v", tt.msg, facts)
			}
			continue
		}
		found := false
		for _, f := range facts {
			if f.Key == tt.wantKey {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("msg %q: expected fact key=%q, got %v", tt.msg, tt.wantKey, facts)
		}
	}
}

// Ensure timestamps are set.
func TestStore_Timestamps(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewStore(dir)
	before := time.Now().UnixNano()
	_ = s.Remember("u1", "APEX", "k", "v")
	after := time.Now().UnixNano()
	facts := s.Recall("u1")
	if len(facts) == 0 {
		t.Fatal("no facts")
	}
	if facts[0].Timestamp < before || facts[0].Timestamp > after {
		t.Errorf("timestamp %d out of range [%d, %d]", facts[0].Timestamp, before, after)
	}
}
