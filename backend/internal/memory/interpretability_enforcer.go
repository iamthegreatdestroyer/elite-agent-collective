// Package memory provides the MNEMONIC memory system for the Elite Agent Collective.
// This file implements the Interpretability Enforcer for ensuring agent decisions are explainable.
package memory

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"strings"
	"sync"
	"time"
)

// InterpretabilityEnforcer ensures agent responses include quality explanations
type InterpretabilityEnforcer struct {
	mu sync.RWMutex

	config  *InterpretabilityConfig
	metrics *InterpretabilityMetrics

	// Evaluation components
	explanationPatterns []*ExplanationPattern
}

// InterpretabilityConfig configures the interpretability enforcer
type InterpretabilityConfig struct {
	MinExplanationLength     int
	MinCoherenceScore        float64
	MinRelevanceScore        float64
	MinFaithfulnessScore     float64
	RequireReasoningChain    bool
	RequireUncertainty       bool
	RequireSourceAttribution bool
}

// DefaultInterpretabilityConfig returns default configuration
func DefaultInterpretabilityConfig() *InterpretabilityConfig {
	return &InterpretabilityConfig{
		MinExplanationLength:     50,
		MinCoherenceScore:        0.5,
		MinRelevanceScore:        0.5,
		MinFaithfulnessScore:     0.5,
		RequireReasoningChain:    false,
		RequireUncertainty:       false,
		RequireSourceAttribution: false,
	}
}

// InterpretabilityMetrics tracks interpretability metrics
type InterpretabilityMetrics struct {
	mu                  sync.RWMutex
	TotalChecks         int64
	PassedChecks        int64
	FailedChecks        int64
	AverageCoherence    float64
	AverageRelevance    float64
	AverageFaithfulness float64
	CoherenceSum        float64
	RelevanceSum        float64
	FaithfulnessSum     float64
}

// ExplanationPattern defines patterns to look for in explanations
type ExplanationPattern struct {
	Name     string
	Pattern  *regexp.Regexp
	Weight   float64
	Required bool
	Category string
}

// NewInterpretabilityEnforcer creates a new interpretability enforcer
func NewInterpretabilityEnforcer(config *InterpretabilityConfig) *InterpretabilityEnforcer {
	if config == nil {
		config = DefaultInterpretabilityConfig()
	}

	return &InterpretabilityEnforcer{
		config:              config,
		metrics:             &InterpretabilityMetrics{},
		explanationPatterns: buildExplanationPatterns(),
	}
}

// buildExplanationPatterns creates patterns for evaluating explanations
func buildExplanationPatterns() []*ExplanationPattern {
	patterns := []*ExplanationPattern{
		{
			Name:     "reasoning_because",
			Pattern:  regexp.MustCompile(`(?i)\bbecause\b`),
			Weight:   0.15,
			Required: false,
			Category: "reasoning",
		},
		{
			Name:     "reasoning_therefore",
			Pattern:  regexp.MustCompile(`(?i)\b(therefore|thus|hence|consequently)\b`),
			Weight:   0.15,
			Required: false,
			Category: "reasoning",
		},
		{
			Name:     "reasoning_since",
			Pattern:  regexp.MustCompile(`(?i)\bsince\b`),
			Weight:   0.1,
			Required: false,
			Category: "reasoning",
		},
		{
			Name:     "step_first",
			Pattern:  regexp.MustCompile(`(?i)\b(first|firstly|step 1)\b`),
			Weight:   0.1,
			Required: false,
			Category: "structure",
		},
		{
			Name:     "step_then",
			Pattern:  regexp.MustCompile(`(?i)\b(then|next|after that|subsequently)\b`),
			Weight:   0.1,
			Required: false,
			Category: "structure",
		},
		{
			Name:     "conclusion",
			Pattern:  regexp.MustCompile(`(?i)\b(finally|in conclusion|to summarize|in summary)\b`),
			Weight:   0.1,
			Required: false,
			Category: "structure",
		},
		{
			Name:     "uncertainty",
			Pattern:  regexp.MustCompile(`(?i)\b(may|might|could|possibly|probably|likely|uncertain|believe|think)\b`),
			Weight:   0.1,
			Required: false,
			Category: "uncertainty",
		},
		{
			Name:     "evidence",
			Pattern:  regexp.MustCompile(`(?i)\b(evidence|based on|according to|research shows|data indicates)\b`),
			Weight:   0.15,
			Required: false,
			Category: "evidence",
		},
		{
			Name:     "alternative",
			Pattern:  regexp.MustCompile(`(?i)\b(alternatively|another approach|other option|however)\b`),
			Weight:   0.05,
			Required: false,
			Category: "alternatives",
		},
	}

	return patterns
}

// ExplainedResponse wraps a response with its explanation
type ExplainedResponse struct {
	Response    *AgentResponse
	Explanation string
	Reasoning   []string
	Sources     []string
	Confidence  float64
	Timestamp   time.Time
}

// ExplanationQuality holds quality scores for an explanation
type ExplanationQuality struct {
	Coherence    float64 // Does explanation make logical sense?
	Relevance    float64 // Does it relate to the response?
	Faithfulness float64 // Does it reflect actual reasoning?
	Completeness float64 // Are all key points explained?
	Clarity      float64 // Is it easy to understand?
	Overall      float64 // Combined score
	Breakdown    map[string]float64
}

// InterpretabilityResult contains the full result of an interpretability check
type InterpretabilityResult struct {
	Passed           bool
	Quality          *ExplanationQuality
	MissingElements  []string
	Suggestions      []string
	RequiredPatterns []string
	FoundPatterns    []string
	Timestamp        time.Time
}

// RequireExplanation checks if a response has an adequate explanation
func (e *InterpretabilityEnforcer) RequireExplanation(ctx context.Context, resp *ExplainedResponse) (*InterpretabilityResult, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	result := &InterpretabilityResult{
		Passed:          true,
		MissingElements: make([]string, 0),
		Suggestions:     make([]string, 0),
		FoundPatterns:   make([]string, 0),
		Timestamp:       time.Now(),
	}

	// Check explanation exists
	if resp.Explanation == "" {
		result.Passed = false
		result.MissingElements = append(result.MissingElements, "explanation")
		result.Suggestions = append(result.Suggestions, "Add an explanation of your reasoning")
		e.recordResult(result, nil)
		return result, nil
	}

	// Check minimum length
	if len(resp.Explanation) < e.config.MinExplanationLength {
		result.Passed = false
		result.MissingElements = append(result.MissingElements, "sufficient_length")
		result.Suggestions = append(result.Suggestions,
			fmt.Sprintf("Explanation should be at least %d characters (current: %d)",
				e.config.MinExplanationLength, len(resp.Explanation)))
	}

	// Evaluate quality
	quality := e.evaluateQuality(resp)
	result.Quality = quality

	// Check quality thresholds
	if quality.Coherence < e.config.MinCoherenceScore {
		result.Passed = false
		result.MissingElements = append(result.MissingElements, "coherence")
		result.Suggestions = append(result.Suggestions, "Improve logical flow of explanation")
	}

	if quality.Relevance < e.config.MinRelevanceScore {
		result.Passed = false
		result.MissingElements = append(result.MissingElements, "relevance")
		result.Suggestions = append(result.Suggestions, "Ensure explanation directly relates to the response")
	}

	if quality.Faithfulness < e.config.MinFaithfulnessScore {
		result.Passed = false
		result.MissingElements = append(result.MissingElements, "faithfulness")
		result.Suggestions = append(result.Suggestions, "Explanation should reflect actual reasoning process")
	}

	// Check required patterns
	if e.config.RequireReasoningChain {
		hasReasoning := false
		for _, p := range e.explanationPatterns {
			if p.Category == "reasoning" && p.Pattern.MatchString(resp.Explanation) {
				hasReasoning = true
				break
			}
		}
		if !hasReasoning {
			result.Passed = false
			result.MissingElements = append(result.MissingElements, "reasoning_chain")
			result.RequiredPatterns = append(result.RequiredPatterns, "reasoning indicators (because, therefore, since)")
			result.Suggestions = append(result.Suggestions, "Include reasoning words like 'because', 'therefore'")
		}
	}

	if e.config.RequireUncertainty && resp.Confidence < 0.9 {
		hasUncertainty := false
		for _, p := range e.explanationPatterns {
			if p.Category == "uncertainty" && p.Pattern.MatchString(resp.Explanation) {
				hasUncertainty = true
				break
			}
		}
		if !hasUncertainty {
			result.Passed = false
			result.MissingElements = append(result.MissingElements, "uncertainty_acknowledgment")
			result.Suggestions = append(result.Suggestions, "Acknowledge uncertainty when confidence is below 90%")
		}
	}

	if e.config.RequireSourceAttribution && len(resp.Sources) == 0 {
		result.Passed = false
		result.MissingElements = append(result.MissingElements, "source_attribution")
		result.Suggestions = append(result.Suggestions, "Cite sources for claims made in the response")
	}

	// Record found patterns
	for _, p := range e.explanationPatterns {
		if p.Pattern.MatchString(resp.Explanation) {
			result.FoundPatterns = append(result.FoundPatterns, p.Name)
		}
	}

	e.recordResult(result, quality)
	return result, nil
}

// evaluateQuality evaluates the quality of an explanation
func (e *InterpretabilityEnforcer) evaluateQuality(resp *ExplainedResponse) *ExplanationQuality {
	quality := &ExplanationQuality{
		Breakdown: make(map[string]float64),
	}

	explanation := resp.Explanation
	responseContent := ""
	if resp.Response != nil {
		responseContent = resp.Response.Content
	}

	// Coherence: Check for logical flow indicators
	coherenceScore := e.evaluateCoherence(explanation)
	quality.Coherence = coherenceScore
	quality.Breakdown["coherence"] = coherenceScore

	// Relevance: Check overlap between explanation and response
	relevanceScore := e.evaluateRelevance(explanation, responseContent)
	quality.Relevance = relevanceScore
	quality.Breakdown["relevance"] = relevanceScore

	// Faithfulness: Check if explanation matches reasoning patterns
	faithfulnessScore := e.evaluateFaithfulness(resp)
	quality.Faithfulness = faithfulnessScore
	quality.Breakdown["faithfulness"] = faithfulnessScore

	// Completeness: Check for required elements
	completenessScore := e.evaluateCompleteness(explanation, resp.Reasoning)
	quality.Completeness = completenessScore
	quality.Breakdown["completeness"] = completenessScore

	// Clarity: Check for clear language
	clarityScore := e.evaluateClarity(explanation)
	quality.Clarity = clarityScore
	quality.Breakdown["clarity"] = clarityScore

	// Overall score (weighted average)
	quality.Overall = (coherenceScore*0.25 + relevanceScore*0.25 +
		faithfulnessScore*0.20 + completenessScore*0.15 + clarityScore*0.15)

	return quality
}

// evaluateCoherence evaluates logical coherence
func (e *InterpretabilityEnforcer) evaluateCoherence(explanation string) float64 {
	score := 0.5 // Base score

	// Check for reasoning patterns
	for _, p := range e.explanationPatterns {
		if p.Category == "reasoning" || p.Category == "structure" {
			if p.Pattern.MatchString(explanation) {
				score += p.Weight
			}
		}
	}

	// Check for sentence structure
	sentences := strings.Split(explanation, ".")
	if len(sentences) >= 2 {
		score += 0.1
	}
	if len(sentences) >= 4 {
		score += 0.1
	}

	return math.Min(1.0, score)
}

// evaluateRelevance evaluates relevance to the response
func (e *InterpretabilityEnforcer) evaluateRelevance(explanation, response string) float64 {
	if response == "" {
		return 0.5 // No response to compare
	}

	// Simple word overlap metric
	explanationWords := strings.Fields(strings.ToLower(explanation))
	responseWords := strings.Fields(strings.ToLower(response))

	if len(explanationWords) == 0 || len(responseWords) == 0 {
		return 0.5
	}

	// Build response word set
	responseSet := make(map[string]bool)
	for _, w := range responseWords {
		if len(w) > 3 { // Skip short words
			responseSet[w] = true
		}
	}

	// Count overlapping words
	overlap := 0
	for _, w := range explanationWords {
		if len(w) > 3 && responseSet[w] {
			overlap++
		}
	}

	// Calculate Jaccard-like similarity
	relevance := float64(overlap) / float64(len(explanationWords))
	return math.Min(1.0, relevance*2+0.3) // Scale and add base
}

// evaluateFaithfulness evaluates if explanation reflects actual reasoning
func (e *InterpretabilityEnforcer) evaluateFaithfulness(resp *ExplainedResponse) float64 {
	score := 0.5 // Base score

	// Check if explicit reasoning steps are present
	if len(resp.Reasoning) > 0 {
		score += 0.2

		// Check if reasoning is reflected in explanation
		for _, reason := range resp.Reasoning {
			if strings.Contains(strings.ToLower(resp.Explanation), strings.ToLower(reason)) {
				score += 0.05
			}
		}
	}

	// Check for evidence patterns
	for _, p := range e.explanationPatterns {
		if p.Category == "evidence" && p.Pattern.MatchString(resp.Explanation) {
			score += p.Weight
		}
	}

	return math.Min(1.0, score)
}

// evaluateCompleteness evaluates if explanation covers key points
func (e *InterpretabilityEnforcer) evaluateCompleteness(explanation string, reasoning []string) float64 {
	if len(reasoning) == 0 {
		// No explicit reasoning to check, use pattern-based evaluation
		patternCount := 0
		for _, p := range e.explanationPatterns {
			if p.Pattern.MatchString(explanation) {
				patternCount++
			}
		}
		return math.Min(1.0, float64(patternCount)/5.0+0.3)
	}

	// Check how many reasoning points are addressed
	covered := 0
	for _, reason := range reasoning {
		if strings.Contains(strings.ToLower(explanation), strings.ToLower(reason)) {
			covered++
		}
	}

	return float64(covered) / float64(len(reasoning))
}

// evaluateClarity evaluates clarity of explanation
func (e *InterpretabilityEnforcer) evaluateClarity(explanation string) float64 {
	score := 0.5

	// Penalize very long sentences
	sentences := strings.Split(explanation, ".")
	avgLength := 0
	for _, s := range sentences {
		avgLength += len(strings.Fields(s))
	}
	if len(sentences) > 0 {
		avgLength /= len(sentences)
	}

	if avgLength < 25 && avgLength > 5 {
		score += 0.2 // Good sentence length
	}

	// Check for clear structure
	if strings.Contains(explanation, "\n") || strings.Contains(explanation, "•") ||
		strings.Contains(explanation, "-") || strings.Contains(explanation, "1.") {
		score += 0.2 // Has some structure
	}

	// Bonus for having concrete examples
	if strings.Contains(strings.ToLower(explanation), "example") ||
		strings.Contains(strings.ToLower(explanation), "for instance") {
		score += 0.1
	}

	return math.Min(1.0, score)
}

// recordResult records the result for metrics
func (e *InterpretabilityEnforcer) recordResult(result *InterpretabilityResult, quality *ExplanationQuality) {
	e.metrics.mu.Lock()
	defer e.metrics.mu.Unlock()

	e.metrics.TotalChecks++

	if result.Passed {
		e.metrics.PassedChecks++
	} else {
		e.metrics.FailedChecks++
	}

	if quality != nil {
		e.metrics.CoherenceSum += quality.Coherence
		e.metrics.RelevanceSum += quality.Relevance
		e.metrics.FaithfulnessSum += quality.Faithfulness

		n := float64(e.metrics.TotalChecks)
		e.metrics.AverageCoherence = e.metrics.CoherenceSum / n
		e.metrics.AverageRelevance = e.metrics.RelevanceSum / n
		e.metrics.AverageFaithfulness = e.metrics.FaithfulnessSum / n
	}
}

// GetMetrics returns current metrics
func (e *InterpretabilityEnforcer) GetMetrics() *InterpretabilityMetrics {
	e.metrics.mu.RLock()
	defer e.metrics.mu.RUnlock()

	return &InterpretabilityMetrics{
		TotalChecks:         e.metrics.TotalChecks,
		PassedChecks:        e.metrics.PassedChecks,
		FailedChecks:        e.metrics.FailedChecks,
		AverageCoherence:    e.metrics.AverageCoherence,
		AverageRelevance:    e.metrics.AverageRelevance,
		AverageFaithfulness: e.metrics.AverageFaithfulness,
	}
}

// PassRate returns the percentage of checks that passed
func (e *InterpretabilityEnforcer) PassRate() float64 {
	e.metrics.mu.RLock()
	defer e.metrics.mu.RUnlock()

	if e.metrics.TotalChecks == 0 {
		return 0
	}
	return float64(e.metrics.PassedChecks) / float64(e.metrics.TotalChecks) * 100
}

// AddPattern adds a custom explanation pattern
func (e *InterpretabilityEnforcer) AddPattern(name, pattern, category string, weight float64, required bool) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	compiled, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid pattern: %w", err)
	}

	e.explanationPatterns = append(e.explanationPatterns, &ExplanationPattern{
		Name:     name,
		Pattern:  compiled,
		Weight:   weight,
		Required: required,
		Category: category,
	})

	return nil
}

// GetPatterns returns all configured patterns
func (e *InterpretabilityEnforcer) GetPatterns() []*ExplanationPattern {
	e.mu.RLock()
	defer e.mu.RUnlock()

	result := make([]*ExplanationPattern, len(e.explanationPatterns))
	copy(result, e.explanationPatterns)
	return result
}
