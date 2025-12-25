package memory

import (
	"fmt"
	"testing"
	"time"
)

func TestNewHierarchicalPlanner(t *testing.T) {
	p := NewHierarchicalPlanner(nil, nil)
	if p == nil {
		t.Fatal("Expected non-nil planner")
	}

	if p.primitiveActions == nil {
		t.Error("Expected non-nil primitive actions map")
	}

	if p.methods == nil {
		t.Error("Expected non-nil methods map")
	}

	if p.maxDepth != 20 {
		t.Errorf("Expected default maxDepth 20, got %d", p.maxDepth)
	}
}

func TestHierarchicalPlanner_RegisterAction(t *testing.T) {
	p := NewHierarchicalPlanner(nil, nil)

	action := &PlannerAction{
		ID:   "action-1",
		Name: "test_action",
		Cost: 1.0,
	}

	p.RegisterAction(action)

	actions := p.GetRegisteredActions()
	if _, exists := actions["test_action"]; !exists {
		t.Error("Expected action to be registered")
	}
}

func TestHierarchicalPlanner_RegisterMethod(t *testing.T) {
	p := NewHierarchicalPlanner(nil, nil)

	method := &Method{
		ID:       "method-1",
		Name:     "test_method",
		TaskName: "composite_task",
		Subtasks: []*Task{
			{ID: "sub-1", Name: "subtask_1", IsPrimitive: true},
		},
		Priority: 1,
	}

	p.RegisterMethod(method)

	methods := p.GetRegisteredMethods()
	if _, exists := methods["composite_task"]; !exists {
		t.Error("Expected method to be registered")
	}
}

func TestHierarchicalPlanner_PlanPrimitiveTask(t *testing.T) {
	p := NewHierarchicalPlanner(nil, nil)

	// Register a primitive action
	action := &PlannerAction{
		ID:            "action-1",
		Name:          "simple_action",
		Cost:          1.0,
		Preconditions: []*Precondition{},
	}
	p.RegisterAction(action)

	// Create state
	state := NewPlannerState()

	// Create task
	task := &Task{
		ID:          "task-1",
		Name:        "simple_action",
		IsPrimitive: true,
	}

	// Plan
	plan, err := p.Plan(task, state)
	if err != nil {
		t.Fatalf("Expected successful plan, got error: %v", err)
	}

	if !plan.Feasible {
		t.Error("Expected feasible plan")
	}

	if len(plan.Actions) != 1 {
		t.Errorf("Expected 1 action in plan, got %d", len(plan.Actions))
	}
}

func TestHierarchicalPlanner_PlanCompositeTask(t *testing.T) {
	p := NewHierarchicalPlanner(nil, nil)

	// Register primitive actions
	p.RegisterAction(&PlannerAction{
		ID:            "action-1",
		Name:          "step_1",
		Cost:          1.0,
		Preconditions: []*Precondition{},
	})

	p.RegisterAction(&PlannerAction{
		ID:            "action-2",
		Name:          "step_2",
		Cost:          2.0,
		Preconditions: []*Precondition{},
	})

	// Register a composite method
	p.RegisterMethod(&Method{
		ID:       "method-1",
		Name:     "composite_method",
		TaskName: "composite_task",
		Subtasks: []*Task{
			{ID: "sub-1", Name: "step_1", IsPrimitive: true},
			{ID: "sub-2", Name: "step_2", IsPrimitive: true},
		},
		Ordering: OrderingSequential,
	})

	state := NewPlannerState()
	task := &Task{ID: "task-1", Name: "composite_task"}

	plan, err := p.Plan(task, state)
	if err != nil {
		t.Fatalf("Expected successful plan, got error: %v", err)
	}

	if len(plan.Actions) != 2 {
		t.Errorf("Expected 2 actions in plan, got %d", len(plan.Actions))
	}

	if plan.TotalCost != 3.0 {
		t.Errorf("Expected total cost 3.0, got %f", plan.TotalCost)
	}
}

func TestHierarchicalPlanner_PlanWithPreconditions(t *testing.T) {
	p := NewHierarchicalPlanner(nil, nil)

	// Register action with precondition
	p.RegisterAction(&PlannerAction{
		ID:   "action-1",
		Name: "conditional_action",
		Preconditions: []*Precondition{
			{Feature: "ready", Operator: "eq", Value: true},
		},
	})

	// Test with precondition NOT met
	state := NewPlannerState()
	task := &Task{ID: "task-1", Name: "conditional_action"}

	plan, err := p.Plan(task, state)
	if err == nil {
		t.Error("Expected error when precondition not met")
	}

	if plan.Feasible {
		t.Error("Expected infeasible plan when precondition not met")
	}

	// Test with precondition met
	state.Set("ready", true)
	plan, err = p.Plan(task, state)
	if err != nil {
		t.Fatalf("Expected successful plan, got error: %v", err)
	}

	if !plan.Feasible {
		t.Error("Expected feasible plan when precondition met")
	}
}

func TestHierarchicalPlanner_PlanWithEffects(t *testing.T) {
	p := NewHierarchicalPlanner(nil, nil)

	// Register action with effect
	p.RegisterAction(&PlannerAction{
		ID:   "action-1",
		Name: "set_action",
		Effects: []*Effect{
			{Feature: "completed", Operation: "set", Value: true},
		},
	})

	state := NewPlannerState()
	task := &Task{ID: "task-1", Name: "set_action"}

	plan, _ := p.Plan(task, state)

	// Execute plan
	finalState, err := p.ExecutePlan(plan, state)
	if err != nil {
		t.Fatalf("Expected successful execution, got error: %v", err)
	}

	val, exists := finalState.Get("completed")
	if !exists || val != true {
		t.Error("Expected 'completed' to be true after execution")
	}
}

func TestHierarchicalPlanner_PlanDeepDecomposition(t *testing.T) {
	p := NewHierarchicalPlanner(nil, nil)

	// Register primitive
	p.RegisterAction(&PlannerAction{
		ID:   "primitive",
		Name: "primitive_action",
	})

	// Create deep decomposition: level3 -> level2 -> level1 -> primitive
	p.RegisterMethod(&Method{
		ID:       "m1",
		Name:     "level1",
		TaskName: "level1_task",
		Subtasks: []*Task{{Name: "primitive_action", IsPrimitive: true}},
	})

	p.RegisterMethod(&Method{
		ID:       "m2",
		Name:     "level2",
		TaskName: "level2_task",
		Subtasks: []*Task{{Name: "level1_task"}},
	})

	p.RegisterMethod(&Method{
		ID:       "m3",
		Name:     "level3",
		TaskName: "level3_task",
		Subtasks: []*Task{{Name: "level2_task"}},
	})

	state := NewPlannerState()
	task := &Task{ID: "top", Name: "level3_task"}

	plan, err := p.Plan(task, state)
	if err != nil {
		t.Fatalf("Expected successful plan, got error: %v", err)
	}

	if len(plan.Actions) != 1 {
		t.Errorf("Expected 1 action after decomposition, got %d", len(plan.Actions))
	}
}

func TestHierarchicalPlanner_MaxDepth(t *testing.T) {
	config := &PlannerConfig{
		MaxDepth:    2,
		MaxPlanTime: 5 * time.Second,
	}
	p := NewHierarchicalPlanner(nil, config)

	// Create recursive method that exceeds depth
	p.RegisterMethod(&Method{
		ID:       "m1",
		Name:     "recursive",
		TaskName: "recursive_task",
		Subtasks: []*Task{{Name: "recursive_task"}}, // Self-reference
	})

	state := NewPlannerState()
	task := &Task{ID: "top", Name: "recursive_task"}

	_, err := p.Plan(task, state)
	if err != ErrMaxDepthReached {
		t.Errorf("Expected ErrMaxDepthReached, got: %v", err)
	}
}

func TestHierarchicalPlanner_Stats(t *testing.T) {
	p := NewHierarchicalPlanner(nil, nil)

	p.RegisterAction(&PlannerAction{
		ID:   "action-1",
		Name: "test_action",
	})

	state := NewPlannerState()
	task := &Task{ID: "task-1", Name: "test_action"}

	// Run a few plans
	for i := 0; i < 3; i++ {
		p.Plan(task, state)
	}

	stats := p.GetStats()
	if stats.TotalPlans != 3 {
		t.Errorf("Expected 3 total plans, got %d", stats.TotalPlans)
	}

	if stats.SuccessfulPlans != 3 {
		t.Errorf("Expected 3 successful plans, got %d", stats.SuccessfulPlans)
	}
}

func TestPrecondition_Evaluate(t *testing.T) {
	tests := []struct {
		name     string
		precond  *Precondition
		state    map[string]interface{}
		expected bool
	}{
		{
			name:     "exists - true",
			precond:  &Precondition{Feature: "x", Operator: "exists"},
			state:    map[string]interface{}{"x": 1},
			expected: true,
		},
		{
			name:     "exists - false",
			precond:  &Precondition{Feature: "x", Operator: "exists"},
			state:    map[string]interface{}{},
			expected: false,
		},
		{
			name:     "not_exists - true",
			precond:  &Precondition{Feature: "x", Operator: "not_exists"},
			state:    map[string]interface{}{},
			expected: true,
		},
		{
			name:     "eq - true",
			precond:  &Precondition{Feature: "x", Operator: "eq", Value: 5},
			state:    map[string]interface{}{"x": 5},
			expected: true,
		},
		{
			name:     "eq - false",
			precond:  &Precondition{Feature: "x", Operator: "eq", Value: 5},
			state:    map[string]interface{}{"x": 3},
			expected: false,
		},
		{
			name:     "gt - true",
			precond:  &Precondition{Feature: "x", Operator: "gt", Value: 5},
			state:    map[string]interface{}{"x": 10},
			expected: true,
		},
		{
			name:     "lt - true",
			precond:  &Precondition{Feature: "x", Operator: "lt", Value: 5},
			state:    map[string]interface{}{"x": 3},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &PlannerState{Features: tt.state}
			result := tt.precond.Evaluate(state)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestEffect_Apply(t *testing.T) {
	tests := []struct {
		name     string
		effect   *Effect
		initial  map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name:     "set",
			effect:   &Effect{Feature: "x", Operation: "set", Value: 10},
			initial:  map[string]interface{}{},
			expected: map[string]interface{}{"x": 10},
		},
		{
			name:     "remove",
			effect:   &Effect{Feature: "x", Operation: "remove"},
			initial:  map[string]interface{}{"x": 5},
			expected: map[string]interface{}{},
		},
		{
			name:     "increment",
			effect:   &Effect{Feature: "x", Operation: "increment", Value: 5},
			initial:  map[string]interface{}{"x": 10},
			expected: map[string]interface{}{"x": float64(15)},
		},
		{
			name:     "decrement",
			effect:   &Effect{Feature: "x", Operation: "decrement", Value: 3},
			initial:  map[string]interface{}{"x": 10},
			expected: map[string]interface{}{"x": float64(7)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &PlannerState{Features: tt.initial}
			tt.effect.Apply(state)

			for k, v := range tt.expected {
				if state.Features[k] != v {
					t.Errorf("Expected %s=%v, got %v", k, v, state.Features[k])
				}
			}

			for k := range tt.initial {
				if _, exists := tt.expected[k]; !exists {
					if _, stillExists := state.Features[k]; stillExists {
						t.Errorf("Expected %s to be removed", k)
					}
				}
			}
		})
	}
}

func TestPlannerState_Clone(t *testing.T) {
	state := NewPlannerState()
	state.Set("a", 1)
	state.Set("b", "test")

	clone := state.Clone()

	// Modify original
	state.Set("a", 999)
	state.Set("c", "new")

	// Clone should be unaffected
	val, _ := clone.Get("a")
	if val != 1 {
		t.Error("Clone should be independent of original")
	}

	_, exists := clone.Get("c")
	if exists {
		t.Error("Clone should not have new keys from original")
	}
}

func TestAgentTaskPlanner(t *testing.T) {
	atp := NewAgentTaskPlanner(nil)

	// Register agents
	atp.RegisterAgent(&AgentInfo{
		ID:           "APEX",
		Tier:         1,
		Capabilities: []string{"coding", "design"},
		Cost:         1.0,
		Availability: true,
	})

	atp.RegisterAgent(&AgentInfo{
		ID:           "CIPHER",
		Tier:         1,
		Capabilities: []string{"security", "crypto"},
		Cost:         1.5,
		Availability: true,
	})

	// Register composite task
	atp.RegisterCompositeTask("secure_development", []string{"APEX", "CIPHER"}, OrderingSequential)

	// Plan
	state := NewPlannerState()
	task := &Task{ID: "dev-task", Name: "secure_development"}

	plan, err := atp.PlanAgentCoordination(task, state)
	if err != nil {
		t.Fatalf("Expected successful plan, got error: %v", err)
	}

	if len(plan.Actions) != 2 {
		t.Errorf("Expected 2 actions in plan, got %d", len(plan.Actions))
	}
}

func TestOrderingType(t *testing.T) {
	if OrderingSequential != "sequential" {
		t.Error("OrderingSequential should be 'sequential'")
	}

	if OrderingParallel != "parallel" {
		t.Error("OrderingParallel should be 'parallel'")
	}

	if OrderingPartial != "partial" {
		t.Error("OrderingPartial should be 'partial'")
	}
}

func TestPlannerErrors(t *testing.T) {
	if ErrPlanningTimeout.Error() == "" {
		t.Error("ErrPlanningTimeout should have message")
	}

	if ErrMaxDepthReached.Error() == "" {
		t.Error("ErrMaxDepthReached should have message")
	}

	if ErrNoPlanFound.Error() == "" {
		t.Error("ErrNoPlanFound should have message")
	}
}

// Benchmarks

func BenchmarkHierarchicalPlanner_PlanPrimitive(b *testing.B) {
	p := NewHierarchicalPlanner(nil, nil)
	p.RegisterAction(&PlannerAction{
		ID:   "action-1",
		Name: "test_action",
	})

	state := NewPlannerState()
	task := &Task{ID: "task-1", Name: "test_action"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.Plan(task, state)
	}
}

func BenchmarkHierarchicalPlanner_PlanComposite(b *testing.B) {
	p := NewHierarchicalPlanner(nil, nil)

	// Register primitives
	for i := 0; i < 5; i++ {
		p.RegisterAction(&PlannerAction{
			ID:   fmt.Sprintf("action-%d", i),
			Name: fmt.Sprintf("step_%d", i),
		})
	}

	// Register composite
	subtasks := make([]*Task, 5)
	for i := 0; i < 5; i++ {
		subtasks[i] = &Task{Name: fmt.Sprintf("step_%d", i), IsPrimitive: true}
	}
	p.RegisterMethod(&Method{
		ID:       "method-1",
		TaskName: "composite",
		Subtasks: subtasks,
	})

	state := NewPlannerState()
	task := &Task{ID: "task-1", Name: "composite"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.Plan(task, state)
	}
}
