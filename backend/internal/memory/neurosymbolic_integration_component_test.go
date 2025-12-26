package memory

import (
	"context"
	"testing"
	"time"
)

// ============================================================================
// Tests for Neurosymbolic Integration Component
// ============================================================================

func TestNeurosymbolicIntegrationComponent_Initialize(t *testing.T) {
	goalStack := NewGoalStack(DefaultGoalStackConfig())
	impasseDetector := NewImpasseDetector(nil, nil)
	workingMemory := NewCognitiveWorkingMemoryComponent(100)

	component := NewNeurosymbolicIntegrationComponent(goalStack, impasseDetector, workingMemory)

	err := component.Initialize(nil)

	if err != nil {
		t.Errorf("Initialize failed: %v", err)
	}

	if component.GetName() != "NeurosymbolicIntegration" {
		t.Errorf("Name mismatch")
	}
}

func TestNeurosymbolicIntegrationComponent_Process_NoGoal(t *testing.T) {
	goalStack := NewGoalStack(DefaultGoalStackConfig())
	impasseDetector := NewImpasseDetector(nil, nil)
	workingMemory := NewCognitiveWorkingMemoryComponent(100)

	component := NewNeurosymbolicIntegrationComponent(goalStack, impasseDetector, workingMemory)
	component.Initialize(nil)

	request := &CognitiveProcessRequest{
		RequestID:   "req-001",
		Timestamp:   time.Now(),
		CurrentGoal: nil, // No goal
	}

	result, err := component.Process(context.Background(), request)

	if err == nil {
		t.Error("Should fail with no goal")
	}

	if result.Status != ProcessSuccess {
		// This is expected behavior
	}
}

func TestNeurosymbolicIntegrationComponent_Process_WithGoal(t *testing.T) {
	goalStack := NewGoalStack(DefaultGoalStackConfig())
	impasseDetector := NewImpasseDetector(nil, nil)
	workingMemory := NewCognitiveWorkingMemoryComponent(100)

	component := NewNeurosymbolicIntegrationComponent(goalStack, impasseDetector, workingMemory)
	component.Initialize(nil)

	goal := &Goal{
		ID:       "goal-neurosym",
		Name:     "Neurosymbolic Test Goal",
		Priority: PriorityHigh,
		Status:   GoalActive,
		Progress: 0.5,
	}

	_ = goalStack.Push(goal)

	request := &CognitiveProcessRequest{
		RequestID:   "req-neurosym",
		Timestamp:   time.Now(),
		CurrentGoal: goal,
	}

	result, err := component.Process(context.Background(), request)

	if err != nil {
		t.Errorf("Process failed: %v", err)
	}

	if result.Status != ProcessSuccess {
		t.Errorf("Status should be success")
	}

	decision := result.Output.(*HybridDecision)
	if decision == nil {
		t.Error("Output should be a HybridDecision")
	}
}

func TestNeurosymbolicIntegrationComponent_GenerateEmbedding(t *testing.T) {
	goalStack := NewGoalStack(DefaultGoalStackConfig())
	impasseDetector := NewImpasseDetector(nil, nil)
	workingMemory := NewCognitiveWorkingMemoryComponent(100)

	component := NewNeurosymbolicIntegrationComponent(goalStack, impasseDetector, workingMemory)
	component.Initialize(nil)

	embedding := component.generateEmbedding(context.Background(), "test-id", "test content")

	if embedding == nil {
		t.Error("Embedding should not be nil")
	}

	if embedding.ID != "test-id" {
		t.Errorf("ID mismatch")
	}

	if len(embedding.Vector) != 768 {
		t.Errorf("Embedding dimension should be 768")
	}

	// Verify it's unit-normalized
	magnitude := 0.0
	for _, v := range embedding.Vector {
		magnitude += v * v
	}
	magnitude = sqrt(magnitude)

	// Should be close to 1.0
	if magnitude < 0.99 || magnitude > 1.01 {
		t.Errorf("Embedding should be unit-normalized, got magnitude %f", magnitude)
	}
}

func TestNeurosymbolicIntegrationComponent_SymbolicReasoning(t *testing.T) {
	goalStack := NewGoalStack(DefaultGoalStackConfig())
	impasseDetector := NewImpasseDetector(nil, nil)
	workingMemory := NewCognitiveWorkingMemoryComponent(100)

	component := NewNeurosymbolicIntegrationComponent(goalStack, impasseDetector, workingMemory)
	component.Initialize(nil)

	tests := []struct {
		goal     *Goal
		minScore float64
		maxScore float64
	}{
		{
			goal: &Goal{
				ID:       "critical",
				Priority: PriorityCritical,
				Status:   GoalActive,
				Progress: 0.8,
			},
			minScore: 0.7,
			maxScore: 1.0,
		},
		{
			goal: &Goal{
				ID:       "normal",
				Priority: PriorityNormal,
				Status:   GoalPending,
				Progress: 0.0,
			},
			minScore: 0.3,
			maxScore: 0.7,
		},
	}

	for _, tt := range tests {
		score := component.symbolicReasoning(context.Background(), tt.goal)

		if score < tt.minScore || score > tt.maxScore {
			t.Errorf("Score %f not in range [%f, %f]", score, tt.minScore, tt.maxScore)
		}
	}
}

func TestNeurosymbolicIntegrationComponent_NeuralReasoning(t *testing.T) {
	goalStack := NewGoalStack(DefaultGoalStackConfig())
	impasseDetector := NewImpasseDetector(nil, nil)
	workingMemory := NewCognitiveWorkingMemoryComponent(100)

	component := NewNeurosymbolicIntegrationComponent(goalStack, impasseDetector, workingMemory)
	component.Initialize(nil)

	embedding := component.generateEmbedding(context.Background(), "test-neural", "test")

	score := component.neuralReasoning(context.Background(), embedding)

	if score < 0 || score > 1 {
		t.Errorf("Neural score should be in [0, 1], got %f", score)
	}
}

func TestNeurosymbolicIntegrationComponent_MakeHybridDecision(t *testing.T) {
	goalStack := NewGoalStack(DefaultGoalStackConfig())
	impasseDetector := NewImpasseDetector(nil, nil)
	workingMemory := NewCognitiveWorkingMemoryComponent(100)

	component := NewNeurosymbolicIntegrationComponent(goalStack, impasseDetector, workingMemory)
	component.Initialize(nil)

	symbolic := 0.8
	neural := 0.6

	decision := component.makeHybridDecision(context.Background(), "test-goal", symbolic, neural)

	if decision == nil {
		t.Error("Decision should not be nil")
	}

	if decision.SymbolicScore != symbolic {
		t.Errorf("Symbolic score mismatch")
	}

	if decision.NeuralScore != neural {
		t.Errorf("Neural score mismatch")
	}

	// Hybrid should be average of weighted scores
	expectedHybrid := 0.5*symbolic + 0.5*neural
	if decision.HybridScore != expectedHybrid {
		t.Errorf("Hybrid score mismatch: expected %f, got %f", expectedHybrid, decision.HybridScore)
	}

	if decision.Confidence < 0 || decision.Confidence > 1 {
		t.Errorf("Confidence should be in [0, 1]")
	}
}

func TestNeurosymbolicIntegrationComponent_CheckSymbolicConstraints(t *testing.T) {
	goalStack := NewGoalStack(DefaultGoalStackConfig())
	impasseDetector := NewImpasseDetector(nil, nil)
	workingMemory := NewCognitiveWorkingMemoryComponent(100)

	component := NewNeurosymbolicIntegrationComponent(goalStack, impasseDetector, workingMemory)
	component.Initialize(nil)

	goal := &Goal{
		ID:       "constraint-test",
		Priority: PriorityHigh,
		Status:   GoalActive,
		Progress: 0.5,
	}

	constraints := component.checkSymbolicConstraints(context.Background(), goal)

	if len(constraints) == 0 {
		t.Error("Should check at least one constraint")
	}

	// All constraints should not be violated for valid goal
	for _, c := range constraints {
		if c.Violated {
			t.Errorf("Constraint %s should not be violated", c.ID)
		}
	}
}

func TestNeurosymbolicIntegrationComponent_CheckSymbolicConstraints_CircularDependency(t *testing.T) {
	goalStack := NewGoalStack(DefaultGoalStackConfig())
	impasseDetector := NewImpasseDetector(nil, nil)
	workingMemory := NewCognitiveWorkingMemoryComponent(100)

	component := NewNeurosymbolicIntegrationComponent(goalStack, impasseDetector, workingMemory)
	component.Initialize(nil)

	goal := &Goal{
		ID:           "self-ref",
		Priority:     PriorityHigh,
		Status:       GoalActive,
		Dependencies: []string{"self-ref"}, // Self-reference!
	}

	constraints := component.checkSymbolicConstraints(context.Background(), goal)

	foundCircular := false
	for _, c := range constraints {
		if c.Condition == "no circular dependencies" && c.Violated {
			foundCircular = true
			break
		}
	}

	if !foundCircular {
		t.Error("Should detect circular dependency")
	}
}

func TestNeurosymbolicIntegrationComponent_RegisterEmbedding(t *testing.T) {
	goalStack := NewGoalStack(DefaultGoalStackConfig())
	impasseDetector := NewImpasseDetector(nil, nil)
	workingMemory := NewCognitiveWorkingMemoryComponent(100)

	component := NewNeurosymbolicIntegrationComponent(goalStack, impasseDetector, workingMemory)
	component.Initialize(nil)

	vector := make([]float64, 768)
	for i := range vector {
		vector[i] = float64(i) / 768.0
	}

	embedding := &SemanticEmbedding{
		ID:     "custom-emb",
		Vector: vector,
	}

	err := component.RegisterEmbedding(embedding)

	if err != nil {
		t.Errorf("RegisterEmbedding failed: %v", err)
	}

	retrieved := component.GetEmbedding("custom-emb")
	if retrieved == nil {
		t.Error("Should retrieve registered embedding")
	}
}

func TestNeurosymbolicIntegrationComponent_RegisterEmbedding_WrongDimension(t *testing.T) {
	goalStack := NewGoalStack(DefaultGoalStackConfig())
	impasseDetector := NewImpasseDetector(nil, nil)
	workingMemory := NewCognitiveWorkingMemoryComponent(100)

	component := NewNeurosymbolicIntegrationComponent(goalStack, impasseDetector, workingMemory)
	component.Initialize(nil)

	embedding := &SemanticEmbedding{
		ID:     "wrong-dim",
		Vector: make([]float64, 512), // Wrong dimension
	}

	err := component.RegisterEmbedding(embedding)

	if err == nil {
		t.Error("Should fail with wrong dimension")
	}
}

func TestNeurosymbolicIntegrationComponent_FindSimilarEmbeddings(t *testing.T) {
	goalStack := NewGoalStack(DefaultGoalStackConfig())
	impasseDetector := NewImpasseDetector(nil, nil)
	workingMemory := NewCognitiveWorkingMemoryComponent(100)

	component := NewNeurosymbolicIntegrationComponent(goalStack, impasseDetector, workingMemory)
	component.Initialize(nil)

	// Register embeddings
	emb1 := component.generateEmbedding(context.Background(), "emb1", "apple")
	_ = component.generateEmbedding(context.Background(), "emb2", "apple fruit")
	_ = component.generateEmbedding(context.Background(), "emb3", "car")

	// Find similar to emb1
	similar := component.FindSimilarEmbeddings(emb1, 10)

	if len(similar) == 0 {
		t.Error("Should find similar embeddings")
	}

	// Should have at least emb1 itself
	found := false
	for _, s := range similar {
		if s.ID == "emb1" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Should find query embedding in results")
	}
}

func TestNeurosymbolicIntegrationComponent_GetMetrics(t *testing.T) {
	goalStack := NewGoalStack(DefaultGoalStackConfig())
	impasseDetector := NewImpasseDetector(nil, nil)
	workingMemory := NewCognitiveWorkingMemoryComponent(100)

	component := NewNeurosymbolicIntegrationComponent(goalStack, impasseDetector, workingMemory)
	component.Initialize(nil)

	goal := &Goal{
		ID:       "metrics-goal",
		Priority: PriorityHigh,
		Status:   GoalActive,
	}

	_ = goalStack.Push(goal)

	request := &CognitiveProcessRequest{
		RequestID:   "metrics-req",
		CurrentGoal: goal,
		Timestamp:   time.Now(),
	}

	component.Process(context.Background(), request)

	metrics := component.GetMetrics()

	if metrics.ComponentName != "NeurosymbolicIntegration" {
		t.Errorf("Component name mismatch")
	}

	if metrics.TotalRequests != 1 {
		t.Errorf("Total requests should be 1")
	}

	if metrics.CustomMetrics == nil {
		t.Error("Custom metrics is nil")
	}
}

func TestNeurosymbolicIntegrationComponent_Shutdown(t *testing.T) {
	goalStack := NewGoalStack(DefaultGoalStackConfig())
	impasseDetector := NewImpasseDetector(nil, nil)
	workingMemory := NewCognitiveWorkingMemoryComponent(100)

	component := NewNeurosymbolicIntegrationComponent(goalStack, impasseDetector, workingMemory)
	component.Initialize(nil)

	err := component.Shutdown()

	if err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}
}

// ============================================================================
// Benchmark Tests
// ============================================================================

func BenchmarkNeurosymbolicIntegrationComponent_Process(b *testing.B) {
	goalStack := NewGoalStack(DefaultGoalStackConfig())
	impasseDetector := NewImpasseDetector(nil, nil)
	workingMemory := NewCognitiveWorkingMemoryComponent(100)

	component := NewNeurosymbolicIntegrationComponent(goalStack, impasseDetector, workingMemory)
	component.Initialize(nil)

	goal := &Goal{
		ID:       "bench-goal",
		Name:     "Benchmark Goal",
		Priority: PriorityHigh,
		Status:   GoalActive,
		Progress: 0.5,
	}

	_ = goalStack.Push(goal)

	request := &CognitiveProcessRequest{
		RequestID:   "bench-req",
		CurrentGoal: goal,
		Timestamp:   time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		component.Process(context.Background(), request)
	}
}

func BenchmarkNeurosymbolicIntegrationComponent_Embedding(b *testing.B) {
	goalStack := NewGoalStack(DefaultGoalStackConfig())
	impasseDetector := NewImpasseDetector(nil, nil)
	workingMemory := NewCognitiveWorkingMemoryComponent(100)

	component := NewNeurosymbolicIntegrationComponent(goalStack, impasseDetector, workingMemory)
	component.Initialize(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		component.generateEmbedding(context.Background(), "bench-id", "test content")
	}
}

func BenchmarkNeurosymbolicIntegrationComponent_CosineSimilarity(b *testing.B) {
	v1 := make([]float64, 768)
	v2 := make([]float64, 768)

	for i := range v1 {
		v1[i] = float64(i) / 768.0
		v2[i] = float64(i) / 768.0
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cosineSimilarity(v1, v2)
	}
}

// ============================================================================
// Integration Tests
// ============================================================================

func TestNeurosymbolicIntegrationComponent_WithCognitiveChain(t *testing.T) {
	goalStack := NewGoalStack(DefaultGoalStackConfig())
	impasseDetector := NewImpasseDetector(nil, nil)
	workingMemory := NewCognitiveWorkingMemoryComponent(100)

	component := NewNeurosymbolicIntegrationComponent(goalStack, impasseDetector, workingMemory)
	component.Initialize(nil)

	chain := NewCognitiveProcessingChain(
		[]CognitiveComponent{component},
		[]string{"NeurosymbolicIntegration"},
	)

	goal := &Goal{
		ID:       "chain-goal",
		Name:     "Chain Goal",
		Priority: PriorityHigh,
		Status:   GoalActive,
	}

	_ = goalStack.Push(goal)

	request := &CognitiveProcessRequest{
		RequestID:   "chain-req",
		CurrentGoal: goal,
		Timestamp:   time.Now(),
	}

	result, err := chain.Execute(context.Background(), request)

	if err != nil {
		t.Errorf("Chain execution failed: %v", err)
	}

	if result.Status != ProcessSuccess {
		t.Errorf("Chain status should be success")
	}
}

func TestNeurosymbolicIntegrationComponent_HybridDecisionWeighting(t *testing.T) {
	goalStack := NewGoalStack(DefaultGoalStackConfig())
	impasseDetector := NewImpasseDetector(nil, nil)
	workingMemory := NewCognitiveWorkingMemoryComponent(100)

	component := NewNeurosymbolicIntegrationComponent(goalStack, impasseDetector, workingMemory)
	component.Initialize(nil)

	// Test equal weighting
	decision := component.makeHybridDecision(context.Background(), "test", 1.0, 0.0)

	// Hybrid should be 0.5 (equal weights)
	expected := 0.5
	if decision.HybridScore != expected {
		t.Errorf("Expected hybrid %f, got %f", expected, decision.HybridScore)
	}
}
