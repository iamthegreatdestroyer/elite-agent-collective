// Package memory provides the MNEMONIC system for the Elite Agent Collective.
// This file contains tests for the Multi-Strategy Planner.

package memory

import (
	"context"
	"testing"
)

// ============================================================================
// Multi-Strategy Planner Tests
// ============================================================================

// TestMultiStrategyPlanner_Initialization tests setup
func TestMultiStrategyPlanner_Initialization(t *testing.T) {
	planner := NewMultiStrategyPlanner(DefaultMultiStrategyPlannerConfig())
	err := planner.Initialize(nil)

	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	metrics := planner.GetMetrics()
	if metrics.ComponentName != "MultiStrategyPlanner" {
		t.Fatalf("Wrong component name: %s", metrics.ComponentName)
	}

	t.Log("Multi-strategy planner initialized successfully")
}

// TestMultiStrategyPlanner_GenerateStrategies tests generation
func TestMultiStrategyPlanner_GenerateStrategies(t *testing.T) {
	planner := NewMultiStrategyPlanner(DefaultMultiStrategyPlannerConfig())
	_ = planner.Initialize(nil)

	goal := &Goal{
		ID:           "strategy-goal",
		Name:         "Strategy Test Goal",
		Priority:     PriorityHigh,
		Status:       GoalActive,
		Dependencies: []string{"dep1", "dep2"},
	}

	strategySet, err := planner.GenerateStrategies(context.Background(), goal)

	if err != nil {
		t.Fatalf("Generation failed: %v", err)
	}

	if strategySet == nil {
		t.Fatal("StrategySet is nil")
	}

	if len(strategySet.Strategies) == 0 {
		t.Fatal("No strategies generated")
	}

	t.Logf("Created %d strategies", len(strategySet.Strategies))
}

// TestMultiStrategyPlanner_StrategyCount tests count
func TestMultiStrategyPlanner_StrategyCount(t *testing.T) {
	planner := NewMultiStrategyPlanner(DefaultMultiStrategyPlannerConfig())
	_ = planner.Initialize(nil)

	goal := &Goal{
		ID:           "count-test",
		Name:         "Count Test Goal",
		Priority:     PriorityHigh,
		Status:       GoalActive,
		Dependencies: []string{"dep1", "dep2", "dep3"},
	}

	strategySet, _ := planner.GenerateStrategies(context.Background(), goal)

	minExpected := 3
	maxExpected := 8

	if len(strategySet.Strategies) < minExpected {
		t.Fatalf("Too few strategies: %d < %d", len(strategySet.Strategies), minExpected)
	}

	if len(strategySet.Strategies) > maxExpected {
		t.Fatalf("Too many strategies: %d > %d", len(strategySet.Strategies), maxExpected)
	}

	t.Logf("Strategy count correct: %d (min: %d, max: %d)",
		len(strategySet.Strategies), minExpected, maxExpected)
}

// TestMultiStrategyPlanner_StrategyProperties tests properties
func TestMultiStrategyPlanner_StrategyProperties(t *testing.T) {
	planner := NewMultiStrategyPlanner(DefaultMultiStrategyPlannerConfig())
	_ = planner.Initialize(nil)

	goal := &Goal{
		ID:       "properties-test",
		Name:     "Properties Test Goal",
		Priority: PriorityHigh,
		Status:   GoalActive,
	}

	strategySet, _ := planner.GenerateStrategies(context.Background(), goal)

	for _, strategy := range strategySet.Strategies {
		// Validate risk
		if strategy.Risk < 0 || strategy.Risk > 1.0 {
			t.Fatalf("Invalid risk: %.2f", strategy.Risk)
		}

		// Validate effectiveness
		if strategy.Effectiveness < 0 || strategy.Effectiveness > 1.0 {
			t.Fatalf("Invalid effectiveness: %.2f", strategy.Effectiveness)
		}

		// Validate resources
		if len(strategy.ResourceNeeds) == 0 {
			t.Fatalf("Strategy %s has no resource needs", strategy.ID)
		}

		// Validate timeline
		if strategy.Timeline <= 0 {
			t.Fatalf("Invalid timeline: %d", strategy.Timeline)
		}

		t.Logf("Strategy %s: Risk=%.2f, Effectiveness=%.2f, Timeline=%d days",
			strategy.ID, strategy.Risk, strategy.Effectiveness, strategy.Timeline)
	}
}

// TestMultiStrategyPlanner_ResourceAllocation tests allocation
func TestMultiStrategyPlanner_ResourceAllocation(t *testing.T) {
	planner := NewMultiStrategyPlanner(DefaultMultiStrategyPlannerConfig())
	_ = planner.Initialize(nil)

	goal := &Goal{
		ID:       "allocation-test",
		Name:     "Allocation Test Goal",
		Priority: PriorityHigh,
		Status:   GoalActive,
	}

	strategySet, _ := planner.GenerateStrategies(context.Background(), goal)

	for _, strategy := range strategySet.Strategies {
		allocation := planner.GetAllocation(strategySet.ID, strategy.ID)

		if allocation == nil {
			t.Fatalf("No allocation for strategy %s", strategy.ID)
		}

		if len(allocation) == 0 {
			t.Fatalf("Empty allocation for strategy %s", strategy.ID)
		}

		t.Logf("Strategy %s: Allocated %d resources", strategy.ID, len(allocation))
	}
}

// TestMultiStrategyPlanner_StrategyComparison tests comparison
func TestMultiStrategyPlanner_StrategyComparison(t *testing.T) {
	planner := NewMultiStrategyPlanner(DefaultMultiStrategyPlannerConfig())
	_ = planner.Initialize(nil)

	goal := &Goal{
		ID:       "comparison-test",
		Name:     "Comparison Test Goal",
		Priority: PriorityHigh,
		Status:   GoalActive,
	}

	strategySet, _ := planner.GenerateStrategies(context.Background(), goal)

	if len(strategySet.ComparisonResults) == 0 {
		t.Fatal("No comparison results generated")
	}

	for _, comparison := range strategySet.ComparisonResults {
		if comparison.Winner == "" {
			t.Fatalf("Comparison %s has no winner", comparison.ID)
		}

		if comparison.Confidence < 0 || comparison.Confidence > 1.0 {
			t.Fatalf("Invalid confidence: %.2f", comparison.Confidence)
		}

		t.Logf("Comparison: %s vs %s, Winner: %s",
			comparison.StrategyA, comparison.StrategyB, comparison.Winner)
	}
}

// TestMultiStrategyPlanner_SelectedStrategy tests selection
func TestMultiStrategyPlanner_SelectedStrategy(t *testing.T) {
	planner := NewMultiStrategyPlanner(DefaultMultiStrategyPlannerConfig())
	_ = planner.Initialize(nil)

	goal := &Goal{
		ID:       "selection-test",
		Name:     "Selection Test Goal",
		Priority: PriorityHigh,
		Status:   GoalActive,
	}

	strategySet, _ := planner.GenerateStrategies(context.Background(), goal)

	selected := planner.GetSelectedStrategy(strategySet.ID)

	if selected == nil {
		t.Fatal("No strategy selected")
	}

	if selected.Status != StrategyOptimized {
		t.Fatalf("Wrong status: %s", selected.Status)
	}

	t.Logf("Selected strategy: %s (Status: %s)", selected.Name, selected.Status)
}

// TestMultiStrategyPlanner_RankStrategies tests ranking
func TestMultiStrategyPlanner_RankStrategies(t *testing.T) {
	planner := NewMultiStrategyPlanner(DefaultMultiStrategyPlannerConfig())
	_ = planner.Initialize(nil)

	goal := &Goal{
		ID:       "ranking-test",
		Name:     "Ranking Test Goal",
		Priority: PriorityHigh,
		Status:   GoalActive,
	}

	strategySet, _ := planner.GenerateStrategies(context.Background(), goal)

	ranked := planner.RankStrategies(strategySet.ID)

	if ranked == nil {
		t.Fatal("Ranked strategies is nil")
	}

	if len(ranked) != len(strategySet.Strategies) {
		t.Fatalf("Ranking count mismatch: %d != %d", len(ranked), len(strategySet.Strategies))
	}

	// Verify descending order
	for i := 0; i < len(ranked)-1; i++ {
		if ranked[i].Effectiveness < ranked[i+1].Effectiveness {
			t.Fatalf("Strategies not properly ranked at index %d", i)
		}
	}

	t.Logf("Ranked %d strategies by effectiveness", len(ranked))
}

// TestMultiStrategyPlanner_ResourceBudget tests budget
func TestMultiStrategyPlanner_ResourceBudget(t *testing.T) {
	planner := NewMultiStrategyPlanner(DefaultMultiStrategyPlannerConfig())
	_ = planner.Initialize(nil)

	goal := &Goal{
		ID:       "budget-test",
		Name:     "Budget Test Goal",
		Priority: PriorityHigh,
		Status:   GoalActive,
	}

	strategySet, _ := planner.GenerateStrategies(context.Background(), goal)

	if len(strategySet.ResourceBudget) == 0 {
		t.Fatal("No resource budget allocated")
	}

	for resource, budget := range strategySet.ResourceBudget {
		if budget <= 0 {
			t.Fatalf("Invalid budget for %s: %.2f", resource, budget)
		}

		t.Logf("Resource %s: Budget=%.2f", resource, budget)
	}
}

// TestMultiStrategyPlanner_StrategyStatus tests status tracking
func TestMultiStrategyPlanner_StrategyStatus(t *testing.T) {
	planner := NewMultiStrategyPlanner(DefaultMultiStrategyPlannerConfig())
	_ = planner.Initialize(nil)

	goal := &Goal{
		ID:       "status-test",
		Name:     "Status Test Goal",
		Priority: PriorityHigh,
		Status:   GoalActive,
	}

	strategySet, _ := planner.GenerateStrategies(context.Background(), goal)

	validStatuses := map[StrategyStatus]bool{
		StrategyProposed:  true,
		StrategyEvaluated: true,
		StrategySelected:  true,
		StrategyRejected:  true,
		StrategyOptimized: true,
	}

	for _, strategy := range strategySet.Strategies {
		if !validStatuses[strategy.Status] {
			t.Fatalf("Invalid status: %s", strategy.Status)
		}

		t.Logf("Strategy %s: Status=%s", strategy.ID, strategy.Status)
	}
}

// TestMultiStrategyPlanner_GetStrategies tests retrieval
func TestMultiStrategyPlanner_GetStrategies(t *testing.T) {
	planner := NewMultiStrategyPlanner(DefaultMultiStrategyPlannerConfig())
	_ = planner.Initialize(nil)

	goal := &Goal{
		ID:       "retrieve-test",
		Name:     "Retrieve Test Goal",
		Priority: PriorityHigh,
		Status:   GoalActive,
	}

	strategySet, _ := planner.GenerateStrategies(context.Background(), goal)

	retrieved := planner.GetStrategies(strategySet.ID)

	if retrieved == nil {
		t.Fatal("Retrieved strategies is nil")
	}

	if len(retrieved) != len(strategySet.Strategies) {
		t.Fatalf("Count mismatch: %d != %d", len(retrieved), len(strategySet.Strategies))
	}

	t.Logf("Retrieved %d strategies", len(retrieved))
}

// TestMultiStrategyPlanner_GetMetrics tests metrics
func TestMultiStrategyPlanner_GetMetrics(t *testing.T) {
	planner := NewMultiStrategyPlanner(DefaultMultiStrategyPlannerConfig())
	_ = planner.Initialize(nil)

	goal := &Goal{
		ID:       "metrics-test",
		Name:     "Metrics Test Goal",
		Priority: PriorityHigh,
		Status:   GoalActive,
	}

	_, _ = planner.GenerateStrategies(context.Background(), goal)

	metrics := planner.GetMetrics()

	if metrics.ComponentName != "MultiStrategyPlanner" {
		t.Fatalf("Wrong component name: %s", metrics.ComponentName)
	}

	if metrics.TotalRequests <= 0 {
		t.Fatal("TotalRequests should be positive")
	}

	t.Logf("Metrics: Requests=%d, Successful=%d, Failed=%d",
		metrics.TotalRequests, metrics.SuccessfulRequests, metrics.FailedRequests)
}

// TestMultiStrategyPlanner_Shutdown tests shutdown
func TestMultiStrategyPlanner_Shutdown(t *testing.T) {
	planner := NewMultiStrategyPlanner(DefaultMultiStrategyPlannerConfig())
	_ = planner.Initialize(nil)

	goal := &Goal{
		ID:       "shutdown-test",
		Name:     "Shutdown Test Goal",
		Priority: PriorityHigh,
		Status:   GoalActive,
	}

	strategySet, _ := planner.GenerateStrategies(context.Background(), goal)
	setID := strategySet.ID

	if planner.GetStrategies(setID) == nil {
		t.Fatal("Strategies should exist before shutdown")
	}

	err := planner.Shutdown()
	if err != nil {
		t.Fatalf("Shutdown failed: %v", err)
	}

	if planner.GetStrategies(setID) != nil {
		t.Fatal("Strategies should be cleared after shutdown")
	}

	t.Log("Shutdown successful")
}

// TestMultiStrategyPlanner_MultipleSets tests multiple sets
func TestMultiStrategyPlanner_MultipleSets(t *testing.T) {
	planner := NewMultiStrategyPlanner(DefaultMultiStrategyPlannerConfig())
	_ = planner.Initialize(nil)

	goals := []*Goal{
		{ID: "goal-1", Name: "Goal 1", Priority: PriorityHigh, Status: GoalActive},
		{ID: "goal-2", Name: "Goal 2", Priority: PriorityHigh, Status: GoalActive},
		{ID: "goal-3", Name: "Goal 3", Priority: PriorityHigh, Status: GoalActive},
	}

	for _, goal := range goals {
		_, err := planner.GenerateStrategies(context.Background(), goal)
		if err != nil {
			t.Fatalf("Generation failed for %s: %v", goal.ID, err)
		}
	}

	t.Logf("Generated strategies for %d goals", len(goals))
}

// BenchmarkMultiStrategyPlanner_GenerateStrategies benchmarks generation
func BenchmarkMultiStrategyPlanner_GenerateStrategies(b *testing.B) {
	planner := NewMultiStrategyPlanner(DefaultMultiStrategyPlannerConfig())
	_ = planner.Initialize(nil)

	goal := &Goal{
		ID:           "bench-goal",
		Name:         "Benchmark Goal",
		Priority:     PriorityHigh,
		Status:       GoalActive,
		Dependencies: []string{"dep1", "dep2"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = planner.GenerateStrategies(context.Background(), goal)
	}
}

// BenchmarkMultiStrategyPlanner_RankStrategies benchmarks ranking
func BenchmarkMultiStrategyPlanner_RankStrategies(b *testing.B) {
	planner := NewMultiStrategyPlanner(DefaultMultiStrategyPlannerConfig())
	_ = planner.Initialize(nil)

	goal := &Goal{
		ID:           "bench-rank",
		Name:         "Benchmark Rank",
		Priority:     PriorityHigh,
		Status:       GoalActive,
		Dependencies: []string{"dep1", "dep2"},
	}

	strategySet, _ := planner.GenerateStrategies(context.Background(), goal)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = planner.RankStrategies(strategySet.ID)
	}
}
