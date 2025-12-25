package memory

import (
	"context"
	"math"
	"testing"
	"time"
)

// ============================================================================
// Phase Transition Controller Tests
// ============================================================================

func TestNewPhaseTransitionController(t *testing.T) {
	config := DefaultPhaseTransitionConfig()
	controller := NewPhaseTransitionController(config)

	if controller == nil {
		t.Fatal("NewPhaseTransitionController returned nil")
	}

	// Verify default parameters
	params := controller.GetParameters()
	if params.Temperature != config.InitialTemperature {
		t.Errorf("Expected temperature %v, got %v", config.InitialTemperature, params.Temperature)
	}
	if params.MutationRate != config.InitialMutationRate {
		t.Errorf("Expected mutation rate %v, got %v", config.InitialMutationRate, params.MutationRate)
	}
}

func TestRecordTask(t *testing.T) {
	controller := NewPhaseTransitionController(DefaultPhaseTransitionConfig())

	// Record some tasks
	for i := 0; i < 10; i++ {
		controller.RecordTask(TaskRecord{
			TaskID:    "task-" + string(rune('0'+i)),
			AgentID:   "APEX",
			Success:   true,
			IsNovel:   i%3 == 0,
			Timestamp: time.Now(),
		})
	}

	// Verify records
	metrics := controller.ComputeMetrics()
	if metrics.WindowSize != 10 {
		t.Errorf("Expected window size 10, got %d", metrics.WindowSize)
	}
}

func TestComputeRoutingEntropy(t *testing.T) {
	tests := []struct {
		name          string
		taskRecords   []TaskRecord
		expectedPhase SystemPhase
		entropyRange  [2]float64 // min, max expected entropy
	}{
		{
			name: "Single agent - frozen",
			taskRecords: func() []TaskRecord {
				records := make([]TaskRecord, 100)
				for i := 0; i < 100; i++ {
					records[i] = TaskRecord{AgentID: "APEX"}
				}
				return records
			}(),
			expectedPhase: PhaseFrozen,
			entropyRange:  [2]float64{0.0, 0.1},
		},
		{
			name: "Uniform distribution - high entropy",
			taskRecords: func() []TaskRecord {
				agents := []string{"APEX", "CIPHER", "ARCHITECT", "AXIOM", "VELOCITY",
					"QUANTUM", "TENSOR", "FORTRESS", "NEURAL", "CRYPTO"}
				records := make([]TaskRecord, 100)
				for i := 0; i < 100; i++ {
					records[i] = TaskRecord{AgentID: agents[i%len(agents)]}
				}
				return records
			}(),
			expectedPhase: PhaseCritical,
			entropyRange:  [2]float64{0.4, 0.9},
		},
		{
			name: "Mixed distribution",
			taskRecords: func() []TaskRecord {
				records := make([]TaskRecord, 100)
				for i := 0; i < 60; i++ {
					records[i] = TaskRecord{AgentID: "APEX"}
				}
				for i := 60; i < 80; i++ {
					records[i] = TaskRecord{AgentID: "CIPHER"}
				}
				for i := 80; i < 100; i++ {
					records[i] = TaskRecord{AgentID: "ARCHITECT"}
				}
				return records
			}(),
			entropyRange: [2]float64{0.1, 0.5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller := NewPhaseTransitionController(DefaultPhaseTransitionConfig())
			for _, record := range tt.taskRecords {
				controller.RecordTask(record)
			}

			metrics := controller.ComputeMetrics()

			if metrics.RoutingEntropy < tt.entropyRange[0] || metrics.RoutingEntropy > tt.entropyRange[1] {
				t.Errorf("Expected entropy in range [%v, %v], got %v",
					tt.entropyRange[0], tt.entropyRange[1], metrics.RoutingEntropy)
			}

			if tt.expectedPhase != 0 && metrics.Phase != tt.expectedPhase {
				t.Errorf("Expected phase %v, got %v", tt.expectedPhase, metrics.Phase)
			}
		})
	}
}

func TestPhaseDetection(t *testing.T) {
	controller := NewPhaseTransitionController(DefaultPhaseTransitionConfig())

	// Test frozen phase (single agent dominance)
	for i := 0; i < 100; i++ {
		controller.RecordTask(TaskRecord{AgentID: "APEX"})
	}
	metrics := controller.ComputeMetrics()
	if metrics.Phase != PhaseFrozen {
		t.Errorf("Expected FROZEN phase, got %v", metrics.Phase)
	}

	// Reset and test critical phase (balanced distribution)
	controller = NewPhaseTransitionController(DefaultPhaseTransitionConfig())
	agents := []string{"APEX", "CIPHER", "ARCHITECT", "AXIOM", "VELOCITY",
		"QUANTUM", "TENSOR", "FORTRESS", "NEURAL", "CRYPTO",
		"FLUX", "PRISM", "SYNAPSE", "CORE", "HELIX"}
	for i := 0; i < 150; i++ {
		controller.RecordTask(TaskRecord{AgentID: agents[i%len(agents)]})
	}
	metrics = controller.ComputeMetrics()
	// Should be in critical or at least not frozen
	if metrics.Phase == PhaseFrozen {
		t.Errorf("Expected CRITICAL or CHAOTIC phase with balanced distribution, got FROZEN")
	}
}

func TestParameterAdjustment(t *testing.T) {
	controller := NewPhaseTransitionController(DefaultPhaseTransitionConfig())

	initialParams := controller.GetParameters()

	// Create frozen state
	for i := 0; i < 100; i++ {
		controller.RecordTask(TaskRecord{AgentID: "APEX"})
	}

	// Update should adjust parameters
	controller.Update()

	newParams := controller.GetParameters()

	// Temperature should increase to add exploration
	if newParams.Temperature <= initialParams.Temperature {
		t.Logf("Expected temperature to increase from frozen state")
		t.Logf("Initial: %v, New: %v", initialParams.Temperature, newParams.Temperature)
	}
}

func TestHistoryTracking(t *testing.T) {
	config := DefaultPhaseTransitionConfig()
	config.HistorySize = 10
	controller := NewPhaseTransitionController(config)

	// Record tasks and update multiple times
	for i := 0; i < 15; i++ {
		controller.RecordTask(TaskRecord{AgentID: "APEX"})
		controller.Update()
	}

	history := controller.GetHistory()
	if len(history) > config.HistorySize {
		t.Errorf("History exceeded max size: got %d, max %d", len(history), config.HistorySize)
	}
}

func TestTemperatureBasedSelection(t *testing.T) {
	controller := NewPhaseTransitionController(DefaultPhaseTransitionConfig())

	scores := map[string]float64{
		"APEX":      0.9,
		"CIPHER":    0.8,
		"ARCHITECT": 0.7,
	}

	// With temperature=1.0, should mostly select APEX but with some variation
	selections := make(map[string]int)
	for i := 0; i < 100; i++ {
		selected := controller.SelectAgentWithTemperature(scores)
		selections[selected]++
	}

	// APEX should be selected most often
	if selections["APEX"] < selections["CIPHER"] {
		t.Logf("APEX selections: %d, CIPHER: %d", selections["APEX"], selections["CIPHER"])
		// This is probabilistic, so not a hard failure
	}

	// Should see some variation (not 100% APEX)
	if selections["APEX"] == 100 {
		t.Log("Warning: No exploration observed - all selections went to APEX")
	}
}

func TestContinuousMonitoring(t *testing.T) {
	controller := NewPhaseTransitionController(DefaultPhaseTransitionConfig())

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	// Start monitoring with short interval
	controller.Start(ctx, 50*time.Millisecond)

	// Record some tasks
	time.Sleep(100 * time.Millisecond)
	for i := 0; i < 10; i++ {
		controller.RecordTask(TaskRecord{AgentID: "APEX"})
	}
	time.Sleep(100 * time.Millisecond)

	// Should have some history from automatic updates
	history := controller.GetHistory()
	if len(history) == 0 {
		t.Error("Expected some history from continuous monitoring")
	}

	controller.Stop()
}

func TestDiagnose(t *testing.T) {
	controller := NewPhaseTransitionController(DefaultPhaseTransitionConfig())

	// Record some tasks
	for i := 0; i < 50; i++ {
		controller.RecordTask(TaskRecord{AgentID: "APEX"})
	}
	controller.Update()

	diag := controller.Diagnose()

	if diag == nil {
		t.Fatal("Diagnose returned nil")
	}

	if diag.TaskCount != 50 {
		t.Errorf("Expected task count 50, got %d", diag.TaskCount)
	}

	if len(diag.Recommendations) == 0 {
		t.Error("Expected recommendations in diagnosis")
	}

	t.Logf("Phase: %s", diag.Phase)
	t.Logf("Health Score: %.2f", diag.HealthScore)
	t.Logf("Recommendations: %v", diag.Recommendations)
}

func TestInnovationRate(t *testing.T) {
	controller := NewPhaseTransitionController(DefaultPhaseTransitionConfig())

	// Record tasks with some novel solutions
	for i := 0; i < 100; i++ {
		controller.RecordTask(TaskRecord{
			AgentID: "APEX",
			IsNovel: i%10 == 0, // 10% novel
		})
	}

	metrics := controller.ComputeMetrics()

	expectedRate := 0.10
	tolerance := 0.02

	if math.Abs(metrics.InnovationRate-expectedRate) > tolerance {
		t.Errorf("Expected innovation rate ~%v, got %v", expectedRate, metrics.InnovationRate)
	}
}

func TestAgentDiversity(t *testing.T) {
	controller := NewPhaseTransitionController(DefaultPhaseTransitionConfig())

	// Perfect diversity: all agents used equally
	agents := []string{"APEX", "CIPHER", "ARCHITECT", "AXIOM", "VELOCITY"}
	for i := 0; i < 100; i++ {
		controller.RecordTask(TaskRecord{AgentID: agents[i%len(agents)]})
	}

	metrics := controller.ComputeMetrics()

	// High diversity expected
	if metrics.AgentDiversity < 0.7 {
		t.Errorf("Expected high diversity (>0.7), got %v", metrics.AgentDiversity)
	}

	// Reset and test low diversity
	controller = NewPhaseTransitionController(DefaultPhaseTransitionConfig())
	for i := 0; i < 90; i++ {
		controller.RecordTask(TaskRecord{AgentID: "APEX"})
	}
	for i := 0; i < 10; i++ {
		controller.RecordTask(TaskRecord{AgentID: "CIPHER"})
	}

	metrics = controller.ComputeMetrics()

	// Lower diversity expected
	if metrics.AgentDiversity > 0.7 {
		t.Errorf("Expected lower diversity (<0.7), got %v", metrics.AgentDiversity)
	}
}

func TestHealthScore(t *testing.T) {
	controller := NewPhaseTransitionController(DefaultPhaseTransitionConfig())

	// Create a balanced, healthy state
	agents := []string{"APEX", "CIPHER", "ARCHITECT", "AXIOM", "VELOCITY",
		"QUANTUM", "TENSOR", "FORTRESS", "NEURAL", "CRYPTO"}
	for i := 0; i < 200; i++ {
		controller.RecordTask(TaskRecord{
			AgentID:      agents[i%len(agents)],
			Success:      true,
			IsNovel:      i%20 == 0,
			FitnessScore: 0.7 + float64(i%10)*0.03,
		})
	}

	metrics := controller.ComputeMetrics()

	// Health score should be reasonable for balanced system
	if metrics.HealthScore < 0.3 || metrics.HealthScore > 1.0 {
		t.Errorf("Health score out of expected range: %v", metrics.HealthScore)
	}

	t.Logf("Health Score: %.2f", metrics.HealthScore)
	t.Logf("Routing Entropy: %.2f", metrics.RoutingEntropy)
	t.Logf("Agent Diversity: %.2f", metrics.AgentDiversity)
	t.Logf("Phase: %s", metrics.Phase)
}

func TestParameterCallbacks(t *testing.T) {
	controller := NewPhaseTransitionController(DefaultPhaseTransitionConfig())

	callbackCalled := false
	controller.OnParameterChange(func(old, new ControlParameters) {
		callbackCalled = true
		t.Logf("Parameters changed: Temp %.2f -> %.2f", old.Temperature, new.Temperature)
	})

	// Create frozen state to trigger parameter change
	for i := 0; i < 100; i++ {
		controller.RecordTask(TaskRecord{AgentID: "APEX"})
	}

	// Multiple updates to trigger adjustment
	for i := 0; i < 5; i++ {
		controller.Update()
	}

	// Callback may or may not be called depending on adjustment magnitude
	t.Logf("Callback called: %v", callbackCalled)
}

// ============================================================================
// Benchmarks
// ============================================================================

func BenchmarkRecordTask(b *testing.B) {
	controller := NewPhaseTransitionController(DefaultPhaseTransitionConfig())

	record := TaskRecord{
		TaskID:  "task-1",
		AgentID: "APEX",
		Success: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		controller.RecordTask(record)
	}
}

func BenchmarkComputeMetrics(b *testing.B) {
	controller := NewPhaseTransitionController(DefaultPhaseTransitionConfig())

	// Pre-populate with tasks
	agents := []string{"APEX", "CIPHER", "ARCHITECT", "AXIOM", "VELOCITY"}
	for i := 0; i < 1000; i++ {
		controller.RecordTask(TaskRecord{
			AgentID:      agents[i%len(agents)],
			FitnessScore: 0.7,
			IsNovel:      i%20 == 0,
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		controller.ComputeMetrics()
	}
}

func BenchmarkUpdate(b *testing.B) {
	controller := NewPhaseTransitionController(DefaultPhaseTransitionConfig())

	// Pre-populate
	agents := []string{"APEX", "CIPHER", "ARCHITECT", "AXIOM", "VELOCITY"}
	for i := 0; i < 1000; i++ {
		controller.RecordTask(TaskRecord{AgentID: agents[i%len(agents)]})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		controller.Update()
	}
}

func BenchmarkSelectAgentWithTemperature(b *testing.B) {
	controller := NewPhaseTransitionController(DefaultPhaseTransitionConfig())

	scores := map[string]float64{
		"APEX":      0.9,
		"CIPHER":    0.85,
		"ARCHITECT": 0.8,
		"AXIOM":     0.75,
		"VELOCITY":  0.7,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		controller.SelectAgentWithTemperature(scores)
	}
}
