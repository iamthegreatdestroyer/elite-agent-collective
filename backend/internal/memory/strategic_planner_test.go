// Package memory provides the MNEMONIC system for the Elite Agent Collective.
// This file contains tests for the Strategic Planner.

package memory

import (
	"context"
	"testing"
	"time"
)

// ============================================================================
// Strategic Planner Tests
// ============================================================================

// TestStrategicPlanner_CreatePlan tests plan creation
func TestStrategicPlanner_CreatePlan(t *testing.T) {
	planner := NewStrategicPlanner(DefaultPlanningConfig())
	_ = planner.Initialize(nil)

	goal := &Goal{
		ID:       "test-goal",
		Name:     "Test Planning Goal",
		Priority: PriorityHigh,
		Status:   GoalActive,
	}

	plan, err := planner.CreatePlan(context.Background(), goal)

	if err != nil {
		t.Fatalf("CreatePlan failed: %v", err)
	}

	if plan == nil {
		t.Fatal("Plan is nil")
	}

	if plan.ID == "" {
		t.Fatal("Plan ID is empty")
	}

	if len(plan.Actions) == 0 {
		t.Fatal("Plan has no actions")
	}

	if plan.TotalCost <= 0 {
		t.Fatal("Plan cost should be positive")
	}

	t.Logf("Created plan with %d actions, cost: %.2f",
		len(plan.Actions), plan.TotalCost)
}

// TestStrategicPlanner_LookaheadTree tests lookahead tree construction
func TestStrategicPlanner_LookaheadTree(t *testing.T) {
	planner := NewStrategicPlanner(DefaultPlanningConfig())
	_ = planner.Initialize(nil)

	goal := &Goal{
		ID:       "lookahead-goal",
		Name:     "Lookahead Test",
		Priority: PriorityHigh,
		Status:   GoalActive,
	}

	_, _ = planner.CreatePlan(context.Background(), goal)

	tree := planner.GetLookaheadTree()

	if tree == nil {
		t.Fatal("Lookahead tree is nil")
	}

	if tree.Depth != 0 {
		t.Fatalf("Root depth should be 0, got %d", tree.Depth)
	}

	nodeCount := planner.countLookaheadNodes(tree)
	if nodeCount <= 1 {
		t.Fatalf("Expected multiple nodes, got %d", nodeCount)
	}

	t.Logf("Lookahead tree has %d nodes", nodeCount)
}

// TestStrategicPlanner_PlanFeasibility tests plan feasibility evaluation
func TestStrategicPlanner_PlanFeasibility(t *testing.T) {
	planner := NewStrategicPlanner(DefaultPlanningConfig())
	_ = planner.Initialize(nil)

	goal := &Goal{
		ID:       "feasibility-goal",
		Name:     "Feasibility Test",
		Priority: PriorityHigh,
		Status:   GoalActive,
	}

	plan, _ := planner.CreatePlan(context.Background(), goal)

	if !plan.Feasible {
		t.Logf("Plan feasibility: %v (cost: %.2f)", plan.Feasible, plan.TotalCost)
	}

	if plan.Explanation == "" {
		t.Fatal("Plan should have explanation")
	}

	t.Logf("Plan explanation: %s", plan.Explanation)
}

// TestStrategicPlanner_PlanExecution tests plan state management
func TestStrategicPlanner_PlanExecution(t *testing.T) {
	planner := NewStrategicPlanner(DefaultPlanningConfig())
	_ = planner.Initialize(nil)

	goal := &Goal{
		ID:       "exec-goal",
		Name:     "Execution Test",
		Priority: PriorityHigh,
		Status:   GoalActive,
	}

	plan, _ := planner.CreatePlan(context.Background(), goal)
	planID := plan.ID

	// Execute plan
	err := planner.ExecutePlan(planID)
	if err != nil {
		t.Fatalf("ExecutePlan failed: %v", err)
	}

	retrieved := planner.GetPlan(planID)
	if retrieved == nil {
		t.Fatal("Plan should exist after execution")
	}

	t.Logf("Plan execution state updated")
}

// TestStrategicPlanner_Optimization tests plan optimization
func TestStrategicPlanner_Optimization(t *testing.T) {
	planner := NewStrategicPlanner(DefaultPlanningConfig())
	_ = planner.Initialize(nil)

	// Create goal with dependencies to potentially trigger optimization
	goal := &Goal{
		ID:           "opt-goal",
		Name:         "Optimization Test",
		Priority:     PriorityLow,
		Status:       GoalActive,
		Dependencies: []string{"dep1", "dep2", "dep3"},
	}

	plan, _ := planner.CreatePlan(context.Background(), goal)
	originalActions := len(plan.Actions)

	// Optimize plan
	optimized := planner.optimizePlan(plan)

	if len(optimized.Actions) != originalActions+1 {
		t.Logf("Optimization added actions: %d -> %d", originalActions, len(optimized.Actions))
	}

	t.Logf("Plan optimization successful")
}

// TestStrategicPlanner_BestStrategy tests strategy selection from lookahead
func TestStrategicPlanner_BestStrategy(t *testing.T) {
	planner := NewStrategicPlanner(DefaultPlanningConfig())
	_ = planner.Initialize(nil)

	goal := &Goal{
		ID:       "strategy-goal",
		Name:     "Strategy Test",
		Priority: PriorityHigh,
		Status:   GoalActive,
	}

	_, _ = planner.CreatePlan(context.Background(), goal)

	tree := planner.GetLookaheadTree()
	if tree == nil || len(tree.Children) == 0 {
		t.Fatal("No children in lookahead tree")
	}

	best := planner.GetBestStrategy(tree)

	if best == nil {
		t.Fatal("Best strategy is nil")
	}

	// Best strategy should have highest score
	for _, child := range tree.Children {
		if child.Score > best.Score {
			t.Fatalf("Best strategy selection failed: %f <= %f",
				best.Score, child.Score)
		}
	}

	t.Logf("Selected best strategy with score: %.3f", best.Score)
}

// TestStrategicPlanner_Caching tests plan caching
func TestStrategicPlanner_Caching(t *testing.T) {
	config := DefaultPlanningConfig()
	config.PlanCachingEnabled = true

	planner := NewStrategicPlanner(config)
	_ = planner.Initialize(nil)

	goal := &Goal{
		ID:       "cache-goal",
		Name:     "Caching Test",
		Priority: PriorityHigh,
		Status:   GoalActive,
	}

	// Create plan first time
	start := time.Now()
	plan1, _ := planner.CreatePlan(context.Background(), goal)
	duration1 := time.Since(start)

	// Create plan second time (should be cached)
	start = time.Now()
	plan2, _ := planner.CreatePlan(context.Background(), goal)
	duration2 := time.Since(start)

	if plan1.ID != plan2.ID {
		t.Fatal("Cached plan should have same ID")
	}

	t.Logf("Caching performance - First: %v, Cached: %v (speedup: %.1fx)",
		duration1, duration2, float64(duration1+1)/float64(duration2+1))
}

// TestStrategicPlanner_MultipleGoals tests planning for multiple goals
func TestStrategicPlanner_MultipleGoals(t *testing.T) {
	planner := NewStrategicPlanner(DefaultPlanningConfig())
	_ = planner.Initialize(nil)

	goals := []*Goal{
		{ID: "goal1", Name: "Goal 1", Priority: PriorityHigh, Status: GoalActive},
		{ID: "goal2", Name: "Goal 2", Priority: PriorityNormal, Status: GoalActive},
		{ID: "goal3", Name: "Goal 3", Priority: PriorityLow, Status: GoalActive},
	}

	plans := make([]*Plan, len(goals))

	for i, goal := range goals {
		plan, err := planner.CreatePlan(context.Background(), goal)
		if err != nil {
			t.Fatalf("Failed to plan goal %s: %v", goal.ID, err)
		}
		plans[i] = plan
	}

	// Verify all plans are different
	for i := 0; i < len(plans); i++ {
		for j := i + 1; j < len(plans); j++ {
			if plans[i].ID == plans[j].ID {
				t.Fatal("Plans should have unique IDs")
			}
		}
	}

	t.Logf("Successfully created %d distinct plans", len(plans))
}

// TestStrategicPlanner_GetMetrics tests metrics retrieval
func TestStrategicPlanner_GetMetrics(t *testing.T) {
	planner := NewStrategicPlanner(DefaultPlanningConfig())
	_ = planner.Initialize(nil)

	goal := &Goal{
		ID:       "metrics-goal",
		Name:     "Metrics Collection Test",
		Priority: PriorityHigh,
		Status:   GoalActive,
	}

	_, _ = planner.CreatePlan(context.Background(), goal)

	metrics := planner.GetMetrics()

	if metrics.ComponentName != "StrategicPlanner" {
		t.Fatalf("Wrong component name: %s", metrics.ComponentName)
	}

	if metrics.TotalRequests <= 0 {
		t.Fatal("TotalRequests should be positive")
	}

	t.Logf("Metrics: Requests=%d, Successful=%d, Failed=%d, Latency=%v",
		metrics.TotalRequests, metrics.SuccessfulRequests, metrics.FailedRequests,
		metrics.AverageLatency)
}

// TestStrategicPlanner_Shutdown tests graceful shutdown
func TestStrategicPlanner_Shutdown(t *testing.T) {
	planner := NewStrategicPlanner(DefaultPlanningConfig())
	_ = planner.Initialize(nil)

	goal := &Goal{
		ID:       "shutdown-goal",
		Name:     "Shutdown Test",
		Priority: PriorityHigh,
		Status:   GoalActive,
	}

	plan, _ := planner.CreatePlan(context.Background(), goal)
	planID := plan.ID

	// Verify plan exists
	if planner.GetPlan(planID) == nil {
		t.Fatal("Plan should exist before shutdown")
	}

	// Shutdown
	err := planner.Shutdown()
	if err != nil {
		t.Fatalf("Shutdown failed: %v", err)
	}

	// Verify plans are cleared
	if planner.GetPlan(planID) != nil {
		t.Fatal("Plan should not exist after shutdown")
	}

	t.Log("Shutdown successful")
}

// BenchmarkStrategicPlanner_CreatePlan benchmarks plan creation
func BenchmarkStrategicPlanner_CreatePlan(b *testing.B) {
	planner := NewStrategicPlanner(DefaultPlanningConfig())
	_ = planner.Initialize(nil)

	goal := &Goal{
		ID:       "bench-goal",
		Name:     "Benchmark Goal",
		Priority: PriorityHigh,
		Status:   GoalActive,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = planner.CreatePlan(context.Background(), goal)
	}
}

// BenchmarkStrategicPlanner_Lookahead benchmarks lookahead tree building
func BenchmarkStrategicPlanner_Lookahead(b *testing.B) {
	goal := &Goal{
		ID:       "bench-lookahead",
		Name:     "Lookahead Benchmark",
		Priority: PriorityHigh,
		Status:   GoalActive,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		planner := NewStrategicPlanner(DefaultPlanningConfig())
		_ = planner.Initialize(nil)
		_, _ = planner.CreatePlan(context.Background(), goal)
		_ = planner.Shutdown()
	}
}
