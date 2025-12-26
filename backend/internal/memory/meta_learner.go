// Package memory implements the MNEMONIC memory system for the Elite Agent Collective.
// This file implements Phase 3: Meta-Learning Enhancement for fast adaptation.
package memory

import (
	"context"
	"errors"
	"math"
	"sort"
	"sync"
	"time"
)

// =============================================================================
// MAML-STYLE FAST ADAPTATION
// =============================================================================

// MetaLearner implements Model-Agnostic Meta-Learning (MAML) style adaptation
// for agents to quickly adapt to new task distributions.
type MetaLearner struct {
	mu sync.RWMutex

	// Base parameters that serve as initialization for adaptation
	baseParameters map[string]*AgentParameters // AgentID -> base params

	// Adaptation configuration
	adaptationSteps  int     // Number of gradient steps during adaptation
	adaptationRate   float64 // Learning rate for inner loop
	metaLearningRate float64 // Learning rate for outer loop (meta-update)
	supportSetSize   int     // Number of examples for few-shot learning
	querySetSize     int     // Number of examples for evaluation

	// Memory integration
	experienceStore ExperienceRetriever

	// Task distribution tracking
	taskDistributions map[string]*TaskDistribution // AgentID -> distribution

	// Metrics
	adaptationMetrics *AdaptationMetrics
}

// AgentParameters represents learnable parameters for an agent
type AgentParameters struct {
	ID               string
	StrategyWeights  []float64          // Weights for different strategies
	ContextBias      []float64          // Bias for context interpretation
	ResponsePatterns map[string]float64 // Pattern -> weight
	LastUpdated      time.Time
	Version          int
}

// TaskDistribution tracks the distribution of tasks for an agent
type TaskDistribution struct {
	AgentID        string
	TaskEmbeddings [][]float64
	TaskClusters   []*TaskCluster
	Mean           []float64
	Variance       []float64
	SampleCount    int
	LastUpdated    time.Time
}

// TaskCluster represents a cluster of similar tasks
type TaskCluster struct {
	ID        string
	Centroid  []float64
	Members   []string // Task IDs
	Weight    float64
	Frequency int
}

// AdaptedAgent represents an agent with adapted parameters
type AdaptedAgent struct {
	BaseAgentID    string
	Parameters     *AgentParameters
	SupportSet     []*Example
	AdaptationTime time.Duration
	Confidence     float64
}

// Example represents a task-solution pair for meta-learning
type Example struct {
	ID             string
	TaskInput      interface{}
	TaskEmbedding  []float64
	ExpectedOutput interface{}
	ActualOutput   interface{}
	Quality        float64
	AgentID        string
	Timestamp      time.Time
}

// AdaptationMetrics tracks meta-learning performance
type AdaptationMetrics struct {
	mu sync.RWMutex

	TotalAdaptations    int
	SuccessfulAdaptions int
	AverageAdaptTime    time.Duration
	FewShotAccuracy     map[int]float64 // K -> accuracy for K-shot
	TaskTransferScore   float64
}

// MetaLearnerConfig holds configuration for the meta-learner
type MetaLearnerConfig struct {
	AdaptationSteps  int
	AdaptationRate   float64
	MetaLearningRate float64
	SupportSetSize   int
	QuerySetSize     int
}

// DefaultMetaLearnerConfig returns sensible defaults
func DefaultMetaLearnerConfig() *MetaLearnerConfig {
	return &MetaLearnerConfig{
		AdaptationSteps:  5,
		AdaptationRate:   0.01,
		MetaLearningRate: 0.001,
		SupportSetSize:   5,
		QuerySetSize:     10,
	}
}

// ExperienceRetriever interface for accessing experiences
type ExperienceRetriever interface {
	RetrieveSimilar(ctx context.Context, embedding []float64, limit int) ([]*ExperienceTuple, error)
	RetrieveByAgent(ctx context.Context, agentID string, limit int) ([]*ExperienceTuple, error)
}

// NewMetaLearner creates a new MAML-style meta-learner
func NewMetaLearner(config *MetaLearnerConfig, store ExperienceRetriever) *MetaLearner {
	if config == nil {
		config = DefaultMetaLearnerConfig()
	}

	return &MetaLearner{
		baseParameters:    make(map[string]*AgentParameters),
		adaptationSteps:   config.AdaptationSteps,
		adaptationRate:    config.AdaptationRate,
		metaLearningRate:  config.MetaLearningRate,
		supportSetSize:    config.SupportSetSize,
		querySetSize:      config.QuerySetSize,
		experienceStore:   store,
		taskDistributions: make(map[string]*TaskDistribution),
		adaptationMetrics: &AdaptationMetrics{
			FewShotAccuracy: make(map[int]float64),
		},
	}
}

// InitializeAgent sets up base parameters for an agent
func (m *MetaLearner) InitializeAgent(agentID string, initialParams *AgentParameters) error {
	if agentID == "" {
		return errors.New("agent ID required")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if initialParams == nil {
		// Create default parameters
		initialParams = &AgentParameters{
			ID:               agentID,
			StrategyWeights:  make([]float64, 10), // 10 strategy dimensions
			ContextBias:      make([]float64, 64), // 64-dim context bias
			ResponsePatterns: make(map[string]float64),
			LastUpdated:      time.Now(),
			Version:          1,
		}

		// Initialize with small random values for symmetry breaking
		for i := range initialParams.StrategyWeights {
			initialParams.StrategyWeights[i] = 0.1 * (float64(i%5) - 2.0) / 10.0
		}
		for i := range initialParams.ContextBias {
			initialParams.ContextBias[i] = 0.01 * float64(i%10-5) / 10.0
		}
	}

	m.baseParameters[agentID] = initialParams
	m.taskDistributions[agentID] = &TaskDistribution{
		AgentID:        agentID,
		TaskEmbeddings: make([][]float64, 0),
		TaskClusters:   make([]*TaskCluster, 0),
		LastUpdated:    time.Now(),
	}

	return nil
}

// Adapt performs fast adaptation given a support set (few-shot learning)
func (m *MetaLearner) Adapt(ctx context.Context, agentID string, supportSet []*Example) (*AdaptedAgent, error) {
	startTime := time.Now()

	m.mu.RLock()
	baseParams, exists := m.baseParameters[agentID]
	m.mu.RUnlock()

	if !exists {
		// Auto-initialize with defaults
		if err := m.InitializeAgent(agentID, nil); err != nil {
			return nil, err
		}
		m.mu.RLock()
		baseParams = m.baseParameters[agentID]
		m.mu.RUnlock()
	}

	// Clone base parameters for adaptation
	adaptedParams := m.cloneParameters(baseParams)

	// Perform adaptation steps (inner loop of MAML)
	for step := 0; step < m.adaptationSteps; step++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Compute gradient on support set
		gradient := m.computeGradient(adaptedParams, supportSet)

		// Update parameters
		adaptedParams = m.applyGradient(adaptedParams, gradient, m.adaptationRate)
	}

	// Compute adaptation confidence based on support set quality
	confidence := m.computeAdaptationConfidence(supportSet, adaptedParams)

	// Update metrics
	m.adaptationMetrics.mu.Lock()
	m.adaptationMetrics.TotalAdaptations++
	adaptTime := time.Since(startTime)
	if m.adaptationMetrics.AverageAdaptTime == 0 {
		m.adaptationMetrics.AverageAdaptTime = adaptTime
	} else {
		m.adaptationMetrics.AverageAdaptTime = (m.adaptationMetrics.AverageAdaptTime + adaptTime) / 2
	}
	m.adaptationMetrics.mu.Unlock()

	return &AdaptedAgent{
		BaseAgentID:    agentID,
		Parameters:     adaptedParams,
		SupportSet:     supportSet,
		AdaptationTime: adaptTime,
		Confidence:     confidence,
	}, nil
}

// MetaUpdate performs the outer loop update (meta-learning across tasks)
func (m *MetaLearner) MetaUpdate(ctx context.Context, agentID string, tasks []*MetaTask) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	baseParams, exists := m.baseParameters[agentID]
	if !exists {
		return errors.New("agent not initialized for meta-learning")
	}

	// Accumulate meta-gradient across tasks
	metaGradient := m.initializeGradient(baseParams)

	for _, task := range tasks {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Fast adaptation on support set (inner loop)
		adaptedParams := m.cloneParameters(baseParams)
		for step := 0; step < m.adaptationSteps; step++ {
			gradient := m.computeGradient(adaptedParams, task.SupportSet)
			adaptedParams = m.applyGradient(adaptedParams, gradient, m.adaptationRate)
		}

		// Evaluate on query set
		queryLoss := m.evaluateLoss(adaptedParams, task.QuerySet)

		// Compute gradient of query loss w.r.t. base parameters
		taskGradient := m.computeMetaGradient(baseParams, adaptedParams, queryLoss, task)
		metaGradient = m.addGradients(metaGradient, taskGradient)
	}

	// Average the gradients
	if len(tasks) > 0 {
		m.scaleGradient(metaGradient, 1.0/float64(len(tasks)))
	}

	// Update base parameters (outer loop)
	m.baseParameters[agentID] = m.applyGradient(baseParams, metaGradient, m.metaLearningRate)
	m.baseParameters[agentID].LastUpdated = time.Now()
	m.baseParameters[agentID].Version++

	return nil
}

// MetaTask represents a task for meta-learning with support and query sets
type MetaTask struct {
	ID         string
	SupportSet []*Example
	QuerySet   []*Example
	TaskType   string
	Difficulty float64
}

// cloneParameters creates a deep copy of agent parameters
func (m *MetaLearner) cloneParameters(p *AgentParameters) *AgentParameters {
	clone := &AgentParameters{
		ID:               p.ID,
		StrategyWeights:  make([]float64, len(p.StrategyWeights)),
		ContextBias:      make([]float64, len(p.ContextBias)),
		ResponsePatterns: make(map[string]float64),
		LastUpdated:      p.LastUpdated,
		Version:          p.Version,
	}
	copy(clone.StrategyWeights, p.StrategyWeights)
	copy(clone.ContextBias, p.ContextBias)
	for k, v := range p.ResponsePatterns {
		clone.ResponsePatterns[k] = v
	}
	return clone
}

// ParameterGradient represents gradients for agent parameters
type ParameterGradient struct {
	StrategyGradient []float64
	ContextGradient  []float64
	PatternGradient  map[string]float64
}

// computeGradient computes the gradient of the loss on examples
func (m *MetaLearner) computeGradient(params *AgentParameters, examples []*Example) *ParameterGradient {
	gradient := &ParameterGradient{
		StrategyGradient: make([]float64, len(params.StrategyWeights)),
		ContextGradient:  make([]float64, len(params.ContextBias)),
		PatternGradient:  make(map[string]float64),
	}

	if len(examples) == 0 {
		return gradient
	}

	// Compute loss and gradient for each example
	for _, ex := range examples {
		// Predict using current parameters
		prediction := m.predict(params, ex)

		// Compute error signal
		errorSignal := ex.Quality - prediction

		// Backprop to strategy weights (simplified gradient)
		for i := range gradient.StrategyGradient {
			if i < len(ex.TaskEmbedding) {
				gradient.StrategyGradient[i] += errorSignal * ex.TaskEmbedding[i]
			}
		}

		// Backprop to context bias
		for i := range gradient.ContextGradient {
			if i < len(ex.TaskEmbedding) {
				gradient.ContextGradient[i] += errorSignal * ex.TaskEmbedding[i] * 0.1
			}
		}
	}

	// Average gradients
	n := float64(len(examples))
	for i := range gradient.StrategyGradient {
		gradient.StrategyGradient[i] /= n
	}
	for i := range gradient.ContextGradient {
		gradient.ContextGradient[i] /= n
	}

	return gradient
}

// predict makes a prediction given parameters and an example
func (m *MetaLearner) predict(params *AgentParameters, ex *Example) float64 {
	// Simple linear prediction model
	prediction := 0.0

	// Strategy contribution
	for i, w := range params.StrategyWeights {
		if i < len(ex.TaskEmbedding) {
			prediction += w * ex.TaskEmbedding[i]
		}
	}

	// Context bias contribution
	for i, b := range params.ContextBias {
		if i < len(ex.TaskEmbedding) {
			prediction += b * ex.TaskEmbedding[i] * 0.1
		}
	}

	// Sigmoid activation for bounded output
	return 1.0 / (1.0 + math.Exp(-prediction))
}

// applyGradient updates parameters using the gradient
func (m *MetaLearner) applyGradient(params *AgentParameters, gradient *ParameterGradient, lr float64) *AgentParameters {
	updated := m.cloneParameters(params)

	for i := range updated.StrategyWeights {
		if i < len(gradient.StrategyGradient) {
			updated.StrategyWeights[i] += lr * gradient.StrategyGradient[i]
		}
	}

	for i := range updated.ContextBias {
		if i < len(gradient.ContextGradient) {
			updated.ContextBias[i] += lr * gradient.ContextGradient[i]
		}
	}

	for pattern, grad := range gradient.PatternGradient {
		updated.ResponsePatterns[pattern] += lr * grad
	}

	return updated
}

// initializeGradient creates a zero gradient
func (m *MetaLearner) initializeGradient(params *AgentParameters) *ParameterGradient {
	return &ParameterGradient{
		StrategyGradient: make([]float64, len(params.StrategyWeights)),
		ContextGradient:  make([]float64, len(params.ContextBias)),
		PatternGradient:  make(map[string]float64),
	}
}

// addGradients combines two gradients
func (m *MetaLearner) addGradients(g1, g2 *ParameterGradient) *ParameterGradient {
	result := &ParameterGradient{
		StrategyGradient: make([]float64, len(g1.StrategyGradient)),
		ContextGradient:  make([]float64, len(g1.ContextGradient)),
		PatternGradient:  make(map[string]float64),
	}

	for i := range g1.StrategyGradient {
		result.StrategyGradient[i] = g1.StrategyGradient[i]
		if i < len(g2.StrategyGradient) {
			result.StrategyGradient[i] += g2.StrategyGradient[i]
		}
	}

	for i := range g1.ContextGradient {
		result.ContextGradient[i] = g1.ContextGradient[i]
		if i < len(g2.ContextGradient) {
			result.ContextGradient[i] += g2.ContextGradient[i]
		}
	}

	for k, v := range g1.PatternGradient {
		result.PatternGradient[k] = v
	}
	for k, v := range g2.PatternGradient {
		result.PatternGradient[k] += v
	}

	return result
}

// scaleGradient multiplies gradient by a scalar
func (m *MetaLearner) scaleGradient(g *ParameterGradient, scale float64) {
	for i := range g.StrategyGradient {
		g.StrategyGradient[i] *= scale
	}
	for i := range g.ContextGradient {
		g.ContextGradient[i] *= scale
	}
	for k := range g.PatternGradient {
		g.PatternGradient[k] *= scale
	}
}

// evaluateLoss computes loss on a set of examples
func (m *MetaLearner) evaluateLoss(params *AgentParameters, examples []*Example) float64 {
	if len(examples) == 0 {
		return 0
	}

	totalLoss := 0.0
	for _, ex := range examples {
		pred := m.predict(params, ex)
		// MSE loss
		diff := pred - ex.Quality
		totalLoss += diff * diff
	}

	return totalLoss / float64(len(examples))
}

// computeMetaGradient computes gradient for the outer loop
func (m *MetaLearner) computeMetaGradient(baseParams, adaptedParams *AgentParameters, queryLoss float64, task *MetaTask) *ParameterGradient {
	// Compute how changes to base parameters affect query loss
	// This is an approximation of the true MAML gradient

	gradient := &ParameterGradient{
		StrategyGradient: make([]float64, len(baseParams.StrategyWeights)),
		ContextGradient:  make([]float64, len(baseParams.ContextBias)),
		PatternGradient:  make(map[string]float64),
	}

	// Approximation: gradient is proportional to difference between adapted and base
	// scaled by query loss
	for i := range gradient.StrategyGradient {
		if i < len(adaptedParams.StrategyWeights) {
			diff := adaptedParams.StrategyWeights[i] - baseParams.StrategyWeights[i]
			gradient.StrategyGradient[i] = -queryLoss * diff
		}
	}

	for i := range gradient.ContextGradient {
		if i < len(adaptedParams.ContextBias) {
			diff := adaptedParams.ContextBias[i] - baseParams.ContextBias[i]
			gradient.ContextGradient[i] = -queryLoss * diff
		}
	}

	return gradient
}

// computeAdaptationConfidence estimates how reliable the adaptation is
func (m *MetaLearner) computeAdaptationConfidence(supportSet []*Example, params *AgentParameters) float64 {
	if len(supportSet) == 0 {
		return 0.0
	}

	// Factors: support set size, quality variance, prediction accuracy
	avgQuality := 0.0
	qualityVariance := 0.0

	for _, ex := range supportSet {
		avgQuality += ex.Quality
	}
	avgQuality /= float64(len(supportSet))

	for _, ex := range supportSet {
		diff := ex.Quality - avgQuality
		qualityVariance += diff * diff
	}
	qualityVariance /= float64(len(supportSet))

	// Confidence based on support set size (diminishing returns)
	sizeConfidence := 1.0 - math.Exp(-float64(len(supportSet))/5.0)

	// Confidence based on quality consistency
	consistencyConfidence := math.Exp(-qualityVariance)

	// Confidence based on prediction accuracy
	predLoss := m.evaluateLoss(params, supportSet)
	accuracyConfidence := math.Exp(-predLoss)

	return (sizeConfidence + consistencyConfidence + accuracyConfidence) / 3.0
}

// GetBaseParameters returns the current base parameters for an agent
func (m *MetaLearner) GetBaseParameters(agentID string) (*AgentParameters, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	params, exists := m.baseParameters[agentID]
	if exists {
		return m.cloneParameters(params), true
	}
	return nil, false
}

// GetMetrics returns adaptation metrics
func (m *MetaLearner) GetMetrics() *AdaptationMetrics {
	m.adaptationMetrics.mu.RLock()
	defer m.adaptationMetrics.mu.RUnlock()

	return &AdaptationMetrics{
		TotalAdaptations:    m.adaptationMetrics.TotalAdaptations,
		SuccessfulAdaptions: m.adaptationMetrics.SuccessfulAdaptions,
		AverageAdaptTime:    m.adaptationMetrics.AverageAdaptTime,
		FewShotAccuracy:     m.adaptationMetrics.FewShotAccuracy,
		TaskTransferScore:   m.adaptationMetrics.TaskTransferScore,
	}
}

// =============================================================================
// PROTOTYPICAL ROUTER
// =============================================================================

// PrototypicalRouter uses prototype-based few-shot learning for agent selection
type PrototypicalRouter struct {
	mu sync.RWMutex

	// Prototypes: agent -> prototype embedding
	agentPrototypes map[string][]float64

	// Prototype metadata
	prototypeStats map[string]*PrototypeStats

	// Configuration
	embeddingDim   int
	distanceMetric DistanceMetric
	updateMomentum float64 // Momentum for online updates
	minExamplesReq int     // Minimum examples to update prototype
}

// PrototypeStats tracks statistics for a prototype
type PrototypeStats struct {
	AgentID     string
	SampleCount int
	LastUpdated time.Time
	Variance    float64
	UsageCount  int
	SuccessRate float64
}

// DistanceMetric defines how distances are computed
type DistanceMetric int

const (
	EuclideanDistance DistanceMetric = iota
	CosineDistance
	ManhattanDistance
)

// NewPrototypicalRouter creates a prototype-based router
func NewPrototypicalRouter(embeddingDim int, metric DistanceMetric) *PrototypicalRouter {
	return &PrototypicalRouter{
		agentPrototypes: make(map[string][]float64),
		prototypeStats:  make(map[string]*PrototypeStats),
		embeddingDim:    embeddingDim,
		distanceMetric:  metric,
		updateMomentum:  0.1,
		minExamplesReq:  3,
	}
}

// UpdatePrototype updates an agent's prototype from examples
func (r *PrototypicalRouter) UpdatePrototype(agentID string, examples []*Example) error {
	if len(examples) < r.minExamplesReq {
		return errors.New("insufficient examples for prototype update")
	}

	// Compute mean embedding
	meanEmbedding := make([]float64, r.embeddingDim)
	validCount := 0

	for _, ex := range examples {
		if len(ex.TaskEmbedding) >= r.embeddingDim {
			for i := 0; i < r.embeddingDim; i++ {
				meanEmbedding[i] += ex.TaskEmbedding[i]
			}
			validCount++
		}
	}

	if validCount == 0 {
		return errors.New("no valid embeddings in examples")
	}

	for i := range meanEmbedding {
		meanEmbedding[i] /= float64(validCount)
	}

	// Compute variance for confidence estimation
	variance := 0.0
	for _, ex := range examples {
		if len(ex.TaskEmbedding) >= r.embeddingDim {
			dist := r.computeDistance(ex.TaskEmbedding[:r.embeddingDim], meanEmbedding)
			variance += dist * dist
		}
	}
	variance /= float64(validCount)

	r.mu.Lock()
	defer r.mu.Unlock()

	// Update with momentum if prototype exists
	if existing, ok := r.agentPrototypes[agentID]; ok {
		for i := range meanEmbedding {
			meanEmbedding[i] = r.updateMomentum*meanEmbedding[i] + (1-r.updateMomentum)*existing[i]
		}
	}

	r.agentPrototypes[agentID] = meanEmbedding
	r.prototypeStats[agentID] = &PrototypeStats{
		AgentID:     agentID,
		SampleCount: validCount,
		LastUpdated: time.Now(),
		Variance:    variance,
	}

	return nil
}

// Route selects the best agent for a task based on prototype distance
func (r *PrototypicalRouter) Route(taskEmbedding []float64) (string, float64, error) {
	if len(taskEmbedding) < r.embeddingDim {
		return "", 0, errors.New("task embedding too short")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.agentPrototypes) == 0 {
		return "", 0, errors.New("no prototypes available")
	}

	bestAgent := ""
	bestDistance := math.MaxFloat64

	for agent, prototype := range r.agentPrototypes {
		dist := r.computeDistance(taskEmbedding[:r.embeddingDim], prototype)
		if dist < bestDistance {
			bestDistance = dist
			bestAgent = agent
		}
	}

	// Convert distance to confidence (inverse relationship)
	confidence := math.Exp(-bestDistance)

	// Update usage stats
	if stats, ok := r.prototypeStats[bestAgent]; ok {
		stats.UsageCount++
	}

	return bestAgent, confidence, nil
}

// RouteTopK returns the top K agents sorted by distance
func (r *PrototypicalRouter) RouteTopK(taskEmbedding []float64, k int) ([]RouteCandidate, error) {
	if len(taskEmbedding) < r.embeddingDim {
		return nil, errors.New("task embedding too short")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	candidates := make([]RouteCandidate, 0, len(r.agentPrototypes))

	for agent, prototype := range r.agentPrototypes {
		dist := r.computeDistance(taskEmbedding[:r.embeddingDim], prototype)
		candidates = append(candidates, RouteCandidate{
			AgentID:    agent,
			Distance:   dist,
			Confidence: math.Exp(-dist),
		})
	}

	// Sort by distance (ascending)
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Distance < candidates[j].Distance
	})

	if k > len(candidates) {
		k = len(candidates)
	}

	return candidates[:k], nil
}

// RouteCandidate represents an agent candidate with routing score
type RouteCandidate struct {
	AgentID    string
	Distance   float64
	Confidence float64
}

// computeDistance computes distance between two embeddings
func (r *PrototypicalRouter) computeDistance(a, b []float64) float64 {
	switch r.distanceMetric {
	case CosineDistance:
		return r.cosineDistance(a, b)
	case ManhattanDistance:
		return r.manhattanDistance(a, b)
	default:
		return r.euclideanDistance(a, b)
	}
}

func (r *PrototypicalRouter) euclideanDistance(a, b []float64) float64 {
	sum := 0.0
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}
	for i := 0; i < minLen; i++ {
		diff := a[i] - b[i]
		sum += diff * diff
	}
	return math.Sqrt(sum)
}

func (r *PrototypicalRouter) cosineDistance(a, b []float64) float64 {
	dotProduct := 0.0
	normA := 0.0
	normB := 0.0

	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}

	for i := 0; i < minLen; i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 1.0 // Maximum distance
	}

	similarity := dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
	return 1.0 - similarity // Convert similarity to distance
}

func (r *PrototypicalRouter) manhattanDistance(a, b []float64) float64 {
	sum := 0.0
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}
	for i := 0; i < minLen; i++ {
		sum += math.Abs(a[i] - b[i])
	}
	return sum
}

// GetPrototype returns the prototype for an agent
func (r *PrototypicalRouter) GetPrototype(agentID string) ([]float64, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	proto, ok := r.agentPrototypes[agentID]
	if !ok {
		return nil, false
	}

	// Return a copy
	result := make([]float64, len(proto))
	copy(result, proto)
	return result, true
}

// GetStats returns prototype statistics
func (r *PrototypicalRouter) GetStats(agentID string) (*PrototypeStats, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stats, ok := r.prototypeStats[agentID]
	return stats, ok
}

// ListAgents returns all agents with prototypes
func (r *PrototypicalRouter) ListAgents() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agents := make([]string, 0, len(r.agentPrototypes))
	for agent := range r.agentPrototypes {
		agents = append(agents, agent)
	}
	return agents
}
