// Package memory provides the MNEMONIC system for the Elite Agent Collective.
// This file implements the Cognitive Working Memory Component for Phase 1.
//
// The Cognitive Working Memory Component:
// - Manages working memory with Miller's capacity limit (7±2)
// - Implements activation-based retrieval
// - Provides spreading activation for priming
// - Enables chunk formation and binding
// - Tracks item decay and recency
// - Integrates with the cognitive framework

package memory

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

// ============================================================================
// Cognitive Working Memory Component
// ============================================================================

// CognitiveWorkingMemoryComponent implements the CognitiveComponent interface
// for managing working memory in the cognitive system.
type CognitiveWorkingMemoryComponent struct {
	mu            sync.RWMutex
	workingMemory *CognitiveWorkingMemory
	metrics       CognitiveMetrics
	lastUpdate    time.Time
	requestCount  int64
	successCount  int64
	errorCount    int64
}

// NewCognitiveWorkingMemoryComponent creates a new working memory component
func NewCognitiveWorkingMemoryComponent(capacity int) *CognitiveWorkingMemoryComponent {
	config := WorkingMemoryConfig{
		Capacity:            capacity,
		DecayRate:           DefaultActivationDecayRate,
		ActivationThreshold: DefaultActivationThreshold,
		SpreadingFactor:     DefaultSpreadingFactor,
	}
	return &CognitiveWorkingMemoryComponent{
		workingMemory: NewCognitiveWorkingMemory(config),
		metrics: CognitiveMetrics{
			ComponentName: "CognitiveWorkingMemory",
			CustomMetrics: make(map[string]interface{}),
		},
		lastUpdate: time.Now(),
	}
}

// Initialize sets up the component
func (cwmc *CognitiveWorkingMemoryComponent) Initialize(config interface{}) error {
	cwmc.mu.Lock()
	defer cwmc.mu.Unlock()

	// Could use config to customize capacity, decay rate, etc.
	// For now, using defaults

	cwmc.metrics.LastUpdated = time.Now()
	return nil
}

// Process handles cognitive working memory requests
func (cwmc *CognitiveWorkingMemoryComponent) Process(
	ctx context.Context,
	request *CognitiveProcessRequest,
) (*CognitiveProcessResult, error) {
	cwmc.mu.Lock()
	cwmc.requestCount++
	cwmc.mu.Unlock()

	startTime := time.Now()
	result := &CognitiveProcessResult{
		Status:             ProcessSuccess,
		ComponentsInvolved: []string{cwmc.GetName()},
		ExecutionSteps:     make([]*ExecutionStep, 0),
		SafetyCheckResults: make([]SafetyValidation, 0),
	}

	// Process the current goal to extract and store items in working memory
	if request.CurrentGoal != nil {
		step := cwmc.processGoal(ctx, request)
		result.ExecutionSteps = append(result.ExecutionSteps, step)
	}

	// Retrieve items most relevant to the current goal
	var retrieved []*WorkingMemoryItem
	if request.CurrentGoal != nil {
		retrieved = cwmc.retrieveRelevantItems(request.CurrentGoal)
		result.Output = retrieved
		result.SelectedAgents = cwmc.extractAgentIDs(retrieved)
	}

	// Update context with working memory state
	request.WorkingMemory = cwmc.workingMemory

	// Build decision trace
	allItems := cwmc.workingMemory.GetAll()
	load := float64(cwmc.workingMemory.Size()) / float64(cwmc.workingMemory.Capacity())

	result.DecisionTrace = &DecisionTrace{
		Steps: []*DecisionStep{
			{
				Index:       0,
				Description: "Loaded working memory state",
				Input:       request.CurrentGoal,
				Output:      allItems,
				Confidence:  1.0,
				Timestamp:   time.Now(),
			},
		},
		InitialState: map[string]interface{}{
			"capacity":   cwmc.workingMemory.Capacity(),
			"load":       load,
			"item_count": len(allItems),
		},
		FinalState: map[string]interface{}{
			"capacity":   cwmc.workingMemory.Capacity(),
			"load":       load,
			"item_count": len(allItems),
		},
	}

	result.Confidence = cwmc.calculateConfidence()
	result.ProcessingTime = time.Since(startTime)
	result.Explanation = fmt.Sprintf(
		"Working memory processed with %d items at %.1f%% capacity",
		len(allItems),
		load*100,
	)

	cwmc.mu.Lock()
	cwmc.successCount++
	cwmc.mu.Unlock()

	return result, nil
}

// processGoal extracts and processes goal information into working memory
func (cwmc *CognitiveWorkingMemoryComponent) processGoal(
	ctx context.Context,
	request *CognitiveProcessRequest,
) *ExecutionStep {
	start := time.Now()
	goal := request.CurrentGoal

	// Extract items from goal: ID, name, description
	items := []string{
		goal.ID,
		goal.Name,
		fmt.Sprintf("Priority: %d", goal.Priority),
	}

	if goal.Description != "" {
		items = append(items, goal.Description)
	}

	if len(goal.Dependencies) > 0 {
		items = append(items, fmt.Sprintf("Dependencies: %d", len(goal.Dependencies)))
	}

	if len(goal.Preconditions) > 0 {
		items = append(items, fmt.Sprintf("Preconditions: %d", len(goal.Preconditions)))
	}

	// Add items to working memory
	for i, itemContent := range items {
		itemID := fmt.Sprintf("goal_%s_%d", goal.ID, i)
		item := &WorkingMemoryItem{
			ID:          itemID,
			Content:     itemContent,
			Activation:  1.0, // New items start with full activation
			Salience:    0.9, // Goals are highly salient
			CreatedAt:   time.Now(),
			LastAccess:  time.Now(),
			AccessCount: 0,
			Source:      SourceGoal,
		}
		cwmc.workingMemory.Add(item)
	}

	return &ExecutionStep{
		ComponentName: cwmc.GetName(),
		StepNumber:    0,
		Input:         goal,
		Output:        len(items),
		Duration:      time.Since(start),
		Status:        "success",
	}
}

// retrieveRelevantItems retrieves items most relevant to the goal
func (cwmc *CognitiveWorkingMemoryComponent) retrieveRelevantItems(goal *Goal) []*WorkingMemoryItem {
	cwmc.mu.RLock()
	defer cwmc.mu.RUnlock()

	items := cwmc.workingMemory.GetAll()

	// Sort by activation (higher first)
	sort.Slice(items, func(i, j int) bool {
		return items[i].Activation > items[j].Activation
	})

	// Return top items based on capacity
	maxItems := cwmc.workingMemory.Capacity()
	if len(items) > maxItems {
		return items[:maxItems]
	}
	return items
}

// extractAgentIDs extracts agent IDs from working memory items
func (cwmc *CognitiveWorkingMemoryComponent) extractAgentIDs(items []*WorkingMemoryItem) []string {
	var agentIDs []string
	for _, item := range items {
		if str, ok := item.Content.(string); ok && len(str) > 0 {
			// Simple heuristic: if content looks like an agent ID, include it
			if len(str) < 100 && len(str) > 0 {
				agentIDs = append(agentIDs, str)
			}
		}
	}
	return agentIDs
}

// calculateConfidence calculates confidence based on working memory state
func (cwmc *CognitiveWorkingMemoryComponent) calculateConfidence() float64 {
	cwmc.mu.RLock()
	defer cwmc.mu.RUnlock()

	// Confidence increases with load (more relevant items in WM)
	load := float64(cwmc.workingMemory.Size()) / float64(cwmc.workingMemory.Capacity())
	// Load ranges 0-1, so confidence is directly proportional
	return load * 0.9 // Max 90% confidence from WM alone
}

// Shutdown gracefully shuts down the component
func (cwmc *CognitiveWorkingMemoryComponent) Shutdown() error {
	cwmc.mu.Lock()
	defer cwmc.mu.Unlock()

	// Cleanup if needed
	cwmc.workingMemory = nil
	return nil
}

// GetMetrics returns current performance metrics
func (cwmc *CognitiveWorkingMemoryComponent) GetMetrics() CognitiveMetrics {
	cwmc.mu.RLock()
	defer cwmc.mu.RUnlock()

	metrics := cwmc.metrics
	metrics.TotalRequests = cwmc.requestCount
	metrics.SuccessfulRequests = cwmc.successCount
	metrics.FailedRequests = cwmc.errorCount
	metrics.LastUpdated = time.Now()

	if cwmc.requestCount > 0 {
		metrics.ErrorRate = float64(cwmc.errorCount) / float64(cwmc.requestCount)
	}

	// Add custom metrics
	load := float64(cwmc.workingMemory.Size()) / float64(cwmc.workingMemory.Capacity())
	metrics.CustomMetrics = map[string]interface{}{
		"capacity":   cwmc.workingMemory.Capacity(),
		"load":       load,
		"item_count": cwmc.workingMemory.Size(),
		"decay_rate": cwmc.workingMemory.decayRate,
	}

	return metrics
}

// GetName returns the component's name
func (cwmc *CognitiveWorkingMemoryComponent) GetName() string {
	return "CognitiveWorkingMemory"
}

// ============================================================================
// Helper Methods
// ============================================================================

// GetWorkingMemoryState returns the current working memory state
func (cwmc *CognitiveWorkingMemoryComponent) GetWorkingMemoryState() *CognitiveWorkingMemory {
	cwmc.mu.RLock()
	defer cwmc.mu.RUnlock()
	return cwmc.workingMemory
}

// DecayActivation applies decay to all items in working memory
func (cwmc *CognitiveWorkingMemoryComponent) DecayActivation() {
	cwmc.mu.Lock()
	defer cwmc.mu.Unlock()
	cwmc.workingMemory.applyDecayLocked()
}

// PrimeItem applies spreading activation to an item
func (cwmc *CognitiveWorkingMemoryComponent) PrimeItem(itemID string, strength float64) bool {
	cwmc.mu.Lock()
	defer cwmc.mu.Unlock()

	item, ok := cwmc.workingMemory.items[itemID]
	if !ok {
		return false
	}

	item.Activation += strength
	if item.Activation > 1.0 {
		item.Activation = 1.0
	}
	return true
}

// GetItemByID retrieves a specific item by ID
func (cwmc *CognitiveWorkingMemoryComponent) GetItemByID(itemID string) *WorkingMemoryItem {
	cwmc.mu.RLock()
	defer cwmc.mu.RUnlock()

	item, ok := cwmc.workingMemory.items[itemID]
	if !ok {
		return nil
	}
	return item
}

// ClearWorkingMemory clears all items (useful for resetting)
func (cwmc *CognitiveWorkingMemoryComponent) ClearWorkingMemory() {
	cwmc.mu.Lock()
	defer cwmc.mu.Unlock()

	config := WorkingMemoryConfig{
		Capacity:            cwmc.workingMemory.Capacity(),
		DecayRate:           cwmc.workingMemory.decayRate,
		ActivationThreshold: cwmc.workingMemory.activationThreshold,
		SpreadingFactor:     cwmc.workingMemory.spreadingFactor,
	}
	cwmc.workingMemory = NewCognitiveWorkingMemory(config)
}
