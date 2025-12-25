package memory

import (
	"testing"
	"time"
)

// ============================================================================
// Impasse Detector Tests
// ============================================================================

func TestNewImpasseDetector(t *testing.T) {
	detector := NewImpasseDetector(nil, nil)

	if detector == nil {
		t.Fatal("NewImpasseDetector returned nil")
	}

	if detector.ActiveCount() != 0 {
		t.Errorf("Expected 0 active impasses, got %d", detector.ActiveCount())
	}
}

func TestImpasseDetector_DetectTie(t *testing.T) {
	detector := NewImpasseDetector(nil, nil)

	// No tie - clear winner
	imp := detector.DetectTie("goal-1", []string{"agent-1", "agent-2"}, []float64{1.0, 0.5})
	if imp != nil {
		t.Error("Should not detect tie when clear winner exists")
	}

	// Tie - very close scores
	imp = detector.DetectTie("goal-1", []string{"agent-1", "agent-2", "agent-3"}, []float64{0.95, 0.97, 0.96})
	if imp == nil {
		t.Fatal("Should detect tie with close scores")
	}

	if imp.Type != ImpasseTie {
		t.Errorf("Expected TIE type, got %s", imp.Type)
	}

	if len(imp.Candidates) < 2 {
		t.Errorf("Expected at least 2 tied candidates, got %d", len(imp.Candidates))
	}

	if imp.Severity != 0.3 {
		t.Errorf("Expected severity 0.3, got %f", imp.Severity)
	}

	t.Logf("Detected tie: %s with candidates %v", imp.ID, imp.Candidates)
}

func TestImpasseDetector_DetectNoMatch(t *testing.T) {
	detector := NewImpasseDetector(nil, nil)

	imp := detector.DetectNoMatch("goal-1", "no agent matches the required capabilities")
	if imp == nil {
		t.Fatal("Should detect no-match impasse")
	}

	if imp.Type != ImpasseNoMatch {
		t.Errorf("Expected NO_MATCH type, got %s", imp.Type)
	}

	if imp.GoalID != "goal-1" {
		t.Errorf("Expected goal ID 'goal-1', got '%s'", imp.GoalID)
	}
}

func TestImpasseDetector_DetectFailure(t *testing.T) {
	detector := NewImpasseDetector(nil, nil)

	imp := detector.DetectFailure("goal-1", "APEX-01", "API rate limit exceeded")
	if imp == nil {
		t.Fatal("Should detect failure impasse")
	}

	if imp.Type != ImpasseFailure {
		t.Errorf("Expected FAILURE type, got %s", imp.Type)
	}

	if imp.FailedAgent != "APEX-01" {
		t.Errorf("Expected failed agent 'APEX-01', got '%s'", imp.FailedAgent)
	}

	if imp.FailureReason != "API rate limit exceeded" {
		t.Error("Failure reason mismatch")
	}
}

func TestImpasseDetector_DetectConflict(t *testing.T) {
	detector := NewImpasseDetector(nil, nil)

	// Single result - no conflict
	imp := detector.DetectConflict("goal-1", map[string]interface{}{"agent-1": "result"})
	if imp != nil {
		t.Error("Should not detect conflict with single result")
	}

	// Multiple results - conflict
	results := map[string]interface{}{
		"agent-1": "use microservices",
		"agent-2": "use monolith",
		"agent-3": "use serverless",
	}
	imp = detector.DetectConflict("goal-1", results)
	if imp == nil {
		t.Fatal("Should detect conflict with multiple results")
	}

	if imp.Type != ImpasseConflict {
		t.Errorf("Expected CONFLICT type, got %s", imp.Type)
	}

	if len(imp.ConflictingResults) != 3 {
		t.Errorf("Expected 3 conflicting results, got %d", len(imp.ConflictingResults))
	}
}

func TestImpasseDetector_DetectCapacity(t *testing.T) {
	detector := NewImpasseDetector(nil, nil)

	imp := detector.DetectCapacity("goal-1", "working memory")
	if imp == nil {
		t.Fatal("Should detect capacity impasse")
	}

	if imp.Type != ImpasseCapacity {
		t.Errorf("Expected CAPACITY type, got %s", imp.Type)
	}

	if imp.Context["resource"] != "working memory" {
		t.Error("Resource not stored in context")
	}
}

func TestImpasseDetector_DetectNoChange(t *testing.T) {
	config := DefaultImpasseDetectorConfig()
	config.NoChangeThreshold = 3
	detector := NewImpasseDetector(config, nil)

	// Below threshold
	imp := detector.DetectNoChange("goal-1", 2)
	if imp != nil {
		t.Error("Should not detect no-change below threshold")
	}

	// At threshold
	imp = detector.DetectNoChange("goal-1", 3)
	if imp == nil {
		t.Fatal("Should detect no-change at threshold")
	}

	if imp.Type != ImpasseNoChange {
		t.Errorf("Expected NO_CHANGE type, got %s", imp.Type)
	}
}

func TestImpasseDetector_DetectConstraint(t *testing.T) {
	detector := NewImpasseDetector(nil, nil)

	imp := detector.DetectConstraint("goal-1", "max_tokens exceeded")
	if imp == nil {
		t.Fatal("Should detect constraint impasse")
	}

	if imp.Type != ImpasseConstraint {
		t.Errorf("Expected CONSTRAINT type, got %s", imp.Type)
	}

	if imp.ConstraintViolated != "max_tokens exceeded" {
		t.Error("Constraint not recorded")
	}
}

func TestImpasseDetector_DetectTimeout(t *testing.T) {
	config := DefaultImpasseDetectorConfig()
	config.TimeoutThreshold = 1 * time.Second
	detector := NewImpasseDetector(config, nil)

	// Below threshold
	imp := detector.DetectTimeout("goal-1", 500*time.Millisecond)
	if imp != nil {
		t.Error("Should not detect timeout below threshold")
	}

	// Above threshold
	imp = detector.DetectTimeout("goal-1", 2*time.Second)
	if imp == nil {
		t.Fatal("Should detect timeout above threshold")
	}

	if imp.Type != ImpasseTimeout {
		t.Errorf("Expected TIMEOUT type, got %s", imp.Type)
	}
}

func TestImpasseDetector_Resolution(t *testing.T) {
	detector := NewImpasseDetector(nil, nil)

	// Create a tie impasse
	imp := detector.DetectTie("goal-1", []string{"agent-1", "agent-2"}, []float64{0.98, 0.97})
	if imp == nil {
		t.Fatal("Should detect tie")
	}

	// Resolve it
	result, err := detector.Resolve(imp.ID)
	if err != nil {
		t.Fatalf("Resolution failed: %v", err)
	}

	if !result.Success {
		t.Error("Resolution should succeed")
	}

	// Impasse should be marked resolved
	resolved, _ := detector.Get(imp.ID)
	if !resolved.IsResolved() {
		t.Error("Impasse should be marked as resolved")
	}

	t.Logf("Resolution: strategy=%s, candidate=%s, message=%s",
		result.Strategy, result.SelectedCandidate, result.Message)
}

func TestImpasseDetector_ResolutionStrategies(t *testing.T) {
	tests := []struct {
		name     string
		create   func(*ImpasseDetector) *Impasse
		expected []ResolutionStrategy
	}{
		{
			name: "Tie resolution uses Random/Consensus",
			create: func(d *ImpasseDetector) *Impasse {
				return d.DetectTie("g", []string{"a", "b"}, []float64{1.0, 1.0})
			},
			expected: []ResolutionStrategy{StrategyRandom, StrategyConsensus, StrategyEscalate},
		},
		{
			name: "NoMatch uses Decompose/Escalate",
			create: func(d *ImpasseDetector) *Impasse {
				return d.DetectNoMatch("g", "no match")
			},
			expected: []ResolutionStrategy{StrategyDecompose, StrategyEscalate, StrategyFallback},
		},
		{
			name: "Failure uses Retry/Backoff",
			create: func(d *ImpasseDetector) *Impasse {
				return d.DetectFailure("g", "agent", "failed")
			},
			expected: []ResolutionStrategy{StrategyRetry, StrategyBackoff, StrategyFallback, StrategyAbort},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := NewImpasseDetector(nil, nil)
			imp := tt.create(detector)
			if imp == nil {
				t.Skip("Impasse not created")
			}

			result, err := detector.Resolve(imp.ID)
			if err != nil {
				t.Fatalf("Resolution failed: %v", err)
			}

			// Check that the used strategy is in expected list
			found := false
			for _, exp := range tt.expected {
				if result.Strategy == exp {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Strategy %s not in expected list %v", result.Strategy, tt.expected)
			}
		})
	}
}

func TestImpasseDetector_EscalationMapping(t *testing.T) {
	detector := NewImpasseDetector(nil, nil)

	// Test that specific impasse types escalate to appropriate agents
	tests := []struct {
		impasseType   ImpasseType
		expectedAgent string
	}{
		{ImpasseTie, "OMNISCIENT-20"},
		{ImpasseConflict, "ARBITER-39"},
		{ImpasseNoChange, "GENESIS-19"},
		{ImpasseTimeout, "VELOCITY-05"},
	}

	for _, tt := range tests {
		t.Run(tt.impasseType.String(), func(t *testing.T) {
			// Create impasse manually and test escalation
			imp := &Impasse{
				ID:         "test-imp",
				Type:       tt.impasseType,
				GoalID:     "test-goal",
				Candidates: []string{"a", "b"}, // Needed for some types
			}
			detector.mu.Lock()
			detector.impasses[imp.ID] = imp
			detector.activeImpasses[imp.ID] = imp
			detector.mu.Unlock()

			// Apply escalation strategy directly
			result, _ := detector.applyStrategy(imp, StrategyEscalate)
			if result.EscalatedTo != tt.expectedAgent {
				t.Errorf("Expected escalation to %s, got %s", tt.expectedAgent, result.EscalatedTo)
			}

			detector.Clear()
		})
	}
}

func TestImpasseDetector_CustomResolver(t *testing.T) {
	detector := NewImpasseDetector(nil, nil)

	customResolved := false
	detector.RegisterResolver(ImpasseTie, func(imp *Impasse) (*ResolutionResult, error) {
		customResolved = true
		return &ResolutionResult{
			Success:           true,
			SelectedCandidate: "custom-selection",
			Message:           "resolved by custom resolver",
		}, nil
	})

	imp := detector.DetectTie("goal-1", []string{"a", "b"}, []float64{1.0, 1.0})
	result, err := detector.Resolve(imp.ID)

	if err != nil {
		t.Fatalf("Resolution failed: %v", err)
	}

	if !customResolved {
		t.Error("Custom resolver should have been called")
	}

	if result.SelectedCandidate != "custom-selection" {
		t.Errorf("Expected 'custom-selection', got '%s'", result.SelectedCandidate)
	}
}

func TestImpasseDetector_Callbacks(t *testing.T) {
	detector := NewImpasseDetector(nil, nil)

	detectedCount := 0
	resolvedCount := 0

	detector.OnImpasseDetected(func(imp *Impasse) {
		detectedCount++
	})

	detector.OnImpasseResolved(func(imp *Impasse, result *ResolutionResult) {
		resolvedCount++
	})

	imp := detector.DetectNoMatch("goal-1", "test")
	if detectedCount != 1 {
		t.Errorf("Expected 1 detection callback, got %d", detectedCount)
	}

	detector.Resolve(imp.ID)
	if resolvedCount != 1 {
		t.Errorf("Expected 1 resolution callback, got %d", resolvedCount)
	}
}

func TestImpasseDetector_CapacityLimit(t *testing.T) {
	config := DefaultImpasseDetectorConfig()
	config.MaxActiveImpasses = 3
	detector := NewImpasseDetector(config, nil)

	// Create impasses up to limit (use unique goal IDs)
	for i := 0; i < 3; i++ {
		detector.DetectNoMatch("goal-"+string(rune('a'+i)), "test")
	}

	if detector.ActiveCount() != 3 {
		t.Errorf("Expected 3 active impasses, got %d", detector.ActiveCount())
	}

	// Next one should trigger capacity impasse
	imp := detector.DetectNoMatch("goal-overflow", "test")
	if imp.Type != ImpasseCapacity {
		t.Errorf("Expected CAPACITY impasse when limit reached, got %s", imp.Type)
	}
}

func TestImpasseDetector_GetByType(t *testing.T) {
	detector := NewImpasseDetector(nil, nil)

	detector.DetectNoMatch("g1", "test1")
	detector.DetectNoMatch("g2", "test2")
	detector.DetectFailure("g3", "agent", "failed")

	noMatchImpasses := detector.GetByType(ImpasseNoMatch)
	if len(noMatchImpasses) != 2 {
		t.Errorf("Expected 2 NO_MATCH impasses, got %d", len(noMatchImpasses))
	}

	failureImpasses := detector.GetByType(ImpasseFailure)
	if len(failureImpasses) != 1 {
		t.Errorf("Expected 1 FAILURE impasse, got %d", len(failureImpasses))
	}
}

func TestImpasseDetector_GetByGoal(t *testing.T) {
	detector := NewImpasseDetector(nil, nil)

	detector.DetectNoMatch("goal-1", "test1")
	detector.DetectFailure("goal-1", "agent", "failed")
	detector.DetectNoMatch("goal-2", "test2")

	goal1Impasses := detector.GetByGoal("goal-1")
	if len(goal1Impasses) != 2 {
		t.Errorf("Expected 2 impasses for goal-1, got %d", len(goal1Impasses))
	}

	goal2Impasses := detector.GetByGoal("goal-2")
	if len(goal2Impasses) != 1 {
		t.Errorf("Expected 1 impasse for goal-2, got %d", len(goal2Impasses))
	}
}

func TestImpasseDetector_Stats(t *testing.T) {
	detector := NewImpasseDetector(nil, nil)

	// Create some impasses
	imp1 := detector.DetectNoMatch("g1", "test")
	detector.DetectFailure("g2", "agent", "failed")
	detector.DetectTie("g3", []string{"a", "b"}, []float64{1.0, 1.0})

	// Resolve one
	detector.Resolve(imp1.ID)

	stats := detector.GetStats()

	if stats.TotalDetected != 3 {
		t.Errorf("Expected 3 detected, got %d", stats.TotalDetected)
	}

	if stats.TotalResolved != 1 {
		t.Errorf("Expected 1 resolved, got %d", stats.TotalResolved)
	}

	if stats.ByType[ImpasseNoMatch] != 1 {
		t.Errorf("Expected 1 NO_MATCH, got %d", stats.ByType[ImpasseNoMatch])
	}

	t.Logf("Stats: Detected=%d, Resolved=%d, Failed=%d",
		stats.TotalDetected, stats.TotalResolved, stats.TotalFailed)
}

func TestImpasseDetector_Snapshot(t *testing.T) {
	detector := NewImpasseDetector(nil, nil)

	detector.DetectNoMatch("g1", "test")
	imp := detector.DetectFailure("g2", "agent", "failed")
	detector.Resolve(imp.ID)

	snapshot := detector.Snapshot()

	if snapshot.ActiveCount != 1 {
		t.Errorf("Expected 1 active, got %d", snapshot.ActiveCount)
	}

	if snapshot.ResolvedCount != 1 {
		t.Errorf("Expected 1 resolved, got %d", snapshot.ResolvedCount)
	}

	t.Logf("Snapshot: Active=%d, Resolved=%d", snapshot.ActiveCount, snapshot.ResolvedCount)
}

func TestImpasseType_String(t *testing.T) {
	tests := []struct {
		impasseType ImpasseType
		expected    string
	}{
		{ImpasseTie, "TIE"},
		{ImpasseNoMatch, "NO_MATCH"},
		{ImpasseFailure, "FAILURE"},
		{ImpasseConflict, "CONFLICT"},
		{ImpasseCapacity, "CAPACITY"},
		{ImpasseNoChange, "NO_CHANGE"},
		{ImpasseConstraint, "CONSTRAINT"},
		{ImpasseTimeout, "TIMEOUT"},
	}

	for _, tt := range tests {
		if tt.impasseType.String() != tt.expected {
			t.Errorf("Expected '%s', got '%s'", tt.expected, tt.impasseType.String())
		}
	}
}

func TestResolutionStrategy_String(t *testing.T) {
	tests := []struct {
		strategy ResolutionStrategy
		expected string
	}{
		{StrategyDecompose, "DECOMPOSE"},
		{StrategyEscalate, "ESCALATE"},
		{StrategyRandom, "RANDOM"},
		{StrategyConsensus, "CONSENSUS"},
		{StrategyRetry, "RETRY"},
		{StrategyBackoff, "BACKOFF"},
		{StrategyFallback, "FALLBACK"},
		{StrategyAbort, "ABORT"},
		{StrategyAsk, "ASK"},
		{StrategyLearn, "LEARN"},
	}

	for _, tt := range tests {
		if tt.strategy.String() != tt.expected {
			t.Errorf("Expected '%s', got '%s'", tt.expected, tt.strategy.String())
		}
	}
}

func TestImpasseDetector_Clear(t *testing.T) {
	detector := NewImpasseDetector(nil, nil)

	detector.DetectNoMatch("g1", "test")
	detector.DetectFailure("g2", "agent", "failed")

	if detector.ActiveCount() != 2 {
		t.Errorf("Expected 2 active before clear, got %d", detector.ActiveCount())
	}

	detector.Clear()

	if detector.ActiveCount() != 0 {
		t.Errorf("Expected 0 active after clear, got %d", detector.ActiveCount())
	}
}

// ============================================================================
// Benchmarks
// ============================================================================

func BenchmarkImpasseDetector_DetectTie(b *testing.B) {
	detector := NewImpasseDetector(nil, nil)
	candidates := []string{"agent-1", "agent-2", "agent-3"}
	scores := []float64{0.95, 0.97, 0.96}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		detector.DetectTie("goal", candidates, scores)
	}
}

func BenchmarkImpasseDetector_Resolve(b *testing.B) {
	detector := NewImpasseDetector(nil, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		imp := detector.DetectTie("goal", []string{"a", "b"}, []float64{1.0, 1.0})
		detector.Resolve(imp.ID)
	}
}

func BenchmarkImpasseDetector_GetActive(b *testing.B) {
	detector := NewImpasseDetector(nil, nil)

	// Create 10 active impasses
	for i := 0; i < 10; i++ {
		detector.DetectNoMatch("goal", "test")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		detector.GetActive()
	}
}
