// Package memory provides the MNEMONIC system for the Elite Agent Collective.
// This file implements Neurosymbolic Reasoning from @NEURAL's Cognitive Architecture Analysis.
//
// Neurosymbolic Reasoning combines:
// - Neural reasoning: LLM-based hypothesis generation, pattern matching
// - Symbolic verification: Logic proofs, constraint checking, formal verification
//
// The key insight is that neural systems are good at generating plausible hypotheses,
// while symbolic systems excel at verifying correctness. Together they provide
// both flexibility and rigor.

package memory

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// ============================================================================
// Core Types
// ============================================================================

// Query represents a reasoning query.
type Query struct {
	ID          string
	Question    string
	Context     map[string]interface{}
	Constraints []Constraint
	MaxDepth    int
	Timeout     time.Duration
}

// Constraint represents a logical constraint on reasoning.
type Constraint struct {
	Type           ConstraintType
	LogicPredicate string
	Args           []interface{}
	Required       bool
}

// ConstraintType categorizes constraints.
type ConstraintType string

const (
	ConstraintMustHold    ConstraintType = "must_hold"     // Invariant
	ConstraintMustNotHold ConstraintType = "must_not_hold" // Forbidden
	ConstraintPrefer      ConstraintType = "prefer"        // Soft preference
	ConstraintImplies     ConstraintType = "implies"       // Conditional
)

// Hypothesis represents a proposed answer or solution.
type Hypothesis struct {
	ID          string
	Statement   string
	Confidence  float64
	Provenance  string   // Where this came from
	Supporting  []string // Supporting evidence
	Assumptions []LogicPredicate
	Timestamp   time.Time
}

// Conclusion represents a verified conclusion.
type Conclusion struct {
	ID              string
	Statement       string
	Confidence      float64
	Verified        bool
	VerifiedAt      time.Time
	HypothesisID    string
	ProofID         string
	Counterexamples []Counterexample
}

// Proof represents a symbolic proof.
type Proof struct {
	ID           string
	HypothesisID string
	Valid        bool
	Steps        []ProofStep
	Assumptions  []LogicPredicate
	Conclusion   LogicPredicate
	Duration     time.Duration
}

// ProofStep is a single step in a proof.
type ProofStep struct {
	StepNum       int
	Rule          string
	Premises      []LogicPredicate
	Conclusion    LogicPredicate
	Justification string
}

// LogicPredicate represents a logical LogicPredicate.
type LogicPredicate struct {
	Name    string
	Args    []interface{}
	Negated bool
}

// String returns a string representation of the LogicPredicate.
func (p LogicPredicate) String() string {
	args := make([]string, len(p.Args))
	for i, arg := range p.Args {
		args[i] = fmt.Sprintf("%v", arg)
	}
	pred := fmt.Sprintf("%s(%s)", p.Name, strings.Join(args, ", "))
	if p.Negated {
		return "¬" + pred
	}
	return pred
}

// Counterexample represents a counterexample to a hypothesis.
type Counterexample struct {
	Description    string
	Bindings       map[string]interface{}
	LogicPredicate LogicPredicate
}

// ============================================================================
// Knowledge Base
// ============================================================================

// LogicKnowledgeBase stores logical facts and rules.
type LogicKnowledgeBase struct {
	facts map[string][]LogicPredicate // LogicPredicate name -> instances
	rules []Rule
	mu    sync.RWMutex
}

// Rule represents a logical inference rule.
type Rule struct {
	ID         string
	Name       string
	Premises   []LogicPredicate
	Conclusion LogicPredicate
	Confidence float64
}

// NewLogicKnowledgeBase creates a new knowledge base.
func NewLogicKnowledgeBase() *LogicKnowledgeBase {
	return &LogicKnowledgeBase{
		facts: make(map[string][]LogicPredicate),
		rules: make([]Rule, 0),
	}
}

// AddFact adds a fact to the knowledge base.
func (kb *LogicKnowledgeBase) AddFact(pred LogicPredicate) {
	kb.mu.Lock()
	defer kb.mu.Unlock()

	if _, exists := kb.facts[pred.Name]; !exists {
		kb.facts[pred.Name] = make([]LogicPredicate, 0)
	}
	kb.facts[pred.Name] = append(kb.facts[pred.Name], pred)
}

// AddRule adds an inference rule.
func (kb *LogicKnowledgeBase) AddRule(rule Rule) {
	kb.mu.Lock()
	defer kb.mu.Unlock()
	kb.rules = append(kb.rules, rule)
}

// Query finds facts matching a pattern.
func (kb *LogicKnowledgeBase) Query(pattern LogicPredicate) []LogicPredicate {
	kb.mu.RLock()
	defer kb.mu.RUnlock()

	facts, exists := kb.facts[pattern.Name]
	if !exists {
		return nil
	}

	// Filter by matching arguments
	matches := make([]LogicPredicate, 0)
	for _, fact := range facts {
		if kb.matches(pattern, fact) {
			matches = append(matches, fact)
		}
	}

	return matches
}

// matches checks if a fact matches a pattern (with variables).
func (kb *LogicKnowledgeBase) matches(pattern, fact LogicPredicate) bool {
	if pattern.Name != fact.Name || len(pattern.Args) != len(fact.Args) {
		return false
	}

	for i, patternArg := range pattern.Args {
		// Check if pattern arg is a variable (starts with ?)
		if str, ok := patternArg.(string); ok && strings.HasPrefix(str, "?") {
			continue // Variable matches anything
		}
		if patternArg != fact.Args[i] {
			return false
		}
	}

	return true
}

// GetApplicableRules returns rules whose premises might be satisfiable.
func (kb *LogicKnowledgeBase) GetApplicableRules(goal LogicPredicate) []Rule {
	kb.mu.RLock()
	defer kb.mu.RUnlock()

	applicable := make([]Rule, 0)
	for _, rule := range kb.rules {
		if kb.unifies(rule.Conclusion, goal) {
			applicable = append(applicable, rule)
		}
	}

	return applicable
}

// unifies checks if two LogicPredicates can be unified.
func (kb *LogicKnowledgeBase) unifies(p1, p2 LogicPredicate) bool {
	if p1.Name != p2.Name || len(p1.Args) != len(p2.Args) {
		return false
	}

	for i := 0; i < len(p1.Args); i++ {
		arg1, arg2 := p1.Args[i], p2.Args[i]

		// Variables unify with anything
		if isVariable(arg1) || isVariable(arg2) {
			continue
		}

		if arg1 != arg2 {
			return false
		}
	}

	return true
}

func isVariable(arg interface{}) bool {
	str, ok := arg.(string)
	return ok && strings.HasPrefix(str, "?")
}

// ============================================================================
// Neural Reasoner (Hypothesis Generator)
// ============================================================================

// NeuralReasoner generates hypotheses using neural/pattern-based methods.
type NeuralReasoner struct {
	experienceRetriever *SubLinearRetriever
	patternMatcher      *PatternMatcher
	confidenceThreshold float64
}

// PatternMatcher matches patterns in experiences.
type PatternMatcher struct {
	patterns map[string]*ReasoningPattern
	mu       sync.RWMutex
}

// ReasoningPattern is a learned reasoning pattern.
type ReasoningPattern struct {
	ID          string
	Trigger     string  // Input pattern that triggers this
	Strategy    string  // Reasoning strategy to apply
	SuccessRate float64 // Historical success rate
	UsageCount  int64
}

// NewNeuralReasoner creates a new neural reasoner.
func NewNeuralReasoner(retriever *SubLinearRetriever) *NeuralReasoner {
	return &NeuralReasoner{
		experienceRetriever: retriever,
		patternMatcher: &PatternMatcher{
			patterns: make(map[string]*ReasoningPattern),
		},
		confidenceThreshold: 0.5,
	}
}

// GenerateHypotheses generates hypotheses for a query.
func (nr *NeuralReasoner) GenerateHypotheses(query *Query) []*Hypothesis {
	hypotheses := make([]*Hypothesis, 0)

	// 1. Pattern-based hypothesis generation
	if patternHyps := nr.generateFromPatterns(query); len(patternHyps) > 0 {
		hypotheses = append(hypotheses, patternHyps...)
	}

	// 2. Experience-based hypothesis generation
	if expHyps := nr.generateFromExperiences(query); len(expHyps) > 0 {
		hypotheses = append(hypotheses, expHyps...)
	}

	// 3. Analogy-based hypothesis generation
	if analogyHyps := nr.generateFromAnalogies(query); len(analogyHyps) > 0 {
		hypotheses = append(hypotheses, analogyHyps...)
	}

	// Sort by confidence
	sort.Slice(hypotheses, func(i, j int) bool {
		return hypotheses[i].Confidence > hypotheses[j].Confidence
	})

	return hypotheses
}

// generateFromPatterns uses learned patterns to generate hypotheses.
func (nr *NeuralReasoner) generateFromPatterns(query *Query) []*Hypothesis {
	nr.patternMatcher.mu.RLock()
	defer nr.patternMatcher.mu.RUnlock()

	hypotheses := make([]*Hypothesis, 0)

	for _, pattern := range nr.patternMatcher.patterns {
		if strings.Contains(query.Question, pattern.Trigger) {
			hyp := &Hypothesis{
				ID:         fmt.Sprintf("hyp-pattern-%s-%d", pattern.ID, time.Now().UnixNano()),
				Statement:  fmt.Sprintf("Based on pattern '%s': %s", pattern.ID, pattern.Strategy),
				Confidence: pattern.SuccessRate,
				Provenance: "pattern_matching",
				Timestamp:  time.Now(),
			}
			hypotheses = append(hypotheses, hyp)
		}
	}

	return hypotheses
}

// generateFromExperiences uses past experiences to generate hypotheses.
func (nr *NeuralReasoner) generateFromExperiences(query *Query) []*Hypothesis {
	if nr.experienceRetriever == nil {
		return nil
	}

	// Query for similar experiences
	queryCtx := &QueryContext{
		AgentID:         "REASONING",
		TaskSignature:   computeTaskSignature(query.Question),
		TopK:            5,
		MinFitnessScore: nr.confidenceThreshold,
	}

	result, err := nr.experienceRetriever.Retrieve(queryCtx)
	if err != nil || len(result.Experiences) == 0 {
		return nil
	}

	hypotheses := make([]*Hypothesis, 0)
	for _, exp := range result.Experiences {
		hyp := &Hypothesis{
			ID:         fmt.Sprintf("hyp-exp-%s-%d", exp.ID, time.Now().UnixNano()),
			Statement:  fmt.Sprintf("Based on similar experience: %s", exp.Strategy),
			Confidence: exp.FitnessScore,
			Provenance: "experience_retrieval",
			Supporting: []string{exp.ID},
			Timestamp:  time.Now(),
		}
		hypotheses = append(hypotheses, hyp)
	}

	return hypotheses
}

// generateFromAnalogies uses analogical reasoning.
func (nr *NeuralReasoner) generateFromAnalogies(query *Query) []*Hypothesis {
	// Analogical reasoning: Find structurally similar problems
	// This is a simplified implementation
	hypotheses := make([]*Hypothesis, 0)

	// Extract key concepts from query
	concepts := extractConcepts(query.Question)

	if len(concepts) >= 2 {
		// Generate analogy-based hypothesis
		hyp := &Hypothesis{
			ID:         fmt.Sprintf("hyp-analogy-%d", time.Now().UnixNano()),
			Statement:  fmt.Sprintf("Analogical reasoning: apply patterns from %s to %s", concepts[0], concepts[1]),
			Confidence: 0.4, // Lower confidence for analogies
			Provenance: "analogical_reasoning",
			Timestamp:  time.Now(),
		}
		hypotheses = append(hypotheses, hyp)
	}

	return hypotheses
}

// RefineHypotheses refines hypotheses based on counterexamples.
func (nr *NeuralReasoner) RefineHypotheses(
	hypotheses []*Hypothesis,
	counterexample *Counterexample,
) []*Hypothesis {
	refined := make([]*Hypothesis, 0)

	for _, hyp := range hypotheses {
		// Lower confidence for hypotheses contradicted by counterexample
		newHyp := &Hypothesis{
			ID:         fmt.Sprintf("%s-refined", hyp.ID),
			Statement:  fmt.Sprintf("%s [refined to avoid: %s]", hyp.Statement, counterexample.Description),
			Confidence: hyp.Confidence * 0.5, // Reduce confidence
			Provenance: hyp.Provenance,
			Supporting: hyp.Supporting,
			Assumptions: append(hyp.Assumptions, LogicPredicate{
				Name:    "avoids",
				Args:    []interface{}{counterexample.Description},
				Negated: false,
			}),
			Timestamp: time.Now(),
		}
		refined = append(refined, newHyp)
	}

	return refined
}

// AddPattern adds a learned reasoning pattern.
func (nr *NeuralReasoner) AddPattern(pattern *ReasoningPattern) {
	nr.patternMatcher.mu.Lock()
	defer nr.patternMatcher.mu.Unlock()
	nr.patternMatcher.patterns[pattern.ID] = pattern
}

// extractConcepts extracts key concepts from text.
func extractConcepts(text string) []string {
	// Simple concept extraction (could be enhanced with NLP)
	words := strings.Fields(strings.ToLower(text))
	concepts := make([]string, 0)

	// Filter for potential concepts (nouns, capitalized words, etc.)
	for _, word := range words {
		if len(word) > 4 && !isStopWord(word) {
			concepts = append(concepts, word)
		}
	}

	return concepts
}

func isStopWord(word string) bool {
	stopWords := map[string]bool{
		"about": true, "above": true, "after": true, "again": true,
		"against": true, "because": true, "before": true, "being": true,
		"between": true, "could": true, "should": true, "would": true,
		"these": true, "those": true, "their": true, "there": true,
		"where": true, "which": true, "while": true, "with": true,
	}
	return stopWords[word]
}

// ============================================================================
// Symbolic Verifier
// ============================================================================

// SymbolicVerifier verifies hypotheses using symbolic reasoning.
type SymbolicVerifier struct {
	knowledgeBase *LogicKnowledgeBase
	maxDepth      int
	timeout       time.Duration
}

// NewSymbolicVerifier creates a new symbolic verifier.
func NewSymbolicVerifier(kb *LogicKnowledgeBase) *SymbolicVerifier {
	return &SymbolicVerifier{
		knowledgeBase: kb,
		maxDepth:      10,
		timeout:       5 * time.Second,
	}
}

// Prove attempts to prove a hypothesis.
func (sv *SymbolicVerifier) Prove(hypothesis *Hypothesis, kb *LogicKnowledgeBase) (*Proof, error) {
	startTime := time.Now()

	// Parse hypothesis into LogicPredicates
	goal := sv.parseHypothesis(hypothesis)

	// Attempt backward chaining proof
	proofSteps, success := sv.backwardChain(goal, kb, 0, make(map[string]bool))

	proof := &Proof{
		ID:           fmt.Sprintf("proof-%s-%d", hypothesis.ID, time.Now().UnixNano()),
		HypothesisID: hypothesis.ID,
		Valid:        success,
		Steps:        proofSteps,
		Conclusion:   goal,
		Duration:     time.Since(startTime),
	}

	if !success {
		// Generate counterexample
		counterexample := sv.findCounterexample(goal, kb)
		return proof, &CounterexampleError{
			Counterexample: counterexample,
		}
	}

	return proof, nil
}

// parseHypothesis converts a hypothesis statement into a LogicPredicate.
func (sv *SymbolicVerifier) parseHypothesis(hypothesis *Hypothesis) LogicPredicate {
	// Simple parsing - in reality this would use NLP
	return LogicPredicate{
		Name:    "holds",
		Args:    []interface{}{hypothesis.Statement},
		Negated: false,
	}
}

// backwardChain attempts to prove a goal using backward chaining.
func (sv *SymbolicVerifier) backwardChain(
	goal LogicPredicate,
	kb *LogicKnowledgeBase,
	depth int,
	visited map[string]bool,
) ([]ProofStep, bool) {
	// Check depth limit
	if depth >= sv.maxDepth {
		return nil, false
	}

	// Check if already visited (avoid cycles)
	goalStr := goal.String()
	if visited[goalStr] {
		return nil, false
	}
	visited[goalStr] = true

	// Check if goal is a known fact
	facts := kb.Query(goal)
	if len(facts) > 0 {
		step := ProofStep{
			StepNum:       depth + 1,
			Rule:          "fact",
			Premises:      nil,
			Conclusion:    goal,
			Justification: "Known fact in knowledge base",
		}
		return []ProofStep{step}, true
	}

	// Try to prove using rules
	rules := kb.GetApplicableRules(goal)
	for _, rule := range rules {
		// Try to prove all premises
		allProved := true
		allSteps := make([]ProofStep, 0)

		for _, premise := range rule.Premises {
			premiseSteps, proved := sv.backwardChain(premise, kb, depth+1, visited)
			if !proved {
				allProved = false
				break
			}
			allSteps = append(allSteps, premiseSteps...)
		}

		if allProved {
			// Add the final derivation step
			step := ProofStep{
				StepNum:       len(allSteps) + 1,
				Rule:          rule.Name,
				Premises:      rule.Premises,
				Conclusion:    goal,
				Justification: fmt.Sprintf("By rule '%s'", rule.Name),
			}
			allSteps = append(allSteps, step)
			return allSteps, true
		}
	}

	delete(visited, goalStr)
	return nil, false
}

// findCounterexample attempts to find a counterexample.
func (sv *SymbolicVerifier) findCounterexample(goal LogicPredicate, kb *LogicKnowledgeBase) Counterexample {
	// Look for negation of goal
	negatedGoal := goal
	negatedGoal.Negated = !goal.Negated

	facts := kb.Query(negatedGoal)
	if len(facts) > 0 {
		return Counterexample{
			Description:    fmt.Sprintf("Found contradicting fact: %s", facts[0].String()),
			LogicPredicate: facts[0],
		}
	}

	return Counterexample{
		Description:    fmt.Sprintf("Could not prove: %s", goal.String()),
		LogicPredicate: goal,
	}
}

// CounterexampleError is returned when verification finds a counterexample.
type CounterexampleError struct {
	Counterexample Counterexample
}

func (e *CounterexampleError) Error() string {
	return fmt.Sprintf("verification failed: %s", e.Counterexample.Description)
}

// ============================================================================
// Neurosymbolic Reasoner (Integration)
// ============================================================================

// NeurosymbolicReasoner combines neural and symbolic reasoning.
type NeurosymbolicReasoner struct {
	neuralReasoner   *NeuralReasoner
	symbolicVerifier *SymbolicVerifier
	knowledgeBase    *LogicKnowledgeBase

	// Configuration
	maxHypotheses    int
	refinementRounds int

	// Statistics
	stats   *ReasoningStats
	statsMu sync.RWMutex
}

// ReasoningStats tracks reasoning statistics.
type ReasoningStats struct {
	TotalQueries         int64
	HypothesesGenerated  int64
	VerificationsSuccess int64
	VerificationsFailed  int64
	AverageProofDepth    float64
	AverageLatency       time.Duration
}

// NeurosymbolicConfig configures the reasoner.
type NeurosymbolicConfig struct {
	MaxHypotheses    int
	RefinementRounds int
	ProofMaxDepth    int
	ProofTimeout     time.Duration
}

// DefaultNeurosymbolicConfig returns default configuration.
func DefaultNeurosymbolicConfig() *NeurosymbolicConfig {
	return &NeurosymbolicConfig{
		MaxHypotheses:    10,
		RefinementRounds: 3,
		ProofMaxDepth:    10,
		ProofTimeout:     5 * time.Second,
	}
}

// NewNeurosymbolicReasoner creates a new neurosymbolic reasoner.
func NewNeurosymbolicReasoner(
	retriever *SubLinearRetriever,
	config *NeurosymbolicConfig,
) *NeurosymbolicReasoner {
	if config == nil {
		config = DefaultNeurosymbolicConfig()
	}

	kb := NewLogicKnowledgeBase()
	nr := NewNeuralReasoner(retriever)
	sv := NewSymbolicVerifier(kb)

	sv.maxDepth = config.ProofMaxDepth
	sv.timeout = config.ProofTimeout

	return &NeurosymbolicReasoner{
		neuralReasoner:   nr,
		symbolicVerifier: sv,
		knowledgeBase:    kb,
		maxHypotheses:    config.MaxHypotheses,
		refinementRounds: config.RefinementRounds,
		stats:            &ReasoningStats{},
	}
}

// Reason performs neurosymbolic reasoning on a query.
func (r *NeurosymbolicReasoner) Reason(query *Query) (*Conclusion, *Proof) {
	startTime := time.Now()
	r.statsMu.Lock()
	r.stats.TotalQueries++
	r.statsMu.Unlock()

	// Step 1: Neural hypothesis generation
	hypotheses := r.neuralReasoner.GenerateHypotheses(query)
	r.statsMu.Lock()
	r.stats.HypothesesGenerated += int64(len(hypotheses))
	r.statsMu.Unlock()

	if len(hypotheses) > r.maxHypotheses {
		hypotheses = hypotheses[:r.maxHypotheses]
	}

	// Step 2: Try to verify each hypothesis
	for round := 0; round < r.refinementRounds; round++ {
		for _, hypothesis := range hypotheses {
			// Skip low confidence hypotheses
			if hypothesis.Confidence < 0.3 {
				continue
			}

			// Attempt symbolic verification
			proof, err := r.symbolicVerifier.Prove(hypothesis, r.knowledgeBase)

			if err == nil && proof.Valid {
				// Verified hypothesis found!
				r.statsMu.Lock()
				r.stats.VerificationsSuccess++
				r.updateAverageLatency(time.Since(startTime))
				if len(proof.Steps) > 0 {
					r.updateAverageDepth(float64(len(proof.Steps)))
				}
				r.statsMu.Unlock()

				conclusion := &Conclusion{
					ID:           fmt.Sprintf("conclusion-%d", time.Now().UnixNano()),
					Statement:    hypothesis.Statement,
					Confidence:   hypothesis.Confidence,
					Verified:     true,
					VerifiedAt:   time.Now(),
					HypothesisID: hypothesis.ID,
					ProofID:      proof.ID,
				}

				return conclusion, proof
			}

			// If verification failed, try to refine using counterexample
			if counterErr, ok := err.(*CounterexampleError); ok {
				hypotheses = r.neuralReasoner.RefineHypotheses(hypotheses, &counterErr.Counterexample)
			}
		}
	}

	// No verified hypothesis found
	r.statsMu.Lock()
	r.stats.VerificationsFailed++
	r.updateAverageLatency(time.Since(startTime))
	r.statsMu.Unlock()

	// Return best unverified hypothesis as conclusion
	if len(hypotheses) > 0 {
		best := hypotheses[0]
		return &Conclusion{
			ID:           fmt.Sprintf("conclusion-unverified-%d", time.Now().UnixNano()),
			Statement:    best.Statement,
			Confidence:   best.Confidence * 0.5, // Lower confidence for unverified
			Verified:     false,
			HypothesisID: best.ID,
		}, nil
	}

	return nil, nil
}

// updateAverageLatency updates running average.
func (r *NeurosymbolicReasoner) updateAverageLatency(latency time.Duration) {
	total := r.stats.VerificationsSuccess + r.stats.VerificationsFailed
	if total == 1 {
		r.stats.AverageLatency = latency
	} else {
		n := float64(total)
		r.stats.AverageLatency = time.Duration(
			(float64(r.stats.AverageLatency)*(n-1) + float64(latency)) / n,
		)
	}
}

// updateAverageDepth updates running average.
func (r *NeurosymbolicReasoner) updateAverageDepth(depth float64) {
	if r.stats.VerificationsSuccess == 1 {
		r.stats.AverageProofDepth = depth
	} else {
		n := float64(r.stats.VerificationsSuccess)
		r.stats.AverageProofDepth = (r.stats.AverageProofDepth*(n-1) + depth) / n
	}
}

// AddFact adds a fact to the knowledge base.
func (r *NeurosymbolicReasoner) AddFact(pred LogicPredicate) {
	r.knowledgeBase.AddFact(pred)
}

// AddRule adds an inference rule to the knowledge base.
func (r *NeurosymbolicReasoner) AddRule(rule Rule) {
	r.knowledgeBase.AddRule(rule)
}

// AddPattern adds a reasoning pattern.
func (r *NeurosymbolicReasoner) AddPattern(pattern *ReasoningPattern) {
	r.neuralReasoner.AddPattern(pattern)
}

// GetStats returns reasoning statistics.
func (r *NeurosymbolicReasoner) GetStats() *ReasoningStats {
	r.statsMu.RLock()
	defer r.statsMu.RUnlock()

	return &ReasoningStats{
		TotalQueries:         r.stats.TotalQueries,
		HypothesesGenerated:  r.stats.HypothesesGenerated,
		VerificationsSuccess: r.stats.VerificationsSuccess,
		VerificationsFailed:  r.stats.VerificationsFailed,
		AverageProofDepth:    r.stats.AverageProofDepth,
		AverageLatency:       r.stats.AverageLatency,
	}
}

// GetKnowledgeBase returns the knowledge base for inspection.
func (r *NeurosymbolicReasoner) GetKnowledgeBase() *LogicKnowledgeBase {
	return r.knowledgeBase
}
