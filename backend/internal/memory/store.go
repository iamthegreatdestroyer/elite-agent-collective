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
	// Vector is the nomic embedding of "key: value", cached once at Remember
	// time for semantic recall. Absent for pre-feature or fail-open records.
	Vector    []float32 `json:"vector,omitempty"`
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
	embedder *EmbedClient
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

// SetEmbedder attaches an embedding client so RecallRelevant can rank facts by
// semantic similarity. A nil or disabled embedder preserves recency behavior.
func (s *Store) SetEmbedder(c *EmbedClient) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.embedder = c
}

// Remember stores a fact for a user. Keeps the most recent 100 facts when
// using the local JSON backend (agentmem has no such cap).
func (s *Store) Remember(userID, agentID, key, value string) error {
	if userID == "" || key == "" || value == "" {
		return nil
	}

	s.mu.Lock()
	agentMem := s.agentMem
	embedder := s.embedder
	s.mu.Unlock()
	if agentMem.Enabled() {
		return agentMem.Remember(userID, agentID, key, value)
	}

	// Embed the fact once, outside the write lock, so a slow gateway can't
	// serialize concurrent Remember calls. Fail-open: on any error the fact is
	// stored without a vector and falls back to recency at recall time.
	var vec []float32
	if embedder.Enabled() {
		if v, err := embedder.Embed(key + ": " + value); err == nil {
			vec = v
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	mem := s.load(userID)
	mem.Facts = append(mem.Facts, Fact{
		Key:       key,
		Value:     value,
		AgentID:   agentID,
		Timestamp: time.Now().UnixNano(),
		Vector:    vec,
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

// relevanceThreshold is the minimum cosine similarity for a stored fact to
// be treated as semantically relevant to the current query. Facts scoring
// below it are only included as recency backfill. Tunable; 0.55 suits short
// "key: value" facts embedded with nomic-embed-text.
const relevanceThreshold = 0.55

// RecallRelevant returns up to k facts for a user, ranked by semantic
// similarity to query when an embedder is configured, and by newest-first
// recency otherwise. It never returns fewer facts than the recency path:
// facts scoring above relevanceThreshold come first, then the newest
// remaining facts backfill to k. Fully fail-open — a disabled/unreachable
// embedder, an empty query, or vectorless (pre-feature) facts all degrade to
// recency.
func (s *Store) RecallRelevant(userID, query string, k int) []Fact {
	s.mu.Lock()
	agentMem := s.agentMem
	embedder := s.embedder
	s.mu.Unlock()
	if agentMem.Enabled() {
		return truncateFacts(agentMem.Recall(userID), k)
	}

	recency := s.Recall(userID)
	if len(recency) == 0 {
		return nil
	}
	if !embedder.Enabled() || query == "" {
		return truncateFacts(recency, k)
	}
	qvec, err := embedder.Embed(query)
	if err != nil || len(qvec) == 0 {
		return truncateFacts(recency, k)
	}

	type scored struct {
		idx   int
		score float64
	}
	var ranked []scored
	for i, f := range recency {
		if len(f.Vector) == len(qvec) && len(f.Vector) > 0 {
			ranked = append(ranked, scored{i, cosineSimilarityFloat32(qvec, f.Vector)})
		}
	}
	sort.SliceStable(ranked, func(i, j int) bool { return ranked[i].score > ranked[j].score })

	included := make([]bool, len(recency))
	out := make([]Fact, 0, len(recency))
	for _, r := range ranked {
		if r.score < relevanceThreshold {
			break
		}
		out = append(out, recency[r.idx])
		included[r.idx] = true
		if k > 0 && len(out) >= k {
			return out
		}
	}
	// Backfill with the newest facts not already included.
	for i, f := range recency {
		if k > 0 && len(out) >= k {
			break
		}
		if included[i] {
			continue
		}
		out = append(out, f)
		included[i] = true
	}
	return out
}

// FormatContextRelevant is FormatContext but query-relevant: it injects the
// facts most semantically related to query (with recency fallback) instead
// of merely the newest.
func (s *Store) FormatContextRelevant(userID, query string, maxFacts int) string {
	facts := s.RecallRelevant(userID, query, maxFacts)
	if len(facts) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("Previously established context:\n")
	for _, f := range facts {
		sb.WriteString(fmt.Sprintf("- %s: %s\n", f.Key, f.Value))
	}
	return sb.String()
}

// truncateFacts caps facts to at most k entries (k<=0 means no cap).
func truncateFacts(facts []Fact, k int) []Fact {
	if k > 0 && len(facts) > k {
		return facts[:k]
	}
	return facts
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
