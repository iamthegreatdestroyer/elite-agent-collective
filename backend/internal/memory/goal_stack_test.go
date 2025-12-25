package memory

import (
	"testing"
)

// ============================================================================
// Goal Stack Tests
// ============================================================================

func TestNewGoalStack(t *testing.T) {
	config := DefaultGoalStackConfig()
	gs := NewGoalStack(config)

	if gs == nil {
		t.Fatal("NewGoalStack returned nil")
	}

	if gs.Size() != 0 {
		t.Errorf("Expected size 0, got %d", gs.Size())
	}

	if !gs.IsEmpty() {
		t.Error("New goal stack should be empty")
	}
}

func TestGoalStack_PushAndPop(t *testing.T) {
	gs := NewGoalStack(DefaultGoalStackConfig())

	goal := &Goal{
		ID:          "goal-1",
		Name:        "Test Goal",
		Description: "A test goal",
		Priority:    PriorityNormal,
	}

	err := gs.Push(goal)
	if err != nil {
		t.Fatalf("Push failed: %v", err)
	}

	if gs.Size() != 1 {
		t.Errorf("Expected size 1, got %d", gs.Size())
	}

	// Goal should be automatically activated
	current := gs.Current()
	if current == nil {
		t.Fatal("Current goal should not be nil")
	}
	if current.ID != "goal-1" {
		t.Errorf("Expected current goal 'goal-1', got '%s'", current.ID)
	}

	// Pop
	popped, err := gs.Pop()
	if err != nil {
		t.Fatalf("Pop failed: %v", err)
	}
	if popped.ID != "goal-1" {
		t.Errorf("Expected popped goal 'goal-1', got '%s'", popped.ID)
	}

	if !gs.IsEmpty() {
		t.Error("Goal stack should be empty after pop")
	}
}

func TestGoalStack_PriorityOrdering(t *testing.T) {
	gs := NewGoalStack(DefaultGoalStackConfig())

	// Add goals with different priorities
	gs.Push(&Goal{ID: "low", Name: "Low Priority", Priority: PriorityLow})
	gs.Push(&Goal{ID: "critical", Name: "Critical", Priority: PriorityCritical})
	gs.Push(&Goal{ID: "normal", Name: "Normal", Priority: PriorityNormal})
	gs.Push(&Goal{ID: "high", Name: "High", Priority: PriorityHigh})

	// Peek should return highest priority
	top, _ := gs.Peek()
	if top.ID != "critical" {
		t.Errorf("Expected 'critical' at top, got '%s'", top.ID)
	}

	// Pop order should be by priority
	expected := []string{"critical", "high", "normal", "low"}
	for _, expID := range expected {
		goal, _ := gs.Pop()
		if goal.ID != expID {
			t.Errorf("Expected '%s', got '%s'", expID, goal.ID)
		}
	}
}

func TestGoalStack_Complete(t *testing.T) {
	gs := NewGoalStack(DefaultGoalStackConfig())

	gs.Push(&Goal{ID: "goal-1", Name: "Test"})

	err := gs.Complete("goal-1")
	if err != nil {
		t.Fatalf("Complete failed: %v", err)
	}

	// Goal should be removed from active
	if gs.Size() != 0 {
		t.Errorf("Expected 0 active goals, got %d", gs.Size())
	}

	// Goal should be in completed
	goal, err := gs.Get("goal-1")
	if err != nil {
		t.Fatalf("Should be able to get completed goal: %v", err)
	}
	if goal.Status != GoalCompleted {
		t.Errorf("Expected COMPLETED status, got %s", goal.Status)
	}
	if goal.Progress != 1.0 {
		t.Errorf("Expected progress 1.0, got %f", goal.Progress)
	}

	// Stats should reflect completion
	stats := gs.GetStats()
	if stats.TotalGoalsCompleted != 1 {
		t.Errorf("Expected 1 completed, got %d", stats.TotalGoalsCompleted)
	}
}

func TestGoalStack_Fail(t *testing.T) {
	gs := NewGoalStack(DefaultGoalStackConfig())

	gs.Push(&Goal{ID: "goal-1", Name: "Test"})

	err := gs.Fail("goal-1", "test failure reason")
	if err != nil {
		t.Fatalf("Fail failed: %v", err)
	}

	goal, _ := gs.Get("goal-1")
	if goal.Status != GoalFailed {
		t.Errorf("Expected FAILED status, got %s", goal.Status)
	}
	if goal.FailureReason != "test failure reason" {
		t.Errorf("Failure reason mismatch")
	}

	stats := gs.GetStats()
	if stats.TotalGoalsFailed != 1 {
		t.Errorf("Expected 1 failed, got %d", stats.TotalGoalsFailed)
	}
}

func TestGoalStack_SuspendAndResume(t *testing.T) {
	gs := NewGoalStack(DefaultGoalStackConfig())

	gs.Push(&Goal{ID: "goal-1", Name: "Test"})

	// Suspend
	err := gs.Suspend("goal-1", "waiting for resource")
	if err != nil {
		t.Fatalf("Suspend failed: %v", err)
	}

	goal, _ := gs.Get("goal-1")
	if goal.Status != GoalSuspended {
		t.Errorf("Expected SUSPENDED status, got %s", goal.Status)
	}
	if goal.SuspensionReason != "waiting for resource" {
		t.Error("Suspension reason mismatch")
	}

	// Should be no active goals
	if gs.Size() != 0 {
		t.Errorf("Expected 0 active goals, got %d", gs.Size())
	}

	// Resume
	err = gs.Resume("goal-1")
	if err != nil {
		t.Fatalf("Resume failed: %v", err)
	}

	goal, _ = gs.Get("goal-1")
	if goal.Status != GoalPending {
		t.Errorf("Expected PENDING status after resume, got %s", goal.Status)
	}

	if gs.Size() != 1 {
		t.Errorf("Expected 1 active goal after resume, got %d", gs.Size())
	}
}

func TestGoalStack_Decompose(t *testing.T) {
	gs := NewGoalStack(DefaultGoalStackConfig())

	parent := &Goal{ID: "parent", Name: "Parent Goal", Priority: PriorityNormal}
	gs.Push(parent)

	// Decompose into subgoals
	subgoals := []*Goal{
		{ID: "sub-1", Name: "Subgoal 1"},
		{ID: "sub-2", Name: "Subgoal 2"},
		{ID: "sub-3", Name: "Subgoal 3"},
	}

	err := gs.Decompose("parent", subgoals)
	if err != nil {
		t.Fatalf("Decompose failed: %v", err)
	}

	// Parent should be suspended (decomposed)
	parentGoal, _ := gs.Get("parent")
	if parentGoal.Status != GoalSuspended {
		t.Errorf("Expected parent SUSPENDED, got %s", parentGoal.Status)
	}
	if len(parentGoal.SubGoalIDs) != 3 {
		t.Errorf("Expected 3 subgoal IDs, got %d", len(parentGoal.SubGoalIDs))
	}

	// Subgoals should be in active queue
	if gs.Size() != 3 {
		t.Errorf("Expected 3 active subgoals, got %d", gs.Size())
	}

	// Subgoals should have parent reference and depth
	sub1, _ := gs.Get("sub-1")
	if sub1.ParentID != "parent" {
		t.Errorf("Expected parent ID 'parent', got '%s'", sub1.ParentID)
	}
	if sub1.Depth != 1 {
		t.Errorf("Expected depth 1, got %d", sub1.Depth)
	}
}

func TestGoalStack_SubgoalCompletion(t *testing.T) {
	gs := NewGoalStack(DefaultGoalStackConfig())

	parent := &Goal{ID: "parent", Name: "Parent Goal"}
	gs.Push(parent)

	subgoals := []*Goal{
		{ID: "sub-1", Name: "Subgoal 1"},
		{ID: "sub-2", Name: "Subgoal 2"},
	}
	gs.Decompose("parent", subgoals)

	// Complete all subgoals
	gs.Complete("sub-1")
	gs.Complete("sub-2")

	// Parent should be auto-completed
	parentGoal, _ := gs.Get("parent")
	if parentGoal.Status != GoalCompleted {
		t.Errorf("Expected parent COMPLETED after all subgoals done, got %s", parentGoal.Status)
	}
}

func TestGoalStack_SubgoalFailure(t *testing.T) {
	gs := NewGoalStack(DefaultGoalStackConfig())

	parent := &Goal{ID: "parent", Name: "Parent Goal"}
	gs.Push(parent)

	subgoals := []*Goal{
		{ID: "sub-1", Name: "Subgoal 1"},
	}
	gs.Decompose("parent", subgoals)

	// Fail the only subgoal
	gs.Fail("sub-1", "subgoal failed")

	// Parent should cascade fail
	parentGoal, _ := gs.Get("parent")
	if parentGoal.Status != GoalFailed {
		t.Errorf("Expected parent FAILED after subgoal failure, got %s", parentGoal.Status)
	}
}

func TestGoalStack_MaxDepth(t *testing.T) {
	config := DefaultGoalStackConfig()
	config.MaxDepth = 3
	gs := NewGoalStack(config)

	// Create nested goals up to max depth
	gs.Push(&Goal{ID: "level-0", Name: "Level 0"})
	gs.Decompose("level-0", []*Goal{{ID: "level-1", Name: "Level 1"}})
	gs.Decompose("level-1", []*Goal{{ID: "level-2", Name: "Level 2"}})
	gs.Decompose("level-2", []*Goal{{ID: "level-3", Name: "Level 3"}})

	// Should fail at max depth
	err := gs.Decompose("level-3", []*Goal{{ID: "level-4", Name: "Level 4"}})
	if err != ErrMaxDepthExceeded {
		t.Errorf("Expected ErrMaxDepthExceeded, got %v", err)
	}
}

func TestGoalStack_GetByStatus(t *testing.T) {
	gs := NewGoalStack(DefaultGoalStackConfig())

	gs.Push(&Goal{ID: "active-1", Name: "Active 1"})
	gs.Push(&Goal{ID: "active-2", Name: "Active 2"})
	gs.Complete("active-1")
	gs.Push(&Goal{ID: "pending", Name: "Pending"})
	gs.Suspend("pending", "suspended")

	completed := gs.GetByStatus(GoalCompleted)
	if len(completed) != 1 {
		t.Errorf("Expected 1 completed, got %d", len(completed))
	}

	suspended := gs.GetByStatus(GoalSuspended)
	// Note: When active-1 was completed, active-2 became current, and pending was suspended
	if len(suspended) < 1 {
		t.Errorf("Expected at least 1 suspended, got %d", len(suspended))
	}
}

func TestGoalStack_GetAncestors(t *testing.T) {
	gs := NewGoalStack(DefaultGoalStackConfig())

	gs.Push(&Goal{ID: "grandparent", Name: "Grandparent"})
	gs.Decompose("grandparent", []*Goal{{ID: "parent", Name: "Parent"}})
	gs.Decompose("parent", []*Goal{{ID: "child", Name: "Child"}})

	ancestors := gs.GetAncestors("child")

	if len(ancestors) != 2 {
		t.Fatalf("Expected 2 ancestors, got %d", len(ancestors))
	}

	if ancestors[0].ID != "parent" {
		t.Errorf("Expected first ancestor 'parent', got '%s'", ancestors[0].ID)
	}
	if ancestors[1].ID != "grandparent" {
		t.Errorf("Expected second ancestor 'grandparent', got '%s'", ancestors[1].ID)
	}
}

func TestGoalStack_UpdateProgress(t *testing.T) {
	gs := NewGoalStack(DefaultGoalStackConfig())

	gs.Push(&Goal{ID: "goal-1", Name: "Test"})

	gs.UpdateProgress("goal-1", 0.5)
	goal, _ := gs.Get("goal-1")
	if goal.Progress != 0.5 {
		t.Errorf("Expected progress 0.5, got %f", goal.Progress)
	}

	// Test clamping
	gs.UpdateProgress("goal-1", 1.5)
	goal, _ = gs.Get("goal-1")
	if goal.Progress != 1.0 {
		t.Errorf("Expected progress clamped to 1.0, got %f", goal.Progress)
	}
}

func TestGoalStack_SetPriority(t *testing.T) {
	gs := NewGoalStack(DefaultGoalStackConfig())

	gs.Push(&Goal{ID: "goal-1", Name: "Test", Priority: PriorityLow})
	gs.Push(&Goal{ID: "goal-2", Name: "Test 2", Priority: PriorityNormal})

	// goal-2 should be higher priority initially
	top, _ := gs.Peek()
	if top.ID != "goal-2" {
		t.Errorf("Expected 'goal-2' at top initially")
	}

	// Boost goal-1's priority
	gs.SetPriority("goal-1", PriorityCritical)

	// Now goal-1 should be at top
	top, _ = gs.Peek()
	if top.ID != "goal-1" {
		t.Errorf("Expected 'goal-1' at top after priority boost, got '%s'", top.ID)
	}
}

func TestGoalStack_Callbacks(t *testing.T) {
	gs := NewGoalStack(DefaultGoalStackConfig())

	activatedCount := 0
	completedCount := 0
	failedCount := 0
	suspendedCount := 0

	gs.OnGoalActivated(func(g *Goal) { activatedCount++ })
	gs.OnGoalCompleted(func(g *Goal) { completedCount++ })
	gs.OnGoalFailed(func(g *Goal) { failedCount++ })
	gs.OnGoalSuspended(func(g *Goal) { suspendedCount++ })

	gs.Push(&Goal{ID: "goal-1"})
	if activatedCount != 1 {
		t.Errorf("Expected 1 activation, got %d", activatedCount)
	}

	gs.Suspend("goal-1", "test")
	if suspendedCount != 1 {
		t.Errorf("Expected 1 suspension, got %d", suspendedCount)
	}

	gs.Resume("goal-1")
	gs.Complete("goal-1")
	if completedCount != 1 {
		t.Errorf("Expected 1 completion, got %d", completedCount)
	}

	gs.Push(&Goal{ID: "goal-2"})
	gs.Fail("goal-2", "test failure")
	if failedCount != 1 {
		t.Errorf("Expected 1 failure, got %d", failedCount)
	}
}

func TestGoalStack_Snapshot(t *testing.T) {
	gs := NewGoalStack(DefaultGoalStackConfig())

	gs.Push(&Goal{ID: "goal-1", Name: "Active"})
	gs.Push(&Goal{ID: "goal-2", Name: "To Suspend"})
	gs.Suspend("goal-2", "test")
	gs.Push(&Goal{ID: "goal-3", Name: "To Complete"})
	gs.Complete("goal-3")

	snapshot := gs.Snapshot()

	if snapshot.ActiveCount != 1 {
		t.Errorf("Expected 1 active, got %d", snapshot.ActiveCount)
	}
	if snapshot.SuspendedCount != 1 {
		t.Errorf("Expected 1 suspended, got %d", snapshot.SuspendedCount)
	}
	if snapshot.CompletedCount != 1 {
		t.Errorf("Expected 1 completed, got %d", snapshot.CompletedCount)
	}

	t.Logf("Snapshot: Active=%d, Suspended=%d, Completed=%d, Current=%s",
		snapshot.ActiveCount, snapshot.SuspendedCount, snapshot.CompletedCount, snapshot.CurrentGoalID)
}

func TestGoalStack_Clear(t *testing.T) {
	gs := NewGoalStack(DefaultGoalStackConfig())

	gs.Push(&Goal{ID: "goal-1"})
	gs.Push(&Goal{ID: "goal-2"})
	gs.Suspend("goal-2", "test")

	gs.Clear()

	if gs.Size() != 0 {
		t.Errorf("Expected 0 active after clear, got %d", gs.Size())
	}
	if gs.TotalSize() != 0 {
		t.Errorf("Expected 0 total after clear, got %d", gs.TotalSize())
	}
	if gs.Current() != nil {
		t.Error("Expected no current goal after clear")
	}
}

func TestGoalStatus_String(t *testing.T) {
	tests := []struct {
		status   GoalStatus
		expected string
	}{
		{GoalPending, "PENDING"},
		{GoalActive, "ACTIVE"},
		{GoalSuspended, "SUSPENDED"},
		{GoalCompleted, "COMPLETED"},
		{GoalFailed, "FAILED"},
		{GoalDecomposed, "DECOMPOSED"},
	}

	for _, tt := range tests {
		if tt.status.String() != tt.expected {
			t.Errorf("Expected '%s', got '%s'", tt.expected, tt.status.String())
		}
	}
}

func TestGoal_IsTerminal(t *testing.T) {
	g := &Goal{Status: GoalActive}
	if g.IsTerminal() {
		t.Error("Active goal should not be terminal")
	}

	g.Status = GoalCompleted
	if !g.IsTerminal() {
		t.Error("Completed goal should be terminal")
	}

	g.Status = GoalFailed
	if !g.IsTerminal() {
		t.Error("Failed goal should be terminal")
	}
}

// ============================================================================
// Benchmarks
// ============================================================================

func BenchmarkGoalStack_Push(b *testing.B) {
	gs := NewGoalStack(DefaultGoalStackConfig())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gs.Push(&Goal{
			ID:       string(rune(i % 100)),
			Name:     "Test Goal",
			Priority: PriorityNormal,
		})
	}
}

func BenchmarkGoalStack_Pop(b *testing.B) {
	gs := NewGoalStack(DefaultGoalStackConfig())

	// Pre-populate
	for i := 0; i < 100; i++ {
		gs.Push(&Goal{
			ID:       string(rune(i)),
			Priority: GoalPriority(i % 10),
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if gs.IsEmpty() {
			// Refill
			for j := 0; j < 100; j++ {
				gs.Push(&Goal{
					ID:       string(rune(j)),
					Priority: GoalPriority(j % 10),
				})
			}
		}
		gs.Pop()
	}
}

func BenchmarkGoalStack_Decompose(b *testing.B) {
	gs := NewGoalStack(DefaultGoalStackConfig())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gs.Clear()
		gs.Push(&Goal{ID: "parent"})
		gs.Decompose("parent", []*Goal{
			{ID: "sub-1"},
			{ID: "sub-2"},
			{ID: "sub-3"},
		})
	}
}

func BenchmarkGoalStack_Complete(b *testing.B) {
	gs := NewGoalStack(DefaultGoalStackConfig())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		id := string(rune(i % 100))
		gs.Push(&Goal{ID: id})
		gs.Complete(id)
	}
}
