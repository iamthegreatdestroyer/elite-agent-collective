// Package memory provides the MNEMONIC system for the Elite Agent Collective.
// This file implements comprehensive integration tests for Phase 1.
//
// The integration test suite validates that all cognitive components work
// seamlessly together in the cognitive processing chain, including:
// - Working Memory
// - Goal Stack
// - Impasse Detection
// - Neurosymbolic Integration

package memory

import (
	"context"
	"testing"
	"time"
)

// ============================================================================
// Comprehensive Integration Tests - Phase 1
// ============================================================================

// ============================================================================
// Phase 1 Integration Tests - Simplified
// ============================================================================

// TestPhase1_Integration_WorkingMemory_Neurosymbolic tests WM + Neurosymbolic
func TestPhase1_Integration_WorkingMemory_Neurosymbolic(t *testing.T) {
	workingMemory := NewCognitiveWorkingMemoryComponent(100)
	goalStack := NewGoalStack(DefaultGoalStackConfig())
	impasseConfig := &ImpasseDetectorConfig{
		MaxRetries:             3,
		BackoffBase:            100 * time.Millisecond,
		BackoffMax:             5 * time.Second,
		TimeoutThreshold:       30 * time.Second,
		NoChangeThreshold:      5,
		TieSimilarityThreshold: 0.9,
		MaxActiveImpasses:      10,
	}
	impasseDetector := NewImpasseDetector(impasseConfig, goalStack)
	neurosymbolic := NewNeurosymbolicIntegrationComponent(goalStack, impasseDetector, workingMemory)

	_ = workingMemory.Initialize(nil)
	_ = neurosymbolic.Initialize(nil)

	// Create test goal
	testGoal := &Goal{
		ID:       "phase1-test-goal",
		Name:     "Complete Phase 1 Integration Test",
		Priority: PriorityHigh,
		Status:   GoalActive,
		Progress: 0.0,
	}

	// Execute both components
	request := &CognitiveProcessRequest{
		RequestID:   "phase1-integration-test",
		Timestamp:   time.Now(),
		CurrentGoal: testGoal,
	}

	// Process through working memory
	wmResult, err := workingMemory.Process(context.Background(), request)
	if err != nil {
		t.Errorf("Working memory processing failed: %v", err)
	}

	if wmResult == nil || wmResult.Status != ProcessSuccess {
		t.Error("Working memory should process successfully")
	}

	// Process through neurosymbolic
	request.WorkingMemory = workingMemory.workingMemory
	nsResult, err := neurosymbolic.Process(context.Background(), request)
	if err != nil {
		t.Errorf("Neurosymbolic processing failed: %v", err)
	}

	if nsResult == nil || nsResult.Status != ProcessSuccess {
		t.Error("Neurosymbolic should process successfully")
	}

	// Cleanup
	_ = workingMemory.Shutdown()
	_ = neurosymbolic.Shutdown()
}

// TestPhase1_Integration_GoalStack_Neurosymbolic tests GoalStack + Neurosymbolic
func TestPhase1_Integration_GoalStack_Neurosymbolic(t *testing.T) {
	workingMemory := NewCognitiveWorkingMemoryComponent(100)
	goalStack := NewGoalStack(DefaultGoalStackConfig())
	impasseConfig := &ImpasseDetectorConfig{
		MaxRetries:             3,
		BackoffBase:            100 * time.Millisecond,
		BackoffMax:             5 * time.Second,
		TimeoutThreshold:       30 * time.Second,
		NoChangeThreshold:      5,
		TieSimilarityThreshold: 0.9,
		MaxActiveImpasses:      10,
	}
	impasseDetector := NewImpasseDetector(impasseConfig, goalStack)
	neurosymbolic := NewNeurosymbolicIntegrationComponent(goalStack, impasseDetector, workingMemory)

	_ = neurosymbolic.Initialize(nil)

	// Push goal to goal stack
	goal := &Goal{
		ID:       "integration-goal",
		Name:     "Integration Test Goal",
		Priority: PriorityHigh,
		Status:   GoalActive,
		Progress: 0.5,
	}

	goalStack.Push(goal)

	// Process through neurosymbolic
	request := &CognitiveProcessRequest{
		RequestID:   "integration-test",
		CurrentGoal: goal,
		Timestamp:   time.Now(),
	}

	result, err := neurosymbolic.Process(context.Background(), request)
	if err != nil {
		t.Errorf("Process failed: %v", err)
	}

	if result == nil || result.Status != ProcessSuccess {
		t.Error("Should succeed")
	}

	_ = neurosymbolic.Shutdown()
}

// TestPhase1_Integration_ImpasseDetector_Neurosymbolic tests ImpasseDetector + Neurosymbolic
func TestPhase1_Integration_ImpasseDetector_Neurosymbolic(t *testing.T) {
	workingMemory := NewCognitiveWorkingMemoryComponent(100)
	goalStack := NewGoalStack(DefaultGoalStackConfig())
	impasseConfig := &ImpasseDetectorConfig{
		MaxRetries:             3,
		BackoffBase:            100 * time.Millisecond,
		BackoffMax:             5 * time.Second,
		TimeoutThreshold:       30 * time.Second,
		NoChangeThreshold:      5,
		TieSimilarityThreshold: 0.9,
		MaxActiveImpasses:      10,
	}
	impasseDetector := NewImpasseDetector(impasseConfig, goalStack)
	neurosymbolic := NewNeurosymbolicIntegrationComponent(goalStack, impasseDetector, workingMemory)

	_ = neurosymbolic.Initialize(nil)

	goal := &Goal{
		ID:       "impasse-neuro-test",
		Name:     "Impasse Neurosymbolic Test",
		Priority: PriorityHigh,
		Status:   GoalActive,
		Progress: 0.0,
	}

	goalStack.Push(goal)

	// Make decision
	request := &CognitiveProcessRequest{
		RequestID:   "impasse-neuro",
		CurrentGoal: goal,
		Timestamp:   time.Now(),
	}

	result, err := neurosymbolic.Process(context.Background(), request)
	if err != nil {
		t.Errorf("Process failed: %v", err)
	}

	if result == nil || result.Status != ProcessSuccess {
		t.Error("Should process successfully")
	}

	_ = neurosymbolic.Shutdown()
}

// TestPhase1_Integration_GoalStackComponent tests GoalStackComponent
func TestPhase1_Integration_GoalStackComponent(t *testing.T) {
	goalStackComp := NewCognitiveGoalStackComponent()

	_ = goalStackComp.Initialize(nil)

	goal := &Goal{
		ID:       "component-test-goal",
		Name:     "Component Test",
		Priority: PriorityHigh,
		Status:   GoalActive,
	}

	request := &CognitiveProcessRequest{
		RequestID:   "component-test",
		CurrentGoal: goal,
		Timestamp:   time.Now(),
	}

	result, err := goalStackComp.Process(context.Background(), request)
	if err != nil {
		t.Errorf("Process failed: %v", err)
	}

	if result == nil || result.Status != ProcessSuccess {
		t.Error("Should process successfully")
	}

	_ = goalStackComp.Shutdown()
}

// ============================================================================
// Performance Benchmarks - Phase 1 System Level
// ============================================================================

func BenchmarkPhase1_WorkingMemory_Process(b *testing.B) {
	workingMemory := NewCognitiveWorkingMemoryComponent(100)
	_ = workingMemory.Initialize(nil)

	goal := &Goal{
		ID:       "bench-goal",
		Name:     "Benchmark Goal",
		Priority: PriorityHigh,
		Status:   GoalActive,
		Progress: 0.5,
	}

	request := &CognitiveProcessRequest{
		RequestID:   "bench-req",
		CurrentGoal: goal,
		Timestamp:   time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		workingMemory.Process(context.Background(), request)
	}
}

func BenchmarkPhase1_Neurosymbolic_Process(b *testing.B) {
	workingMemory := NewCognitiveWorkingMemoryComponent(100)
	goalStack := NewGoalStack(DefaultGoalStackConfig())
	impasseConfig := &ImpasseDetectorConfig{
		MaxRetries:             3,
		BackoffBase:            100 * time.Millisecond,
		BackoffMax:             5 * time.Second,
		TimeoutThreshold:       30 * time.Second,
		NoChangeThreshold:      5,
		TieSimilarityThreshold: 0.9,
		MaxActiveImpasses:      10,
	}
	impasseDetector := NewImpasseDetector(impasseConfig, goalStack)
	neurosymbolic := NewNeurosymbolicIntegrationComponent(goalStack, impasseDetector, workingMemory)

	_ = neurosymbolic.Initialize(nil)

	goal := &Goal{
		ID:       "bench-goal",
		Name:     "Benchmark Goal",
		Priority: PriorityHigh,
		Status:   GoalActive,
		Progress: 0.5,
	}

	goalStack.Push(goal)

	request := &CognitiveProcessRequest{
		RequestID:   "bench-req",
		CurrentGoal: goal,
		Timestamp:   time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		neurosymbolic.Process(context.Background(), request)
	}
}
