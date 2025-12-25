// Package memory provides the MNEMONIC system for the Elite Agent Collective.
// This file implements Phase Transition Engineering (PTE) - Innovation #5 from @GENESIS
//
// The Phase Transition Controller maintains the system at the "edge of chaos" where:
// - Order and chaos are balanced
// - Emergent behaviors are most likely
// - Adaptability is highest
// - Novel solutions are discovered

package memory

import (
	"context"
	"math"
	"sync"
	"time"
)

// ============================================================================
// System Phase Constants
// ============================================================================

// SystemPhase represents the current phase of the multi-agent system.
type SystemPhase int

const (
	// PhaseFrozen indicates excessive order - agents route deterministically
	// Characteristics: Low entropy, predictable, no innovation
	PhaseFrozen SystemPhase = iota

	// PhaseCritical is the target "edge of chaos" state
	// Characteristics: Balanced entropy, maximum emergence, optimal innovation
	PhaseCritical

	// PhaseChaotic indicates excessive randomness - agents route randomly
	// Characteristics: High entropy, unpredictable, unstable
	PhaseChaotic
)

// String returns a human-readable phase name.
func (p SystemPhase) String() string {
	switch p {
	case PhaseFrozen:
		return "FROZEN"
	case PhaseCritical:
		return "CRITICAL"
	case PhaseChaotic:
		return "CHAOTIC"
	default:
		return "UNKNOWN"
	}
}

// ============================================================================
// Criticality Metrics - Measure system state
// ============================================================================

// CriticalityMetrics captures the current state of the multi-agent system.
type CriticalityMetrics struct {
	// RoutingEntropy measures the distribution of agent selection
	// Low = same agents always chosen (frozen), High = random selection (chaotic)
	// Range: 0.0 to 1.0 (normalized by max possible entropy)
	RoutingEntropy float64

	// AgentDiversity measures the spread of capabilities across agents
	// Low = high overlap (redundancy), High = distinct specializations
	// Range: 0.0 to 1.0 (Gini coefficient inverted)
	AgentDiversity float64

	// InnovationRate tracks novel solutions per time window
	// Measures unique patterns discovered / total tasks
	InnovationRate float64

	// AdaptationSpeed measures how quickly fitness improves
	// Calculated as rolling fitness delta over generations
	AdaptationSpeed float64

	// StabilityMetric measures consistency in routing over time
	// Low = chaotic changes, High = stable patterns
	// Range: 0.0 to 1.0
	StabilityMetric float64

	// Phase is the detected system phase based on metrics
	Phase SystemPhase

	// Timestamp when metrics were computed
	Timestamp time.Time

	// WindowSize is the number of tasks in the analysis window
	WindowSize int

	// HealthScore is an aggregate health indicator (0.0 to 1.0)
	HealthScore float64
}

// IsCritical returns true if the system is at the edge of chaos.
func (m *CriticalityMetrics) IsCritical() bool {
	return m.Phase == PhaseCritical
}

// ============================================================================
// Phase Snapshot - Historical record
// ============================================================================

// PhaseSnapshot records system state at a point in time.
type PhaseSnapshot struct {
	Timestamp  time.Time
	Metrics    CriticalityMetrics
	Parameters ControlParameters
}

// ControlParameters holds the tunable system parameters.
type ControlParameters struct {
	// Temperature controls exploration vs exploitation
	// High T = more random agent selection
	// Low T = more deterministic routing
	Temperature float64

	// MutationRate controls prompt evolution speed (for Prompt Genetics)
	// Too high = chaos, Too low = stagnation
	MutationRate float64

	// ConsolidationRate controls memory forgetting speed (for NMC)
	// Too fast = lose knowledge, Too slow = overwhelmed
	ConsolidationRate float64

	// MarketLiquidity controls competition intensity (for EPM)
	// Too high = winner-take-all, Too low = no pressure
	MarketLiquidity float64

	// ExplorationBonus added to less-used agents
	ExplorationBonus float64
}

// ============================================================================
// Task Record - For entropy computation
// ============================================================================

// TaskRecord tracks agent selection for a task.
type TaskRecord struct {
	TaskID       string
	AgentID      string
	Success      bool
	IsNovel      bool
	FitnessScore float64
	Timestamp    time.Time
	ResponseTime time.Duration
}

// ============================================================================
// Phase Transition Controller
// ============================================================================

// PhaseTransitionController maintains the system at the edge of chaos.
// It continuously monitors system metrics and adjusts parameters to maintain
// the optimal "critical" regime where innovation is maximized.
type PhaseTransitionController struct {
	mu sync.RWMutex

	// Current control parameters
	params ControlParameters

	// Target entropy for critical regime (typically 0.5-0.7)
	targetEntropy float64

	// Tolerance for entropy deviation before adjustment
	entropyTolerance float64

	// Rate at which parameters are adjusted
	adaptationRate float64

	// Current metrics
	currentMetrics *CriticalityMetrics

	// Historical snapshots for trend analysis
	history []PhaseSnapshot

	// Maximum history size
	maxHistorySize int

	// Task records for entropy computation
	taskRecords []TaskRecord

	// Maximum task record window
	maxTaskWindow int

	// Agent count (for entropy normalization)
	agentCount int

	// Thresholds for phase detection
	frozenThreshold  float64 // Below this = frozen
	chaoticThreshold float64 // Above this = chaotic

	// Minimum innovation rate before intervention
	minInnovationRate float64

	// Minimum stability before dampening
	minStability float64

	// Callbacks for parameter changes
	onParameterChange func(old, new ControlParameters)

	// Whether the controller is running
	running bool

	// Shutdown channel
	shutdownChan chan struct{}
}

// PhaseTransitionConfig configures the controller.
type PhaseTransitionConfig struct {
	// InitialTemperature for exploration/exploitation
	InitialTemperature float64

	// InitialMutationRate for prompt evolution
	InitialMutationRate float64

	// InitialConsolidationRate for memory forgetting
	InitialConsolidationRate float64

	// InitialMarketLiquidity for competition
	InitialMarketLiquidity float64

	// TargetEntropy for critical regime
	TargetEntropy float64

	// EntropyTolerance before adjustment
	EntropyTolerance float64

	// AdaptationRate for parameter changes
	AdaptationRate float64

	// AgentCount in the system
	AgentCount int

	// TaskWindowSize for metric computation
	TaskWindowSize int

	// HistorySize for trend analysis
	HistorySize int
}

// DefaultPhaseTransitionConfig returns sensible defaults.
func DefaultPhaseTransitionConfig() PhaseTransitionConfig {
	return PhaseTransitionConfig{
		InitialTemperature:       1.0,
		InitialMutationRate:      0.1,
		InitialConsolidationRate: 0.5,
		InitialMarketLiquidity:   0.5,
		TargetEntropy:            0.6, // Edge of chaos
		EntropyTolerance:         0.1,
		AdaptationRate:           0.05,
		AgentCount:               40,
		TaskWindowSize:           1000,
		HistorySize:              100,
	}
}

// NewPhaseTransitionController creates a new controller.
func NewPhaseTransitionController(config PhaseTransitionConfig) *PhaseTransitionController {
	return &PhaseTransitionController{
		params: ControlParameters{
			Temperature:       config.InitialTemperature,
			MutationRate:      config.InitialMutationRate,
			ConsolidationRate: config.InitialConsolidationRate,
			MarketLiquidity:   config.InitialMarketLiquidity,
			ExplorationBonus:  0.1,
		},
		targetEntropy:     config.TargetEntropy,
		entropyTolerance:  config.EntropyTolerance,
		adaptationRate:    config.AdaptationRate,
		agentCount:        config.AgentCount,
		maxHistorySize:    config.HistorySize,
		maxTaskWindow:     config.TaskWindowSize,
		history:           make([]PhaseSnapshot, 0, config.HistorySize),
		taskRecords:       make([]TaskRecord, 0, config.TaskWindowSize),
		frozenThreshold:   0.3,
		chaoticThreshold:  0.8,
		minInnovationRate: 0.01,
		minStability:      0.5,
		shutdownChan:      make(chan struct{}),
	}
}

// ============================================================================
// Task Recording
// ============================================================================

// RecordTask records a task completion for metric computation.
func (c *PhaseTransitionController) RecordTask(record TaskRecord) {
	c.mu.Lock()
	defer c.mu.Unlock()

	record.Timestamp = time.Now()
	c.taskRecords = append(c.taskRecords, record)

	// Trim to window size
	if len(c.taskRecords) > c.maxTaskWindow {
		c.taskRecords = c.taskRecords[len(c.taskRecords)-c.maxTaskWindow:]
	}
}

// ============================================================================
// Metric Computation
// ============================================================================

// ComputeMetrics calculates current criticality metrics.
func (c *PhaseTransitionController) ComputeMetrics() *CriticalityMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.taskRecords) == 0 {
		return &CriticalityMetrics{
			Phase:     PhaseCritical, // Assume critical at start
			Timestamp: time.Now(),
		}
	}

	metrics := &CriticalityMetrics{
		Timestamp:  time.Now(),
		WindowSize: len(c.taskRecords),
	}

	// Compute routing entropy
	metrics.RoutingEntropy = c.computeRoutingEntropy()

	// Compute agent diversity
	metrics.AgentDiversity = c.computeAgentDiversity()

	// Compute innovation rate
	metrics.InnovationRate = c.computeInnovationRate()

	// Compute adaptation speed
	metrics.AdaptationSpeed = c.computeAdaptationSpeed()

	// Compute stability
	metrics.StabilityMetric = c.computeStability()

	// Determine phase
	metrics.Phase = c.detectPhase(metrics)

	// Compute health score
	metrics.HealthScore = c.computeHealthScore(metrics)

	return metrics
}

// computeRoutingEntropy calculates Shannon entropy of agent selection.
func (c *PhaseTransitionController) computeRoutingEntropy() float64 {
	if len(c.taskRecords) == 0 {
		return 0.5 // Neutral
	}

	// Count agent selections
	agentCounts := make(map[string]int)
	total := 0

	for _, record := range c.taskRecords {
		agentCounts[record.AgentID]++
		total++
	}

	if total == 0 {
		return 0.5
	}

	// Calculate Shannon entropy
	entropy := 0.0
	for _, count := range agentCounts {
		p := float64(count) / float64(total)
		if p > 0 {
			entropy -= p * math.Log2(p)
		}
	}

	// Normalize by maximum possible entropy (log2 of agent count)
	maxEntropy := math.Log2(float64(c.agentCount))
	if maxEntropy > 0 {
		return entropy / maxEntropy
	}

	return 0.5
}

// computeAgentDiversity calculates diversity using inverse Gini coefficient.
func (c *PhaseTransitionController) computeAgentDiversity() float64 {
	if len(c.taskRecords) == 0 {
		return 0.5
	}

	// Count agent selections
	agentCounts := make(map[string]int)
	for _, record := range c.taskRecords {
		agentCounts[record.AgentID]++
	}

	if len(agentCounts) == 0 {
		return 0
	}

	// Convert to slice and sort
	counts := make([]float64, 0, len(agentCounts))
	for _, count := range agentCounts {
		counts = append(counts, float64(count))
	}

	// Calculate Gini coefficient
	n := len(counts)
	if n <= 1 {
		return 1.0 // Perfect equality with one agent
	}

	// Sort counts
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			if counts[i] > counts[j] {
				counts[i], counts[j] = counts[j], counts[i]
			}
		}
	}

	// Gini formula
	var sum float64
	var cumulativeSum float64
	for i, count := range counts {
		cumulativeSum += count
		sum += float64(i+1) * count
	}

	if cumulativeSum == 0 {
		return 0.5
	}

	gini := (2*sum)/(float64(n)*cumulativeSum) - float64(n+1)/float64(n)

	// Return inverse (higher = more diverse)
	return 1.0 - gini
}

// computeInnovationRate calculates novel solutions per task.
func (c *PhaseTransitionController) computeInnovationRate() float64 {
	if len(c.taskRecords) == 0 {
		return 0
	}

	novelCount := 0
	for _, record := range c.taskRecords {
		if record.IsNovel {
			novelCount++
		}
	}

	return float64(novelCount) / float64(len(c.taskRecords))
}

// computeAdaptationSpeed calculates fitness improvement rate.
func (c *PhaseTransitionController) computeAdaptationSpeed() float64 {
	if len(c.taskRecords) < 10 {
		return 0.5 // Not enough data
	}

	// Compare first half to second half fitness
	mid := len(c.taskRecords) / 2

	var firstHalfSum, secondHalfSum float64
	for i, record := range c.taskRecords {
		if i < mid {
			firstHalfSum += record.FitnessScore
		} else {
			secondHalfSum += record.FitnessScore
		}
	}

	firstHalfAvg := firstHalfSum / float64(mid)
	secondHalfAvg := secondHalfSum / float64(len(c.taskRecords)-mid)

	// Improvement rate (normalized)
	if firstHalfAvg == 0 {
		return 0.5
	}

	improvement := (secondHalfAvg - firstHalfAvg) / firstHalfAvg
	// Clamp to [-1, 1] then shift to [0, 1]
	return clamp((improvement+1)/2, 0, 1)
}

// computeStability calculates routing consistency over time.
func (c *PhaseTransitionController) computeStability() float64 {
	if len(c.taskRecords) < 20 {
		return 0.5 // Not enough data
	}

	// Split into time windows and compare agent distributions
	windowSize := len(c.taskRecords) / 4
	if windowSize < 5 {
		return 0.5
	}

	// Get distributions for each window
	distributions := make([]map[string]float64, 4)
	for i := 0; i < 4; i++ {
		start := i * windowSize
		end := start + windowSize
		if i == 3 {
			end = len(c.taskRecords)
		}

		dist := make(map[string]float64)
		for j := start; j < end; j++ {
			dist[c.taskRecords[j].AgentID]++
		}
		// Normalize
		total := float64(end - start)
		for k := range dist {
			dist[k] /= total
		}
		distributions[i] = dist
	}

	// Calculate Jensen-Shannon divergence between consecutive windows
	var totalDivergence float64
	for i := 0; i < 3; i++ {
		div := c.jensenShannonDivergence(distributions[i], distributions[i+1])
		totalDivergence += div
	}

	avgDivergence := totalDivergence / 3.0

	// Convert divergence to stability (lower divergence = higher stability)
	// JS divergence is bounded [0, 1], so stability = 1 - divergence
	return 1.0 - avgDivergence
}

// jensenShannonDivergence computes JS divergence between two distributions.
func (c *PhaseTransitionController) jensenShannonDivergence(p, q map[string]float64) float64 {
	// Get all keys
	keys := make(map[string]bool)
	for k := range p {
		keys[k] = true
	}
	for k := range q {
		keys[k] = true
	}

	// Compute mixture distribution
	m := make(map[string]float64)
	for k := range keys {
		m[k] = (p[k] + q[k]) / 2
	}

	// KL divergences
	klPM := c.klDivergence(p, m)
	klQM := c.klDivergence(q, m)

	return (klPM + klQM) / 2
}

// klDivergence computes KL divergence D(p||q).
func (c *PhaseTransitionController) klDivergence(p, q map[string]float64) float64 {
	var kl float64
	for k, pVal := range p {
		if pVal > 0 {
			qVal := q[k]
			if qVal > 0 {
				kl += pVal * math.Log2(pVal/qVal)
			}
		}
	}
	return kl
}

// detectPhase determines system phase from metrics.
func (c *PhaseTransitionController) detectPhase(m *CriticalityMetrics) SystemPhase {
	if m.RoutingEntropy < c.frozenThreshold {
		return PhaseFrozen
	}
	if m.RoutingEntropy > c.chaoticThreshold {
		return PhaseChaotic
	}
	return PhaseCritical
}

// computeHealthScore calculates aggregate health.
func (c *PhaseTransitionController) computeHealthScore(m *CriticalityMetrics) float64 {
	// Weighted combination of metrics
	weights := map[string]float64{
		"entropy":    0.3,
		"diversity":  0.2,
		"innovation": 0.2,
		"adaptation": 0.15,
		"stability":  0.15,
	}

	// Score entropy based on distance from target
	entropyScore := 1.0 - math.Abs(m.RoutingEntropy-c.targetEntropy)/0.5
	if entropyScore < 0 {
		entropyScore = 0
	}

	score := weights["entropy"]*entropyScore +
		weights["diversity"]*m.AgentDiversity +
		weights["innovation"]*math.Min(m.InnovationRate*10, 1.0) + // Scale innovation
		weights["adaptation"]*m.AdaptationSpeed +
		weights["stability"]*m.StabilityMetric

	return clamp(score, 0, 1)
}

// ============================================================================
// Parameter Adjustment - Self-Organized Criticality
// ============================================================================

// Update computes metrics and adjusts parameters to maintain criticality.
func (c *PhaseTransitionController) Update() *CriticalityMetrics {
	metrics := c.ComputeMetrics()

	c.mu.Lock()
	defer c.mu.Unlock()

	c.currentMetrics = metrics
	oldParams := c.params

	// Adjust parameters based on current phase
	c.adjustParameters(metrics)

	// Record snapshot
	snapshot := PhaseSnapshot{
		Timestamp:  time.Now(),
		Metrics:    *metrics,
		Parameters: c.params,
	}
	c.history = append(c.history, snapshot)
	if len(c.history) > c.maxHistorySize {
		c.history = c.history[1:]
	}

	// Trigger callback if parameters changed
	if c.onParameterChange != nil && !c.paramsEqual(oldParams, c.params) {
		c.onParameterChange(oldParams, c.params)
	}

	return metrics
}

// adjustParameters tunes system parameters to maintain edge of chaos.
func (c *PhaseTransitionController) adjustParameters(metrics *CriticalityMetrics) {
	// === ENTROPY CONTROL (Primary) ===
	entropyDelta := metrics.RoutingEntropy - c.targetEntropy

	if math.Abs(entropyDelta) > c.entropyTolerance {
		// Too ordered (low entropy) → increase temperature
		// Too chaotic (high entropy) → decrease temperature
		c.params.Temperature -= entropyDelta * c.adaptationRate
		c.params.Temperature = clamp(c.params.Temperature, 0.1, 3.0)

		// Also adjust exploration bonus
		if entropyDelta < 0 {
			// Frozen: increase exploration
			c.params.ExplorationBonus += c.adaptationRate * 0.5
		} else {
			// Chaotic: decrease exploration
			c.params.ExplorationBonus -= c.adaptationRate * 0.5
		}
		c.params.ExplorationBonus = clamp(c.params.ExplorationBonus, 0, 0.5)
	}

	// === INNOVATION CONTROL ===
	if metrics.InnovationRate < c.minInnovationRate {
		// Stagnation detected → increase mutation rate
		c.params.MutationRate *= 1.0 + c.adaptationRate
		c.params.MutationRate = clamp(c.params.MutationRate, 0.01, 0.3)
	}

	// === STABILITY CONTROL ===
	if metrics.StabilityMetric < c.minStability {
		// Chaos detected → increase consolidation, reduce mutation
		c.params.ConsolidationRate *= 1.0 + c.adaptationRate
		c.params.MutationRate *= 1.0 - c.adaptationRate*0.5
		c.params.ConsolidationRate = clamp(c.params.ConsolidationRate, 0.1, 1.0)
		c.params.MutationRate = clamp(c.params.MutationRate, 0.01, 0.3)
	}

	// === MARKET LIQUIDITY ===
	// Adjust based on agent diversity
	if metrics.AgentDiversity < 0.5 {
		// Low diversity: reduce liquidity to allow more agents to win
		c.params.MarketLiquidity *= 1.0 - c.adaptationRate*0.5
	} else if metrics.AgentDiversity > 0.8 && metrics.InnovationRate < c.minInnovationRate {
		// High diversity but low innovation: increase competition
		c.params.MarketLiquidity *= 1.0 + c.adaptationRate*0.5
	}
	c.params.MarketLiquidity = clamp(c.params.MarketLiquidity, 0.1, 1.0)
}

// paramsEqual checks if two parameter sets are equal.
func (c *PhaseTransitionController) paramsEqual(a, b ControlParameters) bool {
	const epsilon = 0.001
	return math.Abs(a.Temperature-b.Temperature) < epsilon &&
		math.Abs(a.MutationRate-b.MutationRate) < epsilon &&
		math.Abs(a.ConsolidationRate-b.ConsolidationRate) < epsilon &&
		math.Abs(a.MarketLiquidity-b.MarketLiquidity) < epsilon
}

// ============================================================================
// Getters
// ============================================================================

// GetCurrentMetrics returns the most recent metrics.
func (c *PhaseTransitionController) GetCurrentMetrics() *CriticalityMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.currentMetrics
}

// GetParameters returns current control parameters.
func (c *PhaseTransitionController) GetParameters() ControlParameters {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.params
}

// GetHistory returns historical snapshots.
func (c *PhaseTransitionController) GetHistory() []PhaseSnapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]PhaseSnapshot, len(c.history))
	copy(result, c.history)
	return result
}

// IsAtCriticality returns true if system is at edge of chaos.
func (c *PhaseTransitionController) IsAtCriticality() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.currentMetrics == nil {
		return false
	}
	return c.currentMetrics.Phase == PhaseCritical
}

// GetPhase returns the current system phase.
func (c *PhaseTransitionController) GetPhase() SystemPhase {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.currentMetrics == nil {
		return PhaseCritical
	}
	return c.currentMetrics.Phase
}

// ============================================================================
// Continuous Monitoring
// ============================================================================

// Start begins continuous monitoring and adjustment.
func (c *PhaseTransitionController) Start(ctx context.Context, interval time.Duration) {
	c.mu.Lock()
	if c.running {
		c.mu.Unlock()
		return
	}
	c.running = true
	c.mu.Unlock()

	go c.monitorLoop(ctx, interval)
}

// Stop halts continuous monitoring.
func (c *PhaseTransitionController) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.running {
		close(c.shutdownChan)
		c.running = false
	}
}

// monitorLoop runs the continuous monitoring.
func (c *PhaseTransitionController) monitorLoop(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.shutdownChan:
			return
		case <-ticker.C:
			c.Update()
		}
	}
}

// ============================================================================
// Callbacks
// ============================================================================

// OnParameterChange sets a callback for parameter changes.
func (c *PhaseTransitionController) OnParameterChange(fn func(old, new ControlParameters)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onParameterChange = fn
}

// ============================================================================
// Temperature-Based Agent Selection
// ============================================================================

// SelectAgentWithTemperature applies temperature-based softmax selection.
// This is used by the routing system to add controlled randomness.
func (c *PhaseTransitionController) SelectAgentWithTemperature(scores map[string]float64) string {
	c.mu.RLock()
	temperature := c.params.Temperature
	explorationBonus := c.params.ExplorationBonus
	c.mu.RUnlock()

	if len(scores) == 0 {
		return ""
	}

	// Apply exploration bonus to less-used agents
	adjustedScores := make(map[string]float64)
	for agent, score := range scores {
		adjustedScores[agent] = score + explorationBonus*c.getUnderutilizationFactor(agent)
	}

	// Apply temperature-scaled softmax
	expScores := make(map[string]float64)
	var sumExp float64

	for agent, score := range adjustedScores {
		exp := math.Exp(score / temperature)
		expScores[agent] = exp
		sumExp += exp
	}

	if sumExp == 0 {
		// Return highest score agent if softmax fails
		var best string
		var bestScore float64
		for agent, score := range scores {
			if score > bestScore {
				best = agent
				bestScore = score
			}
		}
		return best
	}

	// Sample from distribution
	r := randomFloat64() // Would use crypto/rand in production
	cumulative := 0.0

	for agent, exp := range expScores {
		prob := exp / sumExp
		cumulative += prob
		if r <= cumulative {
			return agent
		}
	}

	// Fallback to highest score
	var best string
	var bestScore float64
	for agent, score := range scores {
		if score > bestScore {
			best = agent
			bestScore = score
		}
	}
	return best
}

// getUnderutilizationFactor returns how underutilized an agent is.
func (c *PhaseTransitionController) getUnderutilizationFactor(agentID string) float64 {
	// Count recent uses
	count := 0
	for _, record := range c.taskRecords {
		if record.AgentID == agentID {
			count++
		}
	}

	if len(c.taskRecords) == 0 {
		return 1.0
	}

	// Expected uniform distribution
	expectedCount := float64(len(c.taskRecords)) / float64(c.agentCount)

	// Underutilization = (expected - actual) / expected
	underutil := (expectedCount - float64(count)) / expectedCount
	return clamp(underutil, 0, 1)
}

// ============================================================================
// Diagnostics
// ============================================================================

// Diagnose returns a diagnostic report of the system state.
func (c *PhaseTransitionController) Diagnose() *PhaseDiagnostic {
	c.mu.RLock()
	defer c.mu.RUnlock()

	diag := &PhaseDiagnostic{
		Timestamp:      time.Now(),
		CurrentMetrics: c.currentMetrics,
		Parameters:     c.params,
		TaskCount:      len(c.taskRecords),
		HistorySize:    len(c.history),
	}

	if c.currentMetrics != nil {
		diag.Phase = c.currentMetrics.Phase.String()
		diag.HealthScore = c.currentMetrics.HealthScore

		// Recommendations
		switch c.currentMetrics.Phase {
		case PhaseFrozen:
			diag.Recommendations = []string{
				"System is too ordered - agents routing deterministically",
				"Increase temperature to add exploration",
				"Increase mutation rate for more variation",
				"Consider forcing some random agent selections",
			}
		case PhaseChaotic:
			diag.Recommendations = []string{
				"System is too chaotic - routing is nearly random",
				"Decrease temperature for more exploitation",
				"Increase consolidation to stabilize patterns",
				"Consider reducing exploration bonus",
			}
		case PhaseCritical:
			diag.Recommendations = []string{
				"System is at optimal criticality",
				"Continue monitoring for drift",
				"Current parameters are effective",
			}
		}
	}

	return diag
}

// PhaseDiagnostic contains diagnostic information.
type PhaseDiagnostic struct {
	Timestamp       time.Time
	Phase           string
	HealthScore     float64
	CurrentMetrics  *CriticalityMetrics
	Parameters      ControlParameters
	TaskCount       int
	HistorySize     int
	Recommendations []string
}

// ============================================================================
// Utility Functions
// ============================================================================

// Note: clamp function is defined in remem_loop.go - reusing it

// randomFloat64 returns a random float64 in [0, 1).
// In production, use crypto/rand for better randomness.
func randomFloat64() float64 {
	// Simple LCG for demonstration - replace with proper RNG
	return float64(time.Now().UnixNano()%1000) / 1000.0
}
