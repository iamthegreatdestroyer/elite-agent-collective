// Package models contains data models for the Elite Agent Collective backend.
// This file defines shared data structures for the MNEMONIC memory system.

package models

// MemoryConfig contains configuration for the MNEMONIC memory system.
type MemoryConfig struct {
	// Enabled determines if memory system is active
	Enabled bool `json:"enabled" yaml:"enabled"`

	// EmbeddingDimension is the size of embedding vectors
	EmbeddingDimension int `json:"embedding_dimension" yaml:"embedding_dimension"`

	// MaxExperiencesPerAgent limits experiences stored per agent
	MaxExperiencesPerAgent int `json:"max_experiences_per_agent" yaml:"max_experiences_per_agent"`

	// MinFitnessThreshold for retrieval filtering
	MinFitnessThreshold float64 `json:"min_fitness_threshold" yaml:"min_fitness_threshold"`

	// BreakthroughThreshold for promoting to collective
	BreakthroughThreshold float64 `json:"breakthrough_threshold" yaml:"breakthrough_threshold"`

	// EvolutionIntervalSeconds between evolution cycles
	EvolutionIntervalSeconds int `json:"evolution_interval_seconds" yaml:"evolution_interval_seconds"`

	// PersistencePath is the directory for memory persistence
	PersistencePath string `json:"persistence_path" yaml:"persistence_path"`

	// LSH configuration
	LSHNumTables    int `json:"lsh_num_tables" yaml:"lsh_num_tables"`
	LSHNumHashFuncs int `json:"lsh_num_hash_funcs" yaml:"lsh_num_hash_funcs"`

	// HNSW configuration
	HNSWMaxConnections int `json:"hnsw_max_connections" yaml:"hnsw_max_connections"`
	HNSWEfConstruction int `json:"hnsw_ef_construction" yaml:"hnsw_ef_construction"`
	HNSWEfSearch       int `json:"hnsw_ef_search" yaml:"hnsw_ef_search"`
}

// DefaultMemoryConfig returns a configuration with sensible defaults.
func DefaultMemoryConfig() *MemoryConfig {
	return &MemoryConfig{
		Enabled:                  true,
		EmbeddingDimension:       384,
		MaxExperiencesPerAgent:   1000,
		MinFitnessThreshold:      0.3,
		BreakthroughThreshold:    0.9,
		EvolutionIntervalSeconds: 3600, // 1 hour
		PersistencePath:          "./data/memory",
		LSHNumTables:             10,
		LSHNumHashFuncs:          12,
		HNSWMaxConnections:       16,
		HNSWEfConstruction:       200,
		HNSWEfSearch:             100,
	}
}

// MemoryRequest represents a request to the memory system.
type MemoryRequest struct {
	// AgentID is the agent making the request
	AgentID string `json:"agent_id"`

	// Query is the text query for retrieval
	Query string `json:"query"`

	// TopK is the number of experiences to retrieve
	TopK int `json:"top_k"`

	// IncludeTierExperiences includes same-tier experiences
	IncludeTierExperiences bool `json:"include_tier_experiences"`

	// IncludeCollective includes cross-tier breakthroughs
	IncludeCollective bool `json:"include_collective"`
}

// MemoryResponse represents a response from the memory system.
type MemoryResponse struct {
	// Experiences are the retrieved experiences
	Experiences []ExperienceSummary `json:"experiences"`

	// RetrievalMethod indicates how experiences were found
	RetrievalMethod string `json:"retrieval_method"`

	// RetrievalLatencyMs is the time taken in milliseconds
	RetrievalLatencyMs float64 `json:"retrieval_latency_ms"`

	// MemoryPrompt is the formatted context injection
	MemoryPrompt string `json:"memory_prompt,omitempty"`
}

// ExperienceSummary is a condensed view of an experience for API responses.
type ExperienceSummary struct {
	// ID is the experience identifier
	ID string `json:"id"`

	// AgentID is the source agent
	AgentID string `json:"agent_id"`

	// TaskType categorizes the task
	TaskType string `json:"task_type"`

	// Strategy describes the approach taken
	Strategy string `json:"strategy"`

	// Success indicates if the task was successful
	Success bool `json:"success"`

	// FitnessScore is the quality metric
	FitnessScore float64 `json:"fitness_score"`

	// Timestamp is when this was created
	Timestamp int64 `json:"timestamp"`
}

// MemoryStatsResponse contains statistics about the memory system.
type MemoryStatsResponse struct {
	// TotalExperiences is the count of all experiences
	TotalExperiences int64 `json:"total_experiences"`

	// ExperiencesByTier maps tier to count
	ExperiencesByTier map[int]int64 `json:"experiences_by_tier"`

	// TotalBreakthroughs is the count of breakthroughs
	TotalBreakthroughs int64 `json:"total_breakthroughs"`

	// EvolutionGeneration is the current generation
	EvolutionGeneration int `json:"evolution_generation"`

	// AvgRetrievalLatencyMs is the average retrieval time
	AvgRetrievalLatencyMs float64 `json:"avg_retrieval_latency_ms"`

	// CacheHitRate is the exact match rate
	CacheHitRate float64 `json:"cache_hit_rate"`
}

// AgentTierMapping defines the tier assignments for all agents.
var AgentTierMapping = map[string]int{
	// Tier 1: Foundational
	"APEX": 1, "CIPHER": 1, "ARCHITECT": 1, "AXIOM": 1, "VELOCITY": 1,
	// Tier 2: Specialists
	"QUANTUM": 2, "TENSOR": 2, "FORTRESS": 2, "NEURAL": 2, "CRYPTO": 2,
	"FLUX": 2, "PRISM": 2, "SYNAPSE": 2, "CORE": 2, "HELIX": 2,
	"VANGUARD": 2, "ECLIPSE": 2,
	// Tier 3: Innovators
	"NEXUS": 3, "GENESIS": 3,
	// Tier 4: Meta
	"OMNISCIENT": 4,
	// Tier 5: Domain Specialists
	"ATLAS": 5, "FORGE": 5, "SENTRY": 5, "VERTEX": 5, "STREAM": 5,
	// Tier 6: Emerging Tech
	"PHOTON": 6, "LATTICE": 6, "MORPH": 6, "PHANTOM": 6, "ORBIT": 6,
	// Tier 7: Human-Centric
	"CANVAS": 7, "LINGUA": 7, "SCRIBE": 7, "MENTOR": 7, "BRIDGE": 7,
	// Tier 8: Enterprise
	"AEGIS": 8, "LEDGER": 8, "PULSE": 8, "ARBITER": 8, "ORACLE": 8,
}

// GetAgentTier returns the tier for a given agent codename.
func GetAgentTier(agentID string) int {
	if tier, ok := AgentTierMapping[agentID]; ok {
		return tier
	}
	return 1 // Default to tier 1
}
