// Package memory provides the MNEMONIC system for the Elite Agent Collective.
// This file implements the Production System from @NEURAL's Cognitive Architecture Analysis.
//
// Production System (SOAR-inspired):
// - Production Rules: IF-THEN rules matching working memory patterns
// - RETE-inspired Matching: Efficient incremental pattern matching
// - Conflict Resolution: Select among matching rules (specificity, recency, priority)
// - Rule Firing: Execute actions and update working memory
// - Chunking: Learn new productions from successful problem-solving sequences
//
// This forms the procedural knowledge component of the cognitive architecture.

package memory

import (
	"errors"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// productionIDCounter provides unique IDs for productions
var productionIDCounter uint64

// chunkIDCounter provides unique IDs for chunks
var chunkIDCounter uint64

// ============================================================================
// Error Definitions
// ============================================================================

var (
	ErrProductionNotFound    = errors.New("production not found")
	ErrInvalidCondition      = errors.New("invalid condition")
	ErrNoMatchingProductions = errors.New("no matching productions")
	ErrProductionDisabled    = errors.New("production is disabled")
)

// ============================================================================
// Condition Types
// ============================================================================

// ConditionType specifies the type of pattern match.
type ConditionType int

const (
	// ConditionEquals matches exact value
	ConditionEquals ConditionType = iota
	// ConditionNotEquals matches if value differs
	ConditionNotEquals
	// ConditionGreaterThan matches if value > threshold
	ConditionGreaterThan
	// ConditionLessThan matches if value < threshold
	ConditionLessThan
	// ConditionContains matches if element contains substring
	ConditionContains
	// ConditionExists matches if element exists
	ConditionExists
	// ConditionNotExists matches if element does not exist
	ConditionNotExists
	// ConditionRegex matches using regex pattern
	ConditionRegex
	// ConditionInRange matches if value is within range
	ConditionInRange
	// ConditionTypeMatch matches by element type
	ConditionTypeMatch
)

// String returns human-readable condition type.
func (ct ConditionType) String() string {
	switch ct {
	case ConditionEquals:
		return "EQUALS"
	case ConditionNotEquals:
		return "NOT_EQUALS"
	case ConditionGreaterThan:
		return "GREATER_THAN"
	case ConditionLessThan:
		return "LESS_THAN"
	case ConditionContains:
		return "CONTAINS"
	case ConditionExists:
		return "EXISTS"
	case ConditionNotExists:
		return "NOT_EXISTS"
	case ConditionRegex:
		return "REGEX"
	case ConditionInRange:
		return "IN_RANGE"
	case ConditionTypeMatch:
		return "TYPE_MATCH"
	default:
		return "UNKNOWN"
	}
}

// ============================================================================
// Condition
// ============================================================================

// Condition represents a single pattern match requirement.
type Condition struct {
	// Type of condition
	Type ConditionType

	// Attribute to match (e.g., "type", "content", "activation")
	Attribute string

	// Value to compare against
	Value interface{}

	// SecondValue for range conditions (min, max)
	SecondValue interface{}

	// Negated inverts the match result
	Negated bool

	// BindVariable captures matched value for use in actions
	BindVariable string
}

// Match tests if a working memory item matches this condition.
func (c *Condition) Match(item *WorkingMemoryItem) bool {
	if item == nil {
		return c.Type == ConditionNotExists
	}

	var attrValue interface{}
	switch c.Attribute {
	case "id":
		attrValue = item.ID
	case "type":
		attrValue = string(item.ContentType)
	case "content":
		attrValue = item.Content
	case "activation":
		attrValue = item.Activation
	case "source":
		attrValue = string(item.Source)
	case "chunk_id":
		attrValue = item.ChunkID
	case "salience":
		attrValue = item.Salience
	default:
		// Check metadata
		if item.Metadata != nil {
			attrValue = item.Metadata[c.Attribute]
		}
	}

	result := c.matchValue(attrValue)

	if c.Negated {
		return !result
	}
	return result
}

// matchValue performs the actual comparison.
func (c *Condition) matchValue(attrValue interface{}) bool {
	switch c.Type {
	case ConditionExists:
		return attrValue != nil

	case ConditionNotExists:
		return attrValue == nil

	case ConditionEquals:
		return fmt.Sprintf("%v", attrValue) == fmt.Sprintf("%v", c.Value)

	case ConditionNotEquals:
		return fmt.Sprintf("%v", attrValue) != fmt.Sprintf("%v", c.Value)

	case ConditionContains:
		strVal := fmt.Sprintf("%v", attrValue)
		pattern := fmt.Sprintf("%v", c.Value)
		return len(strVal) > 0 && len(pattern) > 0 &&
			(strVal == pattern || len(strVal) > len(pattern) &&
				(strVal[:len(pattern)] == pattern || strVal[len(strVal)-len(pattern):] == pattern ||
					findSubstring(strVal, pattern)))

	case ConditionGreaterThan:
		return compareNumeric(attrValue, c.Value) > 0

	case ConditionLessThan:
		return compareNumeric(attrValue, c.Value) < 0

	case ConditionInRange:
		val := toFloat64(attrValue)
		min := toFloat64(c.Value)
		max := toFloat64(c.SecondValue)
		return val >= min && val <= max

	case ConditionTypeMatch:
		return fmt.Sprintf("%v", attrValue) == fmt.Sprintf("%v", c.Value)

	default:
		return false
	}
}

// findSubstring checks if pattern exists in str.
func findSubstring(str, pattern string) bool {
	for i := 0; i <= len(str)-len(pattern); i++ {
		if str[i:i+len(pattern)] == pattern {
			return true
		}
	}
	return false
}

// compareNumeric compares two values numerically.
func compareNumeric(a, b interface{}) int {
	aVal := toFloat64(a)
	bVal := toFloat64(b)
	if aVal > bVal {
		return 1
	} else if aVal < bVal {
		return -1
	}
	return 0
}

// toFloat64 converts interface to float64.
func toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case int32:
		return float64(val)
	default:
		return 0
	}
}

// ============================================================================
// Action Types
// ============================================================================

// ActionType specifies what action to take when production fires.
type ActionType int

const (
	// ActionAdd adds element to working memory
	ActionAdd ActionType = iota
	// ActionRemove removes element from working memory
	ActionRemove
	// ActionModify modifies existing element
	ActionModify
	// ActionPushGoal pushes a new goal
	ActionPushGoal
	// ActionCompleteGoal marks goal complete
	ActionCompleteGoal
	// ActionInvokeAgent invokes a specialist agent
	ActionInvokeAgent
	// ActionEmit emits an event/signal
	ActionEmit
	// ActionLog logs a message
	ActionLog
	// ActionHalt stops production execution
	ActionHalt
)

// String returns human-readable action type.
func (at ActionType) String() string {
	switch at {
	case ActionAdd:
		return "ADD"
	case ActionRemove:
		return "REMOVE"
	case ActionModify:
		return "MODIFY"
	case ActionPushGoal:
		return "PUSH_GOAL"
	case ActionCompleteGoal:
		return "COMPLETE_GOAL"
	case ActionInvokeAgent:
		return "INVOKE_AGENT"
	case ActionEmit:
		return "EMIT"
	case ActionLog:
		return "LOG"
	case ActionHalt:
		return "HALT"
	default:
		return "UNKNOWN"
	}
}

// ============================================================================
// Action
// ============================================================================

// Action represents an action to take when a production fires.
type Action struct {
	// Type of action
	Type ActionType

	// Target element ID (for REMOVE, MODIFY)
	TargetID string

	// Attribute to modify
	Attribute string

	// Value to set or add
	Value interface{}

	// AgentID for INVOKE_AGENT
	AgentID string

	// GoalName for PUSH_GOAL
	GoalName string

	// Priority for new goals
	Priority GoalPriority

	// Message for LOG actions
	Message string

	// UseBinding references a bound variable from conditions
	UseBinding string

	// Metadata for extensions
	Metadata map[string]interface{}
}

// ============================================================================
// Production
// ============================================================================

// Production represents an IF-THEN rule.
type Production struct {
	// ID uniquely identifies this production
	ID string

	// Name is human-readable production name
	Name string

	// Description explains what the production does
	Description string

	// Conditions are the IF part (all must match)
	Conditions []*Condition

	// Actions are the THEN part (all execute on fire)
	Actions []*Action

	// Priority for conflict resolution (higher = more preferred)
	Priority float64

	// Specificity is computed from condition count
	Specificity int

	// Enabled allows disabling without removing
	Enabled bool

	// CreatedAt is when production was created
	CreatedAt time.Time

	// LastFiredAt is when production last fired
	LastFiredAt *time.Time

	// FireCount is how many times production has fired
	FireCount int64

	// SuccessCount is how many times firing led to goal progress
	SuccessCount int64

	// Source indicates where production came from
	Source string // "builtin", "learned", "user"

	// Tags for categorization
	Tags []string

	// Metadata for extensions
	Metadata map[string]interface{}
}

// Utility calculates the utility score (success rate * priority).
func (p *Production) Utility() float64 {
	if p.FireCount == 0 {
		return p.Priority
	}
	successRate := float64(p.SuccessCount) / float64(p.FireCount)
	return successRate * p.Priority
}

// ============================================================================
// Match Result
// ============================================================================

// MatchResult holds information about a production match.
type MatchResult struct {
	// Production that matched
	Production *Production

	// MatchedItems are the WM items that satisfied conditions
	MatchedItems []*WorkingMemoryItem

	// Bindings map variable names to matched values
	Bindings map[string]interface{}

	// Score for conflict resolution
	Score float64

	// MatchTime is when match was computed
	MatchTime time.Time
}

// ============================================================================
// Conflict Resolution Strategy
// ============================================================================

// ProductionConflictStrategy defines how to select among matching productions.
type ProductionConflictStrategy int

const (
	// ConflictStrategySpecificity prefers more specific rules (more conditions)
	ConflictStrategySpecificity ProductionConflictStrategy = iota
	// ConflictStrategyRecency prefers recently fired rules
	ConflictStrategyRecency
	// ConflictStrategyPriority prefers higher priority rules
	ConflictStrategyPriority
	// ConflictStrategyUtility prefers rules with higher utility score
	ConflictStrategyUtility
	// ConflictStrategyRandom randomly selects among matches
	ConflictStrategyRandom
	// ConflictStrategyRefraction prevents same rule firing twice consecutively
	ConflictStrategyRefraction
)

// ============================================================================
// Production System
// ============================================================================

// ProductionSystem manages productions and their execution.
type ProductionSystem struct {
	mu sync.RWMutex

	// productions maps ID to production
	productions map[string]*Production

	// productionsByTag indexes productions by tag
	productionsByTag map[string][]*Production

	// workingMemory reference
	workingMemory *CognitiveWorkingMemory

	// goalStack reference
	goalStack *GoalStack

	// impasseDetector reference
	impasseDetector *ImpasseDetector

	// config holds system configuration
	config *ProductionSystemConfig

	// conflictSet holds currently matching productions
	conflictSet []*MatchResult

	// lastFired tracks last fired production (for refraction)
	lastFired string

	// refractionSet tracks production+items combinations that have fired
	// Key is "productionID:itemID1,itemID2,..."
	refractionSet map[string]bool

	// firingHistory tracks recent firings
	firingHistory []*FiringRecord

	// stats tracks execution statistics
	stats *ProductionStats

	// callbacks
	onProductionFired func(*Production, *MatchResult)
	onConflict        func([]*MatchResult)
	onLearned         func(*Production)
}

// ProductionSystemConfig configures the production system.
type ProductionSystemConfig struct {
	// MaxProductions limits total productions
	MaxProductions int

	// MaxHistorySize limits firing history
	MaxHistorySize int

	// ConflictStrategies in order of application
	ConflictStrategies []ProductionConflictStrategy

	// EnableRefraction prevents same production firing twice in a row
	EnableRefraction bool

	// EnableLearning allows chunking new productions
	EnableLearning bool

	// ChunkingThreshold is minimum success rate to create chunk
	ChunkingThreshold float64

	// MinChunkSequence is minimum sequence length for chunking
	MinChunkSequence int
}

// DefaultProductionSystemConfig returns sensible defaults.
func DefaultProductionSystemConfig() *ProductionSystemConfig {
	return &ProductionSystemConfig{
		MaxProductions:     1000,
		MaxHistorySize:     100,
		ConflictStrategies: []ProductionConflictStrategy{ConflictStrategySpecificity, ConflictStrategyUtility, ConflictStrategyRecency},
		EnableRefraction:   true,
		EnableLearning:     true,
		ChunkingThreshold:  0.8,
		MinChunkSequence:   3,
	}
}

// ProductionStats tracks system statistics.
type ProductionStats struct {
	TotalProductions   int
	TotalFirings       int64
	SuccessfulFirings  int64
	ConflictsResolved  int64
	ProductionsLearned int64
	AverageCycleTime   time.Duration
}

// FiringRecord records a production firing.
type FiringRecord struct {
	ProductionID string
	FiredAt      time.Time
	Success      bool
	GoalID       string
	Bindings     map[string]interface{}
}

// NewProductionSystem creates a new production system.
func NewProductionSystem(config *ProductionSystemConfig, wm *CognitiveWorkingMemory, gs *GoalStack, id *ImpasseDetector) *ProductionSystem {
	if config == nil {
		config = DefaultProductionSystemConfig()
	}

	return &ProductionSystem{
		productions:      make(map[string]*Production),
		productionsByTag: make(map[string][]*Production),
		workingMemory:    wm,
		goalStack:        gs,
		impasseDetector:  id,
		config:           config,
		conflictSet:      make([]*MatchResult, 0),
		refractionSet:    make(map[string]bool),
		firingHistory:    make([]*FiringRecord, 0),
		stats:            &ProductionStats{},
	}
}

// ============================================================================
// Production Management
// ============================================================================

// AddProduction adds a new production to the system.
func (ps *ProductionSystem) AddProduction(prod *Production) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if len(ps.productions) >= ps.config.MaxProductions {
		return errors.New("maximum productions reached")
	}

	if prod.ID == "" {
		prod.ID = fmt.Sprintf("prod-%d", atomic.AddUint64(&productionIDCounter, 1))
	}

	prod.Specificity = len(prod.Conditions)
	if prod.CreatedAt.IsZero() {
		prod.CreatedAt = time.Now()
	}
	if prod.Metadata == nil {
		prod.Metadata = make(map[string]interface{})
	}
	prod.Enabled = true

	ps.productions[prod.ID] = prod
	ps.stats.TotalProductions++

	// Index by tags
	for _, tag := range prod.Tags {
		ps.productionsByTag[tag] = append(ps.productionsByTag[tag], prod)
	}

	return nil
}

// RemoveProduction removes a production.
func (ps *ProductionSystem) RemoveProduction(id string) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	prod, ok := ps.productions[id]
	if !ok {
		return ErrProductionNotFound
	}

	// Remove from tag index
	for _, tag := range prod.Tags {
		prods := ps.productionsByTag[tag]
		for i, p := range prods {
			if p.ID == id {
				ps.productionsByTag[tag] = append(prods[:i], prods[i+1:]...)
				break
			}
		}
	}

	delete(ps.productions, id)
	ps.stats.TotalProductions--

	return nil
}

// GetProduction retrieves a production by ID.
func (ps *ProductionSystem) GetProduction(id string) (*Production, error) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	prod, ok := ps.productions[id]
	if !ok {
		return nil, ErrProductionNotFound
	}
	return prod, nil
}

// GetByTag returns productions with a specific tag.
func (ps *ProductionSystem) GetByTag(tag string) []*Production {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	return ps.productionsByTag[tag]
}

// EnableProduction enables a production.
func (ps *ProductionSystem) EnableProduction(id string) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	prod, ok := ps.productions[id]
	if !ok {
		return ErrProductionNotFound
	}
	prod.Enabled = true
	return nil
}

// DisableProduction disables a production.
func (ps *ProductionSystem) DisableProduction(id string) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	prod, ok := ps.productions[id]
	if !ok {
		return ErrProductionNotFound
	}
	prod.Enabled = false
	return nil
}

// Count returns the number of productions.
func (ps *ProductionSystem) Count() int {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return len(ps.productions)
}

// ============================================================================
// Pattern Matching (RETE-inspired)
// ============================================================================

// Match finds all productions that match current working memory.
func (ps *ProductionSystem) Match() []*MatchResult {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ps.workingMemory == nil {
		return nil
	}

	items := ps.workingMemory.GetAll()
	matches := make([]*MatchResult, 0)

	for _, prod := range ps.productions {
		if !prod.Enabled {
			continue
		}

		// Match all conditions
		matched, bindings, matchedItems := ps.matchProductionWithItems(prod, items)
		if matched {
			// Check refraction - prevent same production+items from firing twice
			if ps.config.EnableRefraction {
				refractionKey := ps.computeRefractionKey(prod.ID, matchedItems)
				if ps.refractionSet[refractionKey] {
					continue
				}
			}

			result := &MatchResult{
				Production:   prod,
				MatchedItems: matchedItems,
				Bindings:     bindings,
				MatchTime:    time.Now(),
			}
			result.Score = ps.calculateScore(result)
			matches = append(matches, result)
		}
	}

	ps.conflictSet = matches
	return matches
}

// matchProduction checks if a production's conditions match.
func (ps *ProductionSystem) matchProduction(prod *Production, items []*WorkingMemoryItem) (bool, map[string]interface{}) {
	bindings := make(map[string]interface{})

	for _, cond := range prod.Conditions {
		matched := false

		if cond.Type == ConditionNotExists {
			// Check no item matches
			exists := false
			for _, item := range items {
				if cond.Match(item) {
					exists = true
					break
				}
			}
			matched = !exists
		} else {
			// Check at least one item matches
			for _, item := range items {
				if cond.Match(item) {
					matched = true
					if cond.BindVariable != "" {
						bindings[cond.BindVariable] = item
					}
					break
				}
			}
		}

		if !matched {
			return false, nil
		}
	}

	return true, bindings
}

// matchProductionWithItems checks if a production's conditions match and returns matched items.
func (ps *ProductionSystem) matchProductionWithItems(prod *Production, items []*WorkingMemoryItem) (bool, map[string]interface{}, []*WorkingMemoryItem) {
	bindings := make(map[string]interface{})
	matchedItems := make([]*WorkingMemoryItem, 0)

	for _, cond := range prod.Conditions {
		matched := false

		if cond.Type == ConditionNotExists {
			// Check no item matches
			exists := false
			for _, item := range items {
				if cond.Match(item) {
					exists = true
					break
				}
			}
			matched = !exists
		} else {
			// Check at least one item matches
			for _, item := range items {
				if cond.Match(item) {
					matched = true
					matchedItems = append(matchedItems, item)
					if cond.BindVariable != "" {
						bindings[cond.BindVariable] = item
					}
					break
				}
			}
		}

		if !matched {
			return false, nil, nil
		}
	}

	return true, bindings, matchedItems
}

// computeRefractionKey creates a unique key for production+items combination.
func (ps *ProductionSystem) computeRefractionKey(prodID string, items []*WorkingMemoryItem) string {
	ids := make([]string, len(items))
	for i, item := range items {
		ids[i] = item.ID
	}
	sort.Strings(ids)
	return fmt.Sprintf("%s:%v", prodID, ids)
}

// calculateScore computes conflict resolution score.
func (ps *ProductionSystem) calculateScore(result *MatchResult) float64 {
	prod := result.Production
	score := 0.0

	// Apply conflict strategies in order
	for i, strategy := range ps.config.ConflictStrategies {
		weight := 1.0 / float64(i+1) // Decreasing weight for later strategies

		switch strategy {
		case ConflictStrategySpecificity:
			score += weight * float64(prod.Specificity)
		case ConflictStrategyPriority:
			score += weight * prod.Priority
		case ConflictStrategyUtility:
			score += weight * prod.Utility()
		case ConflictStrategyRecency:
			if prod.LastFiredAt != nil {
				recency := time.Since(*prod.LastFiredAt).Seconds()
				if recency > 0 {
					score += weight / recency
				}
			}
		}
	}

	return score
}

// ============================================================================
// Conflict Resolution
// ============================================================================

// ResolveConflict selects the best production from the conflict set.
func (ps *ProductionSystem) ResolveConflict() (*MatchResult, error) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if len(ps.conflictSet) == 0 {
		return nil, ErrNoMatchingProductions
	}

	// Sort by score descending
	sort.Slice(ps.conflictSet, func(i, j int) bool {
		return ps.conflictSet[i].Score > ps.conflictSet[j].Score
	})

	// Notify callback
	if ps.onConflict != nil && len(ps.conflictSet) > 1 {
		ps.onConflict(ps.conflictSet)
	}

	ps.stats.ConflictsResolved++
	return ps.conflictSet[0], nil
}

// GetConflictSet returns the current conflict set.
func (ps *ProductionSystem) GetConflictSet() []*MatchResult {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	result := make([]*MatchResult, len(ps.conflictSet))
	copy(result, ps.conflictSet)
	return result
}

// ============================================================================
// Rule Firing
// ============================================================================

// Fire executes the selected production's actions.
func (ps *ProductionSystem) Fire(result *MatchResult) error {
	if result == nil || result.Production == nil {
		return errors.New("invalid match result")
	}

	ps.mu.Lock()
	prod := result.Production

	if !prod.Enabled {
		ps.mu.Unlock()
		return ErrProductionDisabled
	}

	// Update production stats
	now := time.Now()
	prod.LastFiredAt = &now
	prod.FireCount++

	// Record in history
	record := &FiringRecord{
		ProductionID: prod.ID,
		FiredAt:      now,
		Bindings:     result.Bindings,
	}

	if ps.goalStack != nil {
		current := ps.goalStack.Current()
		if current != nil {
			record.GoalID = current.ID
		}
	}

	ps.firingHistory = append(ps.firingHistory, record)
	if len(ps.firingHistory) > ps.config.MaxHistorySize {
		ps.firingHistory = ps.firingHistory[1:]
	}

	ps.lastFired = prod.ID

	// Add to refraction set to prevent same production+items from firing again
	if ps.config.EnableRefraction {
		refractionKey := ps.computeRefractionKey(prod.ID, result.MatchedItems)
		ps.refractionSet[refractionKey] = true
	}

	ps.stats.TotalFirings++

	ps.mu.Unlock()

	// Execute actions (without lock to allow callbacks to access system)
	for _, action := range prod.Actions {
		if err := ps.executeAction(action, result.Bindings); err != nil {
			return err
		}
	}

	// Notify callback
	if ps.onProductionFired != nil {
		ps.onProductionFired(prod, result)
	}

	return nil
}

// executeAction performs a single action.
func (ps *ProductionSystem) executeAction(action *Action, bindings map[string]interface{}) error {
	switch action.Type {
	case ActionAdd:
		if ps.workingMemory != nil {
			item := &WorkingMemoryItem{
				ContentType: WorkingMemoryContentType(fmt.Sprintf("%v", action.Value)),
				Content:     action.Metadata,
				Metadata:    action.Metadata,
			}
			ps.workingMemory.Add(item)
		}

	case ActionRemove:
		if ps.workingMemory != nil {
			ps.workingMemory.Remove(action.TargetID)
		}

	case ActionModify:
		if ps.workingMemory != nil {
			item, exists := ps.workingMemory.Get(action.TargetID)
			if !exists {
				return errors.New("item not found")
			}
			if item.Metadata == nil {
				item.Metadata = make(map[string]interface{})
			}
			item.Metadata[action.Attribute] = action.Value
		}

	case ActionPushGoal:
		if ps.goalStack != nil {
			goal := &Goal{
				Name:     action.GoalName,
				Priority: action.Priority,
			}
			return ps.goalStack.Push(goal)
		}

	case ActionCompleteGoal:
		if ps.goalStack != nil {
			return ps.goalStack.Complete(action.TargetID)
		}

	case ActionInvokeAgent:
		// Emit event for agent invocation (handled externally)
		if ps.onProductionFired != nil {
			// Signaling mechanism through metadata
		}

	case ActionEmit:
		// Emit event (handled by callback)

	case ActionLog:
		// Log message (could use structured logger)

	case ActionHalt:
		// Signal halt (could set flag)
	}

	return nil
}

// ============================================================================
// Recognize-Act Cycle
// ============================================================================

// Cycle performs one recognize-act cycle.
func (ps *ProductionSystem) Cycle() (*MatchResult, error) {
	start := time.Now()

	// Match phase
	matches := ps.Match()
	if len(matches) == 0 {
		return nil, ErrNoMatchingProductions
	}

	// Conflict resolution phase
	selected, err := ps.ResolveConflict()
	if err != nil {
		return nil, err
	}

	// Act phase
	if err := ps.Fire(selected); err != nil {
		return nil, err
	}

	// Update timing stats
	elapsed := time.Since(start)
	ps.mu.Lock()
	if ps.stats.AverageCycleTime == 0 {
		ps.stats.AverageCycleTime = elapsed
	} else {
		ps.stats.AverageCycleTime = (ps.stats.AverageCycleTime + elapsed) / 2
	}
	ps.mu.Unlock()

	return selected, nil
}

// Run executes cycles until no productions match or halt is signaled.
func (ps *ProductionSystem) Run(maxCycles int) (int, error) {
	cycles := 0

	for cycles < maxCycles {
		_, err := ps.Cycle()
		if err != nil {
			if err == ErrNoMatchingProductions {
				return cycles, nil // Normal termination
			}
			return cycles, err
		}
		cycles++
	}

	return cycles, nil
}

// ============================================================================
// Learning (Chunking)
// ============================================================================

// LearnChunk creates a new production from a successful sequence.
func (ps *ProductionSystem) LearnChunk(name string, sequence []*FiringRecord) (*Production, error) {
	if !ps.config.EnableLearning {
		return nil, errors.New("learning disabled")
	}

	if len(sequence) < ps.config.MinChunkSequence {
		return nil, fmt.Errorf("sequence too short: need %d, got %d", ps.config.MinChunkSequence, len(sequence))
	}

	// Collect conditions from all productions in sequence
	conditions := make([]*Condition, 0)
	actions := make([]*Action, 0)

	ps.mu.RLock()
	for i, record := range sequence {
		prod, ok := ps.productions[record.ProductionID]
		if !ok {
			continue
		}

		if i == 0 {
			// First production's conditions become chunk conditions
			conditions = append(conditions, prod.Conditions...)
		}

		if i == len(sequence)-1 {
			// Last production's actions become chunk actions
			actions = append(actions, prod.Actions...)
		}
	}
	ps.mu.RUnlock()

	if len(conditions) == 0 || len(actions) == 0 {
		return nil, errors.New("could not extract conditions/actions from sequence")
	}

	// Create chunk production
	chunk := &Production{
		ID:          fmt.Sprintf("chunk-%d", atomic.AddUint64(&chunkIDCounter, 1)),
		Name:        name,
		Description: fmt.Sprintf("Learned chunk from %d-step sequence", len(sequence)),
		Conditions:  conditions,
		Actions:     actions,
		Priority:    0.5, // Start with medium priority
		Source:      "learned",
		CreatedAt:   time.Now(),
		Tags:        []string{"learned", "chunk"},
		Metadata:    make(map[string]interface{}),
	}
	chunk.Metadata["source_sequence"] = len(sequence)

	if err := ps.AddProduction(chunk); err != nil {
		return nil, err
	}

	ps.mu.Lock()
	ps.stats.ProductionsLearned++
	ps.mu.Unlock()

	if ps.onLearned != nil {
		ps.onLearned(chunk)
	}

	return chunk, nil
}

// GetRecentHistory returns recent firing history.
func (ps *ProductionSystem) GetRecentHistory(n int) []*FiringRecord {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if n > len(ps.firingHistory) {
		n = len(ps.firingHistory)
	}

	result := make([]*FiringRecord, n)
	copy(result, ps.firingHistory[len(ps.firingHistory)-n:])
	return result
}

// MarkSuccess marks the last firing as successful (for learning).
func (ps *ProductionSystem) MarkSuccess() {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if len(ps.firingHistory) == 0 {
		return
	}

	record := ps.firingHistory[len(ps.firingHistory)-1]
	record.Success = true

	if prod, ok := ps.productions[record.ProductionID]; ok {
		prod.SuccessCount++
	}

	ps.stats.SuccessfulFirings++
}

// ============================================================================
// Callbacks
// ============================================================================

// OnProductionFired sets callback for production firing.
func (ps *ProductionSystem) OnProductionFired(fn func(*Production, *MatchResult)) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.onProductionFired = fn
}

// OnConflict sets callback for conflict resolution.
func (ps *ProductionSystem) OnConflict(fn func([]*MatchResult)) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.onConflict = fn
}

// OnLearned sets callback for learned productions.
func (ps *ProductionSystem) OnLearned(fn func(*Production)) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.onLearned = fn
}

// ============================================================================
// Statistics
// ============================================================================

// GetStats returns production system statistics.
func (ps *ProductionSystem) GetStats() *ProductionStats {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	return &ProductionStats{
		TotalProductions:   ps.stats.TotalProductions,
		TotalFirings:       ps.stats.TotalFirings,
		SuccessfulFirings:  ps.stats.SuccessfulFirings,
		ConflictsResolved:  ps.stats.ConflictsResolved,
		ProductionsLearned: ps.stats.ProductionsLearned,
		AverageCycleTime:   ps.stats.AverageCycleTime,
	}
}

// ============================================================================
// Snapshot
// ============================================================================

// ProductionSystemSnapshot represents system state.
type ProductionSystemSnapshot struct {
	Timestamp       time.Time
	ProductionCount int
	ConflictSetSize int
	RecentFirings   int
	LastFiredID     string
	EnabledCount    int
	DisabledCount   int
}

// Snapshot returns current system state.
func (ps *ProductionSystem) Snapshot() *ProductionSystemSnapshot {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	enabled := 0
	disabled := 0
	for _, prod := range ps.productions {
		if prod.Enabled {
			enabled++
		} else {
			disabled++
		}
	}

	return &ProductionSystemSnapshot{
		Timestamp:       time.Now(),
		ProductionCount: len(ps.productions),
		ConflictSetSize: len(ps.conflictSet),
		RecentFirings:   len(ps.firingHistory),
		LastFiredID:     ps.lastFired,
		EnabledCount:    enabled,
		DisabledCount:   disabled,
	}
}

// ============================================================================
// Clear
// ============================================================================

// Clear removes all productions.
func (ps *ProductionSystem) Clear() {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	ps.productions = make(map[string]*Production)
	ps.productionsByTag = make(map[string][]*Production)
	ps.conflictSet = make([]*MatchResult, 0)
	ps.refractionSet = make(map[string]bool)
	ps.firingHistory = make([]*FiringRecord, 0)
	ps.lastFired = ""
	ps.stats = &ProductionStats{}
}
