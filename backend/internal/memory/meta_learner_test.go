package memory

import (
	"context"
	"testing"
	"time"
)

// =============================================================================
// META-LEARNER TESTS
// =============================================================================

func TestMetaLearner_NewMetaLearner(t *testing.T) {
	t.Run("with default config", func(t *testing.T) {
		ml := NewMetaLearner(nil, nil)
		if ml == nil {
			t.Fatal("expected non-nil meta-learner")
		}
		if ml.adaptationSteps != 5 {
			t.Errorf("expected 5 adaptation steps, got %d", ml.adaptationSteps)
		}
	})

	t.Run("with custom config", func(t *testing.T) {
		config := &MetaLearnerConfig{
			AdaptationSteps:  10,
			AdaptationRate:   0.05,
			MetaLearningRate: 0.01,
			SupportSetSize:   8,
			QuerySetSize:     15,
		}
		ml := NewMetaLearner(config, nil)
		if ml.adaptationSteps != 10 {
			t.Errorf("expected 10 adaptation steps, got %d", ml.adaptationSteps)
		}
		if ml.adaptationRate != 0.05 {
			t.Errorf("expected 0.05 adaptation rate, got %f", ml.adaptationRate)
		}
	})
}

func TestMetaLearner_InitializeAgent(t *testing.T) {
	ml := NewMetaLearner(nil, nil)

	t.Run("initialize with nil params creates defaults", func(t *testing.T) {
		err := ml.InitializeAgent("APEX", nil)
		if err != nil {
			t.Fatalf("failed to initialize agent: %v", err)
		}

		params, exists := ml.GetBaseParameters("APEX")
		if !exists {
			t.Fatal("agent parameters should exist")
		}
		if len(params.StrategyWeights) != 10 {
			t.Errorf("expected 10 strategy weights, got %d", len(params.StrategyWeights))
		}
		if len(params.ContextBias) != 64 {
			t.Errorf("expected 64 context bias dims, got %d", len(params.ContextBias))
		}
	})

	t.Run("initialize with custom params", func(t *testing.T) {
		customParams := &AgentParameters{
			ID:               "CIPHER",
			StrategyWeights:  []float64{0.1, 0.2, 0.3},
			ContextBias:      []float64{0.5, 0.5},
			ResponsePatterns: map[string]float64{"security": 0.9},
		}
		err := ml.InitializeAgent("CIPHER", customParams)
		if err != nil {
			t.Fatalf("failed to initialize: %v", err)
		}

		params, exists := ml.GetBaseParameters("CIPHER")
		if !exists || len(params.StrategyWeights) != 3 {
			t.Error("custom params not stored correctly")
		}
	})

	t.Run("empty agent ID fails", func(t *testing.T) {
		err := ml.InitializeAgent("", nil)
		if err == nil {
			t.Error("expected error for empty agent ID")
		}
	})
}

func TestMetaLearner_Adapt(t *testing.T) {
	ml := NewMetaLearner(nil, nil)
	ctx := context.Background()

	t.Run("adapt with support set", func(t *testing.T) {
		supportSet := createTestExamples(5)

		adapted, err := ml.Adapt(ctx, "TENSOR", supportSet)
		if err != nil {
			t.Fatalf("adaptation failed: %v", err)
		}

		if adapted.BaseAgentID != "TENSOR" {
			t.Errorf("expected TENSOR, got %s", adapted.BaseAgentID)
		}
		if adapted.Confidence <= 0 || adapted.Confidence > 1 {
			t.Errorf("confidence should be in (0,1], got %f", adapted.Confidence)
		}
		if len(adapted.SupportSet) != 5 {
			t.Errorf("expected 5 examples in support set")
		}
	})

	t.Run("adapt auto-initializes agent", func(t *testing.T) {
		supportSet := createTestExamples(3)

		adapted, err := ml.Adapt(ctx, "NEW_AGENT", supportSet)
		if err != nil {
			t.Fatalf("adaptation failed: %v", err)
		}

		if adapted == nil {
			t.Fatal("expected adapted agent")
		}

		// Check agent was initialized
		_, exists := ml.GetBaseParameters("NEW_AGENT")
		if !exists {
			t.Error("agent should be auto-initialized")
		}
	})

	t.Run("adapt respects context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		supportSet := createTestExamples(5)
		_, err := ml.Adapt(ctx, "FLUX", supportSet)
		if err == nil {
			t.Error("expected context cancellation error")
		}
	})

	t.Run("adapt with empty support set", func(t *testing.T) {
		adapted, err := ml.Adapt(ctx, "PRISM", []*Example{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Should still work with zero confidence
		if adapted.Confidence != 0 {
			t.Errorf("expected 0 confidence for empty support, got %f", adapted.Confidence)
		}
	})
}

func TestMetaLearner_MetaUpdate(t *testing.T) {
	ml := NewMetaLearner(nil, nil)
	ctx := context.Background()

	t.Run("meta update modifies base parameters", func(t *testing.T) {
		err := ml.InitializeAgent("APEX", nil)
		if err != nil {
			t.Fatalf("failed to initialize: %v", err)
		}

		beforeParams, _ := ml.GetBaseParameters("APEX")
		beforeVersion := beforeParams.Version

		tasks := createTestMetaTasks(3)
		err = ml.MetaUpdate(ctx, "APEX", tasks)
		if err != nil {
			t.Fatalf("meta update failed: %v", err)
		}

		afterParams, _ := ml.GetBaseParameters("APEX")
		if afterParams.Version <= beforeVersion {
			t.Error("version should increment after meta update")
		}
	})

	t.Run("meta update fails for uninitialized agent", func(t *testing.T) {
		tasks := createTestMetaTasks(2)
		err := ml.MetaUpdate(ctx, "NONEXISTENT", tasks)
		if err == nil {
			t.Error("expected error for uninitialized agent")
		}
	})

	t.Run("meta update with empty tasks", func(t *testing.T) {
		err := ml.InitializeAgent("EMPTY_TASKS", nil)
		if err != nil {
			t.Fatalf("failed to initialize: %v", err)
		}

		err = ml.MetaUpdate(ctx, "EMPTY_TASKS", []*MetaTask{})
		if err != nil {
			t.Fatalf("meta update with empty tasks should not error: %v", err)
		}
	})
}

func TestMetaLearner_Metrics(t *testing.T) {
	ml := NewMetaLearner(nil, nil)
	ctx := context.Background()

	// Perform some adaptations
	for i := 0; i < 5; i++ {
		supportSet := createTestExamples(3)
		ml.Adapt(ctx, "METRICS_TEST", supportSet)
	}

	metrics := ml.GetMetrics()
	if metrics.TotalAdaptations != 5 {
		t.Errorf("expected 5 adaptations, got %d", metrics.TotalAdaptations)
	}
	// Note: AverageAdaptTime may be 0 in fast tests, just verify it's non-negative
	if metrics.AverageAdaptTime < 0 {
		t.Error("average adapt time should be non-negative")
	}
}

func TestMetaLearner_CloneParameters(t *testing.T) {
	ml := NewMetaLearner(nil, nil)

	original := &AgentParameters{
		ID:               "TEST",
		StrategyWeights:  []float64{1.0, 2.0, 3.0},
		ContextBias:      []float64{0.1, 0.2},
		ResponsePatterns: map[string]float64{"pattern": 0.5},
		Version:          5,
	}

	cloned := ml.cloneParameters(original)

	// Modify original
	original.StrategyWeights[0] = 999.0
	original.ResponsePatterns["new"] = 1.0

	// Clone should be unchanged
	if cloned.StrategyWeights[0] != 1.0 {
		t.Error("clone should be independent of original")
	}
	if _, ok := cloned.ResponsePatterns["new"]; ok {
		t.Error("clone should not have new pattern")
	}
}

// =============================================================================
// PROTOTYPICAL ROUTER TESTS
// =============================================================================

func TestPrototypicalRouter_New(t *testing.T) {
	router := NewPrototypicalRouter(64, EuclideanDistance)
	if router == nil {
		t.Fatal("expected non-nil router")
	}
	if router.embeddingDim != 64 {
		t.Errorf("expected 64 dim, got %d", router.embeddingDim)
	}
}

func TestPrototypicalRouter_UpdatePrototype(t *testing.T) {
	router := NewPrototypicalRouter(8, EuclideanDistance)

	t.Run("update with valid examples", func(t *testing.T) {
		examples := createTestExamplesWithDim(5, 8)
		err := router.UpdatePrototype("APEX", examples)
		if err != nil {
			t.Fatalf("failed to update prototype: %v", err)
		}

		proto, exists := router.GetPrototype("APEX")
		if !exists {
			t.Fatal("prototype should exist")
		}
		if len(proto) != 8 {
			t.Errorf("expected 8 dim prototype, got %d", len(proto))
		}
	})

	t.Run("update fails with insufficient examples", func(t *testing.T) {
		examples := createTestExamplesWithDim(2, 8)
		err := router.UpdatePrototype("INSUFFICIENT", examples)
		if err == nil {
			t.Error("expected error for insufficient examples")
		}
	})

	t.Run("update with momentum blending", func(t *testing.T) {
		examples1 := make([]*Example, 5)
		for i := range examples1 {
			examples1[i] = &Example{
				TaskEmbedding: make([]float64, 8),
			}
			for j := range examples1[i].TaskEmbedding {
				examples1[i].TaskEmbedding[j] = 1.0 // All ones
			}
		}

		router.UpdatePrototype("BLEND", examples1)
		proto1, _ := router.GetPrototype("BLEND")

		// Second update with different values
		examples2 := make([]*Example, 5)
		for i := range examples2 {
			examples2[i] = &Example{
				TaskEmbedding: make([]float64, 8),
			}
			for j := range examples2[i].TaskEmbedding {
				examples2[i].TaskEmbedding[j] = 0.0 // All zeros
			}
		}

		router.UpdatePrototype("BLEND", examples2)
		proto2, _ := router.GetPrototype("BLEND")

		// With momentum, new prototype should be between old and new
		if proto2[0] == proto1[0] || proto2[0] == 0.0 {
			t.Error("momentum should blend old and new prototypes")
		}
	})
}

func TestPrototypicalRouter_Route(t *testing.T) {
	router := NewPrototypicalRouter(8, EuclideanDistance)

	// Set up prototypes
	apexExamples := make([]*Example, 5)
	for i := range apexExamples {
		apexExamples[i] = &Example{
			TaskEmbedding: []float64{1, 0, 0, 0, 0, 0, 0, 0},
		}
	}
	router.UpdatePrototype("APEX", apexExamples)

	tensorExamples := make([]*Example, 5)
	for i := range tensorExamples {
		tensorExamples[i] = &Example{
			TaskEmbedding: []float64{0, 1, 0, 0, 0, 0, 0, 0},
		}
	}
	router.UpdatePrototype("TENSOR", tensorExamples)

	t.Run("route to nearest prototype", func(t *testing.T) {
		// Query close to APEX
		query := []float64{0.9, 0.1, 0, 0, 0, 0, 0, 0}
		agent, confidence, err := router.Route(query)
		if err != nil {
			t.Fatalf("routing failed: %v", err)
		}
		if agent != "APEX" {
			t.Errorf("expected APEX, got %s", agent)
		}
		if confidence <= 0 || confidence > 1 {
			t.Errorf("confidence should be in (0,1], got %f", confidence)
		}
	})

	t.Run("route to TENSOR", func(t *testing.T) {
		query := []float64{0.1, 0.9, 0, 0, 0, 0, 0, 0}
		agent, _, err := router.Route(query)
		if err != nil {
			t.Fatalf("routing failed: %v", err)
		}
		if agent != "TENSOR" {
			t.Errorf("expected TENSOR, got %s", agent)
		}
	})

	t.Run("route fails with short embedding", func(t *testing.T) {
		_, _, err := router.Route([]float64{0.5})
		if err == nil {
			t.Error("expected error for short embedding")
		}
	})

	t.Run("route fails with no prototypes", func(t *testing.T) {
		emptyRouter := NewPrototypicalRouter(8, EuclideanDistance)
		_, _, err := emptyRouter.Route([]float64{0, 0, 0, 0, 0, 0, 0, 0})
		if err == nil {
			t.Error("expected error for no prototypes")
		}
	})
}

func TestPrototypicalRouter_RouteTopK(t *testing.T) {
	router := NewPrototypicalRouter(4, EuclideanDistance)

	// Set up 4 prototypes
	agents := []string{"A", "B", "C", "D"}
	for i, agent := range agents {
		examples := make([]*Example, 5)
		for j := range examples {
			embedding := []float64{0, 0, 0, 0}
			embedding[i] = 1.0
			examples[j] = &Example{TaskEmbedding: embedding}
		}
		router.UpdatePrototype(agent, examples)
	}

	t.Run("get top 2", func(t *testing.T) {
		query := []float64{0.9, 0.8, 0.1, 0.0}
		candidates, err := router.RouteTopK(query, 2)
		if err != nil {
			t.Fatalf("RouteTopK failed: %v", err)
		}
		if len(candidates) != 2 {
			t.Errorf("expected 2 candidates, got %d", len(candidates))
		}
		// Should be sorted by distance
		if candidates[0].Distance > candidates[1].Distance {
			t.Error("candidates should be sorted by distance")
		}
	})

	t.Run("top K larger than agents", func(t *testing.T) {
		query := []float64{0.5, 0.5, 0.5, 0.5}
		candidates, err := router.RouteTopK(query, 10)
		if err != nil {
			t.Fatalf("RouteTopK failed: %v", err)
		}
		if len(candidates) != 4 {
			t.Errorf("expected 4 candidates (max), got %d", len(candidates))
		}
	})
}

func TestPrototypicalRouter_DistanceMetrics(t *testing.T) {
	t.Run("euclidean distance", func(t *testing.T) {
		router := NewPrototypicalRouter(4, EuclideanDistance)
		a := []float64{1, 0, 0, 0}
		b := []float64{0, 1, 0, 0}
		dist := router.euclideanDistance(a, b)
		expected := 1.41421356 // sqrt(2)
		if dist < expected-0.001 || dist > expected+0.001 {
			t.Errorf("expected ~%f, got %f", expected, dist)
		}
	})

	t.Run("cosine distance", func(t *testing.T) {
		router := NewPrototypicalRouter(4, CosineDistance)
		a := []float64{1, 0, 0, 0}
		b := []float64{0, 1, 0, 0}
		dist := router.cosineDistance(a, b)
		if dist != 1.0 { // Orthogonal vectors have max cosine distance
			t.Errorf("expected 1.0, got %f", dist)
		}

		// Same vectors should have 0 distance
		dist = router.cosineDistance(a, a)
		if dist != 0.0 {
			t.Errorf("expected 0, got %f", dist)
		}
	})

	t.Run("manhattan distance", func(t *testing.T) {
		router := NewPrototypicalRouter(4, ManhattanDistance)
		a := []float64{1, 2, 3, 4}
		b := []float64{0, 0, 0, 0}
		dist := router.manhattanDistance(a, b)
		if dist != 10.0 {
			t.Errorf("expected 10.0, got %f", dist)
		}
	})
}

func TestPrototypicalRouter_ListAgents(t *testing.T) {
	router := NewPrototypicalRouter(4, EuclideanDistance)

	examples := createTestExamplesWithDim(5, 4)
	router.UpdatePrototype("APEX", examples)
	router.UpdatePrototype("CIPHER", examples)
	router.UpdatePrototype("TENSOR", examples)

	agents := router.ListAgents()
	if len(agents) != 3 {
		t.Errorf("expected 3 agents, got %d", len(agents))
	}
}

func TestPrototypicalRouter_GetStats(t *testing.T) {
	router := NewPrototypicalRouter(4, EuclideanDistance)

	examples := createTestExamplesWithDim(5, 4)
	router.UpdatePrototype("APEX", examples)

	stats, ok := router.GetStats("APEX")
	if !ok {
		t.Fatal("stats should exist")
	}
	if stats.SampleCount != 5 {
		t.Errorf("expected 5 samples, got %d", stats.SampleCount)
	}
	if stats.LastUpdated.IsZero() {
		t.Error("last updated should be set")
	}
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

func createTestExamples(n int) []*Example {
	examples := make([]*Example, n)
	for i := 0; i < n; i++ {
		embedding := make([]float64, 64)
		for j := range embedding {
			embedding[j] = float64(i+j) / 100.0
		}
		examples[i] = &Example{
			ID:            string(rune('A' + i)),
			TaskEmbedding: embedding,
			Quality:       0.5 + float64(i)*0.1,
			AgentID:       "TEST",
			Timestamp:     time.Now(),
		}
	}
	return examples
}

func createTestExamplesWithDim(n, dim int) []*Example {
	examples := make([]*Example, n)
	for i := 0; i < n; i++ {
		embedding := make([]float64, dim)
		for j := range embedding {
			embedding[j] = float64(i+j) / 100.0
		}
		examples[i] = &Example{
			ID:            string(rune('A' + i)),
			TaskEmbedding: embedding,
			Quality:       0.5 + float64(i)*0.1,
			AgentID:       "TEST",
			Timestamp:     time.Now(),
		}
	}
	return examples
}

func createTestMetaTasks(n int) []*MetaTask {
	tasks := make([]*MetaTask, n)
	for i := 0; i < n; i++ {
		tasks[i] = &MetaTask{
			ID:         string(rune('T' + i)),
			SupportSet: createTestExamples(5),
			QuerySet:   createTestExamples(10),
			TaskType:   "test",
			Difficulty: float64(i+1) / float64(n),
		}
	}
	return tasks
}

// =============================================================================
// BENCHMARKS
// =============================================================================

func BenchmarkMetaLearner_Adapt(b *testing.B) {
	ml := NewMetaLearner(nil, nil)
	ctx := context.Background()
	supportSet := createTestExamples(5)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ml.Adapt(ctx, "BENCH", supportSet)
	}
}

func BenchmarkPrototypicalRouter_Route(b *testing.B) {
	router := NewPrototypicalRouter(64, EuclideanDistance)

	// Set up 40 prototypes (all agents)
	for i := 0; i < 40; i++ {
		examples := createTestExamplesWithDim(5, 64)
		router.UpdatePrototype(string(rune('A'+i)), examples)
	}

	query := make([]float64, 64)
	for i := range query {
		query[i] = float64(i) / 64.0
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		router.Route(query)
	}
}
