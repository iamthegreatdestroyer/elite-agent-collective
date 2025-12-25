package memory

import (
	"testing"
	"time"
)

func TestNewMemoryConsolidator(t *testing.T) {
	mc := NewMemoryConsolidator(nil)
	if mc == nil {
		t.Fatal("Expected non-nil consolidator")
	}

	if mc.config == nil {
		t.Error("Expected default config")
	}

	if mc.config.BufferCapacity != 100 {
		t.Errorf("Expected buffer capacity 100, got %d", mc.config.BufferCapacity)
	}
}

func TestMemoryConsolidator_AddToBuffer(t *testing.T) {
	mc := NewMemoryConsolidator(nil)

	exp := &ExperienceTuple{
		Input:          "test input",
		Output:         "test output",
		Strategy:       "test strategy",
		AgentID:        "TEST-01",
		TierID:         1,
		FitnessScore:   0.8,
		Timestamp: time.Now().UnixNano(),
		LastAccessTime: time.Now().Add(-1 * time.Hour).UnixNano(),
	}

	mc.AddToBuffer(exp)

	size := mc.GetBufferSize()
	if size != 1 {
		t.Errorf("Expected buffer size 1, got %d", size)
	}
}

func TestMemoryConsolidator_ConsolidateEmpty(t *testing.T) {
	mc := NewMemoryConsolidator(nil)

	result, err := mc.Consolidate()
	if err != nil {
		t.Fatalf("Consolidate failed: %v", err)
	}

	if result.ExperiencesProcessed != 0 {
		t.Errorf("Expected 0 experiences processed, got %d", result.ExperiencesProcessed)
	}
}

func TestMemoryConsolidator_ConsolidateSimple(t *testing.T) {
	config := DefaultConsolidatorConfig()
	config.MinClusterSize = 2
	config.SimilarityThreshold = 0.5
	config.ExemplarCount = 2
	config.MinTimeSinceLastAccess = 1 * time.Minute

	mc := NewMemoryConsolidator(config)

	// Add similar experiences
	baseTime := time.Now().Add(-2 * time.Minute)
	for i := 0; i < 5; i++ {
		exp := &ExperienceTuple{
			Input:          "similar task",
			Output:         "similar output",
			Strategy:       "similar strategy",
			AgentID:        "APEX",
			TierID:         1,
			FitnessScore:   0.8 + float64(i)*0.01,
			Timestamp: baseTime.UnixNano(),
			LastAccessTime: baseTime.UnixNano(),
			Embedding:      []float32{0.1, 0.2, 0.3, 0.4},
		}
		mc.AddToBuffer(exp)
	}

	result, err := mc.Consolidate()
	if err != nil {
		t.Fatalf("Consolidate failed: %v", err)
	}

	if result.ExperiencesProcessed != 5 {
		t.Errorf("Expected 5 experiences processed, got %d", result.ExperiencesProcessed)
	}

	if result.SchemasExtracted < 1 {
		t.Errorf("Expected at least 1 schema, got %d", result.SchemasExtracted)
	}

	if result.CompressionRatio <= 0 {
		t.Errorf("Expected positive compression ratio, got %f", result.CompressionRatio)
	}
}

func TestMemoryConsolidator_ClusteringSimilar(t *testing.T) {
	config := DefaultConsolidatorConfig()
	config.MinClusterSize = 2
	config.SimilarityThreshold = 0.7
	config.MinTimeSinceLastAccess = 1 * time.Minute

	mc := NewMemoryConsolidator(config)

	baseTime := time.Now().Add(-2 * time.Minute)

	// Cluster 1: APEX experiences
	for i := 0; i < 3; i++ {
		exp := &ExperienceTuple{
			Input:          "apex task",
			AgentID:        "APEX",
			TierID:         1,
			FitnessScore:   0.8,
			Timestamp: baseTime.UnixNano(),
			LastAccessTime: baseTime.UnixNano(),
			Embedding:      []float32{1.0, 0.0, 0.0},
		}
		mc.AddToBuffer(exp)
	}

	// Cluster 2: TENSOR experiences
	for i := 0; i < 3; i++ {
		exp := &ExperienceTuple{
			Input:          "tensor task",
			AgentID:        "TENSOR",
			TierID:         2,
			FitnessScore:   0.7,
			Timestamp: baseTime.UnixNano(),
			LastAccessTime: baseTime.UnixNano(),
			Embedding:      []float32{0.0, 1.0, 0.0},
		}
		mc.AddToBuffer(exp)
	}

	result, err := mc.Consolidate()
	if err != nil {
		t.Fatalf("Consolidate failed: %v", err)
	}

	// Should form at least 2 clusters
	if result.ClustersFormed < 2 {
		t.Errorf("Expected at least 2 clusters, got %d", result.ClustersFormed)
	}
}

func TestMemoryConsolidator_ExemplarSelection(t *testing.T) {
	config := DefaultConsolidatorConfig()
	config.ExemplarCount = 2
	mc := NewMemoryConsolidator(config)

	cluster := []*ExperienceTuple{
		{FitnessScore: 0.5},
		{FitnessScore: 0.9},
		{FitnessScore: 0.7},
		{FitnessScore: 0.3},
		{FitnessScore: 0.8},
	}

	exemplars := mc.selectExemplars(cluster, 2)

	if len(exemplars) != 2 {
		t.Errorf("Expected 2 exemplars, got %d", len(exemplars))
	}

	// Should select highest fitness scores
	if exemplars[0].FitnessScore != 0.9 {
		t.Errorf("Expected first exemplar fitness 0.9, got %f", exemplars[0].FitnessScore)
	}

	if exemplars[1].FitnessScore != 0.8 {
		t.Errorf("Expected second exemplar fitness 0.8, got %f", exemplars[1].FitnessScore)
	}
}

func TestMemoryConsolidator_AbstractionLevel(t *testing.T) {
	mc := NewMemoryConsolidator(nil)

	// Low abstraction: all same agent/tier
	lowCluster := []*ExperienceTuple{
		{AgentID: "APEX", TierID: 1},
		{AgentID: "APEX", TierID: 1},
		{AgentID: "APEX", TierID: 1},
	}

	lowLevel := mc.computeAbstractionLevel(lowCluster)
	if lowLevel >= 0.5 {
		t.Errorf("Expected low abstraction level, got %f", lowLevel)
	}

	// High abstraction: diverse agents/tiers
	highCluster := []*ExperienceTuple{
		{AgentID: "APEX", TierID: 1},
		{AgentID: "TENSOR", TierID: 2},
		{AgentID: "CIPHER", TierID: 1},
	}

	highLevel := mc.computeAbstractionLevel(highCluster)
	if highLevel <= 0.5 {
		t.Errorf("Expected high abstraction level, got %f", highLevel)
	}
}

func TestMemoryConsolidator_SchemaExtraction(t *testing.T) {
	mc := NewMemoryConsolidator(nil)

	cluster := []*ExperienceTuple{
		{
			Input:     "task 1",
			Output:    "result 1",
			Strategy:  "strategy 1",
			AgentID:   "APEX",
			TierID:    1,
			Embedding: []float32{0.1, 0.2, 0.3},
		},
		{
			Input:     "task 2",
			Output:    "result 2",
			Strategy:  "strategy 2",
			AgentID:   "APEX",
			TierID:    1,
			Embedding: []float32{0.15, 0.25, 0.35},
		},
	}

	schema := mc.extractSchema(cluster)

	if schema == nil {
		t.Fatal("Expected non-nil schema")
	}

	if schema.Name == "" {
		t.Error("Expected schema name")
	}

	if schema.Confidence <= 0 || schema.Confidence > 1 {
		t.Errorf("Expected confidence in [0,1], got %f", schema.Confidence)
	}

	if len(schema.StrategyPattern.FeatureVector) != 3 {
		t.Errorf("Expected feature vector length 3, got %d", len(schema.StrategyPattern.FeatureVector))
	}
}

func TestMemoryConsolidator_CompressionRatio(t *testing.T) {
	mc := NewMemoryConsolidator(nil)

	original := []*ExperienceTuple{
		{}, {}, {}, {}, {}, {}, {}, {}, {}, {},
	}

	consolidated := []*ConsolidatedMemory{
		{Exemplars: []*ExperienceTuple{{}, {}}},
		{Exemplars: []*ExperienceTuple{{}}},
	}

	ratio := mc.calculateCompressionRatio(original, consolidated)

	// 10 original -> 3 retained = 0.7 compression
	expected := 0.7
	if ratio != expected {
		t.Errorf("Expected compression ratio %f, got %f", expected, ratio)
	}
}

func TestMemoryConsolidator_FilterEligible(t *testing.T) {
	config := DefaultConsolidatorConfig()
	config.MinTimeSinceLastAccess = 10 * time.Minute
	mc := NewMemoryConsolidator(config)

	now := time.Now().UnixNano()
	experiences := []*ExperienceTuple{
		{LastAccessTime: now - (5 * time.Minute).Nanoseconds()},  // Too recent
		{LastAccessTime: now - (15 * time.Minute).Nanoseconds()}, // Eligible
		{LastAccessTime: now - (20 * time.Minute).Nanoseconds()}, // Eligible
		{LastAccessTime: now - (1 * time.Minute).Nanoseconds()},  // Too recent
	}

	eligible := mc.filterEligible(experiences)

	if len(eligible) != 2 {
		t.Errorf("Expected 2 eligible experiences, got %d", len(eligible))
	}
}

func TestMemoryConsolidator_ExperienceSimilarity(t *testing.T) {
	mc := NewMemoryConsolidator(nil)

	// Test with embeddings
	e1 := &ExperienceTuple{
		Embedding: []float32{1.0, 0.0, 0.0},
	}
	e2 := &ExperienceTuple{
		Embedding: []float32{1.0, 0.0, 0.0},
	}
	e3 := &ExperienceTuple{
		Embedding: []float32{0.0, 1.0, 0.0},
	}

	sim12 := mc.experienceSimilarity(e1, e2)
	sim13 := mc.experienceSimilarity(e1, e3)

	if sim12 <= sim13 {
		t.Errorf("Expected identical experiences more similar: %f vs %f", sim12, sim13)
	}

	// Test without embeddings
	e4 := &ExperienceTuple{
		AgentID:      "APEX",
		TierID:       1,
		FitnessScore: 0.8,
	}
	e5 := &ExperienceTuple{
		AgentID:      "APEX",
		TierID:       1,
		FitnessScore: 0.85,
	}

	sim45 := mc.experienceSimilarity(e4, e5)
	if sim45 <= 0 {
		t.Errorf("Expected positive similarity, got %f", sim45)
	}
}

func TestMemoryConsolidator_GetConsolidated(t *testing.T) {
	mc := NewMemoryConsolidator(nil)

	// Manually add a consolidated memory
	mc.consolidatedMu.Lock()
	mc.consolidated["test-id"] = &ConsolidatedMemory{
		ID:        "test-id",
		Frequency: 5,
	}
	mc.consolidatedMu.Unlock()

	consolidated := mc.GetConsolidated()

	if len(consolidated) != 1 {
		t.Errorf("Expected 1 consolidated memory, got %d", len(consolidated))
	}

	cm, ok := consolidated["test-id"]
	if !ok {
		t.Fatal("Expected to find test-id")
	}

	if cm.Frequency != 5 {
		t.Errorf("Expected frequency 5, got %d", cm.Frequency)
	}
}

func TestMemoryConsolidator_GetSchema(t *testing.T) {
	mc := NewMemoryConsolidator(nil)

	mc.consolidatedMu.Lock()
	mc.consolidated["schema-123"] = &ConsolidatedMemory{
		ID: "schema-123",
		Schema: &MemorySchema{
			Name: "test schema",
		},
	}
	mc.consolidatedMu.Unlock()

	cm, ok := mc.GetSchema("schema-123")
	if !ok {
		t.Fatal("Expected to find schema")
	}

	if cm.Schema.Name != "test schema" {
		t.Errorf("Expected schema name 'test schema', got '%s'", cm.Schema.Name)
	}

	_, ok = mc.GetSchema("nonexistent")
	if ok {
		t.Error("Expected not to find nonexistent schema")
	}
}

func TestMemoryConsolidator_Stats(t *testing.T) {
	mc := NewMemoryConsolidator(nil)

	stats := mc.GetStats()
	if stats.TotalConsolidations != 0 {
		t.Errorf("Expected 0 consolidations initially, got %d", stats.TotalConsolidations)
	}

	// Manually update stats
	mc.updateStats(&ConsolidationResult{
		ExperiencesProcessed: 10,
		ClustersFormed:       2,
		SchemasExtracted:     2,
		CompressionRatio:     0.8,
	})

	stats = mc.GetStats()
	if stats.TotalConsolidations != 1 {
		t.Errorf("Expected 1 consolidation, got %d", stats.TotalConsolidations)
	}

	if stats.ExperiencesProcessed != 10 {
		t.Errorf("Expected 10 experiences processed, got %d", stats.ExperiencesProcessed)
	}

	if stats.CompressionRatio != 0.8 {
		t.Errorf("Expected compression ratio 0.8, got %f", stats.CompressionRatio)
	}

	// Second consolidation should average compression ratio
	mc.updateStats(&ConsolidationResult{
		ExperiencesProcessed: 5,
		ClustersFormed:       1,
		SchemasExtracted:     1,
		CompressionRatio:     0.6,
	})

	stats = mc.GetStats()
	expected := (0.8 + 0.6) / 2.0
	if stats.CompressionRatio != expected {
		t.Errorf("Expected average compression ratio %f, got %f", expected, stats.CompressionRatio)
	}
}

func TestMemoryConsolidator_ClearBuffer(t *testing.T) {
	mc := NewMemoryConsolidator(nil)

	mc.AddToBuffer(&ExperienceTuple{Input: "test"})
	mc.AddToBuffer(&ExperienceTuple{Input: "test2"})

	if mc.GetBufferSize() != 2 {
		t.Errorf("Expected buffer size 2, got %d", mc.GetBufferSize())
	}

	mc.ClearBuffer()

	if mc.GetBufferSize() != 0 {
		t.Errorf("Expected buffer size 0 after clear, got %d", mc.GetBufferSize())
	}
}

func TestCosineSimilarity32(t *testing.T) {
	// Identical vectors
	a := []float32{1.0, 0.0, 0.0}
	b := []float32{1.0, 0.0, 0.0}
	sim := cosineSimilarity32(a, b)
	if sim != 1.0 {
		t.Errorf("Expected similarity 1.0 for identical vectors, got %f", sim)
	}

	// Orthogonal vectors
	c := []float32{1.0, 0.0, 0.0}
	d := []float32{0.0, 1.0, 0.0}
	sim = cosineSimilarity32(c, d)
	if sim != 0.0 {
		t.Errorf("Expected similarity 0.0 for orthogonal vectors, got %f", sim)
	}

	// Similar vectors
	e := []float32{1.0, 1.0, 0.0}
	f := []float32{1.0, 0.9, 0.1}
	sim = cosineSimilarity32(e, f)
	if sim < 0.9 {
		t.Errorf("Expected high similarity, got %f", sim)
	}

	// Different lengths
	g := []float32{1.0}
	h := []float32{1.0, 0.0}
	sim = cosineSimilarity32(g, h)
	if sim != 0.0 {
		t.Errorf("Expected similarity 0.0 for different length vectors, got %f", sim)
	}
}

// Benchmark tests
func BenchmarkConsolidator_AddToBuffer(b *testing.B) {
	mc := NewMemoryConsolidator(nil)
	exp := &ExperienceTuple{
		Input:     "test",
		AgentID:   "TEST",
		TierID:    1,
		Embedding: make([]float32, 384),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mc.AddToBuffer(exp)
	}
}

func BenchmarkConsolidator_Consolidate(b *testing.B) {
	mc := NewMemoryConsolidator(&ConsolidatorConfig{
		BufferCapacity:         100,
		MinClusterSize:         3,
		SimilarityThreshold:    0.7,
		ExemplarCount:          3,
		MinTimeSinceLastAccess: 1 * time.Minute,
	})

	baseTime := time.Now().Add(-2 * time.Minute)

	// Pre-populate buffer
	for i := 0; i < 50; i++ {
		exp := &ExperienceTuple{
			Input:          "test input",
			AgentID:        "APEX",
			TierID:         1,
			FitnessScore:   0.8,
			Timestamp: baseTime.UnixNano(),
			LastAccessTime: baseTime.UnixNano(),
			Embedding:      make([]float32, 384),
		}
		for j := 0; j < 384; j++ {
			exp.Embedding[j] = float32(i%10) * 0.1
		}
		mc.AddToBuffer(exp)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mc.Consolidate()
		// Repopulate for next iteration
		for j := 0; j < 50; j++ {
			exp := &ExperienceTuple{
				Input:          "test input",
				AgentID:        "APEX",
				TierID:         1,
				FitnessScore:   0.8,
				Timestamp: baseTime.UnixNano(),
				LastAccessTime: baseTime.UnixNano(),
				Embedding:      make([]float32, 384),
			}
			mc.AddToBuffer(exp)
		}
	}
}
