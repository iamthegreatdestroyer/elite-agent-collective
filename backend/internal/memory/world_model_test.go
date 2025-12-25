package memory

import (
	"testing"
	"time"
)

// ============================================================================
// State Type Tests
// ============================================================================

func TestStateType_String(t *testing.T) {
	tests := []struct {
		stateType StateType
		expected  string
	}{
		{StateInitial, "initial"},
		{StateIntermediate, "intermediate"},
		{StateGoal, "goal"},
		{StateFailure, "failure"},
		{StateUnknown, "unknown"},
		{StateType(99), "undefined"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.stateType.String(); got != tt.expected {
				t.Errorf("StateType.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestStateType_IsTerminal(t *testing.T) {
	if !StateGoal.IsTerminal() {
		t.Error("StateGoal should be terminal")
	}
	if !StateFailure.IsTerminal() {
		t.Error("StateFailure should be terminal")
	}
	if StateInitial.IsTerminal() {
		t.Error("StateInitial should not be terminal")
	}
	if StateIntermediate.IsTerminal() {
		t.Error("StateIntermediate should not be terminal")
	}
}

// ============================================================================
// State Tests
// ============================================================================

func TestNewState(t *testing.T) {
	state := NewState(StateInitial, "Test State")

	if state.ID == "" {
		t.Error("State should have an ID")
	}
	if state.Type != StateInitial {
		t.Errorf("Type = %v, want StateInitial", state.Type)
	}
	if state.Description != "Test State" {
		t.Errorf("Description = %v, want 'Test State'", state.Description)
	}
	if state.Activation != 1.0 {
		t.Errorf("Activation = %v, want 1.0", state.Activation)
	}
	if state.Confidence != 1.0 {
		t.Errorf("Confidence = %v, want 1.0", state.Confidence)
	}
}

func TestState_Clone(t *testing.T) {
	state := NewState(StateGoal, "Original")
	state.SetFeature("key", "value")
	state.Metadata["meta"] = "data"

	clone := state.Clone()

	if clone.ID != state.ID {
		t.Error("Clone should have same ID")
	}
	if clone.Features["key"] != "value" {
		t.Error("Clone should copy features")
	}

	// Modify clone, original should be unaffected
	clone.Features["key"] = "modified"
	if state.Features["key"] == "modified" {
		t.Error("Clone features should be independent")
	}
}

func TestState_Features(t *testing.T) {
	state := NewState(StateInitial, "Test")

	state.SetFeature("string_val", "hello")
	state.SetFeature("int_val", 42)
	state.SetFeature("float_val", 3.14)

	// Get feature
	val, exists := state.GetFeature("string_val")
	if !exists || val != "hello" {
		t.Error("Should retrieve string feature")
	}

	// Get feature as float
	f, ok := state.GetFeatureFloat("int_val")
	if !ok || f != 42.0 {
		t.Error("Should convert int to float")
	}

	f, ok = state.GetFeatureFloat("float_val")
	if !ok || f != 3.14 {
		t.Error("Should get float feature")
	}

	// Non-existent feature
	_, ok = state.GetFeatureFloat("nonexistent")
	if ok {
		t.Error("Should return false for nonexistent feature")
	}
}

func TestState_Similarity(t *testing.T) {
	state1 := NewState(StateInitial, "State 1")
	state1.SetFeature("a", 1)
	state1.SetFeature("b", 2)
	state1.SetFeature("c", 3)

	state2 := NewState(StateInitial, "State 2")
	state2.SetFeature("a", 1)
	state2.SetFeature("b", 2)
	state2.SetFeature("c", 3)

	// Identical features
	sim := state1.Similarity(state2)
	if sim != 1.0 {
		t.Errorf("Identical states should have similarity 1.0, got %v", sim)
	}

	// Different features
	state3 := NewState(StateInitial, "State 3")
	state3.SetFeature("a", 1)
	state3.SetFeature("b", 99)
	state3.SetFeature("d", 4)

	sim = state1.Similarity(state3)
	if sim >= 1.0 || sim <= 0 {
		t.Errorf("Partially similar states should have 0 < similarity < 1, got %v", sim)
	}

	// Nil comparison
	sim = state1.Similarity(nil)
	if sim != 0 {
		t.Errorf("Similarity with nil should be 0, got %v", sim)
	}
}

// ============================================================================
// Action Tests
// ============================================================================

func TestNewSimAction(t *testing.T) {
	action := NewSimAction(SimActionAgent, "Test Action")

	if action.ID == "" {
		t.Error("Action should have an ID")
	}
	if action.Type != SimActionAgent {
		t.Errorf("Type = %v, want SimActionAgent", action.Type)
	}
	if action.Name != "Test Action" {
		t.Errorf("Name = %v, want 'Test Action'", action.Name)
	}
	if action.SuccessProbability != 1.0 {
		t.Errorf("SuccessProbability = %v, want 1.0", action.SuccessProbability)
	}
}

func TestSimActionType_String(t *testing.T) {
	tests := []struct {
		actionType SimActionType
		expected   string
	}{
		{SimActionAgent, "agent"},
		{SimActionQuery, "query"},
		{SimActionTransform, "transform"},
		{SimActionComposite, "composite"},
		{SimActionConditional, "conditional"},
		{SimActionType(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.actionType.String(); got != tt.expected {
			t.Errorf("ActionType.String() = %v, want %v", got, tt.expected)
		}
	}
}

func TestSimAction_IsApplicable(t *testing.T) {
	action := NewSimAction(SimActionTransform, "Test")
	action.Preconditions = []Predicate{
		{Feature: "ready", Operator: "eq", Value: true},
		{Feature: "count", Operator: "gt", Value: 0},
	}

	// State meets preconditions
	state1 := NewState(StateInitial, "Valid")
	state1.SetFeature("ready", true)
	state1.SetFeature("count", 5)

	if !action.IsApplicable(state1) {
		t.Error("Action should be applicable when preconditions met")
	}

	// State doesn't meet preconditions
	state2 := NewState(StateInitial, "Invalid")
	state2.SetFeature("ready", false)
	state2.SetFeature("count", 5)

	if action.IsApplicable(state2) {
		t.Error("Action should not be applicable when preconditions not met")
	}
}

func TestPredicate_Evaluate(t *testing.T) {
	state := NewState(StateInitial, "Test")
	state.SetFeature("count", 10)
	state.SetFeature("name", "test")
	state.SetFeature("active", true)

	tests := []struct {
		predicate Predicate
		expected  bool
	}{
		{Predicate{Feature: "count", Operator: "eq", Value: 10}, true},
		{Predicate{Feature: "count", Operator: "ne", Value: 5}, true},
		{Predicate{Feature: "count", Operator: "gt", Value: 5}, true},
		{Predicate{Feature: "count", Operator: "lt", Value: 15}, true},
		{Predicate{Feature: "count", Operator: "gte", Value: 10}, true},
		{Predicate{Feature: "count", Operator: "lte", Value: 10}, true},
		{Predicate{Feature: "name", Operator: "exists"}, true},
		{Predicate{Feature: "missing", Operator: "not_exists"}, true},
		{Predicate{Feature: "count", Operator: "gt", Value: 15}, false},
		{Predicate{Feature: "missing", Operator: "exists"}, false},
	}

	for _, tt := range tests {
		result := tt.predicate.Evaluate(state)
		if result != tt.expected {
			t.Errorf("Predicate %v.Evaluate() = %v, want %v",
				tt.predicate, result, tt.expected)
		}
	}
}

func TestStateEffect_Apply(t *testing.T) {
	state := NewState(StateInitial, "Test")
	state.SetFeature("count", 10.0)
	state.SetFeature("name", "original")

	// Test set
	effect1 := StateEffect{Feature: "name", Operation: "set", Value: "modified"}
	effect1.Apply(state)
	if state.Features["name"] != "modified" {
		t.Error("Set operation should update value")
	}

	// Test add
	effect2 := StateEffect{Feature: "count", Operation: "add", Value: 5.0}
	effect2.Apply(state)
	if count, _ := state.GetFeatureFloat("count"); count != 15.0 {
		t.Errorf("Add operation should increment value, got %v", count)
	}

	// Test multiply
	effect3 := StateEffect{Feature: "count", Operation: "multiply", Value: 2.0}
	effect3.Apply(state)
	if count, _ := state.GetFeatureFloat("count"); count != 30.0 {
		t.Errorf("Multiply operation should scale value, got %v", count)
	}

	// Test remove
	effect4 := StateEffect{Feature: "name", Operation: "remove"}
	effect4.Apply(state)
	if _, exists := state.Features["name"]; exists {
		t.Error("Remove operation should delete feature")
	}
}

// ============================================================================
// Trajectory Tests
// ============================================================================

func TestNewTrajectory(t *testing.T) {
	state := NewState(StateInitial, "Start")
	traj := NewTrajectory(state)

	if traj.ID == "" {
		t.Error("Trajectory should have an ID")
	}
	if len(traj.States) != 1 {
		t.Errorf("Trajectory should have 1 state, got %d", len(traj.States))
	}
	if len(traj.Actions) != 0 {
		t.Errorf("Trajectory should have 0 actions, got %d", len(traj.Actions))
	}
	if traj.EstimatedSuccess != 1.0 {
		t.Errorf("EstimatedSuccess = %v, want 1.0", traj.EstimatedSuccess)
	}
}

func TestTrajectory_AddStep(t *testing.T) {
	state := NewState(StateInitial, "Start")
	traj := NewTrajectory(state)

	action := NewSimAction(SimActionTransform, "Step 1")
	action.Cost = 2.0
	action.ExpectedDuration = time.Second
	action.SuccessProbability = 0.9

	nextState := NewState(StateIntermediate, "After Step 1")
	traj.AddStep(action, nextState)

	if traj.Length() != 1 {
		t.Errorf("Length() = %d, want 1", traj.Length())
	}
	if traj.TotalCost != 2.0 {
		t.Errorf("TotalCost = %v, want 2.0", traj.TotalCost)
	}
	if traj.EstimatedSuccess != 0.9 {
		t.Errorf("EstimatedSuccess = %v, want 0.9", traj.EstimatedSuccess)
	}
	if traj.CurrentState().Description != "After Step 1" {
		t.Error("CurrentState should be the last added state")
	}
}

func TestTrajectory_IsComplete(t *testing.T) {
	state := NewState(StateInitial, "Start")
	traj := NewTrajectory(state)

	if traj.IsComplete() {
		t.Error("Trajectory starting at initial state should not be complete")
	}

	goalState := NewState(StateGoal, "Goal")
	action := NewSimAction(SimActionTransform, "Reach Goal")
	traj.AddStep(action, goalState)

	if !traj.IsComplete() {
		t.Error("Trajectory ending at goal state should be complete")
	}
	if !traj.IsSuccessful() {
		t.Error("Trajectory ending at goal state should be successful")
	}
}

func TestTrajectory_Clone(t *testing.T) {
	state := NewState(StateInitial, "Start")
	traj := NewTrajectory(state)

	action := NewSimAction(SimActionTransform, "Step")
	nextState := NewState(StateIntermediate, "Next")
	traj.AddStep(action, nextState)

	clone := traj.Clone()

	if clone.ID != traj.ID {
		t.Error("Clone should have same ID")
	}
	if clone.Length() != traj.Length() {
		t.Error("Clone should have same length")
	}

	// Modify clone, original should be unaffected
	clone.TotalCost = 999
	if traj.TotalCost == 999 {
		t.Error("Clone should be independent")
	}
}

// ============================================================================
// State Predictor Tests
// ============================================================================

func TestNewStatePredictor(t *testing.T) {
	sp := NewStatePredictor(nil)

	if sp == nil {
		t.Fatal("NewStatePredictor returned nil")
	}
	if sp.config.DefaultConfidence != 0.7 {
		t.Errorf("DefaultConfidence = %v, want 0.7", sp.config.DefaultConfidence)
	}
}

func TestStatePredictor_Predict(t *testing.T) {
	sp := NewStatePredictor(nil)

	state := NewState(StateInitial, "Current")
	state.SetFeature("count", 10.0)

	action := NewSimAction(SimActionTransform, "Increment")
	action.Effects = []StateEffect{
		{Feature: "count", Operation: "add", Value: 5.0},
	}

	nextState := sp.Predict(state, action)

	if nextState == nil {
		t.Fatal("Predict returned nil")
	}
	if nextState.ParentID != state.ID {
		t.Error("Next state should reference parent")
	}
	if count, _ := nextState.GetFeatureFloat("count"); count != 15.0 {
		t.Errorf("Effect should be applied, count = %v, want 15.0", count)
	}
}

func TestStatePredictor_TransitionRules(t *testing.T) {
	sp := NewStatePredictor(nil)

	// Add transition rule
	sp.AddTransitionRule(StateInitial, SimActionTransform, []StateEffect{
		{Feature: "initialized", Operation: "set", Value: true},
	})

	state := NewState(StateInitial, "Start")
	action := NewSimAction(SimActionTransform, "Initialize")

	nextState := sp.Predict(state, action)

	if initialized, ok := nextState.Features["initialized"].(bool); !ok || !initialized {
		t.Error("Transition rule should be applied")
	}
}

func TestStatePredictor_InferStateType(t *testing.T) {
	sp := NewStatePredictor(nil)

	// Test goal achieved
	state1 := NewState(StateIntermediate, "Test")
	state1.SetFeature("goal_achieved", true)
	action := NewSimAction(SimActionTransform, "Test")

	result := sp.Predict(state1, action)
	if result.Type != StateGoal {
		t.Errorf("State with goal_achieved=true should be StateGoal, got %v", result.Type)
	}

	// Test failed
	state2 := NewState(StateIntermediate, "Test")
	state2.SetFeature("failed", true)

	result = sp.Predict(state2, action)
	if result.Type != StateFailure {
		t.Errorf("State with failed=true should be StateFailure, got %v", result.Type)
	}

	// Test progress complete
	state3 := NewState(StateIntermediate, "Test")
	state3.SetFeature("progress", 1.0)

	result = sp.Predict(state3, action)
	if result.Type != StateGoal {
		t.Errorf("State with progress=1.0 should be StateGoal, got %v", result.Type)
	}
}

// ============================================================================
// Outcome Estimator Tests
// ============================================================================

func TestNewOutcomeEstimator(t *testing.T) {
	oe := NewOutcomeEstimator(nil)

	if oe == nil {
		t.Fatal("NewOutcomeEstimator returned nil")
	}
	if oe.config.BaseSuccessProbability != 0.5 {
		t.Errorf("BaseSuccessProbability = %v, want 0.5", oe.config.BaseSuccessProbability)
	}
}

func TestOutcomeEstimator_Estimate(t *testing.T) {
	oe := NewOutcomeEstimator(nil)

	// Add goal predicates
	oe.AddGoalPredicate(Predicate{Feature: "task_complete", Operator: "eq", Value: true})
	oe.AddGoalPredicate(Predicate{Feature: "quality", Operator: "gte", Value: 0.8})

	// Create trajectory reaching goal
	state := NewState(StateInitial, "Start")
	traj := NewTrajectory(state)

	goalState := NewState(StateGoal, "Goal")
	goalState.SetFeature("task_complete", true)
	goalState.SetFeature("quality", 0.9)

	action := NewSimAction(SimActionTransform, "Complete")
	traj.AddStep(action, goalState)

	estimate := oe.Estimate(traj)
	if estimate <= 0 {
		t.Errorf("Successful trajectory should have positive estimate, got %v", estimate)
	}
}

func TestOutcomeEstimator_IsTerminal(t *testing.T) {
	oe := NewOutcomeEstimator(nil)

	// Goal state is terminal
	goalState := NewState(StateGoal, "Goal")
	if !oe.IsTerminal(goalState) {
		t.Error("Goal state should be terminal")
	}

	// Failure state is terminal
	failState := NewState(StateFailure, "Failed")
	if !oe.IsTerminal(failState) {
		t.Error("Failure state should be terminal")
	}

	// Intermediate state not terminal
	interState := NewState(StateIntermediate, "Working")
	if oe.IsTerminal(interState) {
		t.Error("Intermediate state should not be terminal")
	}

	// Add failure predicate
	oe.AddFailurePredicate(Predicate{Feature: "error", Operator: "exists"})

	errorState := NewState(StateIntermediate, "Error")
	errorState.SetFeature("error", "something went wrong")
	if !oe.IsTerminal(errorState) {
		t.Error("State matching failure predicate should be terminal")
	}
}

func TestOutcomeEstimator_RecordOutcome(t *testing.T) {
	oe := NewOutcomeEstimator(nil)

	state := NewState(StateInitial, "Start")
	traj := NewTrajectory(state)

	goalState := NewState(StateGoal, "Goal")
	action := NewSimAction(SimActionTransform, "Complete")
	traj.AddStep(action, goalState)

	oe.RecordOutcome(traj, true)

	rate := oe.GetSuccessRate()
	if rate != 1.0 {
		t.Errorf("Success rate should be 1.0 after one success, got %v", rate)
	}

	// Record a failure
	traj2 := NewTrajectory(state)
	failState := NewState(StateFailure, "Failed")
	traj2.AddStep(action, failState)
	oe.RecordOutcome(traj2, false)

	rate = oe.GetSuccessRate()
	if rate != 0.5 {
		t.Errorf("Success rate should be 0.5 after 1 success + 1 failure, got %v", rate)
	}
}

// ============================================================================
// World Model Tests
// ============================================================================

func TestNewWorldModel(t *testing.T) {
	wm := NewWorldModel(nil)

	if wm == nil {
		t.Fatal("NewWorldModel returned nil")
	}
	if wm.config.MaxSimulationDepth != 10 {
		t.Errorf("MaxSimulationDepth = %v, want 10", wm.config.MaxSimulationDepth)
	}
	if wm.ActionCount() != 0 {
		t.Error("New world model should have no actions")
	}
}

func TestWorldModel_AddAction(t *testing.T) {
	wm := NewWorldModel(nil)

	action := NewSimAction(SimActionAgent, "Test Action")
	wm.AddAction(action)

	if wm.ActionCount() != 1 {
		t.Errorf("ActionCount = %d, want 1", wm.ActionCount())
	}

	retrieved, exists := wm.GetAction(action.ID)
	if !exists {
		t.Error("Action should be retrievable")
	}
	if retrieved.Name != "Test Action" {
		t.Error("Retrieved action should match original")
	}

	wm.RemoveAction(action.ID)
	if wm.ActionCount() != 0 {
		t.Error("Action should be removed")
	}
}

func TestWorldModel_GetApplicableActions(t *testing.T) {
	wm := NewWorldModel(nil)

	// Action with no preconditions (always applicable)
	action1 := NewSimAction(SimActionTransform, "Always")
	wm.AddAction(action1)

	// Action with precondition
	action2 := NewSimAction(SimActionTransform, "Conditional")
	action2.Preconditions = []Predicate{
		{Feature: "ready", Operator: "eq", Value: true},
	}
	wm.AddAction(action2)

	// State without ready flag
	state1 := NewState(StateInitial, "Not Ready")
	applicable := wm.GetApplicableActions(state1)

	if len(applicable) != 1 {
		t.Errorf("Expected 1 applicable action, got %d", len(applicable))
	}

	// State with ready flag
	state2 := NewState(StateInitial, "Ready")
	state2.SetFeature("ready", true)
	applicable = wm.GetApplicableActions(state2)

	if len(applicable) != 2 {
		t.Errorf("Expected 2 applicable actions, got %d", len(applicable))
	}
}

func TestWorldModel_SimulateAction(t *testing.T) {
	wm := NewWorldModel(nil)

	state := NewState(StateInitial, "Start")
	state.SetFeature("count", 10.0)

	action := NewSimAction(SimActionTransform, "Increment")
	action.Effects = []StateEffect{
		{Feature: "count", Operation: "add", Value: 5.0},
	}

	traj, err := wm.SimulateAction(state, action)
	if err != nil {
		t.Fatalf("SimulateAction failed: %v", err)
	}

	if traj.Length() != 1 {
		t.Errorf("Trajectory length = %d, want 1", traj.Length())
	}

	finalCount, _ := traj.CurrentState().GetFeatureFloat("count")
	if finalCount != 15.0 {
		t.Errorf("Final count = %v, want 15.0", finalCount)
	}
}

func TestWorldModel_SimulateSequence(t *testing.T) {
	wm := NewWorldModel(nil)

	state := NewState(StateInitial, "Start")
	state.SetFeature("count", 0.0)

	actions := make([]*SimAction, 3)
	for i := 0; i < 3; i++ {
		actions[i] = NewSimAction(SimActionTransform, "Add")
		actions[i].Effects = []StateEffect{
			{Feature: "count", Operation: "add", Value: 1.0},
		}
	}

	traj, err := wm.SimulateSequence(state, actions)
	if err != nil {
		t.Fatalf("SimulateSequence failed: %v", err)
	}

	if traj.Length() != 3 {
		t.Errorf("Trajectory length = %d, want 3", traj.Length())
	}

	finalCount, _ := traj.CurrentState().GetFeatureFloat("count")
	if finalCount != 3.0 {
		t.Errorf("Final count = %v, want 3.0", finalCount)
	}
}

func TestWorldModel_SimulateBestPath(t *testing.T) {
	wm := NewWorldModel(nil)

	// Add actions with different success probabilities
	action1 := NewSimAction(SimActionTransform, "Good Path")
	action1.SuccessProbability = 0.9
	action1.Effects = []StateEffect{
		{Feature: "progress", Operation: "add", Value: 0.5},
	}
	wm.AddAction(action1)

	action2 := NewSimAction(SimActionTransform, "Bad Path")
	action2.SuccessProbability = 0.1
	action2.Effects = []StateEffect{
		{Feature: "progress", Operation: "add", Value: 0.1},
	}
	wm.AddAction(action2)

	state := NewState(StateInitial, "Start")
	state.SetFeature("progress", 0.0)

	traj, err := wm.SimulateBestPath(state, 5)
	if err != nil {
		t.Fatalf("SimulateBestPath failed: %v", err)
	}

	if traj.Length() == 0 {
		t.Error("Should find some path")
	}

	// Best path should use high-probability action
	if len(traj.Actions) > 0 && traj.Actions[0].Name != "Good Path" {
		t.Error("Should choose highest probability action")
	}
}

func TestWorldModel_CompareActions(t *testing.T) {
	wm := NewWorldModel(nil)

	state := NewState(StateInitial, "Start")

	action1 := NewSimAction(SimActionTransform, "High Success")
	action1.SuccessProbability = 0.9

	action2 := NewSimAction(SimActionTransform, "Low Success")
	action2.SuccessProbability = 0.3

	comparisons, err := wm.CompareActions(state, []*SimAction{action2, action1})
	if err != nil {
		t.Fatalf("CompareActions failed: %v", err)
	}

	if len(comparisons) != 2 {
		t.Errorf("Expected 2 comparisons, got %d", len(comparisons))
	}

	// Should be sorted by expected success
	if comparisons[0].Action.Name != "High Success" {
		t.Error("Higher success action should be first")
	}
}

func TestWorldModel_ExploreAlternatives(t *testing.T) {
	// Use config with lower pruning threshold to account for multiplicative decay
	config := DefaultWorldModelConfig()
	config.PruningThreshold = 0.01 // Lower threshold since success is multiplicative
	wm := NewWorldModel(config)

	// Add multiple actions
	for i := 0; i < 3; i++ {
		action := NewSimAction(SimActionTransform, "Option")
		action.SuccessProbability = 0.5 + float64(i)*0.1
		action.Effects = []StateEffect{
			{Feature: "path", Operation: "set", Value: i},
		}
		wm.AddAction(action)
	}

	state := NewState(StateInitial, "Start")

	trajectories, err := wm.ExploreAlternatives(state, 2)
	if err != nil {
		t.Fatalf("ExploreAlternatives failed: %v", err)
	}

	if len(trajectories) == 0 {
		t.Error("Should find some trajectories")
	}

	// Should be sorted by success probability
	for i := 1; i < len(trajectories); i++ {
		if trajectories[i].EstimatedSuccess > trajectories[i-1].EstimatedSuccess {
			t.Error("Trajectories should be sorted by success descending")
		}
	}
}

func TestWorldModel_Stats(t *testing.T) {
	wm := NewWorldModel(nil)

	state := NewState(StateInitial, "Start")
	action := NewSimAction(SimActionTransform, "Test")

	wm.SimulateAction(state, action)
	wm.SimulateAction(state, action)

	stats := wm.GetStats()

	if stats.TotalSimulations != 2 {
		t.Errorf("TotalSimulations = %d, want 2", stats.TotalSimulations)
	}
	if stats.TotalTrajectories != 2 {
		t.Errorf("TotalTrajectories = %d, want 2", stats.TotalTrajectories)
	}
}

func TestWorldModel_Clear(t *testing.T) {
	wm := NewWorldModel(nil)

	action := NewSimAction(SimActionTransform, "Test")
	wm.AddAction(action)

	state := NewState(StateInitial, "Start")
	wm.SimulateAction(state, action)

	wm.Clear()

	if wm.ActionCount() != 0 {
		t.Error("Clear should remove all actions")
	}

	stats := wm.GetStats()
	if stats.TotalSimulations != 0 {
		t.Error("Clear should reset stats")
	}
}

// ============================================================================
// Integration Tests
// ============================================================================

func TestWorldModel_CompleteSimulation(t *testing.T) {
	wm := NewWorldModel(nil)

	// Configure outcome estimator
	wm.outcomeEstimator.AddGoalPredicate(Predicate{
		Feature:  "progress",
		Operator: "gte",
		Value:    1.0,
	})

	// Add action that advances progress
	action := NewSimAction(SimActionTransform, "Work")
	action.Effects = []StateEffect{
		{Feature: "progress", Operation: "add", Value: 0.25},
	}
	wm.AddAction(action)

	// Start state
	state := NewState(StateInitial, "Start")
	state.SetFeature("progress", 0.0)

	// Simulate best path to completion
	traj, err := wm.SimulateBestPath(state, 10)
	if err != nil {
		t.Fatalf("Simulation failed: %v", err)
	}

	// Should take 4 steps to reach progress >= 1.0
	if traj.Length() < 4 {
		t.Errorf("Expected at least 4 steps, got %d", traj.Length())
	}

	finalProgress, _ := traj.CurrentState().GetFeatureFloat("progress")
	if finalProgress < 1.0 {
		t.Errorf("Should reach goal (progress >= 1.0), got %v", finalProgress)
	}
}

// ============================================================================
// Benchmark Tests
// ============================================================================

func BenchmarkWorldModel_SimulateAction(b *testing.B) {
	wm := NewWorldModel(nil)

	action := NewSimAction(SimActionTransform, "Test")
	action.Effects = []StateEffect{
		{Feature: "count", Operation: "add", Value: 1.0},
	}

	state := NewState(StateInitial, "Start")
	state.SetFeature("count", 0.0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wm.SimulateAction(state, action)
	}
}

func BenchmarkWorldModel_SimulateBestPath(b *testing.B) {
	wm := NewWorldModel(nil)

	for i := 0; i < 5; i++ {
		action := NewSimAction(SimActionTransform, "Option")
		action.SuccessProbability = 0.5 + float64(i)*0.1
		wm.AddAction(action)
	}

	state := NewState(StateInitial, "Start")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wm.SimulateBestPath(state, 5)
	}
}
