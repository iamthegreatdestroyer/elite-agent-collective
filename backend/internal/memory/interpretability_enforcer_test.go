package memory

import (
	"context"
	"strings"
	"testing"
)

func TestInterpretabilityEnforcer_NewInterpretabilityEnforcer(t *testing.T) {
	t.Run("with default config", func(t *testing.T) {
		e := NewInterpretabilityEnforcer(nil)

		if e == nil {
			t.Fatal("Expected non-nil enforcer")
		}

		if e.config.MinExplanationLength != 50 {
			t.Error("Expected default min explanation length of 50")
		}
	})

	t.Run("with custom config", func(t *testing.T) {
		config := &InterpretabilityConfig{
			MinExplanationLength: 100,
			MinCoherenceScore:    0.7,
		}

		e := NewInterpretabilityEnforcer(config)

		if e.config.MinExplanationLength != 100 {
			t.Error("Expected custom min explanation length")
		}
	})
}

func TestInterpretabilityEnforcer_RequireExplanation(t *testing.T) {
	e := NewInterpretabilityEnforcer(&InterpretabilityConfig{
		MinExplanationLength: 20,
		MinCoherenceScore:    0.3,
		MinRelevanceScore:    0.3,
		MinFaithfulnessScore: 0.3,
	})
	ctx := context.Background()

	t.Run("missing explanation fails", func(t *testing.T) {
		resp := &ExplainedResponse{
			Response:    &AgentResponse{Content: "Hello"},
			Explanation: "",
		}

		result, err := e.RequireExplanation(ctx, resp)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if result.Passed {
			t.Error("Expected missing explanation to fail")
		}

		found := false
		for _, m := range result.MissingElements {
			if m == "explanation" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected 'explanation' in missing elements")
		}
	})

	t.Run("short explanation fails", func(t *testing.T) {
		resp := &ExplainedResponse{
			Response:    &AgentResponse{Content: "Hello world"},
			Explanation: "Short",
		}

		result, err := e.RequireExplanation(ctx, resp)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if result.Passed {
			t.Error("Expected short explanation to fail")
		}
	})

	t.Run("good explanation passes", func(t *testing.T) {
		resp := &ExplainedResponse{
			Response: &AgentResponse{
				Content: "Here is a sorting algorithm implementation in Go.",
			},
			Explanation: "I chose this implementation because it provides O(n log n) time complexity. " +
				"Therefore, it is efficient for large datasets. First, the array is divided, then sorted recursively.",
			Confidence: 0.9,
		}

		result, err := e.RequireExplanation(ctx, resp)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if !result.Passed {
			t.Errorf("Expected good explanation to pass. Missing: %v", result.MissingElements)
		}

		if result.Quality == nil {
			t.Error("Expected quality scores to be set")
		}
	})

	t.Run("explanation with reasoning patterns scores higher", func(t *testing.T) {
		// Without reasoning patterns
		resp1 := &ExplainedResponse{
			Response:    &AgentResponse{Content: "Result"},
			Explanation: "This is the output of the algorithm. The algorithm runs correctly.",
		}

		// With reasoning patterns
		resp2 := &ExplainedResponse{
			Response:    &AgentResponse{Content: "Result"},
			Explanation: "This is the output because the algorithm uses merge sort. Therefore, the time complexity is O(n log n). Since the data is sorted, we can use binary search.",
		}

		result1, _ := e.RequireExplanation(ctx, resp1)
		result2, _ := e.RequireExplanation(ctx, resp2)

		if result2.Quality.Coherence <= result1.Quality.Coherence {
			t.Error("Expected explanation with reasoning patterns to score higher on coherence")
		}
	})
}

func TestInterpretabilityEnforcer_RequireReasoningChain(t *testing.T) {
	e := NewInterpretabilityEnforcer(&InterpretabilityConfig{
		MinExplanationLength:  20,
		RequireReasoningChain: true,
	})
	ctx := context.Background()

	t.Run("explanation without reasoning fails", func(t *testing.T) {
		resp := &ExplainedResponse{
			Response:    &AgentResponse{Content: "Result"},
			Explanation: "Here is the output. It works correctly. The result is accurate.",
		}

		result, _ := e.RequireExplanation(ctx, resp)

		if result.Passed {
			t.Error("Expected explanation without reasoning to fail when required")
		}
	})

	t.Run("explanation with because passes", func(t *testing.T) {
		resp := &ExplainedResponse{
			Response:    &AgentResponse{Content: "Result"},
			Explanation: "Here is the output because the algorithm processes all elements correctly.",
		}

		result, _ := e.RequireExplanation(ctx, resp)

		// Should have reasoning_chain in required but found
		foundReasoning := false
		for _, p := range result.FoundPatterns {
			if p == "reasoning_because" {
				foundReasoning = true
				break
			}
		}
		if !foundReasoning {
			t.Error("Expected to find reasoning_because pattern")
		}
	})
}

func TestInterpretabilityEnforcer_RequireUncertainty(t *testing.T) {
	e := NewInterpretabilityEnforcer(&InterpretabilityConfig{
		MinExplanationLength: 20,
		RequireUncertainty:   true,
	})
	ctx := context.Background()

	t.Run("low confidence without uncertainty fails", func(t *testing.T) {
		resp := &ExplainedResponse{
			Response:    &AgentResponse{Content: "The answer is 42."},
			Explanation: "The answer is definitely 42. This is correct. No doubt about it.",
			Confidence:  0.5, // Low confidence
		}

		result, _ := e.RequireExplanation(ctx, resp)

		if result.Passed {
			t.Error("Expected low confidence without uncertainty acknowledgment to fail")
		}
	})

	t.Run("low confidence with uncertainty passes", func(t *testing.T) {
		resp := &ExplainedResponse{
			Response:    &AgentResponse{Content: "The answer might be 42."},
			Explanation: "I believe the answer is 42, but I'm not completely certain. It could possibly be different.",
			Confidence:  0.5,
		}

		result, _ := e.RequireExplanation(ctx, resp)

		// Check that uncertainty was detected
		foundUncertainty := false
		for _, p := range result.FoundPatterns {
			if p == "uncertainty" {
				foundUncertainty = true
				break
			}
		}
		if !foundUncertainty {
			t.Error("Expected to find uncertainty pattern")
		}
	})

	t.Run("high confidence doesnt need uncertainty", func(t *testing.T) {
		resp := &ExplainedResponse{
			Response:    &AgentResponse{Content: "The answer is 42."},
			Explanation: "The answer is definitely 42. This is mathematically proven and correct.",
			Confidence:  0.95, // High confidence
		}

		result, _ := e.RequireExplanation(ctx, resp)

		// Should pass even without uncertainty words
		missingUncertainty := false
		for _, m := range result.MissingElements {
			if m == "uncertainty_acknowledgment" {
				missingUncertainty = true
				break
			}
		}
		if missingUncertainty {
			t.Error("High confidence should not require uncertainty acknowledgment")
		}
	})
}

func TestInterpretabilityEnforcer_RequireSourceAttribution(t *testing.T) {
	e := NewInterpretabilityEnforcer(&InterpretabilityConfig{
		MinExplanationLength:     20,
		RequireSourceAttribution: true,
	})
	ctx := context.Background()

	t.Run("no sources fails", func(t *testing.T) {
		resp := &ExplainedResponse{
			Response:    &AgentResponse{Content: "Result"},
			Explanation: "This is the explanation of the result that was computed.",
			Sources:     nil,
		}

		result, _ := e.RequireExplanation(ctx, resp)

		if result.Passed {
			t.Error("Expected missing sources to fail when required")
		}
	})

	t.Run("with sources passes", func(t *testing.T) {
		resp := &ExplainedResponse{
			Response:    &AgentResponse{Content: "Result"},
			Explanation: "This is the explanation of the result that was computed using standard algorithms.",
			Sources:     []string{"Cormen et al. Introduction to Algorithms"},
		}

		result, _ := e.RequireExplanation(ctx, resp)

		missingSources := false
		for _, m := range result.MissingElements {
			if m == "source_attribution" {
				missingSources = true
				break
			}
		}
		if missingSources {
			t.Error("Expected sources to be found when provided")
		}
	})
}

func TestExplanationQuality_Evaluation(t *testing.T) {
	e := NewInterpretabilityEnforcer(nil)
	ctx := context.Background()

	t.Run("quality scores are in range", func(t *testing.T) {
		resp := &ExplainedResponse{
			Response: &AgentResponse{
				Content: "This is the algorithm implementation.",
			},
			Explanation: "First, we initialize the data structure. Then, we iterate through the elements. " +
				"Because of the sorting, we can use binary search. Therefore, the time complexity is O(log n). " +
				"Finally, we return the result.",
			Reasoning:  []string{"sorting", "binary search"},
			Confidence: 0.9,
		}

		result, _ := e.RequireExplanation(ctx, resp)

		if result.Quality == nil {
			t.Fatal("Expected quality to be set")
		}

		checkRange := func(name string, value float64) {
			if value < 0 || value > 1 {
				t.Errorf("%s out of range: %f", name, value)
			}
		}

		checkRange("Coherence", result.Quality.Coherence)
		checkRange("Relevance", result.Quality.Relevance)
		checkRange("Faithfulness", result.Quality.Faithfulness)
		checkRange("Completeness", result.Quality.Completeness)
		checkRange("Clarity", result.Quality.Clarity)
		checkRange("Overall", result.Quality.Overall)
	})

	t.Run("better explanation scores higher overall", func(t *testing.T) {
		poorResp := &ExplainedResponse{
			Response:    &AgentResponse{Content: "x"},
			Explanation: "Done. It works.",
		}

		goodResp := &ExplainedResponse{
			Response: &AgentResponse{Content: "Here is a comprehensive implementation."},
			Explanation: "First, I analyzed the problem requirements. Because of the need for efficiency, " +
				"I chose a hash table approach. Therefore, lookups are O(1). Finally, the solution " +
				"handles edge cases properly. For example, empty inputs return an empty result.",
		}

		poorResult, _ := e.RequireExplanation(ctx, poorResp)
		goodResult, _ := e.RequireExplanation(ctx, goodResp)

		if goodResult.Quality.Overall <= poorResult.Quality.Overall {
			t.Error("Expected good explanation to have higher overall score")
		}
	})
}

func TestInterpretabilityEnforcer_Metrics(t *testing.T) {
	e := NewInterpretabilityEnforcer(&InterpretabilityConfig{
		MinExplanationLength: 20,
	})
	ctx := context.Background()

	// Run some checks
	responses := []*ExplainedResponse{
		{
			Response:    &AgentResponse{Content: "Good"},
			Explanation: "This is a good explanation because it explains the reasoning clearly.",
		},
		{
			Response:    &AgentResponse{Content: "Bad"},
			Explanation: "Short",
		},
		{
			Response:    &AgentResponse{Content: "Another good"},
			Explanation: "This explanation is detailed and therefore shows the reasoning process step by step.",
		},
	}

	for _, resp := range responses {
		e.RequireExplanation(ctx, resp)
	}

	metrics := e.GetMetrics()

	t.Run("total checks counted", func(t *testing.T) {
		if metrics.TotalChecks != 3 {
			t.Errorf("Expected 3 total checks, got %d", metrics.TotalChecks)
		}
	})

	t.Run("passed and failed counted", func(t *testing.T) {
		if metrics.PassedChecks+metrics.FailedChecks != metrics.TotalChecks {
			t.Error("Passed + Failed should equal Total")
		}
	})

	t.Run("pass rate calculated", func(t *testing.T) {
		rate := e.PassRate()
		if rate < 0 || rate > 100 {
			t.Errorf("Pass rate out of range: %f", rate)
		}
		t.Logf("Pass rate: %.2f%%", rate)
	})
}

func TestInterpretabilityEnforcer_AddPattern(t *testing.T) {
	e := NewInterpretabilityEnforcer(nil)
	initialCount := len(e.GetPatterns())

	t.Run("add valid pattern", func(t *testing.T) {
		err := e.AddPattern("custom_pattern", `(?i)\bcustom\b`, "custom", 0.1, false)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		patterns := e.GetPatterns()
		if len(patterns) != initialCount+1 {
			t.Error("Pattern was not added")
		}
	})

	t.Run("add invalid pattern fails", func(t *testing.T) {
		err := e.AddPattern("invalid", `[invalid`, "custom", 0.1, false)

		if err == nil {
			t.Error("Expected error for invalid regex")
		}
	})
}

func TestInterpretabilityEnforcer_Suggestions(t *testing.T) {
	e := NewInterpretabilityEnforcer(&InterpretabilityConfig{
		MinExplanationLength:  100,
		RequireReasoningChain: true,
	})
	ctx := context.Background()

	resp := &ExplainedResponse{
		Response:    &AgentResponse{Content: "Result"},
		Explanation: "Short. No reasoning.",
	}

	result, _ := e.RequireExplanation(ctx, resp)

	if len(result.Suggestions) == 0 {
		t.Error("Expected suggestions for failing response")
	}

	// Should have suggestion about length
	foundLengthSuggestion := false
	for _, s := range result.Suggestions {
		if strings.Contains(s, "characters") {
			foundLengthSuggestion = true
			break
		}
	}

	if !foundLengthSuggestion {
		t.Error("Expected suggestion about explanation length")
	}
}

func BenchmarkInterpretabilityEnforcer_RequireExplanation(b *testing.B) {
	e := NewInterpretabilityEnforcer(nil)
	ctx := context.Background()

	resp := &ExplainedResponse{
		Response: &AgentResponse{
			Content: "Here is the implementation of a binary search algorithm in Go.",
		},
		Explanation: "First, I analyzed the requirements and determined that binary search is appropriate. " +
			"Because the input is sorted, we can divide the search space in half with each comparison. " +
			"Therefore, the time complexity is O(log n). The implementation handles edge cases " +
			"like empty arrays and elements not found.",
		Reasoning:  []string{"sorted input", "divide and conquer", "O(log n)"},
		Confidence: 0.95,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.RequireExplanation(ctx, resp)
	}
}

// Helper for string contains check
func strings_Contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && containsSubstr(s, substr)))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
