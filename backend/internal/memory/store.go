// Package memory provides the MNEMONIC system for the Elite Agent Collective.
// This file implements a simple persistent memory store (Mem0 pattern):
// facts are extracted from conversations and stored per-user as JSON on disk.
package memory

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// Fact represents a single remembered key-value pair for a user.
type Fact struct {
	Key       string `json:"key"`
	Value     string `json:"value"`
	AgentID   string `json:"agent_id"`
	Timestamp int64  `json:"timestamp"`
}

// userMemory holds all facts for one user.
type userMemory struct {
	Facts []Fact `json:"facts"`
}

// Store is a persistent memory store backed by per-user JSON files, or by
// a real agentmem semantic-memory backend when one is configured via
// SetAgentMem.
// Thread-safe for concurrent agent access.
type Store struct {
	dataDir  string
	mu       sync.Mutex
	cache    map[string]*userMemory
	agentMem *AgentMemClient
}

// NewStore creates a Store backed by dataDir.
// Creates dataDir if it does not exist.
func NewStore(dataDir string) (*Store, error) {
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return nil, fmt.Errorf("memory store: create dir %s: %w", dataDir, err)
	}
	return &Store{
		dataDir: dataDir,
		cache:   make(map[string]*userMemory),
	}, nil
}

// SetAgentMem attaches a real agentmem backend. When client.Enabled() is
// true, Remember/Recall delegate to it instead of the local JSON files.
// Passing a disabled (empty baseURL) client - or never calling this at all
// - preserves the exact original local-JSON-file behavior.
func (s *Store) SetAgentMem(client *AgentMemClient) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.agentMem = client
}

// Remember stores a fact for a user. Keeps the most recent 100 facts when
// using the local JSON backend (agentmem has no such cap).
func (s *Store) Remember(userID, agentID, key, value string) error {
	if userID == "" || key == "" || value == "" {
		return nil
	}

	s.mu.Lock()
	agentMem := s.agentMem
	s.mu.Unlock()
	if agentMem.Enabled() {
		return agentMem.Remember(userID, agentID, key, value)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	mem := s.load(userID)
	mem.Facts = append(mem.Facts, Fact{
		Key:       key,
		Value:     value,
		AgentID:   agentID,
		Timestamp: time.Now().UnixNano(),
	})
	if len(mem.Facts) > 100 {
		mem.Facts = mem.Facts[len(mem.Facts)-100:]
	}
	return s.save(userID, mem)
}

// Recall returns all facts for a user. Local-JSON results are sorted
// newest-first; agentmem results come back in the order agentmem's
// semantic search returns them.
func (s *Store) Recall(userID string) []Fact {
	s.mu.Lock()
	agentMem := s.agentMem
	s.mu.Unlock()
	if agentMem.Enabled() {
		return agentMem.Recall(userID)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	mem := s.load(userID)
	if len(mem.Facts) == 0 {
		return nil
	}
	result := make([]Fact, len(mem.Facts))
	copy(result, mem.Facts)
	sort.Slice(result, func(i, j int) bool { return result[i].Timestamp > result[j].Timestamp })
	return result
}

// FormatContext returns a formatted block of facts for injection into system prompts.
// Returns empty string if no facts exist.
func (s *Store) FormatContext(userID string, maxFacts int) string {
	facts := s.Recall(userID)
	if len(facts) == 0 {
		return ""
	}
	if maxFacts > 0 && len(facts) > maxFacts {
		facts = facts[:maxFacts]
	}
	var sb strings.Builder
	sb.WriteString("Previously established context:\n")
	for _, f := range facts {
		sb.WriteString(fmt.Sprintf("- %s: %s\n", f.Key, f.Value))
	}
	return sb.String()
}

// ExtractFacts scans a user message for project/technology declarations
// and returns key-value facts worth remembering.
// Pure string matching — no LLM call required.
func ExtractFacts(agentID, message string) []Fact {
	var facts []Fact
	lower := strings.ToLower(message)

	// Detect language/technology declarations.
	langPhrases := []struct{ pattern, key string }{
		{"i'm using rust", "language"},
		{"i'm using go", "language"},
		{"i'm using python", "language"},
		{"i'm using typescript", "language"},
		{"i'm using javascript", "language"},
		{"i'm using java", "language"},
		{"i'm using c++", "language"},
		{"i'm using kotlin", "language"},
		{"i'm using swift", "language"},
		{"we use rust", "language"},
		{"we use go", "language"},
		{"we use python", "language"},
		{"we use typescript", "language"},
	}
	for _, lp := range langPhrases {
		if strings.Contains(lower, lp.pattern) {
			idx := strings.Index(lower, lp.pattern) + len(lp.pattern)
			suffix := strings.TrimSpace(strings.SplitN(message[idx:], "\n", 2)[0])
			lang := strings.TrimRight(lp.pattern[strings.LastIndex(lp.pattern, " ")+1:], " ")
			if suffix != "" {
				lang = lang + " " + suffix
			}
			facts = append(facts, fact(agentID, lp.key, cleanFact(lang)))
			break
		}
	}

	// Detect project type / context.
	projectPhrases := []string{
		"i'm working on ", "we're working on ", "i am working on ",
		"i'm building ", "we're building ", "i am building ",
		"i'm developing ", "we're developing ", "i am developing ",
	}
	for _, pp := range projectPhrases {
		if idx := strings.Index(lower, pp); idx >= 0 {
			rest := message[idx+len(pp):]
			end := strings.IndexAny(rest, ".\n,;")
			if end > 0 {
				rest = rest[:end]
			}
			v := cleanFact(rest)
			if v != "" {
				facts = append(facts, fact(agentID, "project", v))
			}
			break
		}
	}

	// Detect framework / stack declarations.
	frameworkPhrases := []string{
		"using react", "using vue", "using angular", "using next.js",
		"using fastapi", "using django", "using flask", "using actix",
		"using axum", "using gin", "using echo", "using chi",
		"using kubernetes", "using docker",
	}
	for _, fp := range frameworkPhrases {
		if strings.Contains(lower, fp) {
			facts = append(facts, fact(agentID, "framework", fp[len("using "):]))
			break
		}
	}

	return facts
}

// UserID derives a stable, privacy-safe identifier from a bearer token.
// Uses first 16 hex chars of SHA-256.
func UserID(token string) string {
	if token == "" {
		return "anonymous"
	}
	h := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", h[:8])
}

// load reads or initialises a user's in-memory record. Caller must hold s.mu.
func (s *Store) load(userID string) *userMemory {
	if mem, ok := s.cache[userID]; ok {
		return mem
	}
	mem := s.readDisk(userID)
	s.cache[userID] = mem
	return mem
}

func (s *Store) readDisk(userID string) *userMemory {
	data, err := os.ReadFile(filepath.Join(s.dataDir, safe(userID)+".json"))
	if err != nil {
		return &userMemory{}
	}
	var m userMemory
	if err := json.Unmarshal(data, &m); err != nil {
		return &userMemory{}
	}
	return &m
}

func (s *Store) save(userID string, mem *userMemory) error {
	data, err := json.Marshal(mem)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(s.dataDir, safe(userID)+".json"), data, 0600)
}

// safe converts a user ID to a filename-safe string.
func safe(id string) string {
	var sb strings.Builder
	for _, r := range id {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			sb.WriteRune(r)
		}
	}
	if sb.Len() == 0 {
		return "anon"
	}
	return sb.String()
}

func fact(agentID, key, value string) Fact {
	return Fact{Key: key, Value: value, AgentID: agentID, Timestamp: time.Now().UnixNano()}
}

// cleanFact trims and truncates a fact value to a safe max length.
func cleanFact(v string) string {
	v = strings.TrimSpace(v)
	if len(v) > 120 {
		v = v[:120]
	}
	return v
}
