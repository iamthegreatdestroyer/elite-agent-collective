// Package memory provides the MNEMONIC system for the Elite Agent Collective.
// This file implements Goal Stack Management from @NEURAL's Cognitive Architecture Analysis.
//
// The Goal Stack is based on cognitive architectures like SOAR and ACT-R:
// - Hierarchical goal decomposition
// - Goal suspension and resumption
// - Priority-based ordering
// - Subgoal creation during impasses
// - Goal satisfaction tracking

package memory

import (
	"container/heap"
	"errors"
	"fmt"
	"sync"
	"time"
)

// ============================================================================
// Goal Errors
// ============================================================================

var (
	// ErrGoalNotFound indicates a goal was not found
	ErrGoalNotFound = errors.New("goal not found")

	// ErrGoalStackEmpty indicates the goal stack is empty
	ErrGoalStackEmpty = errors.New("goal stack is empty")

	// ErrMaxDepthExceeded indicates goal decomposition is too deep
	ErrMaxDepthExceeded = errors.New("maximum goal stack depth exceeded")

	// ErrCircularDependency indicates a circular goal dependency
	ErrCircularDependency = errors.New("circular goal dependency detected")
)

// ============================================================================
// Goal Status
// ============================================================================

// GoalStatus represents the current state of a goal.
type GoalStatus int

const (
	// GoalPending means the goal has not started
	GoalPending GoalStatus = iota

	// GoalActive means the goal is currently being worked on
	GoalActive

	// GoalSuspended means the goal is paused (e.g., waiting for subgoal)
	GoalSuspended

	// GoalCompleted means the goal was successfully achieved
	GoalCompleted

	// GoalFailed means the goal could not be achieved
	GoalFailed

	// GoalDecomposed means the goal was broken into subgoals
	GoalDecomposed
)

// String returns a human-readable status.
func (s GoalStatus) String() string {
	switch s {
	case GoalPending:
		return "PENDING"
	case GoalActive:
		return "ACTIVE"
	case GoalSuspended:
		return "SUSPENDED"
	case GoalCompleted:
		return "COMPLETED"
	case GoalFailed:
		return "FAILED"
	case GoalDecomposed:
		return "DECOMPOSED"
	default:
		return "UNKNOWN"
	}
}

// ============================================================================
// Goal Priority
// ============================================================================

// GoalPriority defines priority levels.
type GoalPriority int

const (
	PriorityLow      GoalPriority = 1
	PriorityNormal   GoalPriority = 5
	PriorityHigh     GoalPriority = 8
	PriorityCritical GoalPriority = 10
)

// ============================================================================
// Goal
// ============================================================================

// Goal represents a task or objective to be achieved.
type Goal struct {
	// ID uniquely identifies this goal
	ID string

	// Name is a human-readable description
	Name string

	// Description provides detailed context
	Description string

	// Priority determines ordering (higher = more important)
	Priority GoalPriority

	// Status is the current goal state
	Status GoalStatus

	// ParentID links to the parent goal (if this is a subgoal)
	ParentID string

	// SubGoalIDs are child goals created from decomposition
	SubGoalIDs []string

	// Dependencies are goal IDs that must complete first
	Dependencies []string

	// CreatedAt timestamp
	CreatedAt time.Time

	// ActivatedAt is when the goal became active
	ActivatedAt *time.Time

	// CompletedAt is when the goal was completed/failed
	CompletedAt *time.Time

	// Deadline is an optional deadline
	Deadline *time.Time

	// Context holds goal-specific data
	Context map[string]interface{}

	// Preconditions that must be true to start
	Preconditions []string

	// Postconditions that should be true when complete
	Postconditions []string

	// Progress is completion percentage (0.0 to 1.0)
	Progress float64

	// SuspensionReason explains why the goal was suspended
	SuspensionReason string

	// FailureReason explains why the goal failed
	FailureReason string

	// Depth is the nesting level (0 = top-level)
	Depth int

	// Metadata for extensions
	Metadata map[string]interface{}

	// index for heap operations
	index int
}

// IsTerminal returns true if the goal is in a terminal state.
func (g *Goal) IsTerminal() bool {
	return g.Status == GoalCompleted || g.Status == GoalFailed
}

// IsActionable returns true if the goal can be worked on.
func (g *Goal) IsActionable() bool {
	return g.Status == GoalPending || g.Status == GoalActive
}

// ============================================================================
// Goal Priority Queue
// ============================================================================

// goalPriorityQueue implements heap.Interface for priority-based ordering.
type goalPriorityQueue []*Goal

func (pq goalPriorityQueue) Len() int { return len(pq) }

// Less returns true if i has higher priority than j.
// Higher priority value = higher priority.
// Ties are broken by creation time (earlier = higher priority).
func (pq goalPriorityQueue) Less(i, j int) bool {
	if pq[i].Priority != pq[j].Priority {
		return pq[i].Priority > pq[j].Priority
	}
	return pq[i].CreatedAt.Before(pq[j].CreatedAt)
}

func (pq goalPriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *goalPriorityQueue) Push(x interface{}) {
	n := len(*pq)
	goal := x.(*Goal)
	goal.index = n
	*pq = append(*pq, goal)
}

func (pq *goalPriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	goal := old[n-1]
	old[n-1] = nil
	goal.index = -1
	*pq = old[0 : n-1]
	return goal
}

// ============================================================================
// Goal Stack
// ============================================================================

// GoalStack manages hierarchical goals with priority-based processing.
type GoalStack struct {
	mu sync.RWMutex

	// goals stores all goals by ID
	goals map[string]*Goal

	// activeQueue holds pending/active goals in priority order
	activeQueue goalPriorityQueue

	// suspendedGoals holds suspended goals
	suspendedGoals map[string]*Goal

	// completedGoals holds terminal goals for history
	completedGoals map[string]*Goal

	// currentGoal is the actively focused goal
	currentGoal *Goal

	// maxDepth limits goal decomposition depth
	maxDepth int

	// maxGoals limits total goals
	maxGoals int

	// stats tracks goal stack statistics
	stats *GoalStackStats

	// callbacks
	onGoalActivated func(*Goal)
	onGoalCompleted func(*Goal)
	onGoalFailed    func(*Goal)
	onGoalSuspended func(*Goal)
}

// GoalStackStats tracks goal processing statistics.
type GoalStackStats struct {
	TotalGoalsCreated     int64
	TotalGoalsCompleted   int64
	TotalGoalsFailed      int64
	TotalGoalsSuspended   int64
	MaxDepthReached       int
	AverageCompletionTime time.Duration
}

// GoalStackConfig configures the goal stack.
type GoalStackConfig struct {
	MaxDepth int
	MaxGoals int
}

// DefaultGoalStackConfig returns sensible defaults.
func DefaultGoalStackConfig() GoalStackConfig {
	return GoalStackConfig{
		MaxDepth: 10,
		MaxGoals: 100,
	}
}

// NewGoalStack creates a new goal stack.
func NewGoalStack(config GoalStackConfig) *GoalStack {
	gs := &GoalStack{
		goals:          make(map[string]*Goal),
		activeQueue:    make(goalPriorityQueue, 0),
		suspendedGoals: make(map[string]*Goal),
		completedGoals: make(map[string]*Goal),
		maxDepth:       config.MaxDepth,
		maxGoals:       config.MaxGoals,
		stats:          &GoalStackStats{},
	}

	heap.Init(&gs.activeQueue)
	return gs
}

// ============================================================================
// Core Operations
// ============================================================================

// Push adds a new goal to the stack.
func (gs *GoalStack) Push(goal *Goal) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	// Check max goals
	if len(gs.goals) >= gs.maxGoals {
		return fmt.Errorf("maximum goals (%d) exceeded", gs.maxGoals)
	}

	// Check depth
	if goal.Depth > gs.maxDepth {
		return ErrMaxDepthExceeded
	}

	// Initialize goal
	now := time.Now()
	goal.CreatedAt = now
	if goal.Status == 0 {
		goal.Status = GoalPending
	}
	if goal.Priority == 0 {
		goal.Priority = PriorityNormal
	}
	if goal.Context == nil {
		goal.Context = make(map[string]interface{})
	}
	if goal.Metadata == nil {
		goal.Metadata = make(map[string]interface{})
	}

	// Store goal
	gs.goals[goal.ID] = goal

	// Add to active queue if pending
	if goal.Status == GoalPending {
		heap.Push(&gs.activeQueue, goal)
	}

	gs.stats.TotalGoalsCreated++

	// Update depth stats
	if goal.Depth > gs.stats.MaxDepthReached {
		gs.stats.MaxDepthReached = goal.Depth
	}

	// Set as current if none active
	if gs.currentGoal == nil && goal.Status == GoalPending {
		gs.activateGoalLocked(goal)
	}

	return nil
}

// Pop removes and returns the highest-priority goal.
func (gs *GoalStack) Pop() (*Goal, error) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	if len(gs.activeQueue) == 0 {
		return nil, ErrGoalStackEmpty
	}

	goal := heap.Pop(&gs.activeQueue).(*Goal)

	// Clear current goal if it was popped
	if gs.currentGoal != nil && gs.currentGoal.ID == goal.ID {
		gs.currentGoal = nil
	}

	return goal, nil
}

// Peek returns the highest-priority goal without removing it.
func (gs *GoalStack) Peek() (*Goal, error) {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	if len(gs.activeQueue) == 0 {
		return nil, ErrGoalStackEmpty
	}

	return gs.activeQueue[0], nil
}

// Get retrieves a goal by ID.
func (gs *GoalStack) Get(id string) (*Goal, error) {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	goal, ok := gs.goals[id]
	if !ok {
		// Check completed goals
		if goal, ok = gs.completedGoals[id]; ok {
			return goal, nil
		}
		return nil, ErrGoalNotFound
	}

	return goal, nil
}

// Current returns the currently active goal.
func (gs *GoalStack) Current() *Goal {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	return gs.currentGoal
}

// ============================================================================
// Goal State Transitions
// ============================================================================

// Activate makes a goal the current focus.
func (gs *GoalStack) Activate(id string) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	goal, ok := gs.goals[id]
	if !ok {
		return ErrGoalNotFound
	}

	if goal.IsTerminal() {
		return fmt.Errorf("cannot activate terminal goal")
	}

	// Suspend current goal if different
	if gs.currentGoal != nil && gs.currentGoal.ID != id {
		gs.suspendGoalLocked(gs.currentGoal, "preempted by higher priority goal")
	}

	gs.activateGoalLocked(goal)
	return nil
}

// activateGoalLocked activates a goal. Must hold lock.
func (gs *GoalStack) activateGoalLocked(goal *Goal) {
	now := time.Now()
	goal.Status = GoalActive
	goal.ActivatedAt = &now

	// Remove from suspended if present
	delete(gs.suspendedGoals, goal.ID)

	gs.currentGoal = goal

	if gs.onGoalActivated != nil {
		gs.onGoalActivated(goal)
	}
}

// Complete marks a goal as completed.
func (gs *GoalStack) Complete(id string) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	goal, ok := gs.goals[id]
	if !ok {
		return ErrGoalNotFound
	}

	now := time.Now()
	goal.Status = GoalCompleted
	goal.CompletedAt = &now
	goal.Progress = 1.0

	// Remove from active queue
	if goal.index >= 0 && goal.index < len(gs.activeQueue) {
		heap.Remove(&gs.activeQueue, goal.index)
	}

	// Move to completed
	delete(gs.goals, id)
	gs.completedGoals[id] = goal

	gs.stats.TotalGoalsCompleted++

	// Clear current if this was it
	if gs.currentGoal != nil && gs.currentGoal.ID == id {
		gs.currentGoal = nil
		gs.selectNextGoalLocked()
	}

	// Resume parent if all subgoals complete
	if goal.ParentID != "" {
		gs.checkParentCompletionLocked(goal.ParentID)
	}

	if gs.onGoalCompleted != nil {
		gs.onGoalCompleted(goal)
	}

	return nil
}

// Fail marks a goal as failed.
func (gs *GoalStack) Fail(id string, reason string) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	goal, ok := gs.goals[id]
	if !ok {
		return ErrGoalNotFound
	}

	now := time.Now()
	goal.Status = GoalFailed
	goal.CompletedAt = &now
	goal.FailureReason = reason

	// Remove from active queue
	if goal.index >= 0 && goal.index < len(gs.activeQueue) {
		heap.Remove(&gs.activeQueue, goal.index)
	}

	// Move to completed (as failed)
	delete(gs.goals, id)
	gs.completedGoals[id] = goal

	gs.stats.TotalGoalsFailed++

	// Clear current if this was it
	if gs.currentGoal != nil && gs.currentGoal.ID == id {
		gs.currentGoal = nil
		gs.selectNextGoalLocked()
	}

	// Cascade failure to parent if no alternate path
	if goal.ParentID != "" {
		gs.handleSubgoalFailureLocked(goal.ParentID, goal.ID)
	}

	if gs.onGoalFailed != nil {
		gs.onGoalFailed(goal)
	}

	return nil
}

// Suspend pauses a goal.
func (gs *GoalStack) Suspend(id string, reason string) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	goal, ok := gs.goals[id]
	if !ok {
		return ErrGoalNotFound
	}

	gs.suspendGoalLocked(goal, reason)
	return nil
}

// suspendGoalLocked suspends a goal. Must hold lock.
func (gs *GoalStack) suspendGoalLocked(goal *Goal, reason string) {
	goal.Status = GoalSuspended
	goal.SuspensionReason = reason

	// Remove from active queue
	if goal.index >= 0 && goal.index < len(gs.activeQueue) {
		heap.Remove(&gs.activeQueue, goal.index)
	}

	gs.suspendedGoals[goal.ID] = goal

	gs.stats.TotalGoalsSuspended++

	// Clear current if this was it
	if gs.currentGoal != nil && gs.currentGoal.ID == goal.ID {
		gs.currentGoal = nil
	}

	if gs.onGoalSuspended != nil {
		gs.onGoalSuspended(goal)
	}
}

// Resume reactivates a suspended goal.
func (gs *GoalStack) Resume(id string) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	goal, ok := gs.suspendedGoals[id]
	if !ok {
		return ErrGoalNotFound
	}

	goal.Status = GoalPending
	goal.SuspensionReason = ""

	delete(gs.suspendedGoals, id)
	heap.Push(&gs.activeQueue, goal)

	return nil
}

// ============================================================================
// Goal Decomposition
// ============================================================================

// Decompose breaks a goal into subgoals.
func (gs *GoalStack) Decompose(parentID string, subgoals []*Goal) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	parent, ok := gs.goals[parentID]
	if !ok {
		return ErrGoalNotFound
	}

	// Check depth limit
	if parent.Depth+1 > gs.maxDepth {
		return ErrMaxDepthExceeded
	}

	// Mark parent as decomposed
	parent.Status = GoalDecomposed

	// Add subgoals
	subgoalIDs := make([]string, 0, len(subgoals))
	for _, sg := range subgoals {
		sg.ParentID = parentID
		sg.Depth = parent.Depth + 1
		sg.CreatedAt = time.Now()
		sg.Status = GoalPending
		if sg.Priority == 0 {
			sg.Priority = parent.Priority
		}
		if sg.Context == nil {
			sg.Context = make(map[string]interface{})
		}
		if sg.Metadata == nil {
			sg.Metadata = make(map[string]interface{})
		}

		gs.goals[sg.ID] = sg
		heap.Push(&gs.activeQueue, sg)
		subgoalIDs = append(subgoalIDs, sg.ID)

		gs.stats.TotalGoalsCreated++
		if sg.Depth > gs.stats.MaxDepthReached {
			gs.stats.MaxDepthReached = sg.Depth
		}
	}

	parent.SubGoalIDs = subgoalIDs

	// Suspend parent while subgoals execute
	gs.suspendGoalLocked(parent, "decomposed into subgoals")

	// Activate highest priority subgoal
	gs.selectNextGoalLocked()

	return nil
}

// CreateSubgoal creates a single subgoal for a parent.
func (gs *GoalStack) CreateSubgoal(parentID string, subgoal *Goal) error {
	return gs.Decompose(parentID, []*Goal{subgoal})
}

// ============================================================================
// Helper Methods
// ============================================================================

// selectNextGoalLocked selects the next goal to work on. Must hold lock.
func (gs *GoalStack) selectNextGoalLocked() {
	if len(gs.activeQueue) == 0 {
		gs.currentGoal = nil
		return
	}

	// Get highest priority pending goal
	for _, goal := range gs.activeQueue {
		if goal.Status == GoalPending {
			gs.activateGoalLocked(goal)
			return
		}
	}
}

// checkParentCompletionLocked checks if all subgoals are complete. Must hold lock.
func (gs *GoalStack) checkParentCompletionLocked(parentID string) {
	parent, ok := gs.suspendedGoals[parentID]
	if !ok {
		return
	}

	// Check if all subgoals are complete
	allComplete := true
	for _, subID := range parent.SubGoalIDs {
		sub, ok := gs.completedGoals[subID]
		if !ok || sub.Status != GoalCompleted {
			allComplete = false
			break
		}
	}

	if allComplete {
		// Resume and complete parent
		parent.Status = GoalCompleted
		now := time.Now()
		parent.CompletedAt = &now
		parent.Progress = 1.0

		delete(gs.suspendedGoals, parentID)
		delete(gs.goals, parentID)
		gs.completedGoals[parentID] = parent

		gs.stats.TotalGoalsCompleted++

		if gs.onGoalCompleted != nil {
			gs.onGoalCompleted(parent)
		}

		// Recursively check parent's parent
		if parent.ParentID != "" {
			gs.checkParentCompletionLocked(parent.ParentID)
		}
	}
}

// handleSubgoalFailureLocked handles a subgoal failure. Must hold lock.
func (gs *GoalStack) handleSubgoalFailureLocked(parentID string, failedSubgoalID string) {
	parent, ok := gs.suspendedGoals[parentID]
	if !ok {
		return
	}

	// Check if there are other subgoals that might succeed
	hasAlternatives := false
	for _, subID := range parent.SubGoalIDs {
		if subID == failedSubgoalID {
			continue
		}
		if sub, ok := gs.goals[subID]; ok && sub.IsActionable() {
			hasAlternatives = true
			break
		}
		if sub, ok := gs.completedGoals[subID]; ok && sub.Status == GoalCompleted {
			hasAlternatives = true
			break
		}
	}

	if !hasAlternatives {
		// Cascade failure to parent
		parent.Status = GoalFailed
		now := time.Now()
		parent.CompletedAt = &now
		parent.FailureReason = fmt.Sprintf("subgoal %s failed", failedSubgoalID)

		delete(gs.suspendedGoals, parentID)
		delete(gs.goals, parentID)
		gs.completedGoals[parentID] = parent

		gs.stats.TotalGoalsFailed++

		if gs.onGoalFailed != nil {
			gs.onGoalFailed(parent)
		}

		// Recursively cascade
		if parent.ParentID != "" {
			gs.handleSubgoalFailureLocked(parent.ParentID, parentID)
		}
	}
}

// ============================================================================
// Query Operations
// ============================================================================

// Size returns the number of active goals.
func (gs *GoalStack) Size() int {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	return len(gs.activeQueue)
}

// TotalSize returns the total number of goals (active + suspended).
func (gs *GoalStack) TotalSize() int {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	return len(gs.goals) + len(gs.suspendedGoals)
}

// IsEmpty returns true if no active goals.
func (gs *GoalStack) IsEmpty() bool {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	return len(gs.activeQueue) == 0
}

// GetByStatus returns goals with a specific status.
func (gs *GoalStack) GetByStatus(status GoalStatus) []*Goal {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	result := make([]*Goal, 0)

	for _, goal := range gs.goals {
		if goal.Status == status {
			result = append(result, goal)
		}
	}

	if status == GoalSuspended {
		for _, goal := range gs.suspendedGoals {
			result = append(result, goal)
		}
	}

	if status == GoalCompleted || status == GoalFailed {
		for _, goal := range gs.completedGoals {
			if goal.Status == status {
				result = append(result, goal)
			}
		}
	}

	return result
}

// GetSubgoals returns subgoals of a parent goal.
func (gs *GoalStack) GetSubgoals(parentID string) []*Goal {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	parent, ok := gs.goals[parentID]
	if !ok {
		if parent, ok = gs.suspendedGoals[parentID]; !ok {
			return nil
		}
	}

	result := make([]*Goal, 0, len(parent.SubGoalIDs))
	for _, subID := range parent.SubGoalIDs {
		if sub, ok := gs.goals[subID]; ok {
			result = append(result, sub)
		} else if sub, ok := gs.completedGoals[subID]; ok {
			result = append(result, sub)
		}
	}

	return result
}

// GetAncestors returns the goal's parent chain.
func (gs *GoalStack) GetAncestors(id string) []*Goal {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	result := make([]*Goal, 0)
	current := id

	for {
		goal, ok := gs.goals[current]
		if !ok {
			if goal, ok = gs.suspendedGoals[current]; !ok {
				if goal, ok = gs.completedGoals[current]; !ok {
					break
				}
			}
		}

		if goal.ParentID == "" {
			break
		}

		parent, ok := gs.goals[goal.ParentID]
		if !ok {
			if parent, ok = gs.suspendedGoals[goal.ParentID]; !ok {
				if parent, ok = gs.completedGoals[goal.ParentID]; !ok {
					break
				}
			}
		}

		result = append(result, parent)
		current = goal.ParentID
	}

	return result
}

// ============================================================================
// Progress Tracking
// ============================================================================

// UpdateProgress updates a goal's progress.
func (gs *GoalStack) UpdateProgress(id string, progress float64) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	goal, ok := gs.goals[id]
	if !ok {
		return ErrGoalNotFound
	}

	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}

	goal.Progress = progress
	return nil
}

// ============================================================================
// Priority Management
// ============================================================================

// SetPriority changes a goal's priority.
func (gs *GoalStack) SetPriority(id string, priority GoalPriority) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	goal, ok := gs.goals[id]
	if !ok {
		return ErrGoalNotFound
	}

	goal.Priority = priority

	// Reheap to fix ordering
	if goal.index >= 0 {
		heap.Fix(&gs.activeQueue, goal.index)
	}

	return nil
}

// ============================================================================
// Callbacks
// ============================================================================

// OnGoalActivated sets callback for goal activation.
func (gs *GoalStack) OnGoalActivated(fn func(*Goal)) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	gs.onGoalActivated = fn
}

// OnGoalCompleted sets callback for goal completion.
func (gs *GoalStack) OnGoalCompleted(fn func(*Goal)) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	gs.onGoalCompleted = fn
}

// OnGoalFailed sets callback for goal failure.
func (gs *GoalStack) OnGoalFailed(fn func(*Goal)) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	gs.onGoalFailed = fn
}

// OnGoalSuspended sets callback for goal suspension.
func (gs *GoalStack) OnGoalSuspended(fn func(*Goal)) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	gs.onGoalSuspended = fn
}

// ============================================================================
// Statistics
// ============================================================================

// GetStats returns goal stack statistics.
func (gs *GoalStack) GetStats() *GoalStackStats {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	// Return a copy
	stats := *gs.stats
	return &stats
}

// ============================================================================
// Snapshot
// ============================================================================

// GoalStackSnapshot represents the current state.
type GoalStackSnapshot struct {
	Timestamp       time.Time
	ActiveCount     int
	SuspendedCount  int
	CompletedCount  int
	CurrentGoalID   string
	Goals           []*Goal
	MaxDepthReached int
}

// Snapshot returns current state for debugging/monitoring.
func (gs *GoalStack) Snapshot() *GoalStackSnapshot {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	snapshot := &GoalStackSnapshot{
		Timestamp:       time.Now(),
		ActiveCount:     len(gs.activeQueue),
		SuspendedCount:  len(gs.suspendedGoals),
		CompletedCount:  len(gs.completedGoals),
		Goals:           make([]*Goal, 0, len(gs.goals)),
		MaxDepthReached: gs.stats.MaxDepthReached,
	}

	if gs.currentGoal != nil {
		snapshot.CurrentGoalID = gs.currentGoal.ID
	}

	for _, goal := range gs.goals {
		snapshot.Goals = append(snapshot.Goals, goal)
	}

	return snapshot
}

// ============================================================================
// Clear
// ============================================================================

// Clear removes all goals.
func (gs *GoalStack) Clear() {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	gs.goals = make(map[string]*Goal)
	gs.activeQueue = make(goalPriorityQueue, 0)
	gs.suspendedGoals = make(map[string]*Goal)
	gs.completedGoals = make(map[string]*Goal)
	gs.currentGoal = nil

	heap.Init(&gs.activeQueue)
}
