// Package memory provides the MNEMONIC (Multi-Agent Neural Experience Memory with
// Optimized Sub-Linear Inference for Collectives) system for the Elite Agent Collective.
//
// This package implements experience-based memory that enables agents to:
// - Accumulate and evolve strategies across task streams without retraining
// - Share cross-agent experiences through a collective memory pool
// - Achieve sub-linear retrieval using advanced indexing techniques
// - Self-improve at inference time using the ReMem-style Think-Act-Refine loop
package memory

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"sync"
	"time"
)

// ExperienceTuple represents a single experience entry in the MNEMONIC system.
// This is the core data structure inspired by Evo-Memory that captures
// what an agent learned from handling a specific task.
type ExperienceTuple struct {
	// ID is a unique identifier for this experience
	ID string `json:"id"`

	// AgentID identifies which agent created this experience (e.g., "APEX", "CIPHER")
	AgentID string `json:"agent_id"`

	// TierID identifies the agent's tier (1-8)
	TierID int `json:"tier_id"`

	// TaskSignature is a hash of the task type for exact matching
	TaskSignature string `json:"task_signature"`

	// TaskType categorizes the task (e.g., "code_generation", "security_audit")
	TaskType string `json:"task_type"`

	// Input is the original user request/query
	Input string `json:"input"`

	// Output is the agent's response
	Output string `json:"output"`

	// Strategy describes the approach taken to solve the task
	Strategy string `json:"strategy"`

	// Success indicates whether the task was completed successfully
	Success bool `json:"success"`

	// Embedding is the vector representation for semantic search
	Embedding []float32 `json:"embedding"`

	// Metadata contains additional context-specific information
	Metadata map[string]interface{} `json:"metadata"`

	// Timestamp is when this experience was created
	Timestamp int64 `json:"timestamp"`

	// EvolutionGen tracks which generation of evolution this experience belongs to
	EvolutionGen int `json:"evolution_generation"`

	// FitnessScore measures how useful this experience has been
	FitnessScore float64 `json:"fitness_score"`

	// UsageCount tracks how many times this experience was retrieved
	UsageCount int64 `json:"usage_count"`

	// LastAccessTime is when this experience was last retrieved
	LastAccessTime int64 `json:"last_access_time"`
}

// NewExperienceTuple creates a new experience tuple with default values.
func NewExperienceTuple(agentID string, tierID int, input, output, strategy string) *ExperienceTuple {
	now := time.Now().UnixNano()
	return &ExperienceTuple{
		ID:             generateExperienceID(agentID, input, now),
		AgentID:        agentID,
		TierID:         tierID,
		TaskSignature:  computeTaskSignature(input),
		Input:          input,
		Output:         output,
		Strategy:       strategy,
		Success:        true, // Default to true, can be updated after evaluation
		Embedding:      nil,  // To be computed by embedding service
		Metadata:       make(map[string]interface{}),
		Timestamp:      now,
		EvolutionGen:   0,
		FitnessScore:   0.5, // Neutral starting fitness
		UsageCount:     0,
		LastAccessTime: now,
	}
}

// generateExperienceID creates a unique ID for an experience.
func generateExperienceID(agentID, input string, timestamp int64) string {
	h := sha256.New()
	h.Write([]byte(agentID))
	h.Write([]byte(input))
	ts := make([]byte, 8)
	binary.BigEndian.PutUint64(ts, uint64(timestamp))
	h.Write(ts)
	return hex.EncodeToString(h.Sum(nil))[:32]
}

// computeTaskSignature creates a hash signature for a task input.
func computeTaskSignature(input string) string {
	h := sha256.New()
	h.Write([]byte(input))
	return hex.EncodeToString(h.Sum(nil))[:16]
}

// CompressedExperience is a reduced-size representation for efficient storage and transfer.
type CompressedExperience struct {
	// ID is the original experience ID
	ID string `json:"id"`

	// StrategyHash is a compact representation of the strategy
	StrategyHash uint64 `json:"strategy_hash"`

	// PatternVector is a dimensionality-reduced embedding
	PatternVector []float32 `json:"pattern_vector"`

	// SuccessRate aggregates success across similar experiences
	SuccessRate float32 `json:"success_rate"`

	// ApplicabilityScore indicates how broadly applicable this experience is
	ApplicabilityScore float32 `json:"applicability_score"`

	// SourceAgentID tracks the original agent
	SourceAgentID string `json:"source_agent_id"`

	// CompressedAt is when this compression was created
	CompressedAt int64 `json:"compressed_at"`
}

// QueryContext represents the context for a memory retrieval query.
type QueryContext struct {
	// AgentID is the agent making the query
	AgentID string `json:"agent_id"`

	// TierID is the tier of the querying agent
	TierID int `json:"tier_id"`

	// TaskSignature is the hash of the current task for exact matching
	TaskSignature string `json:"task_signature"`

	// TaskType categorizes the current task
	TaskType string `json:"task_type"`

	// Embedding is the vector representation of the current query
	Embedding []float32 `json:"embedding"`

	// TopK is the number of experiences to retrieve
	TopK int `json:"top_k"`

	// IncludeTierExperiences includes experiences from same-tier agents
	IncludeTierExperiences bool `json:"include_tier_experiences"`

	// IncludeCollectiveExperiences includes cross-tier breakthrough experiences
	IncludeCollectiveExperiences bool `json:"include_collective_experiences"`

	// MinFitnessScore filters experiences below this fitness threshold
	MinFitnessScore float64 `json:"min_fitness_score"`

	// MaxAge limits experiences to those created within this duration (nanoseconds)
	MaxAge int64 `json:"max_age"`
}

// NewQueryContext creates a new query context with sensible defaults.
func NewQueryContext(agentID string, tierID int, input string) *QueryContext {
	return &QueryContext{
		AgentID:                      agentID,
		TierID:                       tierID,
		TaskSignature:                computeTaskSignature(input),
		Embedding:                    nil, // To be computed by embedding service
		TopK:                         10,
		IncludeTierExperiences:       true,
		IncludeCollectiveExperiences: true,
		MinFitnessScore:              0.3,
		MaxAge:                       0, // 0 means no age limit
	}
}

// RetrievalResult wraps retrieved experiences with metadata about the retrieval.
type RetrievalResult struct {
	// Experiences contains the retrieved experience tuples
	Experiences []*ExperienceTuple `json:"experiences"`

	// RetrievalMethod indicates how the experiences were found
	// Possible values: "exact", "lsh", "hnsw", "fallback"
	RetrievalMethod string `json:"retrieval_method"`

	// RetrievalLatencyNs is the time taken for retrieval in nanoseconds
	RetrievalLatencyNs int64 `json:"retrieval_latency_ns"`

	// TotalCandidates is the number of candidates considered
	TotalCandidates int `json:"total_candidates"`

	// FilteredCount is how many candidates were filtered out
	FilteredCount int `json:"filtered_count"`
}

// ExecutionTrace captures the details of an agent execution for reflection.
type ExecutionTrace struct {
	// AgentID is the agent that executed
	AgentID string `json:"agent_id"`

	// Strategy describes the approach taken
	Strategy string `json:"strategy"`

	// RetrievedExperiences are the experiences used in context
	RetrievedExperiences []string `json:"retrieved_experiences"` // Experience IDs

	// StartTime is when execution began
	StartTime int64 `json:"start_time"`

	// EndTime is when execution completed
	EndTime int64 `json:"end_time"`

	// TokensUsed counts the tokens consumed
	TokensUsed int `json:"tokens_used"`

	// IntermediateSteps captures any multi-step reasoning
	IntermediateSteps []string `json:"intermediate_steps"`
}

// Evaluation represents the outcome assessment of an execution.
type Evaluation struct {
	// Success indicates whether the task was completed successfully
	Success bool `json:"success"`

	// Score is a numerical quality assessment (0.0 to 1.0)
	Score float64 `json:"score"`

	// Feedback contains any user or automated feedback
	Feedback string `json:"feedback"`

	// Metrics contains task-specific evaluation metrics
	Metrics map[string]float64 `json:"metrics"`

	// EvaluatedAt is when this evaluation was performed
	EvaluatedAt int64 `json:"evaluated_at"`
}

// Breakthrough represents an exceptional experience worth sharing across tiers.
type Breakthrough struct {
	// ID is a unique identifier for this breakthrough
	ID string `json:"id"`

	// OriginExperienceID links to the source experience
	OriginExperienceID string `json:"origin_experience_id"`

	// OriginAgent is the agent that discovered this
	OriginAgent string `json:"origin_agent"`

	// OriginTier is the tier of the origin agent
	OriginTier int `json:"origin_tier"`

	// Strategy describes the breakthrough approach
	Strategy string `json:"strategy"`

	// Embedding is the vector representation
	Embedding []float32 `json:"embedding"`

	// ApplicableTiers lists which tiers can benefit from this
	ApplicableTiers []int `json:"applicable_tiers"`

	// DiscoveredAt is when this breakthrough was identified
	DiscoveredAt int64 `json:"discovered_at"`

	// PropagationCount tracks how many agents received this
	PropagationCount int `json:"propagation_count"`

	// SuccessfulAdoptions counts successful uses by other agents
	SuccessfulAdoptions int `json:"successful_adoptions"`
}

// MemoryStats tracks statistics about the memory system.
type MemoryStats struct {
	mu sync.RWMutex

	// TotalExperiences is the count of all stored experiences
	TotalExperiences int64 `json:"total_experiences"`

	// ExperiencesByAgent maps agent ID to experience count
	ExperiencesByAgent map[string]int64 `json:"experiences_by_agent"`

	// ExperiencesByTier maps tier ID to experience count
	ExperiencesByTier map[int]int64 `json:"experiences_by_tier"`

	// TotalRetrievals counts all retrieval operations
	TotalRetrievals int64 `json:"total_retrievals"`

	// AvgRetrievalLatencyNs tracks average retrieval time
	AvgRetrievalLatencyNs float64 `json:"avg_retrieval_latency_ns"`

	// CacheHitRate tracks how often exact matches are found
	CacheHitRate float64 `json:"cache_hit_rate"`

	// BreakthroughCount tracks total breakthroughs discovered
	BreakthroughCount int64 `json:"breakthrough_count"`

	// EvolutionGeneration is the current global evolution generation
	EvolutionGeneration int `json:"evolution_generation"`

	// LastEvolutionTime is when the last evolution cycle ran
	LastEvolutionTime int64 `json:"last_evolution_time"`
}

// NewMemoryStats creates a new stats tracker with initialized maps.
func NewMemoryStats() *MemoryStats {
	return &MemoryStats{
		ExperiencesByAgent: make(map[string]int64),
		ExperiencesByTier:  make(map[int]int64),
	}
}

// IncrementExperiences safely increments experience counts.
func (s *MemoryStats) IncrementExperiences(agentID string, tierID int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.TotalExperiences++
	s.ExperiencesByAgent[agentID]++
	s.ExperiencesByTier[tierID]++
}

// UpdateRetrievalStats safely updates retrieval statistics.
func (s *MemoryStats) UpdateRetrievalStats(latencyNs int64, cacheHit bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.TotalRetrievals++

	// Rolling average for latency
	if s.TotalRetrievals == 1 {
		s.AvgRetrievalLatencyNs = float64(latencyNs)
	} else {
		s.AvgRetrievalLatencyNs = (s.AvgRetrievalLatencyNs*float64(s.TotalRetrievals-1) + float64(latencyNs)) / float64(s.TotalRetrievals)
	}

	// Rolling average for cache hit rate
	hitVal := 0.0
	if cacheHit {
		hitVal = 1.0
	}
	if s.TotalRetrievals == 1 {
		s.CacheHitRate = hitVal
	} else {
		s.CacheHitRate = (s.CacheHitRate*float64(s.TotalRetrievals-1) + hitVal) / float64(s.TotalRetrievals)
	}
}

// GetStats returns a pointer to a copy of the current statistics.
// Returns a pointer to avoid copying the embedded sync.RWMutex.
func (s *MemoryStats) GetStats() *MemoryStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Create copies of the maps
	agentCopy := make(map[string]int64)
	for k, v := range s.ExperiencesByAgent {
		agentCopy[k] = v
	}
	tierCopy := make(map[int]int64)
	for k, v := range s.ExperiencesByTier {
		tierCopy[k] = v
	}

	return &MemoryStats{
		TotalExperiences:      s.TotalExperiences,
		ExperiencesByAgent:    agentCopy,
		ExperiencesByTier:     tierCopy,
		TotalRetrievals:       s.TotalRetrievals,
		AvgRetrievalLatencyNs: s.AvgRetrievalLatencyNs,
		CacheHitRate:          s.CacheHitRate,
		BreakthroughCount:     s.BreakthroughCount,
		EvolutionGeneration:   s.EvolutionGeneration,
		LastEvolutionTime:     s.LastEvolutionTime,
	}
}
