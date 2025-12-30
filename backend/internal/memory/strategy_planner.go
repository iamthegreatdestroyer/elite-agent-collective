// Package memory provides the MNEMONIC system for the Elite Agent Collective.
// This file implements the Multi-Strategy Planner for Phase 2.

package memory

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

// ============================================================================
// Multi-Strategy Planning Types
// ============================================================================

// Strategy represents a high-level approach to achieve a goal
type Strategy struct {
	ID                  string
	Name                string
	Description         string
	Approach            string
	Risk                float64 // 0-1 risk level
	Effectiveness       float64 // 0-1 expected effectiveness
	ResourceNeeds       map[string]float64
	Timeline            int // estimated days
	Dependencies        []string
	ConfirmedHypotheses []string // hypothesis support
	Status              StrategyStatus
	Priority            int
	CreatedAt           time.Time
}

// StrategyStatus represents strategy state
type StrategyStatus string

const (
	StrategyProposed  StrategyStatus = "proposed"
	StrategyEvaluated StrategyStatus = "evaluated"
	StrategySelected  StrategyStatus = "selected"
	StrategyRejected  StrategyStatus = "rejected"
	StrategyOptimized StrategyStatus = "optimized"
)

// StrategySet represents a collection of strategies for a goal
type StrategySet struct {
	ID                string
	GoalID            string
	Name              string
	Strategies        []*Strategy
	ResourceBudget    map[string]float64
	AllocationMap     map[string]map[string]float64 // strategy -> resource -> amount
	ComparisonResults []*StrategyComparison
	SelectedStrategy  *Strategy
	CreatedAt         time.Time
}

// StrategyComparison represents comparison between two strategies
type StrategyComparison struct {
	ID             string
	StrategyA      string
	StrategyB      string
	EffectivenessA float64
	EffectivenessB float64
	RiskA          float64
	RiskB          float64
	CostA          float64
	CostB          float64
	Winner         string
	Confidence     float64
	Timestamp      time.Time
}

// ResourceAllocation represents resource assignment to strategy
type ResourceAllocation struct {
	StrategyID  string
	Resource    string
	Amount      float64
	Utilization float64
	Efficiency  float64
	Timestamp   time.Time
}

// StrategyOptimization represents optimization applied
type StrategyOptimization struct {
	ID          string
	StrategyID  string
	Type        OptimizationType
	Parameter   string
	OldValue    float64
	NewValue    float64
	Improvement float64
	Timestamp   time.Time
}

// OptimizationType represents optimization category
type OptimizationType string

const (
	OptimizeRisk          OptimizationType = "risk_reduction"
	OptimizeEffectiveness OptimizationType = "effectiveness_increase"
	OptimizeResources     OptimizationType = "resource_optimization"
	OptimizeTimeline      OptimizationType = "timeline_compression"
)

// MultiStrategyPlannerConfig holds configuration
type MultiStrategyPlannerConfig struct {
	MaxStrategiesPerGoal int
	MinEffectiveness     float64
	MaxRisk              float64
	ResourceOptimization bool
	ComparisonDepth      int
}

// ============================================================================
// Multi-Strategy Planner Component
// ============================================================================

// MultiStrategyPlanner implements strategy generation and selection
type MultiStrategyPlanner struct {
	mu           sync.RWMutex
	config       MultiStrategyPlannerConfig
	strategySets map[string]*StrategySet
	strategies   map[string]*Strategy
	comparator   *StrategyComparator
	allocator    *ResourceAllocator
	optimizer    *StrategyOptimizer
	metrics      CognitiveMetrics
	requestCount int64
	successCount int64
	errorCount   int64
}

// StrategyComparator compares strategies
type StrategyComparator struct {
	mu          sync.RWMutex
	comparisons map[string]*StrategyComparison
	scores      map[string]float64
}

// ResourceAllocator manages resource distribution
type ResourceAllocator struct {
	mu          sync.RWMutex
	allocations map[string]*ResourceAllocation
	utilization map[string]float64
}

// StrategyOptimizer optimizes strategy parameters
type StrategyOptimizer struct {
	mu            sync.RWMutex
	optimizations map[string]*StrategyOptimization
	improvements  map[string]float64
}

// NewMultiStrategyPlanner creates a new planner
func NewMultiStrategyPlanner(config MultiStrategyPlannerConfig) *MultiStrategyPlanner {
	if config.MaxStrategiesPerGoal == 0 {
		config.MaxStrategiesPerGoal = 8
	}
	if config.MinEffectiveness == 0 {
		config.MinEffectiveness = 0.65
	}
	if config.MaxRisk == 0 {
		config.MaxRisk = 0.4
	}
	if config.ComparisonDepth == 0 {
		config.ComparisonDepth = 3
	}

	return &MultiStrategyPlanner{
		config:       config,
		strategySets: make(map[string]*StrategySet),
		strategies:   make(map[string]*Strategy),
		metrics: CognitiveMetrics{
			ComponentName: "MultiStrategyPlanner",
			CustomMetrics: make(map[string]interface{}),
		},
		comparator: &StrategyComparator{
			comparisons: make(map[string]*StrategyComparison),
			scores:      make(map[string]float64),
		},
		allocator: &ResourceAllocator{
			allocations: make(map[string]*ResourceAllocation),
			utilization: make(map[string]float64),
		},
		optimizer: &StrategyOptimizer{
			optimizations: make(map[string]*StrategyOptimization),
			improvements:  make(map[string]float64),
		},
	}
}

// Initialize sets up the planner
func (msp *MultiStrategyPlanner) Initialize(config interface{}) error {
	msp.mu.Lock()
	defer msp.mu.Unlock()

	msp.metrics.LastUpdated = time.Now()
	msp.metrics.CustomMetrics["initialized"] = true
	msp.metrics.CustomMetrics["max_strategies"] = msp.config.MaxStrategiesPerGoal

	return nil
}

// GenerateStrategies generates multiple strategies for a goal
func (msp *MultiStrategyPlanner) GenerateStrategies(ctx context.Context, goal *Goal) (*StrategySet, error) {
	msp.mu.Lock()
	defer msp.mu.Unlock()

	startTime := time.Now()
	msp.requestCount++

	strategySet := &StrategySet{
		ID:                fmt.Sprintf("strat-set-%s-%d", goal.ID, time.Now().Unix()),
		GoalID:            goal.ID,
		Name:              fmt.Sprintf("Strategies for %s", goal.Name),
		Strategies:        make([]*Strategy, 0),
		ResourceBudget:    make(map[string]float64),
		AllocationMap:     make(map[string]map[string]float64),
		ComparisonResults: make([]*StrategyComparison, 0),
		CreatedAt:         time.Now(),
	}

	// Generate strategies
	strategies := msp.generateInitialStrategies(goal)
	strategySet.Strategies = append(strategySet.Strategies, strategies...)

	// Initialize resource budgets
	msp.initializeResourceBudgets(strategySet)

	// Allocate resources
	for _, strategy := range strategies {
		allocation := msp.allocateResources(strategy, strategySet.ResourceBudget)
		strategySet.AllocationMap[strategy.ID] = allocation
	}

	// Compare strategies
	comparisons := msp.compareStrategies(strategies)
	strategySet.ComparisonResults = append(strategySet.ComparisonResults, comparisons...)

	// Select best strategy
	selected := msp.selectBestStrategy(strategies, comparisons)
	strategySet.SelectedStrategy = selected

	// Optimize selected strategy
	if selected != nil {
		optimized := msp.optimizeStrategy(selected)
		selected.Status = StrategyOptimized
		msp.optimizer.optimizations[selected.ID] = optimized
	}

	msp.strategySets[strategySet.ID] = strategySet

	duration := time.Since(startTime)
	msp.updateMetrics(duration, true)
	msp.successCount++

	return strategySet, nil
}

// generateInitialStrategies creates initial set of strategies
func (msp *MultiStrategyPlanner) generateInitialStrategies(goal *Goal) []*Strategy {
	strategies := make([]*Strategy, 0)

	numStrategies := 3 + (len(goal.Dependencies) % 5)
	if numStrategies > msp.config.MaxStrategiesPerGoal {
		numStrategies = msp.config.MaxStrategiesPerGoal
	}

	strategyTypes := []string{
		"Direct approach",
		"Incremental strategy",
		"Parallel implementation",
		"Phased rollout",
		"Risk-averse approach",
		"High-reward strategy",
		"Hybrid method",
		"Innovative solution",
	}

	for i := 0; i < numStrategies; i++ {
		strategy := &Strategy{
			ID:                  fmt.Sprintf("strategy-%s-%d", goal.ID, i),
			Name:                fmt.Sprintf("Strategy %d: %s", i+1, strategyTypes[i%len(strategyTypes)]),
			Description:         fmt.Sprintf("Approach for achieving %s", goal.Name),
			Approach:            strategyTypes[i%len(strategyTypes)],
			Risk:                0.1 + (float64(i) * 0.08),
			Effectiveness:       0.7 + (float64(i) * 0.04),
			ResourceNeeds:       msp.generateResourceNeeds(),
			Timeline:            7 + (i * 2),
			Dependencies:        goal.Dependencies,
			ConfirmedHypotheses: []string{},
			Status:              StrategyProposed,
			Priority:            i + 1,
			CreatedAt:           time.Now(),
		}

		if strategy.Risk > 0.4 {
			strategy.Risk = 0.4
		}
		if strategy.Effectiveness > 0.95 {
			strategy.Effectiveness = 0.95
		}

		strategies = append(strategies, strategy)
		msp.strategies[strategy.ID] = strategy
	}

	return strategies
}

// generateResourceNeeds creates resource requirements
func (msp *MultiStrategyPlanner) generateResourceNeeds() map[string]float64 {
	return map[string]float64{
		"development_hours": 40.0 + (float64(int(time.Now().Unix())%100) * 0.5),
		"infrastructure":    10.0 + (float64(int(time.Now().Unix())%50) * 0.2),
		"testing_effort":    20.0 + (float64(int(time.Now().Unix())%30) * 0.3),
		"documentation":     5.0 + (float64(int(time.Now().Unix())%20) * 0.1),
	}
}

// initializeResourceBudgets sets up resource budgets
func (msp *MultiStrategyPlanner) initializeResourceBudgets(strategySet *StrategySet) {
	totalNeeds := make(map[string]float64)

	for _, strategy := range strategySet.Strategies {
		for resource, amount := range strategy.ResourceNeeds {
			totalNeeds[resource] += amount
		}
	}

	// Budget is 1.5x total needs to allow flexibility
	for resource, total := range totalNeeds {
		strategySet.ResourceBudget[resource] = total * 1.5
	}
}

// allocateResources allocates resources to a strategy
func (msp *MultiStrategyPlanner) allocateResources(strategy *Strategy, budget map[string]float64) map[string]float64 {
	allocation := make(map[string]float64)

	totalNeed := 0.0
	for _, amount := range strategy.ResourceNeeds {
		totalNeed += amount
	}

	// Allocate proportionally based on strategy effectiveness
	proportion := (strategy.Effectiveness / 0.95) * 0.8 // 80% of effectiveness-based share

	for resource, need := range strategy.ResourceNeeds {
		available := budget[resource]
		allocated := need * proportion

		if allocated > available {
			allocated = available
		}

		allocation[resource] = allocated
	}

	return allocation
}

// compareStrategies compares strategies pairwise
func (msp *MultiStrategyPlanner) compareStrategies(strategies []*Strategy) []*StrategyComparison {
	comparisons := make([]*StrategyComparison, 0)

	for i := 0; i < len(strategies) && i < msp.config.ComparisonDepth; i++ {
		for j := i + 1; j < len(strategies) && j < msp.config.ComparisonDepth; j++ {
			stratA := strategies[i]
			stratB := strategies[j]

			comparison := &StrategyComparison{
				ID:             fmt.Sprintf("comp-%s-vs-%s", stratA.ID, stratB.ID),
				StrategyA:      stratA.ID,
				StrategyB:      stratB.ID,
				EffectivenessA: stratA.Effectiveness,
				EffectivenessB: stratB.Effectiveness,
				RiskA:          stratA.Risk,
				RiskB:          stratB.Risk,
				CostA:          msp.calculateCost(stratA),
				CostB:          msp.calculateCost(stratB),
				Confidence:     0.85,
				Timestamp:      time.Now(),
			}

			// Determine winner based on effectiveness-to-risk ratio
			scoreA := stratA.Effectiveness / (1.0 + stratA.Risk)
			scoreB := stratB.Effectiveness / (1.0 + stratB.Risk)

			if scoreA > scoreB {
				comparison.Winner = stratA.ID
			} else {
				comparison.Winner = stratB.ID
			}

			comparisons = append(comparisons, comparison)
			msp.comparator.comparisons[comparison.ID] = comparison
		}
	}

	return comparisons
}

// calculateCost calculates total cost of a strategy
func (msp *MultiStrategyPlanner) calculateCost(strategy *Strategy) float64 {
	cost := 0.0
	for _, amount := range strategy.ResourceNeeds {
		cost += amount
	}
	// Add risk factor
	cost *= (1.0 + strategy.Risk)
	return cost
}

// selectBestStrategy selects the optimal strategy
func (msp *MultiStrategyPlanner) selectBestStrategy(strategies []*Strategy, comparisons []*StrategyComparison) *Strategy {
	if len(strategies) == 0 {
		return nil
	}

	scores := make(map[string]float64)

	// Initialize scores
	for _, strategy := range strategies {
		scores[strategy.ID] = 0.0
	}

	// Count wins in comparisons
	for _, comparison := range comparisons {
		if comparison.Winner != "" {
			scores[comparison.Winner] += comparison.Confidence
		}
	}

	// Find highest scoring strategy
	bestStrategy := strategies[0]
	bestScore := scores[bestStrategy.ID]

	for _, strategy := range strategies {
		if scores[strategy.ID] > bestScore {
			bestScore = scores[strategy.ID]
			bestStrategy = strategy
		}
	}

	bestStrategy.Status = StrategySelected
	return bestStrategy
}

// optimizeStrategy optimizes a selected strategy
func (msp *MultiStrategyPlanner) optimizeStrategy(strategy *Strategy) *StrategyOptimization {
	optimization := &StrategyOptimization{
		ID:          fmt.Sprintf("opt-%s", strategy.ID),
		StrategyID:  strategy.ID,
		Type:        OptimizeRisk,
		Parameter:   "risk_level",
		OldValue:    strategy.Risk,
		NewValue:    strategy.Risk * 0.8, // Reduce risk by 20%
		Improvement: strategy.Risk * 0.2,
		Timestamp:   time.Now(),
	}

	// Apply optimization
	if optimization.NewValue < msp.config.MaxRisk {
		strategy.Risk = optimization.NewValue
	}

	return optimization
}

// GetSelectedStrategy returns the selected strategy
func (msp *MultiStrategyPlanner) GetSelectedStrategy(setID string) *Strategy {
	msp.mu.RLock()
	defer msp.mu.RUnlock()

	strategySet, ok := msp.strategySets[setID]
	if !ok {
		return nil
	}

	return strategySet.SelectedStrategy
}

// GetStrategies returns all strategies for a set
func (msp *MultiStrategyPlanner) GetStrategies(setID string) []*Strategy {
	msp.mu.RLock()
	defer msp.mu.RUnlock()

	strategySet, ok := msp.strategySets[setID]
	if !ok {
		return nil
	}

	return strategySet.Strategies
}

// GetAllocation returns resource allocation for a strategy
func (msp *MultiStrategyPlanner) GetAllocation(setID string, strategyID string) map[string]float64 {
	msp.mu.RLock()
	defer msp.mu.RUnlock()

	strategySet, ok := msp.strategySets[setID]
	if !ok {
		return nil
	}

	return strategySet.AllocationMap[strategyID]
}

// RankStrategies returns strategies ranked by effectiveness
func (msp *MultiStrategyPlanner) RankStrategies(setID string) []*Strategy {
	msp.mu.RLock()
	defer msp.mu.RUnlock()

	strategySet, ok := msp.strategySets[setID]
	if !ok {
		return nil
	}

	// Create a copy for sorting
	ranked := make([]*Strategy, len(strategySet.Strategies))
	copy(ranked, strategySet.Strategies)

	// Sort by effectiveness (descending)
	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].Effectiveness > ranked[j].Effectiveness
	})

	return ranked
}

// GetMetrics returns current metrics
func (msp *MultiStrategyPlanner) GetMetrics() CognitiveMetrics {
	msp.mu.RLock()
	defer msp.mu.RUnlock()

	return msp.metrics
}

// updateMetrics updates planner metrics
func (msp *MultiStrategyPlanner) updateMetrics(duration time.Duration, success bool) {
	msp.metrics.LastUpdated = time.Now()
	msp.metrics.TotalRequests = msp.requestCount
	msp.metrics.SuccessfulRequests = msp.successCount
	msp.metrics.FailedRequests = msp.errorCount
	msp.metrics.AverageLatency = duration
	msp.metrics.CustomMetrics["strategy_sets_created"] = len(msp.strategySets)
	msp.metrics.CustomMetrics["strategies_generated"] = len(msp.strategies)
}

// Shutdown gracefully shuts down the planner
func (msp *MultiStrategyPlanner) Shutdown() error {
	msp.mu.Lock()
	defer msp.mu.Unlock()

	msp.strategySets = make(map[string]*StrategySet)
	msp.strategies = make(map[string]*Strategy)
	msp.comparator.comparisons = make(map[string]*StrategyComparison)
	msp.allocator.allocations = make(map[string]*ResourceAllocation)
	msp.optimizer.optimizations = make(map[string]*StrategyOptimization)

	return nil
}

// GetName returns the component name
func (msp *MultiStrategyPlanner) GetName() string {
	return "MultiStrategyPlanner"
}

// DefaultMultiStrategyPlannerConfig returns default configuration
func DefaultMultiStrategyPlannerConfig() MultiStrategyPlannerConfig {
	return MultiStrategyPlannerConfig{
		MaxStrategiesPerGoal: 8,
		MinEffectiveness:     0.65,
		MaxRisk:              0.4,
		ResourceOptimization: true,
		ComparisonDepth:      3,
	}
}
