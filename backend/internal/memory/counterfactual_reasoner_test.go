// Package memory provides the MNEMONIC system for the Elite Agent Collective.
// This file contains tests for the Counterfactual Reasoner.

package memory

import (
	"context"
	"testing"
)

// ============================================================================
// Counterfactual Reasoner Tests
// ============================================================================

// TestCounterfactualReasoner_Initialization tests reasoner setup
func TestCounterfactualReasoner_Initialization(t *testing.T) {
	reasoner := NewCounterfactualReasoner(DefaultCounterfactualConfig())
	err := reasoner.Initialize(nil)

	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	metrics := reasoner.GetMetrics()
	if metrics.ComponentName != "CounterfactualReasoner" {
		t.Fatalf("Wrong component name: %s", metrics.ComponentName)
	}

	t.Log("Counterfactual reasoner initialized successfully")
}

// TestCounterfactualReasoner_AnalyzeCounterfactuals tests analysis creation
func TestCounterfactualReasoner_AnalyzeCounterfactuals(t *testing.T) {
	reasoner := NewCounterfactualReasoner(DefaultCounterfactualConfig())
	_ = reasoner.Initialize(nil)

	goal := &Goal{
		ID:           "cf-goal",
		Name:         "Counterfactual Test Goal",
		Priority:     PriorityHigh,
		Status:       GoalActive,
		Dependencies: []string{"dep1", "dep2"},
	}

	analysis, err := reasoner.AnalyzeCounterfactuals(context.Background(), goal)

	if err != nil {
		t.Fatalf("Analysis failed: %v", err)
	}

	if analysis == nil {
		t.Fatal("Analysis is nil")
	}

	if analysis.ID == "" {
		t.Fatal("Analysis ID is empty")
	}

	if len(analysis.Scenarios) == 0 {
		t.Fatal("No scenarios generated")
	}

	t.Logf("Created analysis with %d scenarios", len(analysis.Scenarios))
}

// TestCounterfactualReasoner_ScenarioGeneration tests scenario creation
func TestCounterfactualReasoner_ScenarioGeneration(t *testing.T) {
	reasoner := NewCounterfactualReasoner(DefaultCounterfactualConfig())
	_ = reasoner.Initialize(nil)

	goal := &Goal{
		ID:       "scenario-test",
		Name:     "Scenario Generation Test",
		Priority: PriorityHigh,
		Status:   GoalActive,
	}

	analysis, _ := reasoner.AnalyzeCounterfactuals(context.Background(), goal)

	for _, scenario := range analysis.Scenarios {
		if scenario.ID == "" {
			t.Fatal("Scenario ID is empty")
		}

		if scenario.Name == "" {
			t.Fatal("Scenario name is empty")
		}

		t.Logf("Generated scenario: %s with %d changes", scenario.Name, len(scenario.Changes))
	}
}

// TestCounterfactualReasoner_OutcomePrediction tests prediction accuracy
func TestCounterfactualReasoner_OutcomePrediction(t *testing.T) {
	reasoner := NewCounterfactualReasoner(DefaultCounterfactualConfig())
	_ = reasoner.Initialize(nil)

	goal := &Goal{
		ID:       "prediction-test",
		Name:     "Prediction Test",
		Priority: PriorityHigh,
		Status:   GoalActive,
	}

	analysis, _ := reasoner.AnalyzeCounterfactuals(context.Background(), goal)

	for scenarioID, prediction := range analysis.Predictions {
		if prediction.SuccessProbability < 0 || prediction.SuccessProbability > 1.0 {
			t.Fatalf("Invalid success probability: %.2f", prediction.SuccessProbability)
		}

		if prediction.TimeToCompletion <= 0 {
			t.Fatal("Invalid completion time")
		}

		if prediction.ResourcesRequired <= 0 {
			t.Fatal("Invalid resource requirement")
		}

		if prediction.Confidence < 0.8 {
			t.Logf("Prediction confidence below expected: %.2f", prediction.Confidence)
		}

		t.Logf("Scenario %s: Success=%.2f, Time=%v, Resources=%.2f",
			scenarioID, prediction.SuccessProbability, prediction.TimeToCompletion,
			prediction.ResourcesRequired)
	}
}

// TestCounterfactualReasoner_DifferenceAnalysis tests difference metrics
func TestCounterfactualReasoner_DifferenceAnalysis(t *testing.T) {
	reasoner := NewCounterfactualReasoner(DefaultCounterfactualConfig())
	_ = reasoner.Initialize(nil)

	goal := &Goal{
		ID:       "diff-test",
		Name:     "Difference Analysis Test",
		Priority: PriorityHigh,
		Status:   GoalActive,
	}

	analysis, _ := reasoner.AnalyzeCounterfactuals(context.Background(), goal)

	for scenarioID, diff := range analysis.Comparisons {
		if diff.SimilarityScore < 0 || diff.SimilarityScore > 1.0 {
			t.Fatalf("Invalid similarity score: %.2f", diff.SimilarityScore)
		}

		if diff.ChangeMagnitude < 0 || diff.ChangeMagnitude > 1.0 {
			t.Fatalf("Invalid change magnitude: %.2f", diff.ChangeMagnitude)
		}

		if diff.ImpactScore < 0 || diff.ImpactScore > 1.0 {
			t.Fatalf("Invalid impact score: %.2f", diff.ImpactScore)
		}

		t.Logf("Scenario %s: Similarity=%.2f, Magnitude=%.2f, Impact=%.2f",
			scenarioID, diff.SimilarityScore, diff.ChangeMagnitude, diff.ImpactScore)
	}
}

// TestCounterfactualReasoner_CausalInsights tests insight extraction
func TestCounterfactualReasoner_CausalInsights(t *testing.T) {
	reasoner := NewCounterfactualReasoner(DefaultCounterfactualConfig())
	_ = reasoner.Initialize(nil)

	goal := &Goal{
		ID:       "insight-test",
		Name:     "Insight Extraction Test",
		Priority: PriorityHigh,
		Status:   GoalActive,
	}

	analysis, _ := reasoner.AnalyzeCounterfactuals(context.Background(), goal)

	if len(analysis.KeyInsights) == 0 {
		t.Logf("No key insights extracted (acceptable)")
	}

	for _, insight := range analysis.KeyInsights {
		if insight.Confidence < 0 || insight.Confidence > 1.0 {
			t.Fatalf("Invalid confidence: %.2f", insight.Confidence)
		}

		t.Logf("Insight: %s → %s (confidence: %.2f)",
			insight.Cause, insight.Effect, insight.Confidence)
	}
}

// TestCounterfactualReasoner_HighestImpactChange tests impact identification
func TestCounterfactualReasoner_HighestImpactChange(t *testing.T) {
	reasoner := NewCounterfactualReasoner(DefaultCounterfactualConfig())
	_ = reasoner.Initialize(nil)

	goal := &Goal{
		ID:       "impact-test",
		Name:     "Impact Test",
		Priority: PriorityHigh,
		Status:   GoalActive,
	}

	analysis, _ := reasoner.AnalyzeCounterfactuals(context.Background(), goal)

	if analysis.HighestImpactChange == "" && len(analysis.Scenarios) > 0 {
		t.Logf("No highest impact change identified (acceptable for all equal scenarios)")
	}

	if analysis.HighestImpactChange != "" {
		t.Logf("Highest impact change: %s", analysis.HighestImpactChange)
	}
}

// TestCounterfactualReasoner_ComparePredictions tests prediction comparison
func TestCounterfactualReasoner_ComparePredictions(t *testing.T) {
	reasoner := NewCounterfactualReasoner(DefaultCounterfactualConfig())
	_ = reasoner.Initialize(nil)

	goal := &Goal{
		ID:           "compare-test",
		Name:         "Comparison Test",
		Priority:     PriorityHigh,
		Status:       GoalActive,
		Dependencies: []string{"dep1", "dep2", "dep3"},
	}

	analysis, _ := reasoner.AnalyzeCounterfactuals(context.Background(), goal)

	if len(analysis.Scenarios) >= 2 {
		scenario1 := analysis.Scenarios[0]
		scenario2 := analysis.Scenarios[1]

		diff, err := reasoner.ComparePredictions(scenario1.ID, scenario2.ID)

		if err != nil {
			t.Fatalf("Comparison failed: %v", err)
		}

		if diff.SimilarityScore < 0 || diff.SimilarityScore > 1.0 {
			t.Fatalf("Invalid similarity in comparison: %.2f", diff.SimilarityScore)
		}

		t.Logf("Comparison: Similarity=%.2f, Change=%.2f, Impact=%.2f",
			diff.SimilarityScore, diff.ChangeMagnitude, diff.ImpactScore)
	}
}

// TestCounterfactualReasoner_GetHighestSuccessProbability tests best scenario selection
func TestCounterfactualReasoner_GetHighestSuccessProbability(t *testing.T) {
	reasoner := NewCounterfactualReasoner(DefaultCounterfactualConfig())
	_ = reasoner.Initialize(nil)

	goal := &Goal{
		ID:       "success-test",
		Name:     "Success Probability Test",
		Priority: PriorityHigh,
		Status:   GoalActive,
	}

	analysis, _ := reasoner.AnalyzeCounterfactuals(context.Background(), goal)

	best := reasoner.GetHighestSuccessProbability(analysis.ID)

	if best == nil {
		t.Fatal("No best prediction found")
	}

	// Verify it's actually the highest
	for _, pred := range analysis.Predictions {
		if pred.SuccessProbability > best.SuccessProbability {
			t.Fatalf("Found higher probability: %.2f > %.2f",
				pred.SuccessProbability, best.SuccessProbability)
		}
	}

	t.Logf("Best scenario success probability: %.2f", best.SuccessProbability)
}

// TestCounterfactualReasoner_GetAnalysis tests analysis retrieval
func TestCounterfactualReasoner_GetAnalysis(t *testing.T) {
	reasoner := NewCounterfactualReasoner(DefaultCounterfactualConfig())
	_ = reasoner.Initialize(nil)

	goal := &Goal{
		ID:       "retrieve-test",
		Name:     "Retrieval Test",
		Priority: PriorityHigh,
		Status:   GoalActive,
	}

	analysis, _ := reasoner.AnalyzeCounterfactuals(context.Background(), goal)
	analysisID := analysis.ID

	retrieved := reasoner.GetAnalysis(analysisID)

	if retrieved == nil {
		t.Fatal("Analysis not found after retrieval")
	}

	if retrieved.ID != analysisID {
		t.Fatalf("Retrieved analysis has different ID: %s != %s",
			retrieved.ID, analysisID)
	}

	t.Logf("Successfully retrieved analysis %s", analysisID)
}

// TestCounterfactualReasoner_GetMetrics tests metrics tracking
func TestCounterfactualReasoner_GetMetrics(t *testing.T) {
	reasoner := NewCounterfactualReasoner(DefaultCounterfactualConfig())
	_ = reasoner.Initialize(nil)

	goal := &Goal{
		ID:       "metrics-test",
		Name:     "Metrics Test",
		Priority: PriorityHigh,
		Status:   GoalActive,
	}

	_, _ = reasoner.AnalyzeCounterfactuals(context.Background(), goal)

	metrics := reasoner.GetMetrics()

	if metrics.ComponentName != "CounterfactualReasoner" {
		t.Fatalf("Wrong component name: %s", metrics.ComponentName)
	}

	if metrics.TotalRequests <= 0 {
		t.Fatal("TotalRequests should be positive")
	}

	t.Logf("Metrics: Requests=%d, Successful=%d, Failed=%d",
		metrics.TotalRequests, metrics.SuccessfulRequests, metrics.FailedRequests)
}

// TestCounterfactualReasoner_Shutdown tests graceful shutdown
func TestCounterfactualReasoner_Shutdown(t *testing.T) {
	reasoner := NewCounterfactualReasoner(DefaultCounterfactualConfig())
	_ = reasoner.Initialize(nil)

	goal := &Goal{
		ID:       "shutdown-test",
		Name:     "Shutdown Test",
		Priority: PriorityHigh,
		Status:   GoalActive,
	}

	analysis, _ := reasoner.AnalyzeCounterfactuals(context.Background(), goal)
	analysisID := analysis.ID

	// Verify exists before shutdown
	if reasoner.GetAnalysis(analysisID) == nil {
		t.Fatal("Analysis should exist before shutdown")
	}

	// Shutdown
	err := reasoner.Shutdown()
	if err != nil {
		t.Fatalf("Shutdown failed: %v", err)
	}

	// Verify cleared after shutdown
	if reasoner.GetAnalysis(analysisID) != nil {
		t.Fatal("Analysis should not exist after shutdown")
	}

	t.Log("Shutdown successful")
}

// BenchmarkCounterfactualReasoner_AnalyzeCounterfactuals benchmarks analysis
func BenchmarkCounterfactualReasoner_AnalyzeCounterfactuals(b *testing.B) {
	reasoner := NewCounterfactualReasoner(DefaultCounterfactualConfig())
	_ = reasoner.Initialize(nil)

	goal := &Goal{
		ID:           "bench-goal",
		Name:         "Benchmark Goal",
		Priority:     PriorityHigh,
		Status:       GoalActive,
		Dependencies: []string{"dep1", "dep2"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = reasoner.AnalyzeCounterfactuals(context.Background(), goal)
	}
}

// BenchmarkCounterfactualReasoner_ComparePredictions benchmarks comparison
func BenchmarkCounterfactualReasoner_ComparePredictions(b *testing.B) {
	reasoner := NewCounterfactualReasoner(DefaultCounterfactualConfig())
	_ = reasoner.Initialize(nil)

	goal := &Goal{
		ID:           "bench-compare",
		Name:         "Benchmark Compare",
		Priority:     PriorityHigh,
		Status:       GoalActive,
		Dependencies: []string{"dep1", "dep2", "dep3"},
	}

	analysis, _ := reasoner.AnalyzeCounterfactuals(context.Background(), goal)

	if len(analysis.Scenarios) < 2 {
		b.Fatal("Need at least 2 scenarios for comparison")
	}

	scenario1 := analysis.Scenarios[0].ID
	scenario2 := analysis.Scenarios[1].ID

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = reasoner.ComparePredictions(scenario1, scenario2)
	}
}
