package memory

import (
	"context"
	"testing"
	"time"
)

// ============================================================================
// Unit Tests for CognitiveGoalStackComponent
// ============================================================================

func TestCognitiveGoalStackComponent_Initialize(t *testing.T) {
	component := NewCognitiveGoalStackComponent()
	err := component.Initialize(nil)

	if err != nil {
		t.Errorf("Initialize failed: %v", err)
	}

	if component.GetName() != "CognitiveGoalStackManagement" {
		t.Errorf("Name mismatch")
	}
}

func TestCognitiveGoalStackComponent_Process_BasicRequest(t *testing.T) {
	component := NewCognitiveGoalStackComponent()
	component.Initialize(nil)

	goal := &Goal{
		ID:       "goal-001",
		Name:     "Test Goal",
		Priority: PriorityHigh,
		Status:   GoalPending,
	}

	request := &CognitiveProcessRequest{
		RequestID:   "req-001",
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

	if result.DecisionTrace == nil {
		t.Error("DecisionTrace is nil")
	}

	if result.Confidence < 0 || result.Confidence > 1 {
		t.Errorf("Confidence out of range: %f", result.Confidence)
	}
}

func TestCognitiveGoalStackComponent_CompleteGoal(t *testing.T) {
	component := NewCognitiveGoalStackComponent()
	component.Initialize(nil)

	goal := &Goal{
		ID:       "goal-complete",
		Name:     "Complete Goal",
		Priority: PriorityHigh,
	}

	request := &CognitiveProcessRequest{
		RequestID:   "req-complete",
		CurrentGoal: goal,
		Timestamp:   time.Now(),
	}

	// Process goal (activate it)
	component.Process(context.Background(), request)

	// Complete the goal
	success := component.CompleteGoal("goal-complete")

	if !success {
		t.Error("CompleteGoal failed")
	}

	// Verify goal is in completed list
	completed := component.GetCompletedGoals()
	if len(completed) == 0 {
		t.Error("Goal not in completed goals")
	}
}

func TestCognitiveGoalStackComponent_FailGoal(t *testing.T) {
	component := NewCognitiveGoalStackComponent()
	component.Initialize(nil)

	goal := &Goal{
		ID:       "goal-fail",
		Name:     "Fail Goal",
		Priority: PriorityNormal,
	}

	request := &CognitiveProcessRequest{
		RequestID:   "req-fail",
		CurrentGoal: goal,
		Timestamp:   time.Now(),
	}

	component.Process(context.Background(), request)

	// Fail the goal
	success := component.FailGoal("goal-fail", "Test failure")

	if !success {
		t.Error("FailGoal failed")
	}

	failed := component.GetGoalByID("goal-fail")
	if failed.Status != GoalFailed {
		t.Errorf("Goal status should be failed, got %v", failed.Status)
	}
}

func TestCognitiveGoalStackComponent_SuspendAndResume(t *testing.T) {
	component := NewCognitiveGoalStackComponent()
	component.Initialize(nil)

	goal := &Goal{
		ID:       "goal-suspend",
		Name:     "Suspend Goal",
		Priority: PriorityNormal,
	}

	request := &CognitiveProcessRequest{
		RequestID:   "req-suspend",
		CurrentGoal: goal,
		Timestamp:   time.Now(),
	}

	component.Process(context.Background(), request)

	// Suspend the goal
	success := component.SuspendGoal("goal-suspend", "Testing suspension")

	if !success {
		t.Error("SuspendGoal failed")
	}

	suspended := component.GetGoalByID("goal-suspend")
	if suspended.Status != GoalSuspended {
		t.Errorf("Goal should be suspended, got %v", suspended.Status)
	}

	// Resume the goal
	success = component.ResumeGoal("goal-suspend")

	if !success {
		t.Error("ResumeGoal failed")
	}

	resumed := component.GetGoalByID("goal-suspend")
	// After resume, the goal goes back to PENDING or ACTIVE state
	if resumed.Status == GoalSuspended {
		t.Errorf("Goal should not be suspended after resume")
	}
}

func TestCognitiveGoalStackComponent_UpdateProgress(t *testing.T) {
	component := NewCognitiveGoalStackComponent()
	component.Initialize(nil)

	goal := &Goal{
		ID:       "goal-progress",
		Name:     "Progress Goal",
		Priority: PriorityHigh,
	}

	request := &CognitiveProcessRequest{
		RequestID:   "req-progress",
		CurrentGoal: goal,
		Timestamp:   time.Now(),
	}

	component.Process(context.Background(), request)

	// Update progress
	success := component.UpdateGoalProgress("goal-progress", 0.5)

	if !success {
		t.Error("UpdateGoalProgress failed")
	}

	updated := component.GetGoalByID("goal-progress")
	if updated.Progress != 0.5 {
		t.Errorf("Progress should be 0.5, got %f", updated.Progress)
	}

	// Update to 1.5 (should be clamped to 1.0)
	component.UpdateGoalProgress("goal-progress", 1.5)
	updated = component.GetGoalByID("goal-progress")
	if updated.Progress != 1.0 {
		t.Errorf("Progress should be clamped to 1.0, got %f", updated.Progress)
	}
}

func TestCognitiveGoalStackComponent_DecomposeGoal(t *testing.T) {
	component := NewCognitiveGoalStackComponent()
	component.Initialize(nil)

	parent := &Goal{
		ID:       "parent-goal",
		Name:     "Parent Goal",
		Priority: PriorityCritical,
	}

	request := &CognitiveProcessRequest{
		RequestID:   "req-parent",
		CurrentGoal: parent,
		Timestamp:   time.Now(),
	}

	component.Process(context.Background(), request)

	// Create subgoals
	subgoal1 := &Goal{
		ID:       "subgoal-1",
		Name:     "Subgoal 1",
		Priority: PriorityHigh,
	}

	subgoal2 := &Goal{
		ID:       "subgoal-2",
		Name:     "Subgoal 2",
		Priority: PriorityNormal,
	}

	// Decompose parent into subgoals
	success := component.DecomposeGoal("parent-goal", []*Goal{subgoal1, subgoal2})

	if !success {
		t.Error("DecomposeGoal failed")
	}

	// Verify parent has subgoals
	updatedParent := component.GetGoalByID("parent-goal")
	if len(updatedParent.SubGoalIDs) != 2 {
		t.Errorf("Parent should have 2 subgoals, got %d", len(updatedParent.SubGoalIDs))
	}
}

func TestCognitiveGoalStackComponent_Metrics(t *testing.T) {
	component := NewCognitiveGoalStackComponent()
	component.Initialize(nil)

	goal := &Goal{
		ID:       "goal-metrics",
		Name:     "Metrics Goal",
		Priority: PriorityNormal,
	}

	request := &CognitiveProcessRequest{
		RequestID:   "req-metrics",
		CurrentGoal: goal,
		Timestamp:   time.Now(),
	}

	component.Process(context.Background(), request)

	metrics := component.GetMetrics()

	if metrics.ComponentName != "CognitiveGoalStackManagement" {
		t.Errorf("Component name mismatch")
	}

	if metrics.TotalRequests != 1 {
		t.Errorf("Total requests should be 1, got %d", metrics.TotalRequests)
	}

	if metrics.SuccessfulRequests != 1 {
		t.Errorf("Successful requests should be 1")
	}

	if metrics.CustomMetrics == nil {
		t.Error("Custom metrics is nil")
	}
}

func TestCognitiveGoalStackComponent_Shutdown(t *testing.T) {
	component := NewCognitiveGoalStackComponent()
	component.Initialize(nil)

	err := component.Shutdown()

	if err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}
}

// ============================================================================
// Benchmark Tests
// ============================================================================

func BenchmarkCognitiveGoalStackComponent_Initialize(b *testing.B) {
	component := NewCognitiveGoalStackComponent()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		component.Initialize(nil)
	}
}

func BenchmarkCognitiveGoalStackComponent_Process(b *testing.B) {
	component := NewCognitiveGoalStackComponent()
	component.Initialize(nil)

	goal := &Goal{
		ID:       "bench-goal",
		Name:     "Benchmark Goal",
		Priority: PriorityHigh,
	}

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

func BenchmarkCognitiveGoalStackComponent_CompleteGoal(b *testing.B) {
	component := NewCognitiveGoalStackComponent()
	component.Initialize(nil)

	goal := &Goal{
		ID:       "bench-complete",
		Name:     "Benchmark Complete",
		Priority: PriorityHigh,
	}

	request := &CognitiveProcessRequest{
		RequestID:   "bench-complete-req",
		CurrentGoal: goal,
		Timestamp:   time.Now(),
	}

	component.Process(context.Background(), request)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		component.CompleteGoal("bench-complete")
	}
}

// ============================================================================
// Integration Tests
// ============================================================================

func TestCognitiveGoalStackComponent_WithChain(t *testing.T) {
	component := NewCognitiveGoalStackComponent()
	component.Initialize(nil)

	chain := NewCognitiveProcessingChain(
		[]CognitiveComponent{component},
		[]string{"CognitiveGoalStackManagement"},
	)

	goal := &Goal{
		ID:       "chain-goal",
		Name:     "Chain Goal",
		Priority: PriorityCritical,
	}

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

func TestCognitiveGoalStackComponent_MultipleGoals(t *testing.T) {
	component := NewCognitiveGoalStackComponent()
	component.Initialize(nil)

	// Process multiple goals
	for i := 0; i < 5; i++ {
		priority := GoalPriority(PriorityHigh - GoalPriority(i))
		goal := &Goal{
			ID:       string(rune('A' + rune(i))),
			Name:     "Goal " + string(rune('A'+rune(i))),
			Priority: priority,
		}

		request := &CognitiveProcessRequest{
			RequestID:   "req-" + string(rune('A'+rune(i))),
			CurrentGoal: goal,
			Timestamp:   time.Now(),
		}

		_, err := component.Process(context.Background(), request)
		if err != nil {
			t.Errorf("Process failed for goal %d: %v", i, err)
		}
	}

	// Verify metrics
	metrics := component.GetMetrics()
	if metrics.TotalRequests != 5 {
		t.Errorf("Total requests should be 5, got %d", metrics.TotalRequests)
	}

	// Complete some goals
	component.CompleteGoal(string(rune('A')))
	component.CompleteGoal(string(rune('B')))

	metrics = component.GetMetrics()
	completedCount := len(component.GetCompletedGoals())
	if completedCount != 2 {
		t.Errorf("Should have 2 completed goals, got %d", completedCount)
	}
}

func TestCognitiveGoalStackComponent_ConcurrentAccess(t *testing.T) {
	component := NewCognitiveGoalStackComponent()
	component.Initialize(nil)

	done := make(chan bool, 5)

	// Concurrent processing
	for i := 0; i < 5; i++ {
		go func(id int) {
			goal := &Goal{
				ID:       string(rune('X' + rune(id))),
				Name:     "Concurrent Goal",
				Priority: PriorityHigh,
			}

			request := &CognitiveProcessRequest{
				RequestID:   "concurrent-req-" + string(rune('X'+rune(id))),
				CurrentGoal: goal,
				Timestamp:   time.Now(),
			}

			_, err := component.Process(context.Background(), request)
			if err != nil {
				t.Errorf("Concurrent process failed: %v", err)
			}

			done <- true
		}(i)
	}

	// Wait for completion
	for i := 0; i < 5; i++ {
		<-done
	}

	metrics := component.GetMetrics()
	if metrics.TotalRequests != 5 {
		t.Errorf("Should have processed 5 requests, got %d", metrics.TotalRequests)
	}
}

func TestCognitiveGoalStackComponent_ClearGoalStack(t *testing.T) {
	component := NewCognitiveGoalStackComponent()
	component.Initialize(nil)

	// Add multiple goals
	for i := 0; i < 3; i++ {
		goal := &Goal{
			ID:       string(rune('1' + rune(i))),
			Name:     "Goal " + string(rune('1'+rune(i))),
			Priority: PriorityNormal,
		}

		request := &CognitiveProcessRequest{
			RequestID:   "req-" + string(rune('1'+rune(i))),
			CurrentGoal: goal,
			Timestamp:   time.Now(),
		}

		component.Process(context.Background(), request)
	}

	// Verify goals added
	stack := component.GetActiveGoalStack()
	if len(stack) != 3 {
		t.Errorf("Should have 3 active goals before clear")
	}

	// Clear
	component.ClearGoalStack()

	// Verify cleared
	stack = component.GetActiveGoalStack()
	if len(stack) != 0 {
		t.Errorf("Stack should be empty after clear")
	}
}

func TestCognitiveGoalStackComponent_GetActiveGoalStack(t *testing.T) {
	component := NewCognitiveGoalStackComponent()
	component.Initialize(nil)

	goal1 := &Goal{
		ID:       "goal-1",
		Name:     "Goal 1",
		Priority: PriorityCritical,
	}

	goal2 := &Goal{
		ID:       "goal-2",
		Name:     "Goal 2",
		Priority: PriorityHigh,
	}

	request := &CognitiveProcessRequest{
		RequestID:   "req-1",
		CurrentGoal: goal1,
		Timestamp:   time.Now(),
	}

	component.Process(context.Background(), request)

	request.CurrentGoal = goal2
	request.RequestID = "req-2"
	component.Process(context.Background(), request)

	activeStack := component.GetActiveGoalStack()
	if len(activeStack) != 2 {
		t.Errorf("Should have 2 active goals, got %d", len(activeStack))
	}

	// Verify sorted by priority (highest first)
	if activeStack[0].Priority < activeStack[1].Priority {
		t.Error("Goals should be sorted by priority descending")
	}
}
