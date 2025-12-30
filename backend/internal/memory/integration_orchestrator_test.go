package memory

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// TestAdvancedIntegrator_Initialization tests component initialization
func TestAdvancedIntegrator_Initialization(t *testing.T) {
	config := DefaultIntegrationConfig()
	integrator := NewAdvancedIntegrator(config)

	if integrator == nil {
		t.Fatal("Expected non-nil integrator")
	}

	err := integrator.Initialize(nil)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	metrics := integrator.GetMetrics()
	if metrics.ComponentName != "AdvancedIntegrator" {
		t.Errorf("Expected component name AdvancedIntegrator, got %s", metrics.ComponentName)
	}

	t.Logf("Initialized successfully: %v", metrics.CustomMetrics)
}

// TestAdvancedIntegrator_ProcessRequest tests complete request processing
func TestAdvancedIntegrator_ProcessRequest(t *testing.T) {
	integrator := NewAdvancedIntegrator(DefaultIntegrationConfig())
	err := integrator.Initialize(nil)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	goal := &Goal{
		ID:          "test-goal",
		Name:        "System Optimization",
		Description: "Optimize system performance",
		Priority:    PriorityHigh,
	}

	request := &IntegrationRequest{
		ID:                   "req-001",
		Goal:                 goal,
		RequireAllComponents: true,
		Timestamp:            time.Now(),
	}

	result, err := integrator.ProcessRequest(context.Background(), request)
	if err != nil {
		t.Fatalf("ProcessRequest failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if !result.Success {
		t.Error("Expected successful result")
	}

	t.Logf("Request processed successfully in %v", result.ExecutionTime)
}

// TestAdvancedIntegrator_ComponentResults tests all component outputs
func TestAdvancedIntegrator_ComponentResults(t *testing.T) {
	integrator := NewAdvancedIntegrator(DefaultIntegrationConfig())
	err := integrator.Initialize(nil)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	goal := &Goal{
		ID:          "comp-test",
		Name:        "Feature Development",
		Description: "Develop new feature",
		Priority:    PriorityNormal,
	}

	request := &IntegrationRequest{
		ID:                   "req-comp",
		Goal:                 goal,
		RequireAllComponents: true,
		Timestamp:            time.Now(),
	}

	result, err := integrator.ProcessRequest(context.Background(), request)
	if err != nil {
		t.Fatalf("ProcessRequest failed: %v", err)
	}

	// Check plans
	if len(result.Plans) == 0 {
		t.Error("Expected plans to be generated")
	}
	t.Logf("Plans generated: %d", len(result.Plans))

	// Check scenarios
	if len(result.Scenarios) == 0 {
		t.Error("Expected scenarios to be generated")
	}
	t.Logf("Scenarios generated: %d", len(result.Scenarios))

	// Check hypotheses
	if len(result.Hypotheses) == 0 {
		t.Error("Expected hypotheses to be generated")
	}
	t.Logf("Hypotheses generated: %d", len(result.Hypotheses))

	// Check strategies
	if len(result.Strategies) == 0 {
		t.Error("Expected strategies to be generated")
	}
	t.Logf("Strategies generated: %d", len(result.Strategies))

	// Check selected strategy
	if result.SelectedStrategy == nil {
		t.Error("Expected selected strategy")
	} else {
		t.Logf("Selected strategy: %s", result.SelectedStrategy.Name)
	}
}

// TestAdvancedIntegrator_DecisionSynthesis tests decision creation
func TestAdvancedIntegrator_DecisionSynthesis(t *testing.T) {
	integrator := NewAdvancedIntegrator(DefaultIntegrationConfig())
	err := integrator.Initialize(nil)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	goal := &Goal{
		ID:          "decision-test",
		Name:        "Critical Decision",
		Description: "Make important decision",
		Priority:    PriorityHigh,
	}

	request := &IntegrationRequest{
		ID:   "req-decision",
		Goal: goal,
	}

	result, err := integrator.ProcessRequest(context.Background(), request)
	if err != nil {
		t.Fatalf("ProcessRequest failed: %v", err)
	}

	// Check decision
	if result.Decision == nil {
		t.Fatal("Expected decision to be generated")
	}

	decision := result.Decision
	t.Logf("Decision ID: %s", decision.ID)
	t.Logf("Recommendation: %s", decision.Recommendation)
	t.Logf("Confidence: %.2f", decision.Confidence)

	if decision.Recommendation == "" {
		t.Error("Expected non-empty recommendation")
	}

	if decision.Confidence < 0 || decision.Confidence > 1 {
		t.Errorf("Invalid confidence: %.2f", decision.Confidence)
	}

	if len(decision.Reasoning) == 0 {
		t.Error("Expected reasoning to be provided")
	}

	t.Logf("Reasoning points: %d", len(decision.Reasoning))
}

// TestAdvancedIntegrator_RiskAssessment tests risk analysis
func TestAdvancedIntegrator_RiskAssessment(t *testing.T) {
	integrator := NewAdvancedIntegrator(DefaultIntegrationConfig())
	err := integrator.Initialize(nil)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	goal := &Goal{
		ID:          "risk-test",
		Name:        "High Risk Project",
		Description: "Project with significant risks",
		Priority:    PriorityHigh,
	}

	request := &IntegrationRequest{
		ID:   "req-risk",
		Goal: goal,
	}

	result, err := integrator.ProcessRequest(context.Background(), request)
	if err != nil {
		t.Fatalf("ProcessRequest failed: %v", err)
	}

	if result.Decision == nil || result.Decision.RiskAssessment == nil {
		t.Fatal("Expected risk assessment")
	}

	risk := result.Decision.RiskAssessment
	t.Logf("Overall risk: %.2f", risk.OverallRisk)
	t.Logf("Risks identified: %d", len(risk.Risks))
	t.Logf("Mitigation steps: %d", len(risk.MitigationSteps))

	if risk.OverallRisk < 0 || risk.OverallRisk > 1 {
		t.Errorf("Invalid overall risk: %.2f", risk.OverallRisk)
	}

	if len(risk.Risks) == 0 {
		t.Error("Expected risks to be identified")
	}

	if len(risk.MitigationSteps) == 0 {
		t.Error("Expected mitigation steps")
	}
}

// TestAdvancedIntegrator_Alternatives tests alternative options
func TestAdvancedIntegrator_Alternatives(t *testing.T) {
	integrator := NewAdvancedIntegrator(DefaultIntegrationConfig())
	err := integrator.Initialize(nil)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	goal := &Goal{
		ID:          "alt-test",
		Name:        "Multi-Option Decision",
		Description: "Decision with multiple alternatives",
		Priority:    PriorityNormal,
	}

	request := &IntegrationRequest{
		ID:   "req-alt",
		Goal: goal,
	}

	result, err := integrator.ProcessRequest(context.Background(), request)
	if err != nil {
		t.Fatalf("ProcessRequest failed: %v", err)
	}

	if result.Decision == nil {
		t.Fatal("Expected decision")
	}

	alternatives := result.Decision.Alternatives
	t.Logf("Alternatives generated: %d", len(alternatives))

	for i, alt := range alternatives {
		t.Logf("Alternative %d: %s (Confidence: %.2f)", i+1, alt.Name, alt.Confidence)

		if alt.Name == "" {
			t.Errorf("Alternative %d has empty name", i)
		}

		if alt.Confidence < 0 || alt.Confidence > 1 {
			t.Errorf("Alternative %d has invalid confidence: %.2f", i, alt.Confidence)
		}
	}
}

// TestAdvancedIntegrator_FormattedOutput tests output formatting
func TestAdvancedIntegrator_FormattedOutput(t *testing.T) {
	integrator := NewAdvancedIntegrator(DefaultIntegrationConfig())
	err := integrator.Initialize(nil)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	goal := &Goal{
		ID:          "output-test",
		Name:        "Output Formatting Test",
		Description: "Test output formatting",
		Priority:    PriorityLow,
	}

	request := &IntegrationRequest{
		ID:   "req-output",
		Goal: goal,
	}

	result, err := integrator.ProcessRequest(context.Background(), request)
	if err != nil {
		t.Fatalf("ProcessRequest failed: %v", err)
	}

	if result.FormattedOutput == nil {
		t.Fatal("Expected formatted output")
	}

	output := result.FormattedOutput
	t.Logf("Summary: %s", output.Summary)
	t.Logf("Confidence Level: %s", output.ConfidenceLevel)
	t.Logf("Key Insights: %d", len(output.KeyInsights))
	t.Logf("Action Items: %d", len(output.ActionItems))

	if output.Summary == "" {
		t.Error("Expected non-empty summary")
	}

	if output.DetailedAnalysis == "" {
		t.Error("Expected detailed analysis")
	}

	if len(output.RecommendationsList) == 0 {
		t.Error("Expected recommendations list")
	}

	if output.ConfidenceLevel == "" {
		t.Error("Expected confidence level")
	}
}

// TestAdvancedIntegrator_ExecutionTime tests performance
func TestAdvancedIntegrator_ExecutionTime(t *testing.T) {
	integrator := NewAdvancedIntegrator(DefaultIntegrationConfig())
	err := integrator.Initialize(nil)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	goal := &Goal{
		ID:          "perf-test",
		Name:        "Performance Test",
		Description: "Test execution speed",
		Priority:    PriorityNormal,
	}

	request := &IntegrationRequest{
		ID:   "req-perf",
		Goal: goal,
	}

	result, err := integrator.ProcessRequest(context.Background(), request)
	if err != nil {
		t.Fatalf("ProcessRequest failed: %v", err)
	}

	t.Logf("Execution time: %v", result.ExecutionTime)

	// Should complete reasonably quickly
	if result.ExecutionTime > 5*time.Second {
		t.Errorf("Execution took too long: %v", result.ExecutionTime)
	}
}

// TestAdvancedIntegrator_GetResult tests result retrieval
func TestAdvancedIntegrator_GetResult(t *testing.T) {
	integrator := NewAdvancedIntegrator(DefaultIntegrationConfig())
	err := integrator.Initialize(nil)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	goal := &Goal{
		ID:          "retrieve-test",
		Name:        "Result Retrieval Test",
		Description: "Test result retrieval",
		Priority:    PriorityLow,
	}

	request := &IntegrationRequest{
		ID:   "req-retrieve",
		Goal: goal,
	}

	result, err := integrator.ProcessRequest(context.Background(), request)
	if err != nil {
		t.Fatalf("ProcessRequest failed: %v", err)
	}

	// Retrieve result
	retrieved := integrator.GetResult(result.ID)
	if retrieved == nil {
		t.Fatal("Expected to retrieve result")
	}

	if retrieved.ID != result.ID {
		t.Errorf("Retrieved wrong result: got %s, want %s", retrieved.ID, result.ID)
	}

	t.Logf("Successfully retrieved result: %s", retrieved.ID)
}

// TestAdvancedIntegrator_Metrics tests metrics tracking
func TestAdvancedIntegrator_Metrics(t *testing.T) {
	integrator := NewAdvancedIntegrator(DefaultIntegrationConfig())
	err := integrator.Initialize(nil)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Process multiple requests
	for i := 0; i < 3; i++ {
		goal := &Goal{
			ID:          fmt.Sprintf("metrics-test-%d", i),
			Name:        fmt.Sprintf("Metrics Test %d", i),
			Description: "Test metrics tracking",
			Priority:    PriorityLow,
		}

		request := &IntegrationRequest{
			ID:   fmt.Sprintf("req-metrics-%d", i),
			Goal: goal,
		}

		_, err := integrator.ProcessRequest(context.Background(), request)
		if err != nil {
			t.Fatalf("ProcessRequest %d failed: %v", i, err)
		}
	}

	metrics := integrator.GetMetrics()
	t.Logf("Total Requests: %d", metrics.TotalRequests)
	t.Logf("Successful: %d", metrics.SuccessfulRequests)
	t.Logf("Failed: %d", metrics.FailedRequests)
	t.Logf("Average Latency: %v", metrics.AverageLatency)

	if metrics.TotalRequests != 3 {
		t.Errorf("Expected 3 total requests, got %d", metrics.TotalRequests)
	}

	if metrics.SuccessfulRequests != 3 {
		t.Errorf("Expected 3 successful requests, got %d", metrics.SuccessfulRequests)
	}
}

// TestAdvancedIntegrator_ComponentDisabling tests selective component execution
func TestAdvancedIntegrator_ComponentDisabling(t *testing.T) {
	config := DefaultIntegrationConfig()
	config.EnableCounterfactual = false
	config.EnableHypothesis = false

	integrator := NewAdvancedIntegrator(config)
	err := integrator.Initialize(nil)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	goal := &Goal{
		ID:          "disable-test",
		Name:        "Component Disable Test",
		Description: "Test selective components",
		Priority:    PriorityLow,
	}

	request := &IntegrationRequest{
		ID:   "req-disable",
		Goal: goal,
	}

	result, err := integrator.ProcessRequest(context.Background(), request)
	if err != nil {
		t.Fatalf("ProcessRequest failed: %v", err)
	}

	// Should still work with disabled components
	if !result.Success {
		t.Error("Expected successful result with disabled components")
	}

	t.Logf("Successfully processed with selective components")
}

// TestAdvancedIntegrator_Shutdown tests graceful shutdown
func TestAdvancedIntegrator_Shutdown(t *testing.T) {
	integrator := NewAdvancedIntegrator(DefaultIntegrationConfig())
	err := integrator.Initialize(nil)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Process a request
	goal := &Goal{
		ID:          "shutdown-test",
		Name:        "Shutdown Test",
		Description: "Test graceful shutdown",
		Priority:    PriorityLow,
	}

	request := &IntegrationRequest{
		ID:   "req-shutdown",
		Goal: goal,
	}

	_, err = integrator.ProcessRequest(context.Background(), request)
	if err != nil {
		t.Fatalf("ProcessRequest failed: %v", err)
	}

	// Shutdown
	err = integrator.Shutdown()
	if err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}

	t.Logf("Shutdown successful")
}

// TestAdvancedIntegrator_MultipleRequests tests handling multiple requests
func TestAdvancedIntegrator_MultipleRequests(t *testing.T) {
	integrator := NewAdvancedIntegrator(DefaultIntegrationConfig())
	err := integrator.Initialize(nil)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	goals := []*Goal{
		{ID: "multi-1", Name: "Goal 1", Description: "First goal", Priority: PriorityHigh},
		{ID: "multi-2", Name: "Goal 2", Description: "Second goal", Priority: PriorityNormal},
		{ID: "multi-3", Name: "Goal 3", Description: "Third goal", Priority: PriorityLow},
	}

	for i, goal := range goals {
		request := &IntegrationRequest{
			ID:   fmt.Sprintf("req-multi-%d", i),
			Goal: goal,
		}

		result, err := integrator.ProcessRequest(context.Background(), request)
		if err != nil {
			t.Fatalf("Request %d failed: %v", i, err)
		}

		if !result.Success {
			t.Errorf("Request %d was not successful", i)
		}

		t.Logf("Request %d processed successfully", i)
	}

	metrics := integrator.GetMetrics()
	if metrics.TotalRequests != 3 {
		t.Errorf("Expected 3 requests, got %d", metrics.TotalRequests)
	}
}

// TestAdvancedIntegrator_EndToEnd tests complete pipeline
func TestAdvancedIntegrator_EndToEnd(t *testing.T) {
	integrator := NewAdvancedIntegrator(DefaultIntegrationConfig())
	err := integrator.Initialize(nil)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	goal := &Goal{
		ID:           "e2e-test",
		Name:         "Complete System Test",
		Description:  "End-to-end integration test",
		Priority:     PriorityHigh,
		Dependencies: []string{"component-a", "component-b"},
	}

	request := &IntegrationRequest{
		ID:                   "req-e2e",
		Goal:                 goal,
		RequireAllComponents: true,
		Context: map[string]interface{}{
			"test_mode": true,
			"user_id":   "test-user",
		},
	}

	result, err := integrator.ProcessRequest(context.Background(), request)
	if err != nil {
		t.Fatalf("End-to-end test failed: %v", err)
	}

	// Verify all components executed
	if len(result.Plans) == 0 {
		t.Error("Plans not generated")
	}
	if len(result.Scenarios) == 0 {
		t.Error("Scenarios not generated")
	}
	if len(result.Hypotheses) == 0 {
		t.Error("Hypotheses not generated")
	}
	if len(result.Strategies) == 0 {
		t.Error("Strategies not generated")
	}
	if result.SelectedStrategy == nil {
		t.Error("Strategy not selected")
	}
	if result.Decision == nil {
		t.Error("Decision not synthesized")
	}
	if result.FormattedOutput == nil {
		t.Error("Output not formatted")
	}

	// Verify decision quality
	if result.Decision.Confidence < 0.5 {
		t.Errorf("Low confidence decision: %.2f", result.Decision.Confidence)
	}

	// Verify output quality
	if len(result.FormattedOutput.KeyInsights) == 0 {
		t.Error("No key insights provided")
	}
	if len(result.FormattedOutput.ActionItems) == 0 {
		t.Error("No action items provided")
	}

	t.Logf("End-to-end test completed successfully")
	t.Logf("Total execution time: %v", result.ExecutionTime)
	t.Logf("Decision confidence: %.2f", result.Decision.Confidence)
	t.Logf("Confidence level: %s", result.FormattedOutput.ConfidenceLevel)
}
