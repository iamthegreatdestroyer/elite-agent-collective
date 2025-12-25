// Package memory provides the MNEMONIC system for the Elite Agent Collective.
// This file implements a Hierarchical Task Network (HTN) Planner from @NEURAL's
// Cognitive Architecture Analysis.
//
// HTN Planning decomposes high-level goals into primitive actions through:
// - Task decomposition: Abstract tasks → subtasks using methods
// - Precondition checking: Verify state requirements before execution
// - Plan synthesis: Combine decomposed tasks into executable sequences
//
// This enables agents to handle complex, multi-step tasks automatically.

package memory

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

// ============================================================================
// Core Types
// ============================================================================

// Task represents a goal to achieve or action to perform.
type Task struct {
	ID           string
	Name         string
	Parameters   map[string]interface{}
	Priority     int
	IsPrimitive  bool // True if this is an executable action
	Dependencies []string
	Deadline     time.Time
	Metadata     map[string]interface{}
}

// Method represents a way to decompose a task into subtasks.
type Method struct {
	ID            string
	Name          string
	TaskName      string // The task this method decomposes
	Preconditions []*Precondition
	Subtasks      []*Task
	Ordering      OrderingType
	Priority      int // Higher priority methods tried first
}

// OrderingType specifies how subtasks should be executed.
type OrderingType string

const (
	OrderingSequential OrderingType = "sequential"
	OrderingParallel   OrderingType = "parallel"
	OrderingPartial    OrderingType = "partial" // Some sequential, some parallel
)

// Precondition represents a condition that must hold for an action/method.
type Precondition struct {
	Feature  string
	Operator string // "eq", "ne", "gt", "lt", "gte", "lte", "exists", "not_exists"
	Value    interface{}
}

// Evaluate checks if precondition holds in the given state.
func (p *Precondition) Evaluate(state *PlannerState) bool {
	val, exists := state.Features[p.Feature]

	switch p.Operator {
	case "exists":
		return exists
	case "not_exists":
		return !exists
	case "eq":
		return exists && val == p.Value
	case "ne":
		return !exists || val != p.Value
	case "gt", "lt", "gte", "lte":
		if !exists {
			return false
		}
		return plannerCompareValues(val, p.Value, p.Operator)
	default:
		return false
	}
}

func plannerCompareValues(a, b interface{}, op string) bool {
	af, aok := toFloat(a)
	bf, bok := toFloat(b)
	if !aok || !bok {
		return false
	}

	switch op {
	case "gt":
		return af > bf
	case "lt":
		return af < bf
	case "gte":
		return af >= bf
	case "lte":
		return af <= bf
	default:
		return false
	}
}

func toFloat(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case int32:
		return float64(n), true
	default:
		return 0, false
	}
}

// PlannerAction represents a primitive, executable action.
type PlannerAction struct {
	ID            string
	Name          string
	Preconditions []*Precondition
	Effects       []*Effect
	Cost          float64
	Duration      time.Duration
	AgentRequired string // Which agent should execute this
}

// ApplicableIn checks if action can be executed in given state.
func (a *PlannerAction) ApplicableIn(state *PlannerState) bool {
	for _, precond := range a.Preconditions {
		if !precond.Evaluate(state) {
			return false
		}
	}
	return true
}

// Effect represents a change to the state after action execution.
type Effect struct {
	Feature   string
	Operation string // "set", "add", "remove", "increment", "decrement"
	Value     interface{}
}

// Apply applies the effect to the state.
func (e *Effect) Apply(state *PlannerState) {
	switch e.Operation {
	case "set":
		state.Features[e.Feature] = e.Value
	case "add":
		if arr, ok := state.Features[e.Feature].([]interface{}); ok {
			state.Features[e.Feature] = append(arr, e.Value)
		} else {
			state.Features[e.Feature] = []interface{}{e.Value}
		}
	case "remove":
		delete(state.Features, e.Feature)
	case "increment":
		if val, ok := state.Features[e.Feature]; ok {
			if f, fok := toFloat(val); fok {
				if inc, iok := toFloat(e.Value); iok {
					state.Features[e.Feature] = f + inc
				}
			}
		}
	case "decrement":
		if val, ok := state.Features[e.Feature]; ok {
			if f, fok := toFloat(val); fok {
				if dec, dok := toFloat(e.Value); dok {
					state.Features[e.Feature] = f - dec
				}
			}
		}
	}
}

// PlannerState represents the current state of the world.
type PlannerState struct {
	Features map[string]interface{}
	mu       sync.RWMutex
}

// NewPlannerState creates a new planner state.
func NewPlannerState() *PlannerState {
	return &PlannerState{
		Features: make(map[string]interface{}),
	}
}

// Clone creates a copy of the state.
func (s *PlannerState) Clone() *PlannerState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	clone := NewPlannerState()
	for k, v := range s.Features {
		clone.Features[k] = v
	}
	return clone
}

// Get retrieves a feature value.
func (s *PlannerState) Get(feature string) (interface{}, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.Features[feature]
	return val, ok
}

// Set sets a feature value.
func (s *PlannerState) Set(feature string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Features[feature] = value
}

// Plan represents a sequence of actions to achieve a goal.
type Plan struct {
	ID                string
	GoalTask          *Task
	Actions           []*PlannerAction
	TotalCost         float64
	EstimatedDuration time.Duration
	CreatedAt         time.Time
	Feasible          bool
	Explanation       string
}

// ============================================================================
// Hierarchical Planner
// ============================================================================

// HierarchicalPlanner implements HTN planning.
type HierarchicalPlanner struct {
	primitiveActions map[string]*PlannerAction
	methods          map[string][]*Method // TaskName → applicable methods
	worldModel       *WorldModel

	// Configuration
	maxDepth    int
	maxPlanTime time.Duration

	// Statistics
	stats   *PlannerStats
	statsMu sync.RWMutex
	mu      sync.RWMutex
}

// PlannerStats tracks planning statistics.
type PlannerStats struct {
	TotalPlans        int64
	SuccessfulPlans   int64
	FailedPlans       int64
	AverageDepth      float64
	AveragePlanLength float64
	AverageLatency    time.Duration
}

// PlannerConfig configures the hierarchical planner.
type PlannerConfig struct {
	MaxDepth    int
	MaxPlanTime time.Duration
}

// DefaultPlannerConfig returns default configuration.
func DefaultPlannerConfig() *PlannerConfig {
	return &PlannerConfig{
		MaxDepth:    20,
		MaxPlanTime: 10 * time.Second,
	}
}

// NewHierarchicalPlanner creates a new HTN planner.
func NewHierarchicalPlanner(worldModel *WorldModel, config *PlannerConfig) *HierarchicalPlanner {
	if config == nil {
		config = DefaultPlannerConfig()
	}

	return &HierarchicalPlanner{
		primitiveActions: make(map[string]*PlannerAction),
		methods:          make(map[string][]*Method),
		worldModel:       worldModel,
		maxDepth:         config.MaxDepth,
		maxPlanTime:      config.MaxPlanTime,
		stats:            &PlannerStats{},
	}
}

// RegisterAction registers a primitive action.
func (p *HierarchicalPlanner) RegisterAction(action *PlannerAction) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.primitiveActions[action.Name] = action
}

// RegisterMethod registers a decomposition method.
func (p *HierarchicalPlanner) RegisterMethod(method *Method) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.methods[method.TaskName]; !exists {
		p.methods[method.TaskName] = make([]*Method, 0)
	}
	p.methods[method.TaskName] = append(p.methods[method.TaskName], method)

	// Sort methods by priority (higher first)
	sort.Slice(p.methods[method.TaskName], func(i, j int) bool {
		return p.methods[method.TaskName][i].Priority > p.methods[method.TaskName][j].Priority
	})
}

// Plan creates a plan to achieve the given task from the current state.
func (p *HierarchicalPlanner) Plan(task *Task, state *PlannerState) (*Plan, error) {
	startTime := time.Now()

	p.statsMu.Lock()
	p.stats.TotalPlans++
	p.statsMu.Unlock()

	// Create deadline context
	deadline := startTime.Add(p.maxPlanTime)

	// Start recursive planning
	actions, depth, err := p.planTask(task, state, 0, deadline)

	plan := &Plan{
		ID:        fmt.Sprintf("plan-%s-%d", task.ID, time.Now().UnixNano()),
		GoalTask:  task,
		Actions:   actions,
		CreatedAt: startTime,
	}

	if err != nil {
		plan.Feasible = false
		plan.Explanation = err.Error()

		p.statsMu.Lock()
		p.stats.FailedPlans++
		p.updatePlannerLatency(time.Since(startTime))
		p.statsMu.Unlock()

		return plan, err
	}

	// Calculate plan metrics
	plan.Feasible = true
	plan.Explanation = fmt.Sprintf("Successfully decomposed into %d actions", len(actions))
	for _, action := range actions {
		plan.TotalCost += action.Cost
		plan.EstimatedDuration += action.Duration
	}

	p.statsMu.Lock()
	p.stats.SuccessfulPlans++
	p.updatePlannerDepth(float64(depth))
	p.updatePlannerLength(float64(len(actions)))
	p.updatePlannerLatency(time.Since(startTime))
	p.statsMu.Unlock()

	return plan, nil
}

// planTask recursively decomposes a task into primitive actions.
func (p *HierarchicalPlanner) planTask(
	task *Task,
	state *PlannerState,
	depth int,
	deadline time.Time,
) ([]*PlannerAction, int, error) {
	// Check deadline
	if time.Now().After(deadline) {
		return nil, depth, ErrPlanningTimeout
	}

	// Check depth limit
	if depth >= p.maxDepth {
		return nil, depth, ErrMaxDepthReached
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	// Check if task is primitive (directly executable)
	if action, ok := p.primitiveActions[task.Name]; ok {
		if action.ApplicableIn(state) {
			return []*PlannerAction{action}, depth + 1, nil
		}
		return nil, depth, fmt.Errorf("action '%s' preconditions not met", task.Name)
	}

	// Non-primitive: find applicable methods
	methods, exists := p.methods[task.Name]
	if !exists {
		return nil, depth, fmt.Errorf("no methods found for task '%s'", task.Name)
	}

	for _, method := range methods {
		// Check method preconditions
		if !p.methodApplicable(method, state) {
			continue
		}

		// Try to plan for all subtasks
		allActions := make([]*PlannerAction, 0)
		currentState := state.Clone()
		allSuccessful := true
		maxSubDepth := depth
		var propagatedErr error

		for _, subtask := range method.Subtasks {
			subActions, subDepth, err := p.planTask(subtask, currentState, depth+1, deadline)
			if err != nil {
				allSuccessful = false
				// Propagate special errors (timeout, max depth)
				if err == ErrMaxDepthReached || err == ErrPlanningTimeout {
					propagatedErr = err
				}
				break
			}

			allActions = append(allActions, subActions...)
			if subDepth > maxSubDepth {
				maxSubDepth = subDepth
			}

			// Apply effects of sub-actions to state for next subtask
			for _, subAction := range subActions {
				for _, effect := range subAction.Effects {
					effect.Apply(currentState)
				}
			}
		}

		if allSuccessful {
			return allActions, maxSubDepth, nil
		}
		// Check for propagated errors (timeout, max depth)
		if propagatedErr != nil {
			return nil, depth, propagatedErr
		}
		// Try next method
	}

	return nil, depth, fmt.Errorf("no applicable method found for task '%s'", task.Name)
}

// methodApplicable checks if a method's preconditions are met.
func (p *HierarchicalPlanner) methodApplicable(method *Method, state *PlannerState) bool {
	for _, precond := range method.Preconditions {
		if !precond.Evaluate(state) {
			return false
		}
	}
	return true
}

// ExecutePlan executes a plan, returning the final state.
func (p *HierarchicalPlanner) ExecutePlan(plan *Plan, initialState *PlannerState) (*PlannerState, error) {
	if !plan.Feasible {
		return nil, fmt.Errorf("cannot execute infeasible plan: %s", plan.Explanation)
	}

	state := initialState.Clone()

	for _, action := range plan.Actions {
		// Check preconditions (they may have changed during execution)
		if !action.ApplicableIn(state) {
			return state, fmt.Errorf("action '%s' preconditions no longer met during execution", action.Name)
		}

		// Apply effects
		for _, effect := range action.Effects {
			effect.Apply(state)
		}
	}

	return state, nil
}

// SimulatePlan simulates plan execution using the world model.
func (p *HierarchicalPlanner) SimulatePlan(plan *Plan, initialState *PlannerState) ([]*PlannerState, error) {
	if p.worldModel == nil {
		return nil, fmt.Errorf("no world model available for simulation")
	}

	states := make([]*PlannerState, 0, len(plan.Actions)+1)
	state := initialState.Clone()
	states = append(states, state)

	for _, action := range plan.Actions {
		// Apply effects
		newState := state.Clone()
		for _, effect := range action.Effects {
			effect.Apply(newState)
		}
		states = append(states, newState)
		state = newState
	}

	return states, nil
}

// Planning errors
var (
	ErrPlanningTimeout = fmt.Errorf("planning exceeded timeout")
	ErrMaxDepthReached = fmt.Errorf("max planning depth reached")
	ErrNoPlanFound     = fmt.Errorf("no valid plan found")
)

// updatePlannerLatency updates running average.
func (p *HierarchicalPlanner) updatePlannerLatency(latency time.Duration) {
	total := p.stats.SuccessfulPlans + p.stats.FailedPlans
	if total == 1 {
		p.stats.AverageLatency = latency
	} else {
		n := float64(total)
		p.stats.AverageLatency = time.Duration(
			(float64(p.stats.AverageLatency)*(n-1) + float64(latency)) / n,
		)
	}
}

// updatePlannerDepth updates running average.
func (p *HierarchicalPlanner) updatePlannerDepth(depth float64) {
	if p.stats.SuccessfulPlans == 1 {
		p.stats.AverageDepth = depth
	} else {
		n := float64(p.stats.SuccessfulPlans)
		p.stats.AverageDepth = (p.stats.AverageDepth*(n-1) + depth) / n
	}
}

// updatePlannerLength updates running average.
func (p *HierarchicalPlanner) updatePlannerLength(length float64) {
	if p.stats.SuccessfulPlans == 1 {
		p.stats.AveragePlanLength = length
	} else {
		n := float64(p.stats.SuccessfulPlans)
		p.stats.AveragePlanLength = (p.stats.AveragePlanLength*(n-1) + length) / n
	}
}

// GetStats returns planning statistics.
func (p *HierarchicalPlanner) GetStats() *PlannerStats {
	p.statsMu.RLock()
	defer p.statsMu.RUnlock()

	return &PlannerStats{
		TotalPlans:        p.stats.TotalPlans,
		SuccessfulPlans:   p.stats.SuccessfulPlans,
		FailedPlans:       p.stats.FailedPlans,
		AverageDepth:      p.stats.AverageDepth,
		AveragePlanLength: p.stats.AveragePlanLength,
		AverageLatency:    p.stats.AverageLatency,
	}
}

// GetRegisteredActions returns all registered primitive actions.
func (p *HierarchicalPlanner) GetRegisteredActions() map[string]*PlannerAction {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make(map[string]*PlannerAction)
	for k, v := range p.primitiveActions {
		result[k] = v
	}
	return result
}

// GetRegisteredMethods returns all registered methods.
func (p *HierarchicalPlanner) GetRegisteredMethods() map[string][]*Method {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make(map[string][]*Method)
	for k, v := range p.methods {
		result[k] = v
	}
	return result
}

// ============================================================================
// Agent Task Planner
// ============================================================================

// AgentTaskPlanner specializes the HTN planner for agent coordination.
type AgentTaskPlanner struct {
	planner   *HierarchicalPlanner
	agentInfo map[string]*AgentInfo
	mu        sync.RWMutex
}

// AgentInfo describes an agent's capabilities.
type AgentInfo struct {
	ID           string
	Tier         int
	Capabilities []string
	Cost         float64
	Availability bool
}

// NewAgentTaskPlanner creates a planner specialized for agent tasks.
func NewAgentTaskPlanner(worldModel *WorldModel) *AgentTaskPlanner {
	return &AgentTaskPlanner{
		planner:   NewHierarchicalPlanner(worldModel, nil),
		agentInfo: make(map[string]*AgentInfo),
	}
}

// RegisterAgent registers an agent's capabilities.
func (atp *AgentTaskPlanner) RegisterAgent(info *AgentInfo) {
	atp.mu.Lock()
	defer atp.mu.Unlock()
	atp.agentInfo[info.ID] = info

	// Register agent as a primitive action
	atp.planner.RegisterAction(&PlannerAction{
		ID:            fmt.Sprintf("action-%s", info.ID),
		Name:          fmt.Sprintf("invoke_%s", info.ID),
		Preconditions: []*Precondition{},
		Effects: []*Effect{
			{Feature: fmt.Sprintf("%s_executed", info.ID), Operation: "set", Value: true},
		},
		Cost:          info.Cost,
		AgentRequired: info.ID,
	})
}

// PlanAgentCoordination creates a plan for coordinating multiple agents.
func (atp *AgentTaskPlanner) PlanAgentCoordination(
	task *Task,
	state *PlannerState,
) (*Plan, error) {
	return atp.planner.Plan(task, state)
}

// RegisterCompositeTask registers a composite task that decomposes into agent invocations.
func (atp *AgentTaskPlanner) RegisterCompositeTask(
	taskName string,
	agentSequence []string,
	ordering OrderingType,
) {
	subtasks := make([]*Task, len(agentSequence))
	for i, agentID := range agentSequence {
		subtasks[i] = &Task{
			ID:          fmt.Sprintf("%s-step-%d", taskName, i),
			Name:        fmt.Sprintf("invoke_%s", agentID),
			IsPrimitive: true,
		}
	}

	method := &Method{
		ID:            fmt.Sprintf("method-%s-default", taskName),
		Name:          fmt.Sprintf("%s_method", taskName),
		TaskName:      taskName,
		Preconditions: []*Precondition{},
		Subtasks:      subtasks,
		Ordering:      ordering,
		Priority:      1,
	}

	atp.planner.RegisterMethod(method)
}

// GetPlanner returns the underlying hierarchical planner.
func (atp *AgentTaskPlanner) GetPlanner() *HierarchicalPlanner {
	return atp.planner
}
