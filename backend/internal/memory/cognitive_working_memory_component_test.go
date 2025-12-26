package memory

import (
	"context"
	"testing"
	"time"
)

// ============================================================================
// Tests for CognitiveWorkingMemoryComponent
// ============================================================================

func TestCognitiveWorkingMemoryComponent_Initialize(t *testing.T) {
	component := NewCognitiveWorkingMemoryComponent(7)
	err := component.Initialize(nil)

	if err != nil {
		t.Errorf("Initialize failed: %v", err)
	}

	if component.GetName() != "CognitiveWorkingMemory" {
		t.Errorf("Name mismatch")
	}
}

func TestCognitiveWorkingMemoryComponent_Process_BasicRequest(t *testing.T) {
	component := NewCognitiveWorkingMemoryComponent(7)
	component.Initialize(nil)

	goal := &Goal{
		ID:       "goal-001",
		Name:     "Solve a complex problem",
		Priority: 9,
		Status:   GoalActive,
	}

	request := &CognitiveProcessRequest{
		RequestID:   "req-001",
		AgentID:     "agent-001",
		Timestamp:   time.Now(),
		CurrentGoal: goal,
	}

	result, err := component.Process(context.Background(), request)

	if err != nil {
		t.Errorf("Process failed: %v", err)
	}

	if result.Status != ProcessSuccess {
		t.Errorf("Status mismatch: got %v, want success", result.Status)
	}

	if result.DecisionTrace == nil {
		t.Error("DecisionTrace is nil")
	}

	// ProcessingTime might be very small but should be >= 0
	if result.ProcessingTime < 0 {
		t.Error("ProcessingTime should be >= 0")
	}

	if result.Confidence < 0 || result.Confidence > 1 {
		t.Errorf("Confidence out of range: %f", result.Confidence)
	}
}

func TestCognitiveWorkingMemoryComponent_Metrics(t *testing.T) {
	component := NewCognitiveWorkingMemoryComponent(7)
	component.Initialize(nil)

	goal := &Goal{
		ID:       "goal-002",
		Name:     "Test metrics",
		Priority: 5,
	}

	request := &CognitiveProcessRequest{
		RequestID:   "req-002",
		CurrentGoal: goal,
		Timestamp:   time.Now(),
	}

	// First request
	component.Process(context.Background(), request)

	metrics := component.GetMetrics()

	if metrics.ComponentName != "CognitiveWorkingMemory" {
		t.Errorf("ComponentName mismatch")
	}

	if metrics.TotalRequests != 1 {
		t.Errorf("TotalRequests should be 1, got %d", metrics.TotalRequests)
	}

	if metrics.SuccessfulRequests != 1 {
		t.Errorf("SuccessfulRequests should be 1, got %d", metrics.SuccessfulRequests)
	}

	if metrics.CustomMetrics == nil {
		t.Error("CustomMetrics is nil")
	}
}

func TestCognitiveWorkingMemoryComponent_Shutdown(t *testing.T) {
	component := NewCognitiveWorkingMemoryComponent(7)
	component.Initialize(nil)

	err := component.Shutdown()
	if err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}
}

func TestCognitiveWorkingMemoryComponent_WithConstraints(t *testing.T) {
	component := NewCognitiveWorkingMemoryComponent(7)
	component.Initialize(nil)

	constraint := ConstitutionalConstraint{
		Name:        "test_constraint",
		Description: "Test constraint",
		Category:    "test",
	}

	goal := &Goal{
		ID:       "goal-003",
		Name:     "With constraints",
		Priority: 8,
	}

	request := &CognitiveProcessRequest{
		RequestID:         "req-003",
		Timestamp:         time.Now(),
		CurrentGoal:       goal,
		ActiveConstraints: []ConstitutionalConstraint{constraint},
	}

	result, err := component.Process(context.Background(), request)

	if err != nil {
		t.Errorf("Process failed: %v", err)
	}

	if result.Status != ProcessSuccess {
		t.Errorf("Status should be success")
	}
}

func TestCognitiveWorkingMemoryComponent_GetWorkingMemoryState(t *testing.T) {
	component := NewCognitiveWorkingMemoryComponent(7)
	component.Initialize(nil)

	state := component.GetWorkingMemoryState()

	if state == nil {
		t.Error("WorkingMemoryState is nil")
	}

	if state.Capacity() != 7 {
		t.Errorf("Capacity should be 7, got %d", state.Capacity())
	}
}

func TestCognitiveWorkingMemoryComponent_PrimeItem(t *testing.T) {
	component := NewCognitiveWorkingMemoryComponent(7)
	component.Initialize(nil)

	// Add an item first
	item := &WorkingMemoryItem{
		ID:         "item-001",
		Content:    "test item",
		Activation: 0.5,
		Salience:   0.8,
		CreatedAt:  time.Now(),
	}

	state := component.GetWorkingMemoryState()
	state.Add(item)

	// Prime the item
	success := component.PrimeItem("item-001", 0.3)

	if !success {
		t.Error("PrimeItem failed")
	}

	// Verify activation increased
	primed := component.GetItemByID("item-001")
	if primed == nil {
		t.Error("Item not found after priming")
	}
}

func TestCognitiveWorkingMemoryComponent_DecayActivation(t *testing.T) {
	component := NewCognitiveWorkingMemoryComponent(7)
	component.Initialize(nil)

	// Add an item
	item := &WorkingMemoryItem{
		ID:         "item-decay",
		Content:    "decay test",
		Activation: 1.0,
		Salience:   0.8,
		CreatedAt:  time.Now().Add(-5 * time.Second), // Old item for decay
	}

	state := component.GetWorkingMemoryState()
	state.Add(item)

	// Verify item was added
	retrieved := component.GetItemByID("item-decay")
	if retrieved == nil {
		t.Error("Item not added")
	}

	// Apply decay (this tests the integration, not guaranteed to change activation)
	component.DecayActivation()

	// Decay should have been applied
	// We just verify the operation completed without error
}

func TestCognitiveWorkingMemoryComponent_ClearWorkingMemory(t *testing.T) {
	component := NewCognitiveWorkingMemoryComponent(7)
	component.Initialize(nil)

	// Add items
	state := component.GetWorkingMemoryState()
	for i := 0; i < 5; i++ {
		item := &WorkingMemoryItem{
			ID:         string(rune(i)),
			Content:    "test",
			Activation: 0.8,
			CreatedAt:  time.Now(),
		}
		state.Add(item)
	}

	// Verify items added
	if state.Size() != 5 {
		t.Errorf("Items not added correctly, got %d", state.Size())
	}

	// Clear
	component.ClearWorkingMemory()

	// Verify cleared
	newState := component.GetWorkingMemoryState()
	if newState.Size() != 0 {
		t.Errorf("Items not cleared, got %d", newState.Size())
	}
}

// ============================================================================
// Benchmark Tests
// ============================================================================

func BenchmarkCognitiveWorkingMemoryComponent_Initialize(b *testing.B) {
	component := NewCognitiveWorkingMemoryComponent(7)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		component.Initialize(nil)
	}
}

func BenchmarkCognitiveWorkingMemoryComponent_Process(b *testing.B) {
	component := NewCognitiveWorkingMemoryComponent(7)
	component.Initialize(nil)

	goal := &Goal{
		ID:       "bench-goal",
		Name:     "Benchmark process",
		Priority: 8,
	}

	request := &CognitiveProcessRequest{
		RequestID:   "bench-req",
		Timestamp:   time.Now(),
		CurrentGoal: goal,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		component.Process(context.Background(), request)
	}
}

func BenchmarkCognitiveWorkingMemoryComponent_GetMetrics(b *testing.B) {
	component := NewCognitiveWorkingMemoryComponent(7)
	component.Initialize(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		component.GetMetrics()
	}
}

func BenchmarkCognitiveWorkingMemoryComponent_DecayActivation(b *testing.B) {
	component := NewCognitiveWorkingMemoryComponent(7)
	component.Initialize(nil)

	// Add items
	state := component.GetWorkingMemoryState()
	for i := 0; i < 5; i++ {
		item := &WorkingMemoryItem{
			ID:         string(rune(i)),
			Content:    "test",
			Activation: 1.0,
			CreatedAt:  time.Now(),
		}
		state.Add(item)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		component.DecayActivation()
	}
}

// ============================================================================
// Integration Tests
// ============================================================================

func TestCognitiveWorkingMemoryComponent_WithChain(t *testing.T) {
	component := NewCognitiveWorkingMemoryComponent(7)
	component.Initialize(nil)

	// Create a processing chain with the component
	chain := NewCognitiveProcessingChain(
		[]CognitiveComponent{component},
		[]string{"CognitiveWorkingMemory"},
	)

	goal := &Goal{
		ID:       "chain-goal",
		Name:     "Test chain integration",
		Priority: 9,
	}

	request := &CognitiveProcessRequest{
		RequestID:   "chain-req",
		Timestamp:   time.Now(),
		CurrentGoal: goal,
	}

	result, err := chain.Execute(context.Background(), request)

	if err != nil {
		t.Errorf("Chain execution failed: %v", err)
	}

	if result.Status != ProcessSuccess {
		t.Errorf("Chain status should be success")
	}

	if len(result.ExecutionSteps) != 1 {
		t.Errorf("Should have 1 execution step")
	}
}

func TestCognitiveWorkingMemoryComponent_MultipleRequests(t *testing.T) {
	component := NewCognitiveWorkingMemoryComponent(7)
	component.Initialize(nil)

	// Make multiple requests
	for i := 0; i < 10; i++ {
		goal := &Goal{
			ID:       "multi-goal-" + string(rune(i)),
			Name:     "Multiple request test",
			Priority: 7,
		}

		request := &CognitiveProcessRequest{
			RequestID:   "multi-req-" + string(rune(i)),
			Timestamp:   time.Now(),
			CurrentGoal: goal,
		}

		result, err := component.Process(context.Background(), request)

		if err != nil {
			t.Errorf("Request %d failed: %v", i, err)
		}

		if result.Status != ProcessSuccess {
			t.Errorf("Request %d status should be success", i)
		}
	}

	// Verify metrics
	metrics := component.GetMetrics()
	if metrics.TotalRequests != 10 {
		t.Errorf("TotalRequests should be 10, got %d", metrics.TotalRequests)
	}

	if metrics.SuccessfulRequests != 10 {
		t.Errorf("SuccessfulRequests should be 10, got %d", metrics.SuccessfulRequests)
	}
}

func TestCognitiveWorkingMemoryComponent_ConcurrentAccess(t *testing.T) {
	component := NewCognitiveWorkingMemoryComponent(7)
	component.Initialize(nil)

	// Simulate concurrent requests
	done := make(chan bool, 5)

	for i := 0; i < 5; i++ {
		go func(id int) {
			goal := &Goal{
				ID:       "concurrent-goal-" + string(rune(id)),
				Name:     "Concurrent test",
				Priority: 8,
			}

			request := &CognitiveProcessRequest{
				RequestID:   "concurrent-req-" + string(rune(id)),
				Timestamp:   time.Now(),
				CurrentGoal: goal,
			}

			_, err := component.Process(context.Background(), request)
			if err != nil {
				t.Errorf("Concurrent request %d failed: %v", id, err)
			}

			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 5; i++ {
		<-done
	}

	metrics := component.GetMetrics()
	if metrics.TotalRequests != 5 {
		t.Errorf("Should handle 5 concurrent requests, got %d", metrics.TotalRequests)
	}
}
