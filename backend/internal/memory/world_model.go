// Package memory provides the MNEMONIC system for the Elite Agent Collective.
// This file implements the World Model / Mental Imagery system from @NEURAL's
// Cognitive Architecture Analysis.
//
// Mental Imagery / Simulation (Important Cognitive Function):
// - Enables "imagining" outcomes before acting
// - Provides forward modeling for planning
// - Predicts state transitions given actions
// - Estimates success probabilities for trajectories
// - Supports comparison of alternative action sequences
//
// Humans use mental simulation to predict outcomes, evaluate options, and plan.
// This module provides the cognitive infrastructure for look-ahead reasoning.

package memory

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// ============================================================================
// Error Definitions
// ============================================================================

var (
	ErrInvalidState            = errors.New("invalid state")
	ErrInvalidAction           = errors.New("invalid action")
	ErrSimulationFailed        = errors.New("simulation failed")
	ErrSimulationDepthExceeded = errors.New("maximum simulation depth exceeded")
	ErrNoTransitionsDefined    = errors.New("no transitions defined for state")
)

// ============================================================================
// State Types
// ============================================================================

// StateType classifies the type of state.
type StateType int

const (
	// StateInitial is the starting state
	StateInitial StateType = iota

	// StateIntermediate is a transitional state
	StateIntermediate

	// StateGoal is a desired end state
	StateGoal

	// StateFailure is an undesired end state
	StateFailure

	// StateUnknown is when state cannot be determined
	StateUnknown
)

// String returns human-readable state type.
func (st StateType) String() string {
	switch st {
	case StateInitial:
		return "initial"
	case StateIntermediate:
		return "intermediate"
	case StateGoal:
		return "goal"
	case StateFailure:
		return "failure"
	case StateUnknown:
		return "unknown"
	default:
		return "undefined"
	}
}

// IsTerminal returns true if this is an end state.
func (st StateType) IsTerminal() bool {
	return st == StateGoal || st == StateFailure
}

// ============================================================================
// State
// ============================================================================

// stateIDCounter provides unique IDs
var stateIDCounter uint64

// State represents a point in the simulation state space.
type State struct {
	// ID uniquely identifies this state
	ID string

	// Type classifies the state
	Type StateType

	// Description is a human-readable label
	Description string

	// Features are the state variables (key-value pairs)
	Features map[string]interface{}

	// Activation represents how "active" or accessible this state is
	Activation float64

	// Confidence in this state representation (0.0 to 1.0)
	Confidence float64

	// Parent is the state this was derived from (if any)
	ParentID string

	// Action that led to this state (if any)
	ActionID string

	// CreatedAt timestamp
	CreatedAt time.Time

	// Metadata holds additional properties
	Metadata map[string]interface{}
}

// NewState creates a new state.
func NewState(stateType StateType, description string) *State {
	return &State{
		ID:          fmt.Sprintf("state-%d", atomic.AddUint64(&stateIDCounter, 1)),
		Type:        stateType,
		Description: description,
		Features:    make(map[string]interface{}),
		Activation:  1.0,
		Confidence:  1.0,
		CreatedAt:   time.Now(),
		Metadata:    make(map[string]interface{}),
	}
}

// Clone creates a deep copy of the state.
func (s *State) Clone() *State {
	clone := &State{
		ID:          s.ID,
		Type:        s.Type,
		Description: s.Description,
		Features:    make(map[string]interface{}),
		Activation:  s.Activation,
		Confidence:  s.Confidence,
		ParentID:    s.ParentID,
		ActionID:    s.ActionID,
		CreatedAt:   s.CreatedAt,
		Metadata:    make(map[string]interface{}),
	}
	for k, v := range s.Features {
		clone.Features[k] = v
	}
	for k, v := range s.Metadata {
		clone.Metadata[k] = v
	}
	return clone
}

// SetFeature sets a feature value.
func (s *State) SetFeature(key string, value interface{}) {
	s.Features[key] = value
}

// GetFeature gets a feature value.
func (s *State) GetFeature(key string) (interface{}, bool) {
	val, exists := s.Features[key]
	return val, exists
}

// GetFeatureFloat gets a feature as float64.
func (s *State) GetFeatureFloat(key string) (float64, bool) {
	val, exists := s.Features[key]
	if !exists {
		return 0, false
	}
	switch v := val.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	default:
		return 0, false
	}
}

// Similarity computes similarity to another state based on features.
func (s *State) Similarity(other *State) float64 {
	if other == nil {
		return 0
	}

	// Count matching features
	matches := 0
	total := 0

	for k, v := range s.Features {
		total++
		if ov, exists := other.Features[k]; exists && v == ov {
			matches++
		}
	}

	// Include features only in other
	for k := range other.Features {
		if _, exists := s.Features[k]; !exists {
			total++
		}
	}

	if total == 0 {
		return 1.0 // Both empty
	}

	return float64(matches) / float64(total)
}

// ============================================================================
// Action
// ============================================================================

// actionIDCounter provides unique IDs
var actionIDCounter uint64

// SimActionType classifies simulation actions.
type SimActionType int

const (
	// SimActionAgent invokes an agent
	SimActionAgent SimActionType = iota

	// SimActionQuery performs a query/retrieval
	SimActionQuery

	// SimActionTransform modifies state
	SimActionTransform

	// SimActionComposite is a sequence of actions
	SimActionComposite

	// SimActionConditional depends on state
	SimActionConditional
)

// String returns human-readable action type.
func (at SimActionType) String() string {
	switch at {
	case SimActionAgent:
		return "agent"
	case SimActionQuery:
		return "query"
	case SimActionTransform:
		return "transform"
	case SimActionComposite:
		return "composite"
	case SimActionConditional:
		return "conditional"
	default:
		return "unknown"
	}
}

// SimAction represents an action that can be simulated.
type SimAction struct {
	// ID uniquely identifies this action
	ID string

	// Type classifies the action
	Type SimActionType

	// Name is a human-readable label
	Name string

	// Description explains what the action does
	Description string

	// Parameters for the action
	Parameters map[string]interface{}

	// Preconditions that must hold for action to be applicable
	Preconditions []Predicate

	// Effects describe how the action changes state
	Effects []StateEffect

	// Cost represents the resource cost of this action
	Cost float64

	// ExpectedDuration estimates execution time
	ExpectedDuration time.Duration

	// SuccessProbability is the base success rate
	SuccessProbability float64

	// Metadata holds additional properties
	Metadata map[string]interface{}
}

// NewSimAction creates a new simulation action.
func NewSimAction(actionType SimActionType, name string) *SimAction {
	return &SimAction{
		ID:                 fmt.Sprintf("action-%d", atomic.AddUint64(&actionIDCounter, 1)),
		Type:               actionType,
		Name:               name,
		Parameters:         make(map[string]interface{}),
		Preconditions:      make([]Predicate, 0),
		Effects:            make([]StateEffect, 0),
		Cost:               1.0,
		SuccessProbability: 1.0,
		Metadata:           make(map[string]interface{}),
	}
}

// Clone creates a deep copy of the action.
func (a *SimAction) Clone() *SimAction {
	clone := &SimAction{
		ID:                 a.ID,
		Type:               a.Type,
		Name:               a.Name,
		Description:        a.Description,
		Parameters:         make(map[string]interface{}),
		Preconditions:      make([]Predicate, len(a.Preconditions)),
		Effects:            make([]StateEffect, len(a.Effects)),
		Cost:               a.Cost,
		ExpectedDuration:   a.ExpectedDuration,
		SuccessProbability: a.SuccessProbability,
		Metadata:           make(map[string]interface{}),
	}
	for k, v := range a.Parameters {
		clone.Parameters[k] = v
	}
	copy(clone.Preconditions, a.Preconditions)
	copy(clone.Effects, a.Effects)
	for k, v := range a.Metadata {
		clone.Metadata[k] = v
	}
	return clone
}

// IsApplicable checks if action can be applied in given state.
func (a *SimAction) IsApplicable(state *State) bool {
	for _, precond := range a.Preconditions {
		if !precond.Evaluate(state) {
			return false
		}
	}
	return true
}

// Predicate represents a condition on state.
type Predicate struct {
	// Feature to check
	Feature string

	// Operator for comparison
	Operator string // "eq", "ne", "gt", "lt", "gte", "lte", "exists", "not_exists"

	// Value to compare against
	Value interface{}
}

// Evaluate checks if predicate holds for state.
func (p *Predicate) Evaluate(state *State) bool {
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
		return simCompareValues(val, p.Value, p.Operator)
	default:
		return false
	}
}

// simCompareValues compares two values with the given operator.
func simCompareValues(a, b interface{}, op string) bool {
	af, aok := simToFloat64(a)
	bf, bok := simToFloat64(b)
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

// simToFloat64 converts interface to float64.
func simToFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case int32:
		return float64(val), true
	default:
		return 0, false
	}
}

// StateEffect describes how an action changes state.
type StateEffect struct {
	// Feature to modify
	Feature string

	// Operation to perform
	Operation string // "set", "add", "multiply", "remove"

	// Value for the operation
	Value interface{}

	// Probability that this effect occurs (0.0 to 1.0)
	Probability float64
}

// Apply applies the effect to a state.
func (e *StateEffect) Apply(state *State) {
	switch e.Operation {
	case "set":
		state.Features[e.Feature] = e.Value
	case "add":
		if current, ok := state.GetFeatureFloat(e.Feature); ok {
			if delta, ok := simToFloat64(e.Value); ok {
				state.Features[e.Feature] = current + delta
			}
		}
	case "multiply":
		if current, ok := state.GetFeatureFloat(e.Feature); ok {
			if factor, ok := simToFloat64(e.Value); ok {
				state.Features[e.Feature] = current * factor
			}
		}
	case "remove":
		delete(state.Features, e.Feature)
	}
}

// ============================================================================
// Trajectory
// ============================================================================

// trajectoryIDCounter provides unique IDs
var trajectoryIDCounter uint64

// Trajectory represents a sequence of states and actions.
type Trajectory struct {
	// ID uniquely identifies this trajectory
	ID string

	// States in order
	States []*State

	// Actions taken between states
	Actions []*SimAction

	// EstimatedSuccess probability (0.0 to 1.0)
	EstimatedSuccess float64

	// TotalCost accumulated
	TotalCost float64

	// TotalDuration estimated
	TotalDuration time.Duration

	// Confidence in this trajectory
	Confidence float64

	// CreatedAt timestamp
	CreatedAt time.Time

	// Metadata holds additional properties
	Metadata map[string]interface{}
}

// NewTrajectory creates a new trajectory starting from a state.
func NewTrajectory(initialState *State) *Trajectory {
	return &Trajectory{
		ID:               fmt.Sprintf("traj-%d", atomic.AddUint64(&trajectoryIDCounter, 1)),
		States:           []*State{initialState.Clone()},
		Actions:          make([]*SimAction, 0),
		EstimatedSuccess: 1.0,
		Confidence:       1.0,
		CreatedAt:        time.Now(),
		Metadata:         make(map[string]interface{}),
	}
}

// Clone creates a deep copy of the trajectory.
func (t *Trajectory) Clone() *Trajectory {
	clone := &Trajectory{
		ID:               t.ID,
		States:           make([]*State, len(t.States)),
		Actions:          make([]*SimAction, len(t.Actions)),
		EstimatedSuccess: t.EstimatedSuccess,
		TotalCost:        t.TotalCost,
		TotalDuration:    t.TotalDuration,
		Confidence:       t.Confidence,
		CreatedAt:        t.CreatedAt,
		Metadata:         make(map[string]interface{}),
	}
	for i, s := range t.States {
		clone.States[i] = s.Clone()
	}
	for i, a := range t.Actions {
		clone.Actions[i] = a.Clone()
	}
	for k, v := range t.Metadata {
		clone.Metadata[k] = v
	}
	return clone
}

// AddStep adds a state-action pair to the trajectory.
func (t *Trajectory) AddStep(action *SimAction, resultState *State) {
	t.Actions = append(t.Actions, action)
	t.States = append(t.States, resultState)
	t.TotalCost += action.Cost
	t.TotalDuration += action.ExpectedDuration
	t.EstimatedSuccess *= action.SuccessProbability
	t.Confidence *= resultState.Confidence
}

// CurrentState returns the last state in the trajectory.
func (t *Trajectory) CurrentState() *State {
	if len(t.States) == 0 {
		return nil
	}
	return t.States[len(t.States)-1]
}

// Length returns the number of steps (actions) in the trajectory.
func (t *Trajectory) Length() int {
	return len(t.Actions)
}

// IsComplete checks if trajectory reaches a terminal state.
func (t *Trajectory) IsComplete() bool {
	current := t.CurrentState()
	return current != nil && current.Type.IsTerminal()
}

// IsSuccessful checks if trajectory reaches a goal state.
func (t *Trajectory) IsSuccessful() bool {
	current := t.CurrentState()
	return current != nil && current.Type == StateGoal
}

// ============================================================================
// State Predictor
// ============================================================================

// StatePredictor predicts next state given current state and action.
type StatePredictor struct {
	mu sync.RWMutex

	// transitionRules map state type + action to effects
	transitionRules map[string][]StateEffect

	// stateHistory for learning from past predictions
	stateHistory []*StatePrediction

	// config
	config *PredictorConfig
}

// PredictorConfig configures the state predictor.
type PredictorConfig struct {
	// MaxHistory limits stored predictions
	MaxHistory int

	// DefaultConfidence for predictions without history
	DefaultConfidence float64

	// LearningRate for updating predictions
	LearningRate float64
}

// DefaultPredictorConfig returns sensible defaults.
func DefaultPredictorConfig() *PredictorConfig {
	return &PredictorConfig{
		MaxHistory:        1000,
		DefaultConfidence: 0.7,
		LearningRate:      0.1,
	}
}

// StatePrediction records a prediction and its outcome.
type StatePrediction struct {
	FromState  *State
	Action     *SimAction
	Predicted  *State
	Actual     *State // nil until observed
	Confidence float64
	Timestamp  time.Time
}

// NewStatePredictor creates a new state predictor.
func NewStatePredictor(config *PredictorConfig) *StatePredictor {
	if config == nil {
		config = DefaultPredictorConfig()
	}
	return &StatePredictor{
		transitionRules: make(map[string][]StateEffect),
		stateHistory:    make([]*StatePrediction, 0),
		config:          config,
	}
}

// AddTransitionRule adds a rule for state transitions.
func (sp *StatePredictor) AddTransitionRule(stateType StateType, actionType SimActionType, effects []StateEffect) {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	key := fmt.Sprintf("%s:%s", stateType.String(), actionType.String())
	sp.transitionRules[key] = effects
}

// Predict predicts the next state given current state and action.
func (sp *StatePredictor) Predict(currentState *State, action *SimAction) *State {
	sp.mu.RLock()
	defer sp.mu.RUnlock()

	if currentState == nil || action == nil {
		return nil
	}

	// Create new state as copy of current
	nextState := currentState.Clone()
	nextState.ID = fmt.Sprintf("state-%d", atomic.AddUint64(&stateIDCounter, 1))
	nextState.ParentID = currentState.ID
	nextState.ActionID = action.ID
	nextState.CreatedAt = time.Now()

	// Apply action effects
	for _, effect := range action.Effects {
		// Check probability of effect occurring
		if effect.Probability > 0 && effect.Probability < 1.0 {
			// For simulation, we apply deterministically but reduce confidence
			nextState.Confidence *= effect.Probability
		}
		effect.Apply(nextState)
	}

	// Look for transition rules
	key := fmt.Sprintf("%s:%s", currentState.Type.String(), action.Type.String())
	if rules, exists := sp.transitionRules[key]; exists {
		for _, effect := range rules {
			effect.Apply(nextState)
		}
	}

	// Determine state type based on features
	nextState.Type = sp.inferStateType(nextState)

	// Apply default confidence
	if nextState.Confidence == currentState.Confidence {
		nextState.Confidence *= sp.config.DefaultConfidence
	}

	return nextState
}

// inferStateType determines state type from features.
func (sp *StatePredictor) inferStateType(state *State) StateType {
	// Check for goal indicators
	if complete, ok := state.Features["goal_achieved"]; ok {
		if b, isBool := complete.(bool); isBool && b {
			return StateGoal
		}
	}

	// Check for failure indicators
	if failed, ok := state.Features["failed"]; ok {
		if b, isBool := failed.(bool); isBool && b {
			return StateFailure
		}
	}

	// Check completion percentage
	if progress, ok := state.GetFeatureFloat("progress"); ok {
		if progress >= 1.0 {
			return StateGoal
		}
		if progress < 0 {
			return StateFailure
		}
	}

	return StateIntermediate
}

// RecordPrediction stores a prediction for learning.
func (sp *StatePredictor) RecordPrediction(from *State, action *SimAction, predicted *State) {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	pred := &StatePrediction{
		FromState:  from.Clone(),
		Action:     action.Clone(),
		Predicted:  predicted.Clone(),
		Confidence: predicted.Confidence,
		Timestamp:  time.Now(),
	}

	sp.stateHistory = append(sp.stateHistory, pred)
	if len(sp.stateHistory) > sp.config.MaxHistory {
		sp.stateHistory = sp.stateHistory[1:]
	}
}

// UpdateWithActual updates a prediction with the actual observed state.
func (sp *StatePredictor) UpdateWithActual(predictionIndex int, actual *State) {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	if predictionIndex < 0 || predictionIndex >= len(sp.stateHistory) {
		return
	}

	pred := sp.stateHistory[predictionIndex]
	pred.Actual = actual.Clone()

	// Could use this to improve predictions over time
	// (future enhancement: adjust confidence based on accuracy)
}

// ============================================================================
// Outcome Estimator
// ============================================================================

// OutcomeEstimator estimates success probability for trajectories.
type OutcomeEstimator struct {
	mu sync.RWMutex

	// goalPredicates define what constitutes success
	goalPredicates []Predicate

	// failurePredicates define what constitutes failure
	failurePredicates []Predicate

	// historicalOutcomes for learning
	historicalOutcomes []*OutcomeRecord

	// config
	config *EstimatorConfig
}

// EstimatorConfig configures the outcome estimator.
type EstimatorConfig struct {
	// MaxHistory limits stored outcomes
	MaxHistory int

	// BaseSuccessProbability when no history
	BaseSuccessProbability float64

	// ConfidenceDecayPerStep reduces confidence with longer trajectories
	ConfidenceDecayPerStep float64
}

// DefaultEstimatorConfig returns sensible defaults.
func DefaultEstimatorConfig() *EstimatorConfig {
	return &EstimatorConfig{
		MaxHistory:             1000,
		BaseSuccessProbability: 0.5,
		ConfidenceDecayPerStep: 0.95,
	}
}

// OutcomeRecord stores outcome history.
type OutcomeRecord struct {
	TrajectoryID  string
	FinalState    *State
	Success       bool
	Confidence    float64
	TrajectoryLen int
	TotalCost     float64
	Timestamp     time.Time
}

// NewOutcomeEstimator creates a new outcome estimator.
func NewOutcomeEstimator(config *EstimatorConfig) *OutcomeEstimator {
	if config == nil {
		config = DefaultEstimatorConfig()
	}
	return &OutcomeEstimator{
		goalPredicates:     make([]Predicate, 0),
		failurePredicates:  make([]Predicate, 0),
		historicalOutcomes: make([]*OutcomeRecord, 0),
		config:             config,
	}
}

// AddGoalPredicate adds a condition that indicates success.
func (oe *OutcomeEstimator) AddGoalPredicate(pred Predicate) {
	oe.mu.Lock()
	defer oe.mu.Unlock()
	oe.goalPredicates = append(oe.goalPredicates, pred)
}

// AddFailurePredicate adds a condition that indicates failure.
func (oe *OutcomeEstimator) AddFailurePredicate(pred Predicate) {
	oe.mu.Lock()
	defer oe.mu.Unlock()
	oe.failurePredicates = append(oe.failurePredicates, pred)
}

// Estimate computes success probability for a trajectory.
func (oe *OutcomeEstimator) Estimate(trajectory *Trajectory) float64 {
	oe.mu.RLock()
	defer oe.mu.RUnlock()

	if trajectory == nil || len(trajectory.States) == 0 {
		return 0
	}

	finalState := trajectory.CurrentState()
	if finalState == nil {
		return 0
	}

	// Check for explicit terminal states
	if finalState.Type == StateGoal {
		return trajectory.Confidence
	}
	if finalState.Type == StateFailure {
		return 0
	}

	// Check goal predicates
	goalsMet := 0
	for _, pred := range oe.goalPredicates {
		if pred.Evaluate(finalState) {
			goalsMet++
		}
	}

	// Check failure predicates
	for _, pred := range oe.failurePredicates {
		if pred.Evaluate(finalState) {
			return 0 // Any failure predicate triggers failure
		}
	}

	// Compute base probability
	var probability float64
	if len(oe.goalPredicates) > 0 {
		probability = float64(goalsMet) / float64(len(oe.goalPredicates))
	} else {
		probability = oe.config.BaseSuccessProbability
	}

	// Apply confidence decay for trajectory length
	for i := 0; i < trajectory.Length(); i++ {
		probability *= oe.config.ConfidenceDecayPerStep
	}

	// Factor in trajectory confidence and accumulated success probability
	probability *= trajectory.Confidence
	probability *= trajectory.EstimatedSuccess

	return math.Max(0, math.Min(1, probability))
}

// IsTerminal checks if state is a terminal state.
func (oe *OutcomeEstimator) IsTerminal(state *State) bool {
	if state.Type.IsTerminal() {
		return true
	}

	oe.mu.RLock()
	defer oe.mu.RUnlock()

	// Check if all goal predicates are met
	if len(oe.goalPredicates) > 0 {
		allMet := true
		for _, pred := range oe.goalPredicates {
			if !pred.Evaluate(state) {
				allMet = false
				break
			}
		}
		if allMet {
			return true
		}
	}

	// Check if any failure predicate is met
	for _, pred := range oe.failurePredicates {
		if pred.Evaluate(state) {
			return true
		}
	}

	return false
}

// RecordOutcome stores an outcome for learning.
func (oe *OutcomeEstimator) RecordOutcome(trajectory *Trajectory, success bool) {
	oe.mu.Lock()
	defer oe.mu.Unlock()

	record := &OutcomeRecord{
		TrajectoryID:  trajectory.ID,
		FinalState:    trajectory.CurrentState().Clone(),
		Success:       success,
		Confidence:    trajectory.Confidence,
		TrajectoryLen: trajectory.Length(),
		TotalCost:     trajectory.TotalCost,
		Timestamp:     time.Now(),
	}

	oe.historicalOutcomes = append(oe.historicalOutcomes, record)
	if len(oe.historicalOutcomes) > oe.config.MaxHistory {
		oe.historicalOutcomes = oe.historicalOutcomes[1:]
	}
}

// GetSuccessRate returns historical success rate.
func (oe *OutcomeEstimator) GetSuccessRate() float64 {
	oe.mu.RLock()
	defer oe.mu.RUnlock()

	if len(oe.historicalOutcomes) == 0 {
		return oe.config.BaseSuccessProbability
	}

	successes := 0
	for _, record := range oe.historicalOutcomes {
		if record.Success {
			successes++
		}
	}

	return float64(successes) / float64(len(oe.historicalOutcomes))
}

// ============================================================================
// World Model
// ============================================================================

// WorldModel enables mental simulation of actions and outcomes.
type WorldModel struct {
	mu sync.RWMutex

	// statePredictor predicts state transitions
	statePredictor *StatePredictor

	// outcomeEstimator estimates success probabilities
	outcomeEstimator *OutcomeEstimator

	// availableActions that can be simulated
	availableActions map[string]*SimAction

	// config
	config *WorldModelConfig

	// stats
	stats *WorldModelStats
}

// WorldModelConfig configures the world model.
type WorldModelConfig struct {
	// MaxSimulationDepth limits look-ahead
	MaxSimulationDepth int

	// MaxBranchingFactor limits alternatives per step
	MaxBranchingFactor int

	// MinConfidenceThreshold stops exploration below this
	MinConfidenceThreshold float64

	// EnablePruning removes unlikely branches
	EnablePruning bool

	// PruningThreshold for branch removal
	PruningThreshold float64
}

// DefaultWorldModelConfig returns sensible defaults.
func DefaultWorldModelConfig() *WorldModelConfig {
	return &WorldModelConfig{
		MaxSimulationDepth:     10,
		MaxBranchingFactor:     5,
		MinConfidenceThreshold: 0.1,
		EnablePruning:          true,
		PruningThreshold:       0.2,
	}
}

// WorldModelStats tracks simulation statistics.
type WorldModelStats struct {
	TotalSimulations       int64
	TotalStepsSimulated    int64
	TotalTrajectories      int64
	SuccessfulTrajectories int64
	AverageDepth           float64
	AverageSuccess         float64
}

// NewWorldModel creates a new world model.
func NewWorldModel(config *WorldModelConfig) *WorldModel {
	if config == nil {
		config = DefaultWorldModelConfig()
	}
	return &WorldModel{
		statePredictor:   NewStatePredictor(nil),
		outcomeEstimator: NewOutcomeEstimator(nil),
		availableActions: make(map[string]*SimAction),
		config:           config,
		stats:            &WorldModelStats{},
	}
}

// SetStatePredictor sets a custom state predictor.
func (wm *WorldModel) SetStatePredictor(sp *StatePredictor) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.statePredictor = sp
}

// SetOutcomeEstimator sets a custom outcome estimator.
func (wm *WorldModel) SetOutcomeEstimator(oe *OutcomeEstimator) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.outcomeEstimator = oe
}

// AddAction registers an action for simulation.
func (wm *WorldModel) AddAction(action *SimAction) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.availableActions[action.ID] = action
}

// RemoveAction removes an action from simulation.
func (wm *WorldModel) RemoveAction(actionID string) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	delete(wm.availableActions, actionID)
}

// GetAction retrieves an action by ID.
func (wm *WorldModel) GetAction(actionID string) (*SimAction, bool) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	action, exists := wm.availableActions[actionID]
	if !exists {
		return nil, false
	}
	return action.Clone(), true
}

// GetApplicableActions returns actions applicable in the given state.
func (wm *WorldModel) GetApplicableActions(state *State) []*SimAction {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	applicable := make([]*SimAction, 0)
	for _, action := range wm.availableActions {
		if action.IsApplicable(state) {
			applicable = append(applicable, action.Clone())
		}
	}

	// Sort by success probability descending
	sort.Slice(applicable, func(i, j int) bool {
		return applicable[i].SuccessProbability > applicable[j].SuccessProbability
	})

	// Limit to max branching factor
	if len(applicable) > wm.config.MaxBranchingFactor {
		applicable = applicable[:wm.config.MaxBranchingFactor]
	}

	return applicable
}

// ============================================================================
// Simulation Methods
// ============================================================================

// SimulateAction simulates a single action from current state.
func (wm *WorldModel) SimulateAction(currentState *State, action *SimAction) (*Trajectory, error) {
	if currentState == nil {
		return nil, ErrInvalidState
	}
	if action == nil {
		return nil, ErrInvalidAction
	}

	wm.mu.Lock()
	wm.stats.TotalSimulations++
	wm.stats.TotalStepsSimulated++
	wm.mu.Unlock()

	trajectory := NewTrajectory(currentState)

	// Predict next state
	nextState := wm.statePredictor.Predict(currentState, action)
	if nextState == nil {
		return nil, ErrSimulationFailed
	}

	trajectory.AddStep(action, nextState)
	trajectory.EstimatedSuccess = wm.outcomeEstimator.Estimate(trajectory)

	wm.mu.Lock()
	wm.stats.TotalTrajectories++
	if trajectory.IsSuccessful() {
		wm.stats.SuccessfulTrajectories++
	}
	wm.mu.Unlock()

	return trajectory, nil
}

// SimulateSequence simulates a sequence of actions.
func (wm *WorldModel) SimulateSequence(currentState *State, actions []*SimAction) (*Trajectory, error) {
	if currentState == nil {
		return nil, ErrInvalidState
	}
	if len(actions) == 0 {
		return nil, ErrInvalidAction
	}
	if len(actions) > wm.config.MaxSimulationDepth {
		return nil, ErrMaxDepthExceeded
	}

	wm.mu.Lock()
	wm.stats.TotalSimulations++
	wm.mu.Unlock()

	trajectory := NewTrajectory(currentState)
	state := currentState

	for _, action := range actions {
		// Check if action is applicable
		if !action.IsApplicable(state) {
			// Stop at first inapplicable action
			break
		}

		// Predict next state
		nextState := wm.statePredictor.Predict(state, action)
		if nextState == nil {
			break
		}

		trajectory.AddStep(action, nextState)
		state = nextState

		wm.mu.Lock()
		wm.stats.TotalStepsSimulated++
		wm.mu.Unlock()

		// Check for terminal state
		if wm.outcomeEstimator.IsTerminal(state) {
			break
		}

		// Check confidence threshold
		if trajectory.Confidence < wm.config.MinConfidenceThreshold {
			break
		}
	}

	trajectory.EstimatedSuccess = wm.outcomeEstimator.Estimate(trajectory)

	wm.mu.Lock()
	wm.stats.TotalTrajectories++
	if trajectory.IsSuccessful() {
		wm.stats.SuccessfulTrajectories++
	}
	wm.mu.Unlock()

	return trajectory, nil
}

// SimulateBestPath finds the best trajectory using greedy search.
func (wm *WorldModel) SimulateBestPath(currentState *State, maxDepth int) (*Trajectory, error) {
	if currentState == nil {
		return nil, ErrInvalidState
	}
	if maxDepth <= 0 {
		maxDepth = wm.config.MaxSimulationDepth
	}
	if maxDepth > wm.config.MaxSimulationDepth {
		maxDepth = wm.config.MaxSimulationDepth
	}

	wm.mu.Lock()
	wm.stats.TotalSimulations++
	wm.mu.Unlock()

	trajectory := NewTrajectory(currentState)
	state := currentState

	for depth := 0; depth < maxDepth; depth++ {
		// Get applicable actions
		actions := wm.GetApplicableActions(state)
		if len(actions) == 0 {
			break
		}

		// Find best action (highest expected success)
		var bestAction *SimAction
		var bestScore float64 = -1

		for _, action := range actions {
			nextState := wm.statePredictor.Predict(state, action)
			if nextState == nil {
				continue
			}

			// Score = success probability * confidence
			score := action.SuccessProbability * nextState.Confidence
			if score > bestScore {
				bestScore = score
				bestAction = action
			}
		}

		if bestAction == nil {
			break
		}

		// Apply best action
		nextState := wm.statePredictor.Predict(state, bestAction)
		trajectory.AddStep(bestAction, nextState)
		state = nextState

		wm.mu.Lock()
		wm.stats.TotalStepsSimulated++
		wm.mu.Unlock()

		// Check for terminal state
		if wm.outcomeEstimator.IsTerminal(state) {
			break
		}

		// Check confidence threshold
		if trajectory.Confidence < wm.config.MinConfidenceThreshold {
			break
		}
	}

	trajectory.EstimatedSuccess = wm.outcomeEstimator.Estimate(trajectory)

	wm.mu.Lock()
	wm.stats.TotalTrajectories++
	wm.stats.AverageDepth = (wm.stats.AverageDepth*float64(wm.stats.TotalTrajectories-1) + float64(trajectory.Length())) / float64(wm.stats.TotalTrajectories)
	if trajectory.IsSuccessful() {
		wm.stats.SuccessfulTrajectories++
	}
	wm.mu.Unlock()

	return trajectory, nil
}

// CompareActions simulates multiple actions and returns them ranked by expected outcome.
func (wm *WorldModel) CompareActions(currentState *State, actions []*SimAction) ([]*ActionComparison, error) {
	if currentState == nil {
		return nil, ErrInvalidState
	}
	if len(actions) == 0 {
		return nil, ErrInvalidAction
	}

	comparisons := make([]*ActionComparison, 0, len(actions))

	for _, action := range actions {
		traj, err := wm.SimulateAction(currentState, action)
		if err != nil {
			continue
		}

		comparisons = append(comparisons, &ActionComparison{
			Action:           action,
			Trajectory:       traj,
			ExpectedSuccess:  traj.EstimatedSuccess,
			ExpectedCost:     traj.TotalCost,
			ExpectedDuration: traj.TotalDuration,
			Confidence:       traj.Confidence,
		})
	}

	// Sort by expected success descending
	sort.Slice(comparisons, func(i, j int) bool {
		return comparisons[i].ExpectedSuccess > comparisons[j].ExpectedSuccess
	})

	return comparisons, nil
}

// ActionComparison holds comparison results for an action.
type ActionComparison struct {
	Action           *SimAction
	Trajectory       *Trajectory
	ExpectedSuccess  float64
	ExpectedCost     float64
	ExpectedDuration time.Duration
	Confidence       float64
}

// ExploreAlternatives explores multiple trajectories from current state.
func (wm *WorldModel) ExploreAlternatives(currentState *State, depth int) ([]*Trajectory, error) {
	if currentState == nil {
		return nil, ErrInvalidState
	}
	if depth <= 0 {
		depth = 3
	}
	if depth > wm.config.MaxSimulationDepth {
		depth = wm.config.MaxSimulationDepth
	}

	trajectories := make([]*Trajectory, 0)

	// Get initial actions
	actions := wm.GetApplicableActions(currentState)
	if len(actions) == 0 {
		return trajectories, nil
	}

	// Explore each action
	for _, action := range actions {
		traj, err := wm.exploreRecursive(currentState, action, depth-1)
		if err != nil {
			continue
		}
		trajectories = append(trajectories, traj)
	}

	// Sort by success probability
	sort.Slice(trajectories, func(i, j int) bool {
		return trajectories[i].EstimatedSuccess > trajectories[j].EstimatedSuccess
	})

	// Prune if enabled
	if wm.config.EnablePruning {
		pruned := make([]*Trajectory, 0)
		for _, traj := range trajectories {
			if traj.EstimatedSuccess >= wm.config.PruningThreshold {
				pruned = append(pruned, traj)
			}
		}
		trajectories = pruned
	}

	return trajectories, nil
}

// exploreRecursive explores trajectories recursively.
func (wm *WorldModel) exploreRecursive(state *State, action *SimAction, remainingDepth int) (*Trajectory, error) {
	traj := NewTrajectory(state)

	// Apply action
	nextState := wm.statePredictor.Predict(state, action)
	if nextState == nil {
		return traj, ErrSimulationFailed
	}

	traj.AddStep(action, nextState)

	wm.mu.Lock()
	wm.stats.TotalStepsSimulated++
	wm.mu.Unlock()

	// Check terminal or depth limit
	if remainingDepth <= 0 || wm.outcomeEstimator.IsTerminal(nextState) {
		traj.EstimatedSuccess = wm.outcomeEstimator.Estimate(traj)
		return traj, nil
	}

	// Continue with best action
	nextActions := wm.GetApplicableActions(nextState)
	if len(nextActions) == 0 {
		traj.EstimatedSuccess = wm.outcomeEstimator.Estimate(traj)
		return traj, nil
	}

	// Find best continuation
	var bestContinuation *Trajectory
	var bestSuccess float64 = -1

	for _, nextAction := range nextActions {
		continuation, err := wm.exploreRecursive(nextState, nextAction, remainingDepth-1)
		if err != nil {
			continue
		}
		if continuation.EstimatedSuccess > bestSuccess {
			bestSuccess = continuation.EstimatedSuccess
			bestContinuation = continuation
		}
	}

	// Merge best continuation
	if bestContinuation != nil && len(bestContinuation.Actions) > 0 {
		for i, a := range bestContinuation.Actions {
			traj.AddStep(a, bestContinuation.States[i+1])
		}
	}

	traj.EstimatedSuccess = wm.outcomeEstimator.Estimate(traj)
	return traj, nil
}

// ============================================================================
// Statistics and Debugging
// ============================================================================

// GetStats returns simulation statistics.
func (wm *WorldModel) GetStats() WorldModelStats {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	stats := *wm.stats
	if stats.TotalTrajectories > 0 {
		stats.AverageSuccess = float64(stats.SuccessfulTrajectories) / float64(stats.TotalTrajectories)
	}
	return stats
}

// Clear resets the world model state.
func (wm *WorldModel) Clear() {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	wm.availableActions = make(map[string]*SimAction)
	wm.stats = &WorldModelStats{}
}

// ActionCount returns the number of registered actions.
func (wm *WorldModel) ActionCount() int {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	return len(wm.availableActions)
}
