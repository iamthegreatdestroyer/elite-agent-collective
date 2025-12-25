// Package memory provides the MNEMONIC system for the Elite Agent Collective.
// This file implements Memory Consolidation from @NEURAL's Cognitive Architecture Analysis.
//
// Memory Consolidation mimics the human brain's offline processing during sleep:
// - Short-term buffer accumulates recent experiences
// - Periodic consolidation extracts patterns and schemas
// - Similar experiences are clustered and compressed
// - Exemplars are preserved while details are abstracted
// - Schemas capture reusable patterns for future use

package memory

import (
	"fmt"
	"math"
	"sort"
	"sync"
	"time"
)

// ============================================================================
// Consolidated Memory Structures
// ============================================================================

// ConsolidatedMemory represents a compressed pattern extracted from multiple experiences.
type ConsolidatedMemory struct {
	// ID uniquely identifies this consolidated memory
	ID string

	// Schema is the extracted pattern/template
	Schema *MemorySchema

	// Exemplars are the best representative experiences
	Exemplars []*ExperienceTuple

	// Frequency is how many experiences contributed to this
	Frequency int

	// AbstractionLevel indicates how general vs specific (0.0 = specific, 1.0 = abstract)
	AbstractionLevel float64

	// AgentIDs that contributed experiences
	AgentIDs []string

	// TierIDs that contributed experiences
	TierIDs []int

	// AverageFitness of contributing experiences
	AverageFitness float64

	// CreatedAt timestamp
	CreatedAt time.Time

	// LastUpdated timestamp
	LastUpdated time.Time

	// UsageCount tracks how often this schema is applied
	UsageCount int64
}

// MemorySchema represents an abstracted pattern.
type MemorySchema struct {
	// Name of the schema
	Name string

	// InputPattern describes the input characteristics
	InputPattern SchemaPattern

	// StrategyPattern describes the strategy characteristics
	StrategyPattern SchemaPattern

	// OutputPattern describes the output characteristics
	OutputPattern SchemaPattern

	// Preconditions that should hold
	Preconditions []string

	// Effects that typically result
	Effects []string

	// Confidence in this schema (0.0-1.0)
	Confidence float64
}

// SchemaPattern represents a pattern with variables and constraints.
type SchemaPattern struct {
	// Variables are placeholders in the pattern
	Variables map[string]string // name -> type

	// Constraints on variables
	Constraints []PatternConstraint

	// FeatureVector is a semantic representation
	FeatureVector []float32
}

// PatternConstraint constrains pattern variables.
type PatternConstraint struct {
	Variable string
	Type     string // "type", "range", "regex", "semantic"
	Value    interface{}
}

// ============================================================================
// Consolidator Configuration
// ============================================================================

// ConsolidatorConfig configures memory consolidation.
type ConsolidatorConfig struct {
	// BufferCapacity before triggering consolidation
	BufferCapacity int

	// MinClusterSize minimum experiences to form a cluster
	MinClusterSize int

	// SimilarityThreshold for clustering (0.0-1.0)
	SimilarityThreshold float64

	// ExemplarCount how many exemplars to keep per cluster
	ExemplarCount int

	// CompressionRatio target (0.0-1.0, higher = more compression)
	CompressionRatio float64

	// ConsolidationInterval how often to run (0 = manual only)
	ConsolidationInterval time.Duration

	// MinTimeSinceLastAccess before experiences are eligible
	MinTimeSinceLastAccess time.Duration

	// EnableAutoConsolidation whether to run automatically
	EnableAutoConsolidation bool
}

// DefaultConsolidatorConfig returns sensible defaults.
func DefaultConsolidatorConfig() *ConsolidatorConfig {
	return &ConsolidatorConfig{
		BufferCapacity:          100,
		MinClusterSize:          3,
		SimilarityThreshold:     0.75,
		ExemplarCount:           3,
		CompressionRatio:        0.7,
		ConsolidationInterval:   1 * time.Hour,
		MinTimeSinceLastAccess:  10 * time.Minute,
		EnableAutoConsolidation: false, // Manual trigger for now
	}
}

// ============================================================================
// Memory Consolidator
// ============================================================================

// MemoryConsolidator consolidates experiences into patterns and schemas.
type MemoryConsolidator struct {
	config *ConsolidatorConfig

	// Short-term buffer of recent experiences
	shortTermBuffer []*ExperienceTuple
	bufferMu        sync.RWMutex

	// Consolidated memories
	consolidated   map[string]*ConsolidatedMemory
	consolidatedMu sync.RWMutex

	// Statistics
	stats   *ConsolidationStats
	statsMu sync.RWMutex

	// Control
	stopChan chan struct{}
	doneChan chan struct{}
}

// ConsolidationStats tracks consolidation metrics.
type ConsolidationStats struct {
	TotalConsolidations   int64
	ExperiencesProcessed  int64
	ClustersFormed        int64
	SchemasExtracted      int64
	CompressionRatio      float64
	LastConsolidationTime time.Time
	AverageClusterSize    float64
}

// NewMemoryConsolidator creates a new memory consolidator.
func NewMemoryConsolidator(config *ConsolidatorConfig) *MemoryConsolidator {
	if config == nil {
		config = DefaultConsolidatorConfig()
	}

	mc := &MemoryConsolidator{
		config:          config,
		shortTermBuffer: make([]*ExperienceTuple, 0, config.BufferCapacity),
		consolidated:    make(map[string]*ConsolidatedMemory),
		stats:           &ConsolidationStats{},
		stopChan:        make(chan struct{}),
		doneChan:        make(chan struct{}),
	}

	// Start automatic consolidation if enabled
	if config.EnableAutoConsolidation && config.ConsolidationInterval > 0 {
		go mc.runAutoConsolidation()
	}

	return mc
}

// AddToBuffer adds an experience to the short-term buffer.
func (mc *MemoryConsolidator) AddToBuffer(exp *ExperienceTuple) {
	mc.bufferMu.Lock()
	defer mc.bufferMu.Unlock()

	mc.shortTermBuffer = append(mc.shortTermBuffer, exp)

	// Check if we should trigger consolidation
	if len(mc.shortTermBuffer) >= mc.config.BufferCapacity {
		// In manual mode, just track that buffer is full
		// User must call Consolidate()
		if !mc.config.EnableAutoConsolidation {
			// Could log or notify here
		}
	}
}

// Consolidate runs the consolidation process.
func (mc *MemoryConsolidator) Consolidate() (*ConsolidationResult, error) {
	mc.bufferMu.Lock()
	if len(mc.shortTermBuffer) == 0 {
		mc.bufferMu.Unlock()
		return &ConsolidationResult{}, nil
	}

	// Take snapshot of buffer and clear it
	buffer := make([]*ExperienceTuple, len(mc.shortTermBuffer))
	copy(buffer, mc.shortTermBuffer)
	mc.shortTermBuffer = mc.shortTermBuffer[:0]
	mc.bufferMu.Unlock()

	startTime := time.Now()

	// 1. Filter experiences by recency
	eligible := mc.filterEligible(buffer)
	if len(eligible) < mc.config.MinClusterSize {
		// Not enough experiences, return them to buffer
		mc.bufferMu.Lock()
		mc.shortTermBuffer = append(mc.shortTermBuffer, eligible...)
		mc.bufferMu.Unlock()
		return &ConsolidationResult{}, nil
	}

	// 2. Cluster similar experiences
	clusters := mc.clusterExperiences(eligible)

	// 3. Extract schemas from each cluster
	var newConsolidated []*ConsolidatedMemory
	for _, cluster := range clusters {
		if len(cluster) >= mc.config.MinClusterSize {
			schema := mc.extractSchema(cluster)
			exemplars := mc.selectExemplars(cluster, mc.config.ExemplarCount)
			abstraction := mc.computeAbstractionLevel(cluster)

			consolidated := &ConsolidatedMemory{
				ID:               fmt.Sprintf("consolidated-%d", time.Now().UnixNano()),
				Schema:           schema,
				Exemplars:        exemplars,
				Frequency:        len(cluster),
				AbstractionLevel: abstraction,
				AgentIDs:         mc.extractAgentIDs(cluster),
				TierIDs:          mc.extractTierIDs(cluster),
				AverageFitness:   mc.computeAverageFitness(cluster),
				CreatedAt:        time.Now(),
				LastUpdated:      time.Now(),
			}

			newConsolidated = append(newConsolidated, consolidated)
		}
	}

	// 4. Store consolidated memories
	mc.consolidatedMu.Lock()
	for _, cm := range newConsolidated {
		mc.consolidated[cm.ID] = cm
	}
	mc.consolidatedMu.Unlock()

	// 5. Update statistics
	result := &ConsolidationResult{
		ExperiencesProcessed: len(eligible),
		ClustersFormed:       len(clusters),
		SchemasExtracted:     len(newConsolidated),
		ConsolidatedMemories: newConsolidated,
		Duration:             time.Since(startTime),
		CompressionRatio:     mc.calculateCompressionRatio(eligible, newConsolidated),
	}

	mc.updateStats(result)

	return result, nil
}

// ConsolidationResult contains the results of a consolidation run.
type ConsolidationResult struct {
	ExperiencesProcessed int
	ClustersFormed       int
	SchemasExtracted     int
	ConsolidatedMemories []*ConsolidatedMemory
	Duration             time.Duration
	CompressionRatio     float64
}

// filterEligible filters experiences by access recency.
func (mc *MemoryConsolidator) filterEligible(experiences []*ExperienceTuple) []*ExperienceTuple {
	now := time.Now().UnixNano()
	cutoffNano := now - mc.config.MinTimeSinceLastAccess.Nanoseconds()

	eligible := make([]*ExperienceTuple, 0)
	for _, exp := range experiences {
		if exp.LastAccessTime < cutoffNano {
			eligible = append(eligible, exp)
		}
	}
	return eligible
}

// clusterExperiences groups similar experiences using hierarchical clustering.
func (mc *MemoryConsolidator) clusterExperiences(experiences []*ExperienceTuple) [][]*ExperienceTuple {
	if len(experiences) < mc.config.MinClusterSize {
		return nil
	}

	// Simple agglomerative clustering
	// Start with each experience as its own cluster
	clusters := make([][]*ExperienceTuple, len(experiences))
	for i, exp := range experiences {
		clusters[i] = []*ExperienceTuple{exp}
	}

	// Iteratively merge most similar clusters
	for len(clusters) > 1 {
		bestI, bestJ := -1, -1
		bestSim := mc.config.SimilarityThreshold

		// Find most similar pair
		for i := 0; i < len(clusters); i++ {
			for j := i + 1; j < len(clusters); j++ {
				sim := mc.clusterSimilarity(clusters[i], clusters[j])
				if sim > bestSim {
					bestSim = sim
					bestI, bestJ = i, j
				}
			}
		}

		// No more merges possible
		if bestI == -1 {
			break
		}

		// Merge clusters i and j
		clusters[bestI] = append(clusters[bestI], clusters[bestJ]...)
		clusters = append(clusters[:bestJ], clusters[bestJ+1:]...)
	}

	return clusters
}

// clusterSimilarity computes average similarity between two clusters.
func (mc *MemoryConsolidator) clusterSimilarity(c1, c2 []*ExperienceTuple) float64 {
	if len(c1) == 0 || len(c2) == 0 {
		return 0
	}

	totalSim := 0.0
	count := 0

	for _, e1 := range c1 {
		for _, e2 := range c2 {
			totalSim += mc.experienceSimilarity(e1, e2)
			count++
		}
	}

	return totalSim / float64(count)
}

// experienceSimilarity computes similarity between two experiences.
func (mc *MemoryConsolidator) experienceSimilarity(e1, e2 *ExperienceTuple) float64 {
	// Use embedding similarity if available
	if len(e1.Embedding) > 0 && len(e2.Embedding) > 0 {
		return cosineSimilarity32(e1.Embedding, e2.Embedding)
	}

	// Fallback: compare agent and tier
	sim := 0.0
	if e1.AgentID == e2.AgentID {
		sim += 0.5
	}
	if e1.TierID == e2.TierID {
		sim += 0.3
	}
	if math.Abs(e1.FitnessScore-e2.FitnessScore) < 0.1 {
		sim += 0.2
	}

	return sim
}

// extractSchema extracts a common pattern from a cluster.
func (mc *MemoryConsolidator) extractSchema(cluster []*ExperienceTuple) *MemorySchema {
	if len(cluster) == 0 {
		return nil
	}

	// Compute centroid of embeddings
	var centroid []float32
	if len(cluster[0].Embedding) > 0 {
		dim := len(cluster[0].Embedding)
		centroid = make([]float32, dim)
		for _, exp := range cluster {
			for i, val := range exp.Embedding {
				centroid[i] += val
			}
		}
		for i := range centroid {
			centroid[i] /= float32(len(cluster))
		}
	}

	// Extract common strategy pattern
	strategyPattern := SchemaPattern{
		Variables:     make(map[string]string),
		Constraints:   []PatternConstraint{},
		FeatureVector: centroid,
	}

	// Compute confidence based on cluster cohesion
	confidence := mc.computeClusterCohesion(cluster)

	schema := &MemorySchema{
		Name:            fmt.Sprintf("schema-%d", time.Now().UnixNano()),
		InputPattern:    SchemaPattern{Variables: make(map[string]string)},
		StrategyPattern: strategyPattern,
		OutputPattern:   SchemaPattern{Variables: make(map[string]string)},
		Preconditions:   []string{},
		Effects:         []string{},
		Confidence:      confidence,
	}

	return schema
}

// computeClusterCohesion measures how similar cluster members are.
func (mc *MemoryConsolidator) computeClusterCohesion(cluster []*ExperienceTuple) float64 {
	if len(cluster) < 2 {
		return 1.0
	}

	totalSim := 0.0
	count := 0
	for i := 0; i < len(cluster); i++ {
		for j := i + 1; j < len(cluster); j++ {
			totalSim += mc.experienceSimilarity(cluster[i], cluster[j])
			count++
		}
	}

	return totalSim / float64(count)
}

// selectExemplars chooses the best representative experiences.
func (mc *MemoryConsolidator) selectExemplars(cluster []*ExperienceTuple, count int) []*ExperienceTuple {
	if len(cluster) <= count {
		return cluster
	}

	// Sort by fitness score (descending)
	sorted := make([]*ExperienceTuple, len(cluster))
	copy(sorted, cluster)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].FitnessScore > sorted[j].FitnessScore
	})

	// Take top N
	return sorted[:count]
}

// computeAbstractionLevel determines how abstract vs specific a pattern is.
func (mc *MemoryConsolidator) computeAbstractionLevel(cluster []*ExperienceTuple) float64 {
	// More diverse cluster = higher abstraction
	agentSet := make(map[string]bool)
	tierSet := make(map[int]bool)

	for _, exp := range cluster {
		agentSet[exp.AgentID] = true
		tierSet[exp.TierID] = true
	}

	// Normalized by max possible diversity
	agentDiversity := float64(len(agentSet)) / float64(len(cluster))
	tierDiversity := float64(len(tierSet)) / float64(len(cluster))

	return (agentDiversity + tierDiversity) / 2.0
}

// extractAgentIDs extracts unique agent IDs from cluster.
func (mc *MemoryConsolidator) extractAgentIDs(cluster []*ExperienceTuple) []string {
	agentSet := make(map[string]bool)
	for _, exp := range cluster {
		agentSet[exp.AgentID] = true
	}

	agents := make([]string, 0, len(agentSet))
	for agent := range agentSet {
		agents = append(agents, agent)
	}
	return agents
}

// extractTierIDs extracts unique tier IDs from cluster.
func (mc *MemoryConsolidator) extractTierIDs(cluster []*ExperienceTuple) []int {
	tierSet := make(map[int]bool)
	for _, exp := range cluster {
		tierSet[exp.TierID] = true
	}

	tiers := make([]int, 0, len(tierSet))
	for tier := range tierSet {
		tiers = append(tiers, tier)
	}
	return tiers
}

// computeAverageFitness computes average fitness of cluster.
func (mc *MemoryConsolidator) computeAverageFitness(cluster []*ExperienceTuple) float64 {
	if len(cluster) == 0 {
		return 0
	}

	total := 0.0
	for _, exp := range cluster {
		total += exp.FitnessScore
	}
	return total / float64(len(cluster))
}

// calculateCompressionRatio computes how much compression was achieved.
func (mc *MemoryConsolidator) calculateCompressionRatio(
	original []*ExperienceTuple,
	consolidated []*ConsolidatedMemory,
) float64 {
	if len(original) == 0 {
		return 0
	}

	// Count retained exemplars
	retained := 0
	for _, cm := range consolidated {
		retained += len(cm.Exemplars)
	}

	return 1.0 - (float64(retained) / float64(len(original)))
}

// updateStats updates consolidation statistics.
func (mc *MemoryConsolidator) updateStats(result *ConsolidationResult) {
	mc.statsMu.Lock()
	defer mc.statsMu.Unlock()

	mc.stats.TotalConsolidations++
	mc.stats.ExperiencesProcessed += int64(result.ExperiencesProcessed)
	mc.stats.ClustersFormed += int64(result.ClustersFormed)
	mc.stats.SchemasExtracted += int64(result.SchemasExtracted)
	mc.stats.LastConsolidationTime = time.Now()

	// Running average of compression ratio
	if mc.stats.TotalConsolidations == 1 {
		mc.stats.CompressionRatio = result.CompressionRatio
	} else {
		n := float64(mc.stats.TotalConsolidations)
		mc.stats.CompressionRatio = (mc.stats.CompressionRatio*(n-1) + result.CompressionRatio) / n
	}

	// Running average of cluster size
	if result.ClustersFormed > 0 {
		avgClusterSize := float64(result.ExperiencesProcessed) / float64(result.ClustersFormed)
		if mc.stats.TotalConsolidations == 1 {
			mc.stats.AverageClusterSize = avgClusterSize
		} else {
			n := float64(mc.stats.TotalConsolidations)
			mc.stats.AverageClusterSize = (mc.stats.AverageClusterSize*(n-1) + avgClusterSize) / n
		}
	}
}

// runAutoConsolidation runs periodic consolidation in background.
func (mc *MemoryConsolidator) runAutoConsolidation() {
	ticker := time.NewTicker(mc.config.ConsolidationInterval)
	defer ticker.Stop()
	defer close(mc.doneChan)

	for {
		select {
		case <-ticker.C:
			mc.Consolidate()
		case <-mc.stopChan:
			return
		}
	}
}

// GetConsolidated returns all consolidated memories.
func (mc *MemoryConsolidator) GetConsolidated() map[string]*ConsolidatedMemory {
	mc.consolidatedMu.RLock()
	defer mc.consolidatedMu.RUnlock()

	// Return copy
	result := make(map[string]*ConsolidatedMemory, len(mc.consolidated))
	for k, v := range mc.consolidated {
		result[k] = v
	}
	return result
}

// GetSchema retrieves a specific schema by ID.
func (mc *MemoryConsolidator) GetSchema(id string) (*ConsolidatedMemory, bool) {
	mc.consolidatedMu.RLock()
	defer mc.consolidatedMu.RUnlock()

	cm, ok := mc.consolidated[id]
	return cm, ok
}

// GetStats returns consolidation statistics.
func (mc *MemoryConsolidator) GetStats() *ConsolidationStats {
	mc.statsMu.RLock()
	defer mc.statsMu.RUnlock()

	// Return copy
	return &ConsolidationStats{
		TotalConsolidations:   mc.stats.TotalConsolidations,
		ExperiencesProcessed:  mc.stats.ExperiencesProcessed,
		ClustersFormed:        mc.stats.ClustersFormed,
		SchemasExtracted:      mc.stats.SchemasExtracted,
		CompressionRatio:      mc.stats.CompressionRatio,
		LastConsolidationTime: mc.stats.LastConsolidationTime,
		AverageClusterSize:    mc.stats.AverageClusterSize,
	}
}

// GetBufferSize returns current short-term buffer size.
func (mc *MemoryConsolidator) GetBufferSize() int {
	mc.bufferMu.RLock()
	defer mc.bufferMu.RUnlock()
	return len(mc.shortTermBuffer)
}

// ClearBuffer clears the short-term buffer.
func (mc *MemoryConsolidator) ClearBuffer() {
	mc.bufferMu.Lock()
	defer mc.bufferMu.Unlock()
	mc.shortTermBuffer = mc.shortTermBuffer[:0]
}

// Stop stops automatic consolidation.
func (mc *MemoryConsolidator) Stop() {
	if mc.config.EnableAutoConsolidation {
		close(mc.stopChan)
		<-mc.doneChan
	}
}

// Helper function for float32 cosine similarity
func cosineSimilarity32(a, b []float32) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := 0; i < len(a); i++ {
		dotProduct += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}
