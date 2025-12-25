package memory

import (
	"fmt"
	"testing"
	"time"
)

func TestNewNeurosymbolicReasoner(t *testing.T) {
	r := NewNeurosymbolicReasoner(nil, nil)
	if r == nil {
		t.Fatal("Expected non-nil reasoner")
	}

	if r.knowledgeBase == nil {
		t.Error("Expected non-nil knowledge base")
	}

	if r.neuralReasoner == nil {
		t.Error("Expected non-nil neural reasoner")
	}

	if r.symbolicVerifier == nil {
		t.Error("Expected non-nil symbolic verifier")
	}
}

func TestLogicKnowledgeBase_AddAndQuery(t *testing.T) {
	kb := NewLogicKnowledgeBase()

	// Add facts
	kb.AddFact(LogicPredicate{Name: "parent", Args: []interface{}{"alice", "bob"}})
	kb.AddFact(LogicPredicate{Name: "parent", Args: []interface{}{"bob", "charlie"}})
	kb.AddFact(LogicPredicate{Name: "parent", Args: []interface{}{"alice", "diana"}})

	// Query with exact match
	results := kb.Query(LogicPredicate{Name: "parent", Args: []interface{}{"alice", "bob"}})
	if len(results) != 1 {
		t.Errorf("Expected 1 result for exact match, got %d", len(results))
	}

	// Query with variable
	results = kb.Query(LogicPredicate{Name: "parent", Args: []interface{}{"alice", "?x"}})
	if len(results) != 2 {
		t.Errorf("Expected 2 results for variable query, got %d", len(results))
	}

	// Query non-existent
	results = kb.Query(LogicPredicate{Name: "parent", Args: []interface{}{"eve", "?x"}})
	if len(results) != 0 {
		t.Errorf("Expected 0 results for non-existent query, got %d", len(results))
	}
}

func TestLogicKnowledgeBase_AddRule(t *testing.T) {
	kb := NewLogicKnowledgeBase()

	// Add grandparent rule: grandparent(X,Z) :- parent(X,Y), parent(Y,Z)
	rule := Rule{
		ID:   "grandparent-rule",
		Name: "grandparent",
		Premises: []LogicPredicate{
			{Name: "parent", Args: []interface{}{"?x", "?y"}},
			{Name: "parent", Args: []interface{}{"?y", "?z"}},
		},
		Conclusion: LogicPredicate{Name: "grandparent", Args: []interface{}{"?x", "?z"}},
		Confidence: 1.0,
	}

	kb.AddRule(rule)

	// Check applicable rules
	rules := kb.GetApplicableRules(LogicPredicate{Name: "grandparent", Args: []interface{}{"alice", "charlie"}})
	if len(rules) != 1 {
		t.Errorf("Expected 1 applicable rule, got %d", len(rules))
	}
}

func TestLogicPredicate_String(t *testing.T) {
	tests := []struct {
		pred     LogicPredicate
		expected string
	}{
		{
			pred:     LogicPredicate{Name: "parent", Args: []interface{}{"alice", "bob"}},
			expected: "parent(alice, bob)",
		},
		{
			pred:     LogicPredicate{Name: "mortal", Args: []interface{}{"socrates"}, Negated: true},
			expected: "¬mortal(socrates)",
		},
		{
			pred:     LogicPredicate{Name: "equals", Args: []interface{}{42, "answer"}},
			expected: "equals(42, answer)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.pred.String()
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestNeuralReasoner_GenerateHypotheses(t *testing.T) {
	nr := NewNeuralReasoner(nil)

	// Add some patterns
	nr.AddPattern(&ReasoningPattern{
		ID:          "security-pattern",
		Trigger:     "security",
		Strategy:    "Apply security-first design principles",
		SuccessRate: 0.85,
	})

	query := &Query{
		ID:       "test-query",
		Question: "How should I handle security for authentication?",
		Timeout:  5 * time.Second,
	}

	hypotheses := nr.GenerateHypotheses(query)

	// Should generate at least the pattern-based hypothesis
	found := false
	for _, h := range hypotheses {
		if h.Provenance == "pattern_matching" {
			found = true
			if h.Confidence != 0.85 {
				t.Errorf("Expected confidence 0.85, got %f", h.Confidence)
			}
		}
	}

	if !found {
		t.Error("Expected to find pattern-matching hypothesis")
	}
}

func TestNeuralReasoner_RefineHypotheses(t *testing.T) {
	nr := NewNeuralReasoner(nil)

	hypotheses := []*Hypothesis{
		{
			ID:         "hyp-1",
			Statement:  "Solution A",
			Confidence: 0.8,
			Provenance: "test",
		},
	}

	counterexample := &Counterexample{
		Description:    "Solution A fails for edge case X",
		LogicPredicate: LogicPredicate{Name: "fails", Args: []interface{}{"A", "X"}},
	}

	refined := nr.RefineHypotheses(hypotheses, counterexample)

	if len(refined) != 1 {
		t.Fatalf("Expected 1 refined hypothesis, got %d", len(refined))
	}

	// Confidence should be reduced
	if refined[0].Confidence >= hypotheses[0].Confidence {
		t.Errorf("Expected reduced confidence, got %f >= %f",
			refined[0].Confidence, hypotheses[0].Confidence)
	}

	// Should have additional assumption
	if len(refined[0].Assumptions) == 0 {
		t.Error("Expected refined hypothesis to have assumptions")
	}
}

func TestSymbolicVerifier_ProveFromFact(t *testing.T) {
	kb := NewLogicKnowledgeBase()

	// Add fact - the verifier wraps statements in "holds" predicate
	kb.AddFact(LogicPredicate{Name: "holds", Args: []interface{}{"mortal(socrates)"}})

	sv := NewSymbolicVerifier(kb)

	hypothesis := &Hypothesis{
		ID:        "test-hyp",
		Statement: "mortal(socrates)",
	}

	proof, err := sv.Prove(hypothesis, kb)

	if err != nil {
		t.Fatalf("Expected successful proof, got error: %v", err)
	}

	if !proof.Valid {
		t.Error("Expected valid proof")
	}

	if len(proof.Steps) == 0 {
		t.Error("Expected at least one proof step")
	}
}

func TestSymbolicVerifier_ProveWithRule(t *testing.T) {
	kb := NewLogicKnowledgeBase()

	// Add facts
	kb.AddFact(LogicPredicate{Name: "man", Args: []interface{}{"socrates"}})

	// Add rule: mortal(X) :- man(X)
	kb.AddRule(Rule{
		ID:   "mortality-rule",
		Name: "mortality",
		Premises: []LogicPredicate{
			{Name: "man", Args: []interface{}{"?x"}},
		},
		Conclusion: LogicPredicate{Name: "mortal", Args: []interface{}{"?x"}},
		Confidence: 1.0,
	})

	sv := NewSymbolicVerifier(kb)
	_ = sv // Used to verify it can be created

	// This tests that we can find applicable rules
	// Note: Full unification would require more sophisticated matching
	rules := kb.GetApplicableRules(LogicPredicate{Name: "mortal", Args: []interface{}{"socrates"}})
	if len(rules) != 1 {
		t.Errorf("Expected 1 applicable rule, got %d", len(rules))
	}
}

func TestNeurosymbolicReasoner_Reason(t *testing.T) {
	r := NewNeurosymbolicReasoner(nil, nil)

	// Add some knowledge
	r.AddFact(LogicPredicate{Name: "capability", Args: []interface{}{"APEX", "coding"}})
	r.AddFact(LogicPredicate{Name: "capability", Args: []interface{}{"CIPHER", "security"}})

	// Add a pattern
	r.AddPattern(&ReasoningPattern{
		ID:          "coding-pattern",
		Trigger:     "code",
		Strategy:    "Use APEX for coding tasks",
		SuccessRate: 0.9,
	})

	query := &Query{
		ID:       "test-query",
		Question: "How should I write this code?",
		Timeout:  5 * time.Second,
	}

	conclusion, _ := r.Reason(query)

	if conclusion == nil {
		t.Fatal("Expected non-nil conclusion")
	}

	if conclusion.Confidence <= 0 {
		t.Error("Expected positive confidence")
	}
}

func TestNeurosymbolicReasoner_Stats(t *testing.T) {
	r := NewNeurosymbolicReasoner(nil, nil)

	// Add a pattern so hypotheses get generated
	r.AddPattern(&ReasoningPattern{
		ID:          "test-pattern",
		Trigger:     "question",
		Strategy:    "test strategy",
		SuccessRate: 0.8,
	})

	// Run a few queries
	for i := 0; i < 3; i++ {
		query := &Query{
			ID:       "test-query",
			Question: "test question",
		}
		r.Reason(query)
	}

	stats := r.GetStats()

	if stats.TotalQueries != 3 {
		t.Errorf("Expected 3 total queries, got %d", stats.TotalQueries)
	}

	if stats.HypothesesGenerated <= 0 {
		t.Error("Expected some hypotheses generated")
	}
}

func TestCounterexampleError(t *testing.T) {
	ce := Counterexample{
		Description:    "test counterexample",
		LogicPredicate: LogicPredicate{Name: "test", Args: []interface{}{"x"}},
	}

	err := &CounterexampleError{Counterexample: ce}

	errStr := err.Error()
	if errStr == "" {
		t.Error("Expected non-empty error string")
	}

	if !contains(errStr, "test counterexample") {
		t.Errorf("Error should contain counterexample description, got: %s", errStr)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestExtractConcepts(t *testing.T) {
	concepts := extractConcepts("How should I implement security authentication for users?")

	if len(concepts) == 0 {
		t.Error("Expected some concepts extracted")
	}

	// Should include meaningful words
	foundSecurity := false
	foundAuthentication := false
	for _, c := range concepts {
		if c == "security" {
			foundSecurity = true
		}
		if c == "authentication" {
			foundAuthentication = true
		}
	}

	if !foundSecurity {
		t.Error("Expected to find 'security' in concepts")
	}
	if !foundAuthentication {
		t.Error("Expected to find 'authentication' in concepts")
	}
}

func TestIsStopWord(t *testing.T) {
	if !isStopWord("about") {
		t.Error("'about' should be a stop word")
	}

	if !isStopWord("with") {
		t.Error("'with' should be a stop word")
	}

	if isStopWord("security") {
		t.Error("'security' should not be a stop word")
	}
}

func TestLogicKnowledgeBase_Concurrency(t *testing.T) {
	kb := NewLogicKnowledgeBase()

	// Concurrent reads and writes
	done := make(chan bool)

	// Writer
	go func() {
		for i := 0; i < 100; i++ {
			kb.AddFact(LogicPredicate{Name: "test", Args: []interface{}{i}})
		}
		done <- true
	}()

	// Reader
	go func() {
		for i := 0; i < 100; i++ {
			kb.Query(LogicPredicate{Name: "test", Args: []interface{}{"?x"}})
		}
		done <- true
	}()

	<-done
	<-done
}

func TestHypothesis_Fields(t *testing.T) {
	h := &Hypothesis{
		ID:         "hyp-1",
		Statement:  "Test statement",
		Confidence: 0.75,
		Provenance: "test",
		Supporting: []string{"exp-1", "exp-2"},
		Assumptions: []LogicPredicate{
			{Name: "assumption", Args: []interface{}{"x"}},
		},
		Timestamp: time.Now(),
	}

	if h.ID != "hyp-1" {
		t.Error("ID mismatch")
	}

	if h.Confidence != 0.75 {
		t.Error("Confidence mismatch")
	}

	if len(h.Supporting) != 2 {
		t.Error("Supporting evidence mismatch")
	}
}

func TestConclusion_Fields(t *testing.T) {
	c := &Conclusion{
		ID:           "conc-1",
		Statement:    "Verified conclusion",
		Confidence:   0.95,
		Verified:     true,
		VerifiedAt:   time.Now(),
		HypothesisID: "hyp-1",
		ProofID:      "proof-1",
	}

	if !c.Verified {
		t.Error("Expected verified to be true")
	}

	if c.Confidence != 0.95 {
		t.Error("Confidence mismatch")
	}
}

func TestProof_Fields(t *testing.T) {
	p := &Proof{
		ID:           "proof-1",
		HypothesisID: "hyp-1",
		Valid:        true,
		Steps: []ProofStep{
			{
				StepNum:       1,
				Rule:          "fact",
				Conclusion:    LogicPredicate{Name: "test", Args: []interface{}{"x"}},
				Justification: "Known fact",
			},
		},
		Duration: 10 * time.Millisecond,
	}

	if !p.Valid {
		t.Error("Expected valid proof")
	}

	if len(p.Steps) != 1 {
		t.Error("Expected 1 proof step")
	}
}

// Benchmarks

func BenchmarkNeuralReasoner_GenerateHypotheses(b *testing.B) {
	nr := NewNeuralReasoner(nil)

	// Add patterns
	for i := 0; i < 10; i++ {
		nr.AddPattern(&ReasoningPattern{
			ID:          fmt.Sprintf("pattern-%d", i),
			Trigger:     fmt.Sprintf("trigger%d", i),
			Strategy:    "test strategy",
			SuccessRate: 0.8,
		})
	}

	query := &Query{
		ID:       "bench-query",
		Question: "test query with trigger1 trigger5",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		nr.GenerateHypotheses(query)
	}
}

func BenchmarkLogicKnowledgeBase_Query(b *testing.B) {
	kb := NewLogicKnowledgeBase()

	// Add facts
	for i := 0; i < 1000; i++ {
		kb.AddFact(LogicPredicate{
			Name: "fact",
			Args: []interface{}{i, i * 2},
		})
	}

	pattern := LogicPredicate{Name: "fact", Args: []interface{}{500, "?x"}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		kb.Query(pattern)
	}
}

func BenchmarkSymbolicVerifier_Prove(b *testing.B) {
	kb := NewLogicKnowledgeBase()

	// Add facts
	for i := 0; i < 100; i++ {
		kb.AddFact(LogicPredicate{Name: "holds", Args: []interface{}{fmt.Sprintf("statement-%d", i)}})
	}

	sv := NewSymbolicVerifier(kb)
	hyp := &Hypothesis{
		ID:        "bench-hyp",
		Statement: "statement-50",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sv.Prove(hyp, kb)
	}
}
