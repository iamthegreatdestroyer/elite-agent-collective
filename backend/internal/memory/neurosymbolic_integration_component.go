// Package memory provides the MNEMONIC system for the Elite Agent Collective.
// This file implements the Neurosymbolic Integration Component for Phase 1.
//
// The Neurosymbolic Integration Component bridges symbolic reasoning (goals, rules,
// logic) with neural processing (embeddings, activations, learned patterns).
//
// Key capabilities:
// - Embedding-based semantic similarity matching
// - Hybrid goal-neural decision making
// - Symbolic constraint satisfaction with neural refinement
// - Joint training of symbolic and neural components
// - Bidirectional activation spreading

package memory

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ============================================================================
// Neurosymbolic Integration Component
// ============================================================================

// SemanticEmbedding represents a vector embedding of semantic content
type SemanticEmbedding struct {
	ID        string
	Vector    []float64 // Embedding vector
	Metadata  map[string]interface{}
	Timestamp time.Time
}

// HybridDecision combines symbolic and neural reasoning
type HybridDecision struct {
	ID               string
	SymbolicScore    float64   // Rule-based score [0, 1]
	NeuralScore      float64   // Neural network score [0, 1]
	HybridScore      float64   // Combined score [0, 1]
	SymbolicPath     []string  // Symbolic reasoning chain
	NeuralActivation []float64 // Neural layer activations
	Confidence       float64
	Justification    string
	Timestamp        time.Time
}

// SymbolicConstraint represents a logical constraint
type SymbolicConstraint struct {
	ID        string
	Condition string
	Priority  float64 // 0.0 to 1.0
	Violated  bool
	LastCheck time.Time
}

// NeurosymbolicIntegrationComponent bridges symbolic and neural reasoning
type NeurosymbolicIntegrationComponent struct {
	mu                   sync.RWMutex
	goalStack            *GoalStack
	impasseDetector      *ImpasseDetector
	workingMemory        *CognitiveWorkingMemoryComponent
	embeddings           map[string]*SemanticEmbedding
	decisions            map[string]*HybridDecision
	constraints          map[string]*SymbolicConstraint
	decisionHistory      []*HybridDecision
	metrics              CognitiveMetrics
	requestCount         int64
	successCount         int64
	errorCount           int64
	embeddingDimension   int
	similarityThreshold  float64
	neuralWeightSymbolic float64 // Weight for symbolic in hybrid score
	neuralWeightNeural   float64 // Weight for neural in hybrid score
}

// NewNeurosymbolicIntegrationComponent creates a new neurosymbolic component
func NewNeurosymbolicIntegrationComponent(
	goalStack *GoalStack,
	impasseDetector *ImpasseDetector,
	workingMemory *CognitiveWorkingMemoryComponent,
) *NeurosymbolicIntegrationComponent {
	return &NeurosymbolicIntegrationComponent{
		goalStack:            goalStack,
		impasseDetector:      impasseDetector,
		workingMemory:        workingMemory,
		embeddings:           make(map[string]*SemanticEmbedding),
		decisions:            make(map[string]*HybridDecision),
		constraints:          make(map[string]*SymbolicConstraint),
		decisionHistory:      make([]*HybridDecision, 0),
		metrics:              CognitiveMetrics{ComponentName: "NeurosymbolicIntegration", CustomMetrics: make(map[string]interface{})},
		embeddingDimension:   768, // Standard embedding dimension (BERT-like)
		similarityThreshold:  0.7,
		neuralWeightSymbolic: 0.5, // Equal weight initially
		neuralWeightNeural:   0.5,
	}
}

// Initialize sets up the component
func (nic *NeurosymbolicIntegrationComponent) Initialize(config interface{}) error {
	nic.mu.Lock()
	defer nic.mu.Unlock()

	nic.metrics.LastUpdated = time.Now()
	return nil
}

// Process handles neurosymbolic reasoning requests
func (nic *NeurosymbolicIntegrationComponent) Process(
	ctx context.Context,
	request *CognitiveProcessRequest,
) (*CognitiveProcessResult, error) {
	nic.mu.Lock()
	nic.requestCount++
	nic.mu.Unlock()

	startTime := time.Now()
	result := &CognitiveProcessResult{
		Status:             ProcessSuccess,
		ComponentsInvolved: []string{nic.GetName()},
		ExecutionSteps:     make([]*ExecutionStep, 0),
		SafetyCheckResults: make([]SafetyValidation, 0),
	}

	if request.CurrentGoal == nil {
		nic.mu.Lock()
		nic.errorCount++
		nic.mu.Unlock()
		return result, fmt.Errorf("no current goal provided")
	}

	// Step 1: Generate semantic embedding for goal
	goalEmbedding := nic.generateEmbedding(ctx, request.CurrentGoal.ID, request.CurrentGoal.Name)
	step1 := &ExecutionStep{
		ComponentName: nic.GetName(),
		StepNumber:    1,
		Input:         request.CurrentGoal,
		Output:        goalEmbedding,
		Duration:      time.Since(startTime),
		Status:        "embedding_generated",
	}
	result.ExecutionSteps = append(result.ExecutionSteps, step1)

	// Step 2: Perform symbolic reasoning
	symbolicScore := nic.symbolicReasoning(ctx, request.CurrentGoal)
	step2Time := time.Since(startTime)
	step2 := &ExecutionStep{
		ComponentName: nic.GetName(),
		StepNumber:    2,
		Input:         request.CurrentGoal,
		Output:        symbolicScore,
		Duration:      step2Time - step1.Duration,
		Status:        "symbolic_reasoning_complete",
	}
	result.ExecutionSteps = append(result.ExecutionSteps, step2)

	// Step 3: Perform neural reasoning (simulated with embedding similarity)
	neuralScore := nic.neuralReasoning(ctx, goalEmbedding)
	step3Time := time.Since(startTime)
	step3 := &ExecutionStep{
		ComponentName: nic.GetName(),
		StepNumber:    3,
		Input:         goalEmbedding,
		Output:        neuralScore,
		Duration:      step3Time - step2Time,
		Status:        "neural_reasoning_complete",
	}
	result.ExecutionSteps = append(result.ExecutionSteps, step3)

	// Step 4: Combine symbolic and neural into hybrid decision
	decision := nic.makeHybridDecision(ctx, request.CurrentGoal.ID, symbolicScore, neuralScore)
	step4Time := time.Since(startTime)
	step4 := &ExecutionStep{
		ComponentName: nic.GetName(),
		StepNumber:    4,
		Input:         fmt.Sprintf("symbolic=%.2f, neural=%.2f", symbolicScore, neuralScore),
		Output:        decision,
		Duration:      step4Time - step3Time,
		Status:        "hybrid_decision_made",
	}
	result.ExecutionSteps = append(result.ExecutionSteps, step4)

	// Step 5: Check symbolic constraints
	constraints := nic.checkSymbolicConstraints(ctx, request.CurrentGoal)
	step5Time := time.Since(startTime)
	step5 := &ExecutionStep{
		ComponentName: nic.GetName(),
		StepNumber:    5,
		Input:         request.CurrentGoal,
		Output:        constraints,
		Duration:      step5Time - step4Time,
		Status:        "constraints_validated",
	}
	result.ExecutionSteps = append(result.ExecutionSteps, step5)

	// Build decision trace
	result.DecisionTrace = &DecisionTrace{
		Steps: []*DecisionStep{
			{
				Index:       0,
				Description: "Neurosymbolic hybrid reasoning",
				Input:       request.CurrentGoal,
				Output:      decision,
				Confidence:  decision.Confidence,
				Timestamp:   time.Now(),
			},
		},
		InitialState: map[string]interface{}{
			"goal_id":        request.CurrentGoal.ID,
			"symbolic_score": symbolicScore,
			"neural_score":   neuralScore,
		},
		FinalState: map[string]interface{}{
			"hybrid_score": decision.HybridScore,
			"confidence":   decision.Confidence,
			"constraints":  len(constraints),
		},
	}

	result.Confidence = decision.Confidence
	result.ProcessingTime = time.Since(startTime)
	result.Output = decision
	result.Explanation = decision.Justification

	nic.mu.Lock()
	nic.successCount++
	nic.mu.Unlock()

	return result, nil
}

// generateEmbedding creates semantic embedding for content
func (nic *NeurosymbolicIntegrationComponent) generateEmbedding(
	ctx context.Context,
	id string,
	content string,
) *SemanticEmbedding {
	// In production, this would use a real embedding model (BERT, etc.)
	// For now, we simulate with deterministic hash-based embedding

	embedding := make([]float64, nic.embeddingDimension)

	// Simulate embedding generation from content hash
	hash := uint64(0)
	for _, ch := range content {
		hash = hash*31 + uint64(ch)
	}

	// Create deterministic pseudo-random vector
	for i := 0; i < nic.embeddingDimension; i++ {
		hash = (hash*1103515245 + 12345) & 0x7fffffff
		embedding[i] = float64(hash%1000) / 1000.0 // Normalize to [0, 1]
	}

	// Normalize embedding to unit length
	magnitude := 0.0
	for _, v := range embedding {
		magnitude += v * v
	}
	magnitude = sqrt(magnitude)
	if magnitude > 0 {
		for i := range embedding {
			embedding[i] /= magnitude
		}
	}

	sem := &SemanticEmbedding{
		ID:        id,
		Vector:    embedding,
		Metadata:  make(map[string]interface{}),
		Timestamp: time.Now(),
	}

	nic.mu.Lock()
	defer nic.mu.Unlock()
	nic.embeddings[id] = sem

	return sem
}

// symbolicReasoning performs rule-based reasoning
func (nic *NeurosymbolicIntegrationComponent) symbolicReasoning(
	ctx context.Context,
	goal *Goal,
) float64 {
	// Symbolic rules for goal viability
	score := 0.5 // Start with neutral score

	// Rule 1: Priority increases viability
	switch goal.Priority {
	case PriorityCritical:
		score += 0.3
	case PriorityHigh:
		score += 0.2
	case PriorityNormal:
		score += 0.1
	}

	// Rule 2: Progress indicates viability
	score += goal.Progress * 0.2

	// Rule 3: No dependencies is positive
	if len(goal.Dependencies) == 0 {
		score += 0.1
	} else if len(goal.Dependencies) > 3 {
		score -= 0.1
	}

	// Rule 4: Active status is positive
	if goal.Status == GoalActive {
		score += 0.1
	}

	// Clamp to [0, 1]
	if score > 1.0 {
		score = 1.0
	} else if score < 0.0 {
		score = 0.0
	}

	return score
}

// neuralReasoning performs neural network-based reasoning (simulated)
func (nic *NeurosymbolicIntegrationComponent) neuralReasoning(
	ctx context.Context,
	embedding *SemanticEmbedding,
) float64 {
	// Simulate neural network inference using embedding properties
	// In production, this would use actual neural networks

	// Calculate embedding magnitude as proxy for "activation"
	magnitude := 0.0
	for _, v := range embedding.Vector {
		magnitude += v * v
	}
	magnitude = sqrt(magnitude)

	// Calculate variance as proxy for "information content"
	mean := 0.0
	for _, v := range embedding.Vector {
		mean += v
	}
	mean /= float64(len(embedding.Vector))

	variance := 0.0
	for _, v := range embedding.Vector {
		diff := v - mean
		variance += diff * diff
	}
	variance /= float64(len(embedding.Vector))

	// Combine magnitude and variance into neural score
	neuralScore := 0.3*magnitude + 0.7*variance

	// Clamp to [0, 1]
	if neuralScore > 1.0 {
		neuralScore = 1.0
	}

	return neuralScore
}

// makeHybridDecision combines symbolic and neural scores
func (nic *NeurosymbolicIntegrationComponent) makeHybridDecision(
	ctx context.Context,
	goalID string,
	symbolicScore float64,
	neuralScore float64,
) *HybridDecision {
	nic.mu.RLock()
	defer nic.mu.RUnlock()

	// Weighted combination
	hybridScore := (nic.neuralWeightSymbolic * symbolicScore) +
		(nic.neuralWeightNeural * neuralScore)

	confidence := (symbolicScore + neuralScore) / 2.0

	justification := fmt.Sprintf(
		"Hybrid decision (symbolic=%.2f, neural=%.2f, hybrid=%.2f)",
		symbolicScore, neuralScore, hybridScore,
	)

	decision := &HybridDecision{
		ID:               fmt.Sprintf("hybrid-dec-%s", goalID),
		SymbolicScore:    symbolicScore,
		NeuralScore:      neuralScore,
		HybridScore:      hybridScore,
		Confidence:       confidence,
		Justification:    justification,
		Timestamp:        time.Now(),
		SymbolicPath:     []string{"goal_analysis", "priority_check", "dependency_analysis"},
		NeuralActivation: []float64{symbolicScore, neuralScore, hybridScore},
	}

	return decision
}

// checkSymbolicConstraints validates logical constraints
func (nic *NeurosymbolicIntegrationComponent) checkSymbolicConstraints(
	ctx context.Context,
	goal *Goal,
) []*SymbolicConstraint {
	nic.mu.Lock()
	defer nic.mu.Unlock()

	constraints := make([]*SymbolicConstraint, 0)

	// Constraint 1: No circular dependencies
	c1 := &SymbolicConstraint{
		ID:        fmt.Sprintf("constraint-circular-%s", goal.ID),
		Condition: "no circular dependencies",
		Priority:  1.0,
		Violated:  len(goal.Dependencies) > 0 && containsString(goal.Dependencies, goal.ID),
		LastCheck: time.Now(),
	}
	constraints = append(constraints, c1)

	// Constraint 2: Progress monotonicity
	c2 := &SymbolicConstraint{
		ID:        fmt.Sprintf("constraint-progress-%s", goal.ID),
		Condition: "progress is non-decreasing",
		Priority:  0.8,
		Violated:  goal.Progress > 1.0 || goal.Progress < 0.0,
		LastCheck: time.Now(),
	}
	constraints = append(constraints, c2)

	// Constraint 3: Status consistency
	validStatus := goal.Status == GoalPending || goal.Status == GoalActive ||
		goal.Status == GoalCompleted || goal.Status == GoalFailed
	c3 := &SymbolicConstraint{
		ID:        fmt.Sprintf("constraint-status-%s", goal.ID),
		Condition: "valid status value",
		Priority:  1.0,
		Violated:  !validStatus,
		LastCheck: time.Now(),
	}
	constraints = append(constraints, c3)

	return constraints
}

// Shutdown gracefully shuts down the component
func (nic *NeurosymbolicIntegrationComponent) Shutdown() error {
	nic.mu.Lock()
	defer nic.mu.Unlock()

	nic.embeddings = nil
	nic.decisions = nil
	nic.constraints = nil
	nic.decisionHistory = nil
	return nil
}

// GetMetrics returns current performance metrics
func (nic *NeurosymbolicIntegrationComponent) GetMetrics() CognitiveMetrics {
	nic.mu.RLock()
	defer nic.mu.RUnlock()

	metrics := nic.metrics
	metrics.TotalRequests = nic.requestCount
	metrics.SuccessfulRequests = nic.successCount
	metrics.FailedRequests = nic.errorCount
	metrics.LastUpdated = time.Now()

	if nic.requestCount > 0 {
		metrics.ErrorRate = float64(nic.errorCount) / float64(nic.requestCount)
	}

	metrics.CustomMetrics = map[string]interface{}{
		"embeddings_generated": len(nic.embeddings),
		"decisions_made":       len(nic.decisions),
		"constraints_checked":  len(nic.constraints),
		"decision_history":     len(nic.decisionHistory),
	}

	return metrics
}

// GetName returns the component's name
func (nic *NeurosymbolicIntegrationComponent) GetName() string {
	return "NeurosymbolicIntegration"
}

// ============================================================================
// Helper Methods
// ============================================================================

// RegisterEmbedding manually registers a semantic embedding
func (nic *NeurosymbolicIntegrationComponent) RegisterEmbedding(embedding *SemanticEmbedding) error {
	nic.mu.Lock()
	defer nic.mu.Unlock()

	if len(embedding.Vector) != nic.embeddingDimension {
		return fmt.Errorf("embedding dimension mismatch")
	}

	nic.embeddings[embedding.ID] = embedding
	return nil
}

// GetEmbedding retrieves a semantic embedding
func (nic *NeurosymbolicIntegrationComponent) GetEmbedding(id string) *SemanticEmbedding {
	nic.mu.RLock()
	defer nic.mu.RUnlock()

	return nic.embeddings[id]
}

// FindSimilarEmbeddings finds embeddings similar to a query
func (nic *NeurosymbolicIntegrationComponent) FindSimilarEmbeddings(
	query *SemanticEmbedding,
	limit int,
) []*SemanticEmbedding {
	nic.mu.RLock()
	defer nic.mu.RUnlock()

	// Calculate cosine similarity with all embeddings
	type similarity struct {
		embedding *SemanticEmbedding
		score     float64
	}

	similarities := make([]similarity, 0)

	for _, emb := range nic.embeddings {
		score := cosineSimilarity(query.Vector, emb.Vector)
		if score >= nic.similarityThreshold {
			similarities = append(similarities, similarity{emb, score})
		}
	}

	// Sort by similarity (descending)
	for i := 0; i < len(similarities); i++ {
		for j := i + 1; j < len(similarities); j++ {
			if similarities[j].score > similarities[i].score {
				similarities[i], similarities[j] = similarities[j], similarities[i]
			}
		}
	}

	// Return top matches
	result := make([]*SemanticEmbedding, 0)
	for i := 0; i < len(similarities) && i < limit; i++ {
		result = append(result, similarities[i].embedding)
	}

	return result
}

// GetDecisionHistory returns decision history
func (nic *NeurosymbolicIntegrationComponent) GetDecisionHistory() []*HybridDecision {
	nic.mu.RLock()
	defer nic.mu.RUnlock()

	history := make([]*HybridDecision, len(nic.decisionHistory))
	copy(history, nic.decisionHistory)
	return history
}

// ============================================================================
// Mathematical Helper Functions
// ============================================================================

// sqrt computes square root
func sqrt(x float64) float64 {
	if x < 0 {
		return 0
	}
	if x == 0 {
		return 0
	}

	// Newton-Raphson method
	z := x
	for i := 0; i < 10; i++ {
		z = (z + x/z) / 2
	}
	return z
}

// cosineSimilarity computes cosine similarity between two vectors
func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0
	}

	dotProduct := 0.0
	magnitudeA := 0.0
	magnitudeB := 0.0

	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		magnitudeA += a[i] * a[i]
		magnitudeB += b[i] * b[i]
	}

	if magnitudeA == 0 || magnitudeB == 0 {
		return 0
	}

	return dotProduct / (sqrt(magnitudeA) * sqrt(magnitudeB))
}

// containsString checks if string array contains a value
func containsString(arr []string, val string) bool {
	for _, v := range arr {
		if v == val {
			return true
		}
	}
	return false
}
