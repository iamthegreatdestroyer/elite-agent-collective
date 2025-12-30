// Package memory provides the MNEMONIC system for the Elite Agent Collective.
// This file contains tests for the Scientific Hypothesis Generator.

package memory

import (
	"context"
	"testing"
)

// ============================================================================
// Scientific Hypothesis Generator Tests
// ============================================================================

// TestScientificHypothesisGenerator_Initialization tests setup
func TestScientificHypothesisGenerator_Initialization(t *testing.T) {
	generator := NewScientificHypothesisGenerator(DefaultScientificHypothesisGeneratorConfig())
	err := generator.Initialize(nil)

	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	metrics := generator.GetMetrics()
	if metrics.ComponentName != "ScientificHypothesisGenerator" {
		t.Fatalf("Wrong component name: %s", metrics.ComponentName)
	}

	t.Log("Scientific hypothesis generator initialized successfully")
}

// TestScientificHypothesisGenerator_GenerateHypotheses tests generation
func TestScientificHypothesisGenerator_GenerateHypotheses(t *testing.T) {
	generator := NewScientificHypothesisGenerator(DefaultScientificHypothesisGeneratorConfig())
	_ = generator.Initialize(nil)

	goal := &Goal{
		ID:           "hyp-goal",
		Name:         "Hypothesis Test Goal",
		Priority:     PriorityHigh,
		Status:       GoalActive,
		Dependencies: []string{"dep1", "dep2"},
	}

	hypothesisSet, err := generator.GenerateHypotheses(context.Background(), goal)

	if err != nil {
		t.Fatalf("Generation failed: %v", err)
	}

	if hypothesisSet == nil {
		t.Fatal("HypothesisSet is nil")
	}

	if len(hypothesisSet.Hypotheses) == 0 {
		t.Fatal("No hypotheses generated")
	}

	t.Logf("Created %d hypotheses", len(hypothesisSet.Hypotheses))
}

// TestScientificHypothesisGenerator_HypothesisCount tests count
func TestScientificHypothesisGenerator_HypothesisCount(t *testing.T) {
	generator := NewScientificHypothesisGenerator(DefaultScientificHypothesisGeneratorConfig())
	_ = generator.Initialize(nil)

	goal := &Goal{
		ID:           "count-test",
		Name:         "Count Test Goal",
		Priority:     PriorityHigh,
		Status:       GoalActive,
		Dependencies: []string{"dep1", "dep2", "dep3"},
	}

	hypothesisSet, _ := generator.GenerateHypotheses(context.Background(), goal)

	minExpected := 5
	maxExpected := 15

	if len(hypothesisSet.Hypotheses) < minExpected {
		t.Fatalf("Too few hypotheses: %d < %d", len(hypothesisSet.Hypotheses), minExpected)
	}

	if len(hypothesisSet.Hypotheses) > maxExpected {
		t.Fatalf("Too many hypotheses: %d > %d", len(hypothesisSet.Hypotheses), maxExpected)
	}

	t.Logf("Hypothesis count correct: %d (min: %d, max: %d)",
		len(hypothesisSet.Hypotheses), minExpected, maxExpected)
}

// TestScientificHypothesisGenerator_Evidence tests evidence
func TestScientificHypothesisGenerator_Evidence(t *testing.T) {
	generator := NewScientificHypothesisGenerator(DefaultScientificHypothesisGeneratorConfig())
	_ = generator.Initialize(nil)

	goal := &Goal{
		ID:       "evidence-test",
		Name:     "Evidence Test Goal",
		Priority: PriorityHigh,
		Status:   GoalActive,
	}

	hypothesisSet, _ := generator.GenerateHypotheses(context.Background(), goal)

	for _, hypothesis := range hypothesisSet.Hypotheses {
		if len(hypothesis.Evidence) == 0 {
			t.Fatalf("Hypothesis %s has no evidence", hypothesis.ID)
		}

		for _, ev := range hypothesis.Evidence {
			if ev.Strength < 0 || ev.Strength > 1.0 {
				t.Fatalf("Invalid evidence strength: %.2f", ev.Strength)
			}

			if ev.Confidence < 0 || ev.Confidence > 1.0 {
				t.Fatalf("Invalid evidence confidence: %.2f", ev.Confidence)
			}
		}

		t.Logf("Hypothesis %s: %d evidence pieces", hypothesis.ID, len(hypothesis.Evidence))
	}
}

// TestScientificHypothesisGenerator_Validation tests validation
func TestScientificHypothesisGenerator_Validation(t *testing.T) {
	generator := NewScientificHypothesisGenerator(DefaultScientificHypothesisGeneratorConfig())
	_ = generator.Initialize(nil)

	goal := &Goal{
		ID:       "validation-test",
		Name:     "Validation Test Goal",
		Priority: PriorityHigh,
		Status:   GoalActive,
	}

	hypothesisSet, _ := generator.GenerateHypotheses(context.Background(), goal)

	for _, hypothesis := range hypothesisSet.Hypotheses {
		if len(hypothesis.Validations) == 0 {
			t.Fatalf("Hypothesis %s has no validations", hypothesis.ID)
		}

		for _, val := range hypothesis.Validations {
			if val.Confidence < 0 || val.Confidence > 1.0 {
				t.Fatalf("Invalid validation confidence: %.2f", val.Confidence)
			}

			t.Logf("Hypothesis %s: Result=%s, Confidence=%.2f",
				hypothesis.ID, val.Result, val.Confidence)
		}
	}
}

// TestScientificHypothesisGenerator_Status tests status
func TestScientificHypothesisGenerator_Status(t *testing.T) {
	generator := NewScientificHypothesisGenerator(DefaultScientificHypothesisGeneratorConfig())
	_ = generator.Initialize(nil)

	goal := &Goal{
		ID:       "status-test",
		Name:     "Status Test Goal",
		Priority: PriorityHigh,
		Status:   GoalActive,
	}

	hypothesisSet, _ := generator.GenerateHypotheses(context.Background(), goal)

	validStatuses := map[ScientificStatus]bool{
		ScientificPending:    true,
		ScientificValidating: true,
		ScientificConfirmed:  true,
		ScientificRejected:   true,
		ScientificRefined:    true,
	}

	for _, hypothesis := range hypothesisSet.Hypotheses {
		if !validStatuses[hypothesis.Status] {
			t.Fatalf("Invalid status: %s", hypothesis.Status)
		}

		t.Logf("Hypothesis %s: Status=%s", hypothesis.ID, hypothesis.Status)
	}
}

// TestScientificHypothesisGenerator_BeliefState tests belief state
func TestScientificHypothesisGenerator_BeliefState(t *testing.T) {
	generator := NewScientificHypothesisGenerator(DefaultScientificHypothesisGeneratorConfig())
	_ = generator.Initialize(nil)

	goal := &Goal{
		ID:       "belief-test",
		Name:     "Belief Test Goal",
		Priority: PriorityHigh,
		Status:   GoalActive,
	}

	hypothesisSet, _ := generator.GenerateHypotheses(context.Background(), goal)
	beliefState := generator.GetBeliefState(hypothesisSet.ID)

	if beliefState == nil {
		t.Fatal("BeliefState is nil")
	}

	if len(beliefState) == 0 {
		t.Fatal("BeliefState is empty")
	}

	for hypothesisID, belief := range beliefState {
		if belief < 0 || belief > 1.0 {
			t.Fatalf("Invalid belief: %.2f", belief)
		}

		t.Logf("Belief for %s: %.2f", hypothesisID, belief)
	}
}

// TestScientificHypothesisGenerator_ValidationStats tests stats
func TestScientificHypothesisGenerator_ValidationStats(t *testing.T) {
	generator := NewScientificHypothesisGenerator(DefaultScientificHypothesisGeneratorConfig())
	_ = generator.Initialize(nil)

	goal := &Goal{
		ID:       "stats-test",
		Name:     "Stats Test Goal",
		Priority: PriorityHigh,
		Status:   GoalActive,
	}

	hypothesisSet, _ := generator.GenerateHypotheses(context.Background(), goal)

	stats := hypothesisSet.ValidationStats

	if stats.TotalHypotheses == 0 {
		t.Fatal("TotalHypotheses is 0")
	}

	if stats.AverageConfidence < 0 || stats.AverageConfidence > 1.0 {
		t.Fatalf("Invalid average confidence: %.2f", stats.AverageConfidence)
	}

	if stats.ConfirmationRate < 0 || stats.ConfirmationRate > 1.0 {
		t.Fatalf("Invalid confirmation rate: %.2f", stats.ConfirmationRate)
	}

	t.Logf("Stats: Total=%d, Confirmed=%d, Rejected=%d, Rate=%.2f%%",
		stats.TotalHypotheses, stats.ConfirmedCount, stats.RejectedCount,
		stats.ConfirmationRate*100)
}

// TestScientificHypothesisGenerator_RefineHypothesis tests refinement
func TestScientificHypothesisGenerator_RefineHypothesis(t *testing.T) {
	generator := NewScientificHypothesisGenerator(DefaultScientificHypothesisGeneratorConfig())
	_ = generator.Initialize(nil)

	goal := &Goal{
		ID:       "refine-test",
		Name:     "Refine Test Goal",
		Priority: PriorityHigh,
		Status:   GoalActive,
	}

	hypothesisSet, _ := generator.GenerateHypotheses(context.Background(), goal)
	hypothesis := hypothesisSet.Hypotheses[0]

	originalEvidenceCount := len(hypothesis.Evidence)

	newEvidence := &ScientificEvidence{
		ID:          "new-ev-1",
		Description: "New evidence for refinement",
		Type:        EvidenceExperimental,
		Strength:    0.9,
		Confidence:  0.85,
	}

	refined, err := generator.RefineHypothesis(hypothesis.ID, newEvidence)

	if err != nil {
		t.Fatalf("Refinement failed: %v", err)
	}

	if refined == nil {
		t.Fatal("Refined hypothesis is nil")
	}

	if len(refined.Evidence) != originalEvidenceCount+1 {
		t.Fatalf("Evidence not properly added: expected %d, got %d",
			originalEvidenceCount+1, len(refined.Evidence))
	}

	t.Logf("Refined hypothesis %s: Evidence count increased from %d to %d",
		refined.ID, originalEvidenceCount, len(refined.Evidence))
}

// TestScientificHypothesisGenerator_GetConfirmedHypotheses tests retrieval
func TestScientificHypothesisGenerator_GetConfirmedHypotheses(t *testing.T) {
	generator := NewScientificHypothesisGenerator(DefaultScientificHypothesisGeneratorConfig())
	_ = generator.Initialize(nil)

	goal := &Goal{
		ID:       "confirmed-test",
		Name:     "Confirmed Test Goal",
		Priority: PriorityHigh,
		Status:   GoalActive,
	}

	hypothesisSet, _ := generator.GenerateHypotheses(context.Background(), goal)

	confirmed := generator.GetConfirmedHypotheses(hypothesisSet.ID)

	if confirmed == nil {
		t.Fatal("Confirmed hypotheses is nil")
	}

	for _, hyp := range confirmed {
		if hyp.Status != ScientificConfirmed {
			t.Fatalf("Non-confirmed hypothesis returned: %s", hyp.Status)
		}
	}

	t.Logf("Found %d confirmed hypotheses", len(confirmed))
}

// TestScientificHypothesisGenerator_GetMetrics tests metrics
func TestScientificHypothesisGenerator_GetMetrics(t *testing.T) {
	generator := NewScientificHypothesisGenerator(DefaultScientificHypothesisGeneratorConfig())
	_ = generator.Initialize(nil)

	goal := &Goal{
		ID:       "metrics-test",
		Name:     "Metrics Test Goal",
		Priority: PriorityHigh,
		Status:   GoalActive,
	}

	_, _ = generator.GenerateHypotheses(context.Background(), goal)

	metrics := generator.GetMetrics()

	if metrics.ComponentName != "ScientificHypothesisGenerator" {
		t.Fatalf("Wrong component name: %s", metrics.ComponentName)
	}

	if metrics.TotalRequests <= 0 {
		t.Fatal("TotalRequests should be positive")
	}

	t.Logf("Metrics: Requests=%d, Successful=%d, Failed=%d",
		metrics.TotalRequests, metrics.SuccessfulRequests, metrics.FailedRequests)
}

// TestScientificHypothesisGenerator_Shutdown tests shutdown
func TestScientificHypothesisGenerator_Shutdown(t *testing.T) {
	generator := NewScientificHypothesisGenerator(DefaultScientificHypothesisGeneratorConfig())
	_ = generator.Initialize(nil)

	goal := &Goal{
		ID:       "shutdown-test",
		Name:     "Shutdown Test Goal",
		Priority: PriorityHigh,
		Status:   GoalActive,
	}

	hypothesisSet, _ := generator.GenerateHypotheses(context.Background(), goal)
	setID := hypothesisSet.ID

	if generator.GetBeliefState(setID) == nil {
		t.Fatal("BeliefState should exist before shutdown")
	}

	err := generator.Shutdown()
	if err != nil {
		t.Fatalf("Shutdown failed: %v", err)
	}

	if generator.GetBeliefState(setID) != nil {
		t.Fatal("BeliefState should be cleared after shutdown")
	}

	t.Log("Shutdown successful")
}

// TestScientificHypothesisGenerator_ConfidenceLevels tests confidence
func TestScientificHypothesisGenerator_ConfidenceLevels(t *testing.T) {
	generator := NewScientificHypothesisGenerator(DefaultScientificHypothesisGeneratorConfig())
	_ = generator.Initialize(nil)

	goal := &Goal{
		ID:           "confidence-test",
		Name:         "Confidence Test Goal",
		Priority:     PriorityHigh,
		Status:       GoalActive,
		Dependencies: []string{"dep1", "dep2", "dep3"},
	}

	hypothesisSet, _ := generator.GenerateHypotheses(context.Background(), goal)

	for _, hypothesis := range hypothesisSet.Hypotheses {
		if hypothesis.Confidence < 0.6 || hypothesis.Confidence > 1.0 {
			t.Fatalf("Invalid confidence: %.2f", hypothesis.Confidence)
		}

		t.Logf("Hypothesis %s: Confidence=%.2f", hypothesis.ID, hypothesis.Confidence)
	}
}

// BenchmarkScientificHypothesisGenerator_GenerateHypotheses benchmarks generation
func BenchmarkScientificHypothesisGenerator_GenerateHypotheses(b *testing.B) {
	generator := NewScientificHypothesisGenerator(DefaultScientificHypothesisGeneratorConfig())
	_ = generator.Initialize(nil)

	goal := &Goal{
		ID:           "bench-goal",
		Name:         "Benchmark Goal",
		Priority:     PriorityHigh,
		Status:       GoalActive,
		Dependencies: []string{"dep1", "dep2"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = generator.GenerateHypotheses(context.Background(), goal)
	}
}

// BenchmarkScientificHypothesisGenerator_RefineHypothesis benchmarks refinement
func BenchmarkScientificHypothesisGenerator_RefineHypothesis(b *testing.B) {
	generator := NewScientificHypothesisGenerator(DefaultScientificHypothesisGeneratorConfig())
	_ = generator.Initialize(nil)

	goal := &Goal{
		ID:           "bench-refine",
		Name:         "Benchmark Refine",
		Priority:     PriorityHigh,
		Status:       GoalActive,
		Dependencies: []string{"dep1", "dep2"},
	}

	hypothesisSet, _ := generator.GenerateHypotheses(context.Background(), goal)
	hypothesis := hypothesisSet.Hypotheses[0]

	newEvidence := &ScientificEvidence{
		ID:          "bench-ev",
		Description: "Benchmark evidence",
		Type:        EvidenceExperimental,
		Strength:    0.9,
		Confidence:  0.85,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = generator.RefineHypothesis(hypothesis.ID, newEvidence)
	}
}
