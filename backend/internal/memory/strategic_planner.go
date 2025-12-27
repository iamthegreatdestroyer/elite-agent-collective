// Package memory provides the MNEMONIC system for the Elite Agent Collective.
// This file implements the Strategic Planner for Phase 2.

package memory

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

// ============================================================================
// Strategic Planner Component
// ============================================================================

// LookaheadNode represents a node in the lookahead tree
type LookaheadNode struct {
	ID       string
	Goal     *Goal
	State    interface{}
	Children []*LookaheadNode
	Parent   *LookaheadNode
	Depth    int
	Score    float64
	Visits   int
	Reward   float64
}

// PlanningConfig holds configuration for the planner
type PlanningConfig struct {
	MaxLookaheadDepth  int
	MaxBranchingFactor int
	TimeoutSeconds     int
	PlanCachingEnabled bool
	OptimizationLevel  int
}

// StrategicPlanner implements strategic planning for the cognitive system
type StrategicPlanner struct {
	mu            sync.RWMutex
	config        PlanningConfig
	plans         map[string]*Plan
	lookaheadTree *LookaheadNode
	planCache     map[string]*Plan
	metrics       CognitiveMetrics
	lastUpdate    time.Time
	requestCount  int64
	successCount  int64
	errorCount    int64
	workingMemory *CognitiveWorkingMemory
	goalStack     *GoalStack
}

// NewStrategicPlanner creates a new strategic planner
func NewStrategicPlanner(config PlanningConfig) *StrategicPlanner {
	if config.MaxLookaheadDepth == 0 {
		config.MaxLookaheadDepth = 5
	}
	if config.MaxBranchingFactor == 0 {
		config.MaxBranchingFactor = 3
	}
	if config.TimeoutSeconds == 0 {
		config.TimeoutSeconds = 30
	}

	return &StrategicPlanner{
		config:        config,
		plans:         make(map[string]*Plan),
		planCache:     make(map[string]*Plan),
		lookaheadTree: nil,
		metrics: CognitiveMetrics{
			ComponentName: "StrategicPlanner",
			CustomMetrics: make(map[string]interface{}),
		},
		lastUpdate: time.Now(),
	}
}

// Initialize sets up the planner
func (sp *StrategicPlanner) Initialize(config interface{}) error {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	sp.metrics.LastUpdated = time.Now()
	sp.metrics.CustomMetrics["initialized"] = true
	sp.metrics.CustomMetrics["lookahead_depth"] = sp.config.MaxLookaheadDepth

	return nil
}

// CreatePlan creates a strategic plan for a goal
func (sp *StrategicPlanner) CreatePlan(ctx context.Context, goal *Goal) (*Plan, error) {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	startTime := time.Now()
	sp.requestCount++

	// Check cache first
	if sp.config.PlanCachingEnabled {
		if cached, ok := sp.planCache[goal.ID]; ok {
			sp.successCount++
			return cached, nil
		}
	}

	plan := &Plan{
		ID:          fmt.Sprintf("plan-%s-%d", goal.ID, time.Now().Unix()),
		GoalTask:    nil,
		Actions:     make([]*PlannerAction, 0),
		TotalCost:   0.0,
		CreatedAt:   time.Now(),
		Feasible:    true,
		Explanation: "Strategic plan created",
	}

	// Generate initial actions
	actions := sp.generateInitialActions(goal)
	plan.Actions = append(plan.Actions, actions...)

	// Calculate plan metrics
	for _, action := range plan.Actions {
		plan.TotalCost += action.Cost
	}

	// Evaluate plan feasibility
	plan.Feasible = sp.evaluatePlan(plan)

	// Build lookahead tree for deeper analysis
	sp.buildLookaheadTree(ctx, goal, plan)

	// Optimize plan if not feasible
	if !plan.Feasible {
		plan = sp.optimizePlan(plan)
	}

	// Cache the plan
	if sp.config.PlanCachingEnabled {
		sp.planCache[goal.ID] = plan
	}

	// Store the plan
	sp.plans[plan.ID] = plan

	duration := time.Since(startTime)
	sp.updateMetrics(duration, true)
	sp.successCount++

	return plan, nil
}

// generateInitialActions generates the first actions toward a goal
func (sp *StrategicPlanner) generateInitialActions(goal *Goal) []*PlannerAction {
	actions := make([]*PlannerAction, 0)

	// Generate 1-3 initial actions based on goal characteristics
	numActions := 1 + (len(goal.Dependencies) % 2)

	for i := 0; i < numActions; i++ {
		action := &PlannerAction{
			ID:            fmt.Sprintf("action-%s-%d", goal.ID, i),
			Name:          fmt.Sprintf("Action for %s (step %d)", goal.Name, i),
			Cost:          0.5 + float64(i)*0.1,
			Preconditions: make([]*Precondition, 0),
			Effects:       make([]*Effect, 0),
		}
		actions = append(actions, action)
	}

	return actions
}

// evaluatePlan evaluates the quality of a plan
func (sp *StrategicPlanner) evaluatePlan(plan *Plan) bool {
	if len(plan.Actions) == 0 {
		return false
	}

	// Plan is feasible if it has actions and reasonable cost
	return plan.TotalCost < 100.0 && len(plan.Actions) > 0
}

// buildLookaheadTree builds a tree of lookahead nodes
func (sp *StrategicPlanner) buildLookaheadTree(ctx context.Context, goal *Goal, plan *Plan) {
	root := &LookaheadNode{
		ID:       fmt.Sprintf("lookahead-root-%s", goal.ID),
		Goal:     goal,
		Depth:    0,
		Score:    0.8,
		Children: make([]*LookaheadNode, 0),
	}

	sp.expandNode(root, 0)
	sp.lookaheadTree = root
}

// expandNode recursively expands a lookahead node
func (sp *StrategicPlanner) expandNode(node *LookaheadNode, depth int) {
	if depth >= sp.config.MaxLookaheadDepth {
		return
	}

	// Create child nodes
	numChildren := 1 + (depth % sp.config.MaxBranchingFactor)

	for i := 0; i < numChildren; i++ {
		child := &LookaheadNode{
			ID:       fmt.Sprintf("lookahead-node-%s-%d-%d", node.ID, depth, i),
			Goal:     node.Goal,
			Depth:    depth + 1,
			Score:    node.Score * (0.9 - float64(i)*0.05),
			Parent:   node,
			Children: make([]*LookaheadNode, 0),
		}

		node.Children = append(node.Children, child)

		// Recursively expand child
		sp.expandNode(child, depth+1)
	}
}

// optimizePlan improves a plan's feasibility
func (sp *StrategicPlanner) optimizePlan(plan *Plan) *Plan {
	// Add extra actions to increase feasibility
	extraAction := &PlannerAction{
		ID:            fmt.Sprintf("extra-action-%s", plan.ID),
		Name:          "Optimization action",
		Cost:          0.3,
		Preconditions: make([]*Precondition, 0),
		Effects:       make([]*Effect, 0),
	}

	plan.Actions = append(plan.Actions, extraAction)
	plan.TotalCost += extraAction.Cost
	plan.Feasible = sp.evaluatePlan(plan)

	return plan
}

// GetPlan retrieves a stored plan
func (sp *StrategicPlanner) GetPlan(planID string) *Plan {
	sp.mu.RLock()
	defer sp.mu.RUnlock()

	return sp.plans[planID]
}

// ExecutePlan marks a plan as executing
func (sp *StrategicPlanner) ExecutePlan(planID string) error {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	plan, ok := sp.plans[planID]
	if !ok {
		return fmt.Errorf("plan not found: %s", planID)
	}

	plan.Explanation = "Plan execution started"

	return nil
}

// GetLookaheadTree returns the current lookahead tree
func (sp *StrategicPlanner) GetLookaheadTree() *LookaheadNode {
	sp.mu.RLock()
	defer sp.mu.RUnlock()

	return sp.lookaheadTree
}

// GetBestStrategy selects the best strategy from lookahead tree
func (sp *StrategicPlanner) GetBestStrategy(node *LookaheadNode) *LookaheadNode {
	if node == nil || len(node.Children) == 0 {
		return node
	}

	// Sort children by score
	children := make([]*LookaheadNode, len(node.Children))
	copy(children, node.Children)

	sort.Slice(children, func(i, j int) bool {
		return children[i].Score > children[j].Score
	})

	return children[0]
}

// updateMetrics updates planning metrics
func (sp *StrategicPlanner) updateMetrics(duration time.Duration, success bool) {
	sp.metrics.LastUpdated = time.Now()
	sp.metrics.TotalRequests = sp.requestCount
	sp.metrics.SuccessfulRequests = sp.successCount
	sp.metrics.FailedRequests = sp.errorCount
	sp.metrics.AverageLatency = duration
	sp.metrics.CustomMetrics["plans_created"] = len(sp.plans)
	sp.metrics.CustomMetrics["cache_size"] = len(sp.planCache)

	if sp.lookaheadTree != nil {
		sp.metrics.CustomMetrics["lookahead_nodes"] = sp.countLookaheadNodes(sp.lookaheadTree)
	}
}

// countLookaheadNodes counts total nodes in lookahead tree
func (sp *StrategicPlanner) countLookaheadNodes(node *LookaheadNode) int {
	if node == nil {
		return 0
	}

	count := 1
	for _, child := range node.Children {
		count += sp.countLookaheadNodes(child)
	}

	return count
}

// GetMetrics returns current metrics
func (sp *StrategicPlanner) GetMetrics() CognitiveMetrics {
	sp.mu.RLock()
	defer sp.mu.RUnlock()

	return sp.metrics
}

// Shutdown gracefully shuts down the planner
func (sp *StrategicPlanner) Shutdown() error {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	sp.plans = make(map[string]*Plan)
	sp.planCache = make(map[string]*Plan)
	sp.lookaheadTree = nil

	return nil
}

// GetName returns the component name
func (sp *StrategicPlanner) GetName() string {
	return "StrategicPlanner"
}

// DefaultPlanningConfig returns default configuration
func DefaultPlanningConfig() PlanningConfig {
	return PlanningConfig{
		MaxLookaheadDepth:  5,
		MaxBranchingFactor: 3,
		TimeoutSeconds:     30,
		PlanCachingEnabled: true,
		OptimizationLevel:  2,
	}
}
