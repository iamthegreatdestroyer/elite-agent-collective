// Package memory provides the MNEMONIC system for the Elite Agent Collective.
// This file implements the Cognitive Goal Stack Management Component for Phase 1.
//
// The Cognitive Goal Stack Component:
// - Manages hierarchical goal structures using the existing GoalStack
// - Supports goal decomposition and aggregation
// - Tracks goal status, progress, and dependencies
// - Enables goal suspension and resumption
// - Manages goal activation and completion
// - Integrates with cognitive framework

package memory

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

// ============================================================================
// Cognitive Goal Stack Component
// ============================================================================

// CognitiveGoalStackComponent implements the CognitiveComponent interface
// for managing hierarchical goal structures in the cognitive system.
type CognitiveGoalStackComponent struct {
	mu             sync.RWMutex
	goalStack      *GoalStack
	goalTree       map[string]*Goal // ID -> Goal mapping for fast lookup
	metrics        CognitiveMetrics
	lastUpdate     time.Time
	requestCount   int64
	successCount   int64
	errorCount     int64
	goalsProcessed int64
	goalsCompleted int64
	goalsSuspended int64
	activatedGoals []string // Stack of activated goals
	completedGoals []string // Stack of completed goals
}

// NewCognitiveGoalStackComponent creates a new goal stack management component
func NewCognitiveGoalStackComponent() *CognitiveGoalStackComponent {
	config := DefaultGoalStackConfig()
	return &CognitiveGoalStackComponent{
		goalStack:      NewGoalStack(config),
		goalTree:       make(map[string]*Goal),
		activatedGoals: make([]string, 0),
		completedGoals: make([]string, 0),
		metrics: CognitiveMetrics{
			ComponentName: "CognitiveGoalStackManagement",
			CustomMetrics: make(map[string]interface{}),
		},
		lastUpdate: time.Now(),
	}
}

// Initialize sets up the component
func (cgsc *CognitiveGoalStackComponent) Initialize(config interface{}) error {
	cgsc.mu.Lock()
	defer cgsc.mu.Unlock()

	cgsc.metrics.LastUpdated = time.Now()
	return nil
}

// Process handles cognitive goal stack management requests
func (cgsc *CognitiveGoalStackComponent) Process(
	ctx context.Context,
	request *CognitiveProcessRequest,
) (*CognitiveProcessResult, error) {
	cgsc.mu.Lock()
	cgsc.requestCount++
	cgsc.mu.Unlock()

	startTime := time.Now()
	result := &CognitiveProcessResult{
		Status:             ProcessSuccess,
		ComponentsInvolved: []string{cgsc.GetName()},
		ExecutionSteps:     make([]*ExecutionStep, 0),
		SafetyCheckResults: make([]SafetyValidation, 0),
	}

	// Process current goal and manage goal stack
	if request.CurrentGoal != nil {
		step := cgsc.processGoalActivation(ctx, request)
		result.ExecutionSteps = append(result.ExecutionSteps, step)
	}

	// Get active goal stack
	activeGoals := cgsc.getActiveGoals()
	result.Output = activeGoals

	// Build decision trace
	stackSize := cgsc.goalStack.TotalSize()
	result.DecisionTrace = &DecisionTrace{
		Steps: []*DecisionStep{
			{
				Index:       0,
				Description: "Processed goal stack",
				Input:       request.CurrentGoal,
				Output:      activeGoals,
				Confidence:  1.0,
				Timestamp:   time.Now(),
			},
		},
		InitialState: map[string]interface{}{
			"stack_size":      stackSize,
			"active_count":    len(activeGoals),
			"goals_processed": cgsc.goalsProcessed,
			"goals_completed": cgsc.goalsCompleted,
		},
		FinalState: map[string]interface{}{
			"stack_size":      stackSize,
			"active_count":    len(activeGoals),
			"goals_processed": cgsc.goalsProcessed,
			"goals_completed": cgsc.goalsCompleted,
		},
	}

	result.Confidence = cgsc.calculateConfidence()
	result.ProcessingTime = time.Since(startTime)
	result.Explanation = fmt.Sprintf(
		"Goal stack managed with %d active goals out of %d total",
		len(activeGoals),
		stackSize,
	)

	cgsc.mu.Lock()
	cgsc.successCount++
	cgsc.mu.Unlock()

	return result, nil
}

// processGoalActivation activates a goal and manages the goal stack
func (cgsc *CognitiveGoalStackComponent) processGoalActivation(
	ctx context.Context,
	request *CognitiveProcessRequest,
) *ExecutionStep {
	start := time.Now()
	goal := request.CurrentGoal

	// Push goal onto stack
	_ = cgsc.goalStack.Push(goal)
	cgsc.goalTree[goal.ID] = goal
	cgsc.goalsProcessed++

	// Track activation
	cgsc.activatedGoals = append(cgsc.activatedGoals, goal.ID)

	return &ExecutionStep{
		ComponentName: cgsc.GetName(),
		StepNumber:    0,
		Input:         goal,
		Output:        len(cgsc.activatedGoals),
		Duration:      time.Since(start),
		Status:        "success",
	}
}

// CompleteGoal marks a goal as completed
func (cgsc *CognitiveGoalStackComponent) CompleteGoal(goalID string) bool {
	cgsc.mu.Lock()
	defer cgsc.mu.Unlock()

	goal, ok := cgsc.goalTree[goalID]
	if !ok {
		return false
	}

	// Mark as completed using GoalStack API
	_ = cgsc.goalStack.Complete(goalID)

	// Remove from active goals
	for i, id := range cgsc.activatedGoals {
		if id == goalID {
			cgsc.activatedGoals = append(cgsc.activatedGoals[:i], cgsc.activatedGoals[i+1:]...)
			break
		}
	}

	// Track completion
	cgsc.completedGoals = append(cgsc.completedGoals, goalID)
	cgsc.goalsCompleted++
	goal.Progress = 1.0

	return true
}

// FailGoal marks a goal as failed
func (cgsc *CognitiveGoalStackComponent) FailGoal(goalID string, reason string) bool {
	cgsc.mu.Lock()
	defer cgsc.mu.Unlock()

	_, ok := cgsc.goalTree[goalID]
	if !ok {
		return false
	}

	// Mark as failed using GoalStack API
	_ = cgsc.goalStack.Fail(goalID, reason)

	// Remove from active goals
	for i, id := range cgsc.activatedGoals {
		if id == goalID {
			cgsc.activatedGoals = append(cgsc.activatedGoals[:i], cgsc.activatedGoals[i+1:]...)
			break
		}
	}

	return true
}

// SuspendGoal suspends a goal (can be resumed later)
func (cgsc *CognitiveGoalStackComponent) SuspendGoal(goalID string, reason string) bool {
	cgsc.mu.Lock()
	defer cgsc.mu.Unlock()

	_, ok := cgsc.goalTree[goalID]
	if !ok {
		return false
	}

	// Mark as suspended using GoalStack API
	_ = cgsc.goalStack.Suspend(goalID, reason)

	cgsc.goalsSuspended++
	return true
}

// ResumeGoal resumes a suspended goal
func (cgsc *CognitiveGoalStackComponent) ResumeGoal(goalID string) bool {
	cgsc.mu.Lock()
	defer cgsc.mu.Unlock()

	_, ok := cgsc.goalTree[goalID]
	if !ok {
		return false
	}

	// Resume using GoalStack API
	_ = cgsc.goalStack.Resume(goalID)

	// Re-add to activated goals
	cgsc.activatedGoals = append(cgsc.activatedGoals, goalID)

	return true
}

// UpdateGoalProgress updates the progress of a goal
func (cgsc *CognitiveGoalStackComponent) UpdateGoalProgress(goalID string, progress float64) bool {
	cgsc.mu.Lock()
	defer cgsc.mu.Unlock()

	goal, ok := cgsc.goalTree[goalID]
	if !ok {
		return false
	}

	// Clamp progress to [0, 1]
	if progress < 0 {
		progress = 0
	} else if progress > 1 {
		progress = 1
	}

	goal.Progress = progress
	return true
}

// DecomposeGoal creates subgoals from a parent goal
func (cgsc *CognitiveGoalStackComponent) DecomposeGoal(parentID string, subgoals []*Goal) bool {
	cgsc.mu.Lock()
	defer cgsc.mu.Unlock()

	parent, ok := cgsc.goalTree[parentID]
	if !ok {
		return false
	}

	// Decompose using GoalStack API
	_ = cgsc.goalStack.Decompose(parentID, subgoals)

	// Add subgoals to tree
	for _, subgoal := range subgoals {
		subgoal.ParentID = parentID
		subgoal.Depth = parent.Depth + 1
		cgsc.goalTree[subgoal.ID] = subgoal
	}

	return true
}

// getActiveGoals returns all currently active goals
func (cgsc *CognitiveGoalStackComponent) getActiveGoals() []*Goal {
	cgsc.mu.RLock()
	defer cgsc.mu.RUnlock()

	activeGoals := make([]*Goal, 0)
	for _, goalID := range cgsc.activatedGoals {
		if goal, ok := cgsc.goalTree[goalID]; ok && goal.Status != GoalCompleted && goal.Status != GoalFailed {
			activeGoals = append(activeGoals, goal)
		}
	}

	// Sort by priority (descending)
	sort.Slice(activeGoals, func(i, j int) bool {
		return activeGoals[i].Priority > activeGoals[j].Priority
	})

	return activeGoals
}

// calculateConfidence calculates confidence based on goal stack state
func (cgsc *CognitiveGoalStackComponent) calculateConfidence() float64 {
	cgsc.mu.RLock()
	defer cgsc.mu.RUnlock()

	if cgsc.goalsProcessed == 0 {
		return 0.5 // Default confidence when no goals processed
	}

	// Confidence increases with completion rate
	completionRate := float64(cgsc.goalsCompleted) / float64(cgsc.goalsProcessed)
	return completionRate * 0.9 // Max 90% confidence
}

// Shutdown gracefully shuts down the component
func (cgsc *CognitiveGoalStackComponent) Shutdown() error {
	cgsc.mu.Lock()
	defer cgsc.mu.Unlock()

	cgsc.goalStack = nil
	cgsc.goalTree = nil
	return nil
}

// GetMetrics returns current performance metrics
func (cgsc *CognitiveGoalStackComponent) GetMetrics() CognitiveMetrics {
	cgsc.mu.RLock()
	defer cgsc.mu.RUnlock()

	metrics := cgsc.metrics
	metrics.TotalRequests = cgsc.requestCount
	metrics.SuccessfulRequests = cgsc.successCount
	metrics.FailedRequests = cgsc.errorCount
	metrics.LastUpdated = time.Now()

	if cgsc.requestCount > 0 {
		metrics.ErrorRate = float64(cgsc.errorCount) / float64(cgsc.requestCount)
	}

	// Add custom metrics
	metrics.CustomMetrics = map[string]interface{}{
		"stack_size":      cgsc.goalStack.TotalSize(),
		"active_goals":    len(cgsc.activatedGoals),
		"completed_goals": cgsc.goalsCompleted,
		"suspended_goals": cgsc.goalsSuspended,
		"total_goals":     len(cgsc.goalTree),
		"completion_rate": cgsc.calculateConfidence(),
	}

	return metrics
}

// GetName returns the component's name
func (cgsc *CognitiveGoalStackComponent) GetName() string {
	return "CognitiveGoalStackManagement"
}

// ============================================================================
// Helper Methods
// ============================================================================

// GetGoalByID retrieves a goal by ID
func (cgsc *CognitiveGoalStackComponent) GetGoalByID(goalID string) *Goal {
	cgsc.mu.RLock()
	defer cgsc.mu.RUnlock()

	goal, ok := cgsc.goalTree[goalID]
	if !ok {
		return nil
	}
	return goal
}

// GetActiveGoalStack returns the current active goal stack
func (cgsc *CognitiveGoalStackComponent) GetActiveGoalStack() []*Goal {
	cgsc.mu.RLock()
	defer cgsc.mu.RUnlock()

	stack := make([]*Goal, 0)
	for _, goalID := range cgsc.activatedGoals {
		if goal, ok := cgsc.goalTree[goalID]; ok && goal.Status != GoalCompleted && goal.Status != GoalFailed {
			stack = append(stack, goal)
		}
	}
	return stack
}

// GetCompletedGoals returns all completed goals
func (cgsc *CognitiveGoalStackComponent) GetCompletedGoals() []*Goal {
	cgsc.mu.RLock()
	defer cgsc.mu.RUnlock()

	completed := make([]*Goal, 0)
	for _, goalID := range cgsc.completedGoals {
		if goal, ok := cgsc.goalTree[goalID]; ok && goal.Status == GoalCompleted {
			completed = append(completed, goal)
		}
	}
	return completed
}

// ClearGoalStack clears all goals (useful for reset)
func (cgsc *CognitiveGoalStackComponent) ClearGoalStack() {
	cgsc.mu.Lock()
	defer cgsc.mu.Unlock()

	config := DefaultGoalStackConfig()
	cgsc.goalStack = NewGoalStack(config)
	cgsc.goalTree = make(map[string]*Goal)
	cgsc.activatedGoals = make([]string, 0)
	cgsc.completedGoals = make([]string, 0)
}
