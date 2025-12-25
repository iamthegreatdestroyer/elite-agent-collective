// Package memory provides the MNEMONIC system for the Elite Agent Collective.
// This file implements Impasse Detection & Resolution from @NEURAL's Cognitive Architecture Analysis.
//
// Impasses occur when the cognitive system cannot make progress. Based on SOAR:
// - Tie Impasse: Multiple operators/agents have equal preference
// - No-Change Impasse: Current goal cannot make progress
// - Operator No-Result: Applied operator produces no result
// - State No-Change: State doesn't change after operator application
// - Constraint Failure: Violates declared constraints
//
// Resolution strategies create subgoals to resolve impasses.

package memory

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"time"
)

// ============================================================================
// Impasse Types
// ============================================================================

// ImpasseType classifies the nature of the impasse.
type ImpasseType int

const (
	// ImpasseTie occurs when multiple agents/operators have equal preference
	ImpasseTie ImpasseType = iota

	// ImpasseNoMatch occurs when no agent/operator matches the current goal
	ImpasseNoMatch

	// ImpasseFailure occurs when the selected operator/agent fails
	ImpasseFailure

	// ImpasseConflict occurs when agents disagree on the solution
	ImpasseConflict

	// ImpasseCapacity occurs when resources are exhausted
	ImpasseCapacity

	// ImpasseNoChange occurs when no progress is being made
	ImpasseNoChange

	// ImpasseConstraint occurs when a constraint is violated
	ImpasseConstraint

	// ImpasseTimeout occurs when processing takes too long
	ImpasseTimeout
)

// String returns a human-readable impasse type.
func (t ImpasseType) String() string {
	switch t {
	case ImpasseTie:
		return "TIE"
	case ImpasseNoMatch:
		return "NO_MATCH"
	case ImpasseFailure:
		return "FAILURE"
	case ImpasseConflict:
		return "CONFLICT"
	case ImpasseCapacity:
		return "CAPACITY"
	case ImpasseNoChange:
		return "NO_CHANGE"
	case ImpasseConstraint:
		return "CONSTRAINT"
	case ImpasseTimeout:
		return "TIMEOUT"
	default:
		return "UNKNOWN"
	}
}

// ============================================================================
// Resolution Strategy
// ============================================================================

// ResolutionStrategy defines how to resolve an impasse.
type ResolutionStrategy int

const (
	// StrategyDecompose breaks the goal into subgoals
	StrategyDecompose ResolutionStrategy = iota

	// StrategyEscalate escalates to a higher-tier agent
	StrategyEscalate

	// StrategyRandom randomly selects among tied options
	StrategyRandom

	// StrategyConsensus builds consensus among conflicting agents
	StrategyConsensus

	// StrategyRetry retries the failed operation
	StrategyRetry

	// StrategyBackoff backs off and waits before retry
	StrategyBackoff

	// StrategyFallback uses a fallback agent/approach
	StrategyFallback

	// StrategyAbort abandons the goal
	StrategyAbort

	// StrategyAsk asks for external input
	StrategyAsk

	// StrategyLearn triggers learning from the impasse
	StrategyLearn
)

// String returns a human-readable strategy name.
func (s ResolutionStrategy) String() string {
	switch s {
	case StrategyDecompose:
		return "DECOMPOSE"
	case StrategyEscalate:
		return "ESCALATE"
	case StrategyRandom:
		return "RANDOM"
	case StrategyConsensus:
		return "CONSENSUS"
	case StrategyRetry:
		return "RETRY"
	case StrategyBackoff:
		return "BACKOFF"
	case StrategyFallback:
		return "FALLBACK"
	case StrategyAbort:
		return "ABORT"
	case StrategyAsk:
		return "ASK"
	case StrategyLearn:
		return "LEARN"
	default:
		return "UNKNOWN"
	}
}

// ============================================================================
// Impasse
// ============================================================================

// Impasse represents a detected impasse in the cognitive system.
type Impasse struct {
	// ID uniquely identifies this impasse
	ID string

	// Type classifies the impasse
	Type ImpasseType

	// GoalID is the goal that encountered the impasse
	GoalID string

	// Description explains the impasse
	Description string

	// Candidates are the tied agents/operators (for Tie impasses)
	Candidates []string

	// FailedAgent is the agent that failed (for Failure impasses)
	FailedAgent string

	// FailureReason explains why the agent failed
	FailureReason string

	// ConflictingResults holds conflicting outputs (for Conflict impasses)
	ConflictingResults map[string]interface{}

	// ConstraintViolated is the violated constraint (for Constraint impasses)
	ConstraintViolated string

	// DetectedAt is when the impasse was detected
	DetectedAt time.Time

	// ResolvedAt is when the impasse was resolved (nil if unresolved)
	ResolvedAt *time.Time

	// Resolution is the strategy used to resolve
	Resolution ResolutionStrategy

	// ResolutionDetails provides additional resolution context
	ResolutionDetails string

	// Severity indicates how serious the impasse is (0.0 to 1.0)
	Severity float64

	// RetryCount is how many times resolution has been attempted
	RetryCount int

	// MaxRetries is the maximum number of retries
	MaxRetries int

	// Context holds additional impasse-specific data
	Context map[string]interface{}

	// Metadata for extensions
	Metadata map[string]interface{}
}

// IsResolved returns true if the impasse has been resolved.
func (imp *Impasse) IsResolved() bool {
	return imp.ResolvedAt != nil
}

// ============================================================================
// Resolution Result
// ============================================================================

// ResolutionResult is the outcome of attempting to resolve an impasse.
type ResolutionResult struct {
	// Success indicates if the resolution succeeded
	Success bool

	// Strategy used for resolution
	Strategy ResolutionStrategy

	// SelectedCandidate is the chosen agent (for Tie resolution)
	SelectedCandidate string

	// NewGoals are subgoals created (for Decompose resolution)
	NewGoals []*Goal

	// EscalatedTo is the higher-tier agent (for Escalate resolution)
	EscalatedTo string

	// ConsensusResult is the agreed solution (for Consensus resolution)
	ConsensusResult interface{}

	// Message provides additional context
	Message string

	// LearnedPattern is a pattern extracted (for Learn resolution)
	LearnedPattern string

	// Duration is how long resolution took
	Duration time.Duration
}

// ============================================================================
// Impasse Detector
// ============================================================================

// ImpasseDetector detects and resolves impasses.
type ImpasseDetector struct {
	mu sync.RWMutex

	// impasses holds all detected impasses by ID
	impasses map[string]*Impasse

	// activeImpasses holds unresolved impasses
	activeImpasses map[string]*Impasse

	// resolvedImpasses holds resolved impasses for history
	resolvedImpasses map[string]*Impasse

	// goalStack for subgoal creation
	goalStack *GoalStack

	// config holds detector configuration
	config *ImpasseDetectorConfig

	// stats tracks detection statistics
	stats *ImpasseStats

	// strategyHandlers maps impasse types to resolution strategies
	strategyHandlers map[ImpasseType][]ResolutionStrategy

	// callbacks
	onImpasseDetected func(*Impasse)
	onImpasseResolved func(*Impasse, *ResolutionResult)

	// customResolvers allow custom resolution logic
	customResolvers map[ImpasseType]func(*Impasse) (*ResolutionResult, error)
}

// ImpasseDetectorConfig configures the detector.
type ImpasseDetectorConfig struct {
	// MaxRetries for resolution attempts
	MaxRetries int

	// BackoffBase is the base delay for exponential backoff (ms)
	BackoffBase time.Duration

	// BackoffMax is the maximum backoff delay
	BackoffMax time.Duration

	// TimeoutThreshold for detecting timeout impasses
	TimeoutThreshold time.Duration

	// NoChangeThreshold for detecting no-change impasses
	NoChangeThreshold int

	// TieSimilarityThreshold for detecting ties
	TieSimilarityThreshold float64

	// MaxActiveImpasses before triggering capacity impasse
	MaxActiveImpasses int
}

// DefaultImpasseDetectorConfig returns sensible defaults.
func DefaultImpasseDetectorConfig() *ImpasseDetectorConfig {
	return &ImpasseDetectorConfig{
		MaxRetries:             3,
		BackoffBase:            100 * time.Millisecond,
		BackoffMax:             5 * time.Second,
		TimeoutThreshold:       30 * time.Second,
		NoChangeThreshold:      5,
		TieSimilarityThreshold: 0.95,
		MaxActiveImpasses:      10,
	}
}

// ImpasseStats tracks impasse statistics.
type ImpasseStats struct {
	TotalDetected         int64
	TotalResolved         int64
	TotalFailed           int64
	ByType                map[ImpasseType]int64
	ByResolution          map[ResolutionStrategy]int64
	AverageResolutionTime time.Duration
}

// NewImpasseDetector creates a new impasse detector.
func NewImpasseDetector(config *ImpasseDetectorConfig, goalStack *GoalStack) *ImpasseDetector {
	if config == nil {
		config = DefaultImpasseDetectorConfig()
	}

	detector := &ImpasseDetector{
		impasses:         make(map[string]*Impasse),
		activeImpasses:   make(map[string]*Impasse),
		resolvedImpasses: make(map[string]*Impasse),
		goalStack:        goalStack,
		config:           config,
		stats: &ImpasseStats{
			ByType:       make(map[ImpasseType]int64),
			ByResolution: make(map[ResolutionStrategy]int64),
		},
		strategyHandlers: make(map[ImpasseType][]ResolutionStrategy),
		customResolvers:  make(map[ImpasseType]func(*Impasse) (*ResolutionResult, error)),
	}

	// Set default resolution strategies per impasse type
	detector.strategyHandlers[ImpasseTie] = []ResolutionStrategy{StrategyRandom, StrategyConsensus, StrategyEscalate}
	detector.strategyHandlers[ImpasseNoMatch] = []ResolutionStrategy{StrategyDecompose, StrategyEscalate, StrategyFallback}
	detector.strategyHandlers[ImpasseFailure] = []ResolutionStrategy{StrategyRetry, StrategyBackoff, StrategyFallback, StrategyAbort}
	detector.strategyHandlers[ImpasseConflict] = []ResolutionStrategy{StrategyConsensus, StrategyEscalate, StrategyRandom}
	detector.strategyHandlers[ImpasseCapacity] = []ResolutionStrategy{StrategyBackoff, StrategyAbort}
	detector.strategyHandlers[ImpasseNoChange] = []ResolutionStrategy{StrategyDecompose, StrategyEscalate, StrategyAbort}
	detector.strategyHandlers[ImpasseConstraint] = []ResolutionStrategy{StrategyFallback, StrategyDecompose, StrategyAbort}
	detector.strategyHandlers[ImpasseTimeout] = []ResolutionStrategy{StrategyBackoff, StrategyAbort}

	return detector
}

// ============================================================================
// Impasse Detection
// ============================================================================

// DetectTie detects when multiple candidates have equal preference.
func (d *ImpasseDetector) DetectTie(goalID string, candidates []string, scores []float64) *Impasse {
	if len(candidates) < 2 || len(candidates) != len(scores) {
		return nil
	}

	// Find max score
	maxScore := scores[0]
	for _, s := range scores[1:] {
		if s > maxScore {
			maxScore = s
		}
	}

	// Find candidates within threshold of max
	tiedCandidates := make([]string, 0)
	for i, s := range scores {
		if maxScore > 0 && s/maxScore >= d.config.TieSimilarityThreshold {
			tiedCandidates = append(tiedCandidates, candidates[i])
		}
	}

	if len(tiedCandidates) < 2 {
		return nil
	}

	return d.createImpasse(ImpasseTie, goalID, fmt.Sprintf("%d candidates tied", len(tiedCandidates)), func(imp *Impasse) {
		imp.Candidates = tiedCandidates
		imp.Severity = 0.3 // Ties are low severity
	})
}

// DetectNoMatch detects when no agent matches the goal.
func (d *ImpasseDetector) DetectNoMatch(goalID string, description string) *Impasse {
	return d.createImpasse(ImpasseNoMatch, goalID, description, func(imp *Impasse) {
		imp.Severity = 0.6
	})
}

// DetectFailure detects when an agent fails to complete a task.
func (d *ImpasseDetector) DetectFailure(goalID, agentID, reason string) *Impasse {
	return d.createImpasse(ImpasseFailure, goalID, reason, func(imp *Impasse) {
		imp.FailedAgent = agentID
		imp.FailureReason = reason
		imp.Severity = 0.7
	})
}

// DetectConflict detects when agents produce conflicting results.
func (d *ImpasseDetector) DetectConflict(goalID string, results map[string]interface{}) *Impasse {
	if len(results) < 2 {
		return nil
	}

	return d.createImpasse(ImpasseConflict, goalID, fmt.Sprintf("%d conflicting results", len(results)), func(imp *Impasse) {
		imp.ConflictingResults = results
		imp.Severity = 0.5
		imp.Candidates = make([]string, 0, len(results))
		for agent := range results {
			imp.Candidates = append(imp.Candidates, agent)
		}
	})
}

// DetectCapacity detects when system capacity is exhausted.
func (d *ImpasseDetector) DetectCapacity(goalID, resource string) *Impasse {
	return d.createImpasse(ImpasseCapacity, goalID, fmt.Sprintf("capacity exhausted: %s", resource), func(imp *Impasse) {
		imp.Severity = 0.8
		imp.Context["resource"] = resource
	})
}

// DetectNoChange detects when no progress is being made.
func (d *ImpasseDetector) DetectNoChange(goalID string, iterations int) *Impasse {
	if iterations < d.config.NoChangeThreshold {
		return nil
	}

	return d.createImpasse(ImpasseNoChange, goalID, fmt.Sprintf("no change after %d iterations", iterations), func(imp *Impasse) {
		imp.Severity = 0.6
		imp.Context["iterations"] = iterations
	})
}

// DetectConstraint detects when a constraint is violated.
func (d *ImpasseDetector) DetectConstraint(goalID, constraint string) *Impasse {
	return d.createImpasse(ImpasseConstraint, goalID, fmt.Sprintf("constraint violated: %s", constraint), func(imp *Impasse) {
		imp.ConstraintViolated = constraint
		imp.Severity = 0.7
	})
}

// DetectTimeout detects when processing exceeds time limit.
func (d *ImpasseDetector) DetectTimeout(goalID string, elapsed time.Duration) *Impasse {
	if elapsed < d.config.TimeoutThreshold {
		return nil
	}

	return d.createImpasse(ImpasseTimeout, goalID, fmt.Sprintf("timeout after %v", elapsed), func(imp *Impasse) {
		imp.Severity = 0.9
		imp.Context["elapsed"] = elapsed.String()
	})
}

// createImpasse creates and registers an impasse.
func (d *ImpasseDetector) createImpasse(impasseType ImpasseType, goalID, description string, configure func(*Impasse)) *Impasse {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Check capacity
	if len(d.activeImpasses) >= d.config.MaxActiveImpasses {
		// Create capacity impasse instead
		if impasseType != ImpasseCapacity {
			// Avoid infinite recursion
			cap := &Impasse{
				ID:          fmt.Sprintf("imp-capacity-%d", time.Now().UnixNano()),
				Type:        ImpasseCapacity,
				GoalID:      goalID,
				Description: "too many active impasses",
				DetectedAt:  time.Now(),
				Severity:    0.9,
				MaxRetries:  d.config.MaxRetries,
				Context:     make(map[string]interface{}),
				Metadata:    make(map[string]interface{}),
			}
			d.impasses[cap.ID] = cap
			d.activeImpasses[cap.ID] = cap
			d.stats.TotalDetected++
			d.stats.ByType[ImpasseCapacity]++
			return cap
		}
	}

	imp := &Impasse{
		ID:          fmt.Sprintf("imp-%s-%d", impasseType.String(), time.Now().UnixNano()),
		Type:        impasseType,
		GoalID:      goalID,
		Description: description,
		DetectedAt:  time.Now(),
		MaxRetries:  d.config.MaxRetries,
		Context:     make(map[string]interface{}),
		Metadata:    make(map[string]interface{}),
	}

	if configure != nil {
		configure(imp)
	}

	d.impasses[imp.ID] = imp
	d.activeImpasses[imp.ID] = imp
	d.stats.TotalDetected++
	d.stats.ByType[impasseType]++

	if d.onImpasseDetected != nil {
		d.onImpasseDetected(imp)
	}

	return imp
}

// ============================================================================
// Impasse Resolution
// ============================================================================

// Resolve attempts to resolve an impasse.
func (d *ImpasseDetector) Resolve(impasseID string) (*ResolutionResult, error) {
	d.mu.Lock()
	imp, ok := d.activeImpasses[impasseID]
	if !ok {
		d.mu.Unlock()
		return nil, errors.New("impasse not found or already resolved")
	}
	d.mu.Unlock()

	// Check for custom resolver
	if resolver, ok := d.customResolvers[imp.Type]; ok {
		result, err := resolver(imp)
		if err == nil && result.Success {
			d.markResolved(imp, result)
			return result, nil
		}
		// Fall through to default strategies
	}

	// Try strategies in order
	strategies := d.strategyHandlers[imp.Type]
	for _, strategy := range strategies {
		result, err := d.applyStrategy(imp, strategy)
		if err != nil {
			continue
		}
		if result.Success {
			d.markResolved(imp, result)
			return result, nil
		}
	}

	// All strategies failed
	imp.RetryCount++
	if imp.RetryCount >= imp.MaxRetries {
		result := &ResolutionResult{
			Success:  false,
			Strategy: StrategyAbort,
			Message:  "all resolution strategies exhausted",
		}
		d.markResolved(imp, result)
		d.stats.TotalFailed++
		return result, nil
	}

	return nil, errors.New("resolution strategies failed, will retry")
}

// applyStrategy applies a specific resolution strategy.
func (d *ImpasseDetector) applyStrategy(imp *Impasse, strategy ResolutionStrategy) (*ResolutionResult, error) {
	start := time.Now()

	var result *ResolutionResult

	switch strategy {
	case StrategyDecompose:
		result = d.resolveDecompose(imp)
	case StrategyEscalate:
		result = d.resolveEscalate(imp)
	case StrategyRandom:
		result = d.resolveRandom(imp)
	case StrategyConsensus:
		result = d.resolveConsensus(imp)
	case StrategyRetry:
		result = d.resolveRetry(imp)
	case StrategyBackoff:
		result = d.resolveBackoff(imp)
	case StrategyFallback:
		result = d.resolveFallback(imp)
	case StrategyAbort:
		result = d.resolveAbort(imp)
	case StrategyAsk:
		result = d.resolveAsk(imp)
	case StrategyLearn:
		result = d.resolveLearn(imp)
	default:
		return nil, fmt.Errorf("unknown strategy: %v", strategy)
	}

	result.Strategy = strategy
	result.Duration = time.Since(start)

	return result, nil
}

// resolveDecompose creates subgoals to address the impasse.
func (d *ImpasseDetector) resolveDecompose(imp *Impasse) *ResolutionResult {
	if d.goalStack == nil {
		return &ResolutionResult{Success: false, Message: "no goal stack available"}
	}

	// Create diagnostic subgoal
	subgoal := &Goal{
		ID:          fmt.Sprintf("resolve-%s", imp.ID),
		Name:        fmt.Sprintf("Resolve %s impasse", imp.Type),
		Description: imp.Description,
		Priority:    PriorityHigh,
	}

	// Would normally add subgoal to stack, but for now just signal success
	return &ResolutionResult{
		Success:  true,
		NewGoals: []*Goal{subgoal},
		Message:  "created diagnostic subgoal",
	}
}

// resolveEscalate escalates to a higher-tier agent.
func (d *ImpasseDetector) resolveEscalate(imp *Impasse) *ResolutionResult {
	// Map impasse type to appropriate higher-tier agent
	escalationMap := map[ImpasseType]string{
		ImpasseTie:        "OMNISCIENT-20", // Meta agent for tie-breaking
		ImpasseNoMatch:    "NEXUS-18",      // Cross-domain for novel problems
		ImpasseFailure:    "GENESIS-19",    // Innovation for novel approaches
		ImpasseConflict:   "ARBITER-39",    // Conflict resolution specialist
		ImpasseCapacity:   "FLUX-11",       // Infrastructure for scaling
		ImpasseNoChange:   "GENESIS-19",    // Innovation for breakthroughs
		ImpasseConstraint: "AXIOM-04",      // Formal analysis
		ImpasseTimeout:    "VELOCITY-05",   // Performance optimization
	}

	agent, ok := escalationMap[imp.Type]
	if !ok {
		agent = "OMNISCIENT-20" // Default to meta agent
	}

	return &ResolutionResult{
		Success:     true,
		EscalatedTo: agent,
		Message:     fmt.Sprintf("escalated to %s", agent),
	}
}

// resolveRandom randomly selects among tied candidates.
func (d *ImpasseDetector) resolveRandom(imp *Impasse) *ResolutionResult {
	if len(imp.Candidates) == 0 {
		return &ResolutionResult{Success: false, Message: "no candidates to choose from"}
	}

	// Use time-based "random" selection
	selected := imp.Candidates[time.Now().UnixNano()%int64(len(imp.Candidates))]

	return &ResolutionResult{
		Success:           true,
		SelectedCandidate: selected,
		Message:           fmt.Sprintf("randomly selected %s", selected),
	}
}

// resolveConsensus builds consensus among conflicting agents.
func (d *ImpasseDetector) resolveConsensus(imp *Impasse) *ResolutionResult {
	if len(imp.ConflictingResults) < 2 && len(imp.Candidates) < 2 {
		return &ResolutionResult{Success: false, Message: "not enough participants for consensus"}
	}

	// Simulate consensus by taking majority or first result
	// In a real system, this would involve more sophisticated voting
	var consensusResult interface{}
	if len(imp.ConflictingResults) > 0 {
		for _, result := range imp.ConflictingResults {
			consensusResult = result
			break
		}
	}

	return &ResolutionResult{
		Success:         true,
		ConsensusResult: consensusResult,
		Message:         "consensus reached",
	}
}

// resolveRetry simply indicates that retry is appropriate.
func (d *ImpasseDetector) resolveRetry(imp *Impasse) *ResolutionResult {
	if imp.RetryCount >= imp.MaxRetries {
		return &ResolutionResult{Success: false, Message: "max retries exceeded"}
	}

	return &ResolutionResult{
		Success: true,
		Message: fmt.Sprintf("retry attempt %d/%d", imp.RetryCount+1, imp.MaxRetries),
	}
}

// resolveBackoff waits with exponential backoff.
func (d *ImpasseDetector) resolveBackoff(imp *Impasse) *ResolutionResult {
	backoff := d.config.BackoffBase * time.Duration(math.Pow(2, float64(imp.RetryCount)))
	if backoff > d.config.BackoffMax {
		backoff = d.config.BackoffMax
	}

	// In a real system, we'd schedule a delayed retry
	// For now, just indicate the backoff duration
	return &ResolutionResult{
		Success: true,
		Message: fmt.Sprintf("backing off for %v", backoff),
	}
}

// resolveFallback uses a fallback approach.
func (d *ImpasseDetector) resolveFallback(imp *Impasse) *ResolutionResult {
	// Map impasse types to fallback agents
	fallbacks := map[ImpasseType]string{
		ImpasseNoMatch:    "APEX-01",    // General-purpose
		ImpasseFailure:    "APEX-01",    // Try general approach
		ImpasseConstraint: "GENESIS-19", // Novel approach needed
	}

	agent, ok := fallbacks[imp.Type]
	if !ok {
		return &ResolutionResult{Success: false, Message: "no fallback available"}
	}

	return &ResolutionResult{
		Success:           true,
		SelectedCandidate: agent,
		Message:           fmt.Sprintf("falling back to %s", agent),
	}
}

// resolveAbort abandons the goal.
func (d *ImpasseDetector) resolveAbort(imp *Impasse) *ResolutionResult {
	if d.goalStack != nil && imp.GoalID != "" {
		_ = d.goalStack.Fail(imp.GoalID, fmt.Sprintf("aborted due to %s impasse: %s", imp.Type, imp.Description))
	}

	return &ResolutionResult{
		Success: true,
		Message: "goal aborted",
	}
}

// resolveAsk indicates external input is needed.
func (d *ImpasseDetector) resolveAsk(imp *Impasse) *ResolutionResult {
	return &ResolutionResult{
		Success: true,
		Message: "awaiting external input",
	}
}

// resolveLearn extracts a pattern from the impasse for learning.
func (d *ImpasseDetector) resolveLearn(imp *Impasse) *ResolutionResult {
	// Generate a pattern description
	pattern := fmt.Sprintf("%s on goal=%s: %s", imp.Type, imp.GoalID, imp.Description)

	return &ResolutionResult{
		Success:        true,
		LearnedPattern: pattern,
		Message:        "pattern extracted for learning",
	}
}

// markResolved marks an impasse as resolved.
func (d *ImpasseDetector) markResolved(imp *Impasse, result *ResolutionResult) {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now()
	imp.ResolvedAt = &now
	imp.Resolution = result.Strategy
	imp.ResolutionDetails = result.Message

	delete(d.activeImpasses, imp.ID)
	d.resolvedImpasses[imp.ID] = imp

	d.stats.TotalResolved++
	d.stats.ByResolution[result.Strategy]++

	if d.onImpasseResolved != nil {
		d.onImpasseResolved(imp, result)
	}
}

// ============================================================================
// Query Operations
// ============================================================================

// Get retrieves an impasse by ID.
func (d *ImpasseDetector) Get(id string) (*Impasse, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if imp, ok := d.impasses[id]; ok {
		return imp, true
	}
	if imp, ok := d.resolvedImpasses[id]; ok {
		return imp, true
	}
	return nil, false
}

// GetActive returns all active (unresolved) impasses.
func (d *ImpasseDetector) GetActive() []*Impasse {
	d.mu.RLock()
	defer d.mu.RUnlock()

	result := make([]*Impasse, 0, len(d.activeImpasses))
	for _, imp := range d.activeImpasses {
		result = append(result, imp)
	}
	return result
}

// GetByType returns impasses of a specific type.
func (d *ImpasseDetector) GetByType(impasseType ImpasseType) []*Impasse {
	d.mu.RLock()
	defer d.mu.RUnlock()

	result := make([]*Impasse, 0)
	for _, imp := range d.impasses {
		if imp.Type == impasseType {
			result = append(result, imp)
		}
	}
	for _, imp := range d.resolvedImpasses {
		if imp.Type == impasseType {
			result = append(result, imp)
		}
	}
	return result
}

// GetByGoal returns impasses for a specific goal.
func (d *ImpasseDetector) GetByGoal(goalID string) []*Impasse {
	d.mu.RLock()
	defer d.mu.RUnlock()

	result := make([]*Impasse, 0)
	for _, imp := range d.impasses {
		if imp.GoalID == goalID {
			result = append(result, imp)
		}
	}
	for _, imp := range d.resolvedImpasses {
		if imp.GoalID == goalID {
			result = append(result, imp)
		}
	}
	return result
}

// ActiveCount returns the number of active impasses.
func (d *ImpasseDetector) ActiveCount() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.activeImpasses)
}

// ============================================================================
// Custom Resolvers
// ============================================================================

// RegisterResolver registers a custom resolver for an impasse type.
func (d *ImpasseDetector) RegisterResolver(impasseType ImpasseType, resolver func(*Impasse) (*ResolutionResult, error)) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.customResolvers[impasseType] = resolver
}

// ============================================================================
// Callbacks
// ============================================================================

// OnImpasseDetected sets callback for impasse detection.
func (d *ImpasseDetector) OnImpasseDetected(fn func(*Impasse)) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.onImpasseDetected = fn
}

// OnImpasseResolved sets callback for impasse resolution.
func (d *ImpasseDetector) OnImpasseResolved(fn func(*Impasse, *ResolutionResult)) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.onImpasseResolved = fn
}

// ============================================================================
// Statistics
// ============================================================================

// GetStats returns impasse detection statistics.
func (d *ImpasseDetector) GetStats() *ImpasseStats {
	d.mu.RLock()
	defer d.mu.RUnlock()

	stats := &ImpasseStats{
		TotalDetected: d.stats.TotalDetected,
		TotalResolved: d.stats.TotalResolved,
		TotalFailed:   d.stats.TotalFailed,
		ByType:        make(map[ImpasseType]int64),
		ByResolution:  make(map[ResolutionStrategy]int64),
	}

	for k, v := range d.stats.ByType {
		stats.ByType[k] = v
	}
	for k, v := range d.stats.ByResolution {
		stats.ByResolution[k] = v
	}

	return stats
}

// ============================================================================
// Snapshot
// ============================================================================

// ImpasseSnapshot represents the current detector state.
type ImpasseSnapshot struct {
	Timestamp      time.Time
	ActiveCount    int
	ResolvedCount  int
	ActiveImpasses []*Impasse
}

// Snapshot returns current state for debugging/monitoring.
func (d *ImpasseDetector) Snapshot() *ImpasseSnapshot {
	d.mu.RLock()
	defer d.mu.RUnlock()

	snapshot := &ImpasseSnapshot{
		Timestamp:      time.Now(),
		ActiveCount:    len(d.activeImpasses),
		ResolvedCount:  len(d.resolvedImpasses),
		ActiveImpasses: make([]*Impasse, 0, len(d.activeImpasses)),
	}

	for _, imp := range d.activeImpasses {
		snapshot.ActiveImpasses = append(snapshot.ActiveImpasses, imp)
	}

	return snapshot
}

// ============================================================================
// Clear
// ============================================================================

// Clear removes all impasses.
func (d *ImpasseDetector) Clear() {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.impasses = make(map[string]*Impasse)
	d.activeImpasses = make(map[string]*Impasse)
	d.resolvedImpasses = make(map[string]*Impasse)
}
