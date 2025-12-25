// Package memory provides the MNEMONIC system for the Elite Agent Collective.
// This file implements the Self-Model from @NEURAL's Cognitive Architecture Analysis.
//
// The Self-Model enables metacognition - agents knowing what they know and don't know:
// - Capability awareness: What tasks can this agent handle well?
// - Limitation awareness: What are the agent's weaknesses?
// - Uncertainty estimation: How confident is the agent in its predictions?
// - Performance tracking: Historical success/failure patterns
//
// This is crucial for proper task routing, knowing when to ask for help,
// and avoiding overconfidence in uncertain situations.

package memory

import (
	"fmt"
	"math"
	"sort"
	"sync"
	"time"
)

// ============================================================================
// Core Types
// ============================================================================

// SelfModel represents an agent's metacognitive self-awareness.
type SelfModel struct {
	AgentID            string
	Tier               int
	CapabilityModel    *CapabilityModel
	LimitationModel    *LimitationModel
	UncertaintyModel   *UncertaintyModel
	PerformanceTracker *PerformanceTracker
	mu                 sync.RWMutex
}

// CapabilityModel tracks an agent's abilities.
type CapabilityModel struct {
	capabilities map[string]*Capability
	mu           sync.RWMutex
}

// Capability represents a specific agent capability.
type Capability struct {
	Name        string
	Confidence  float64 // 0.0 to 1.0
	UsageCount  int64
	SuccessRate float64
	LastUsed    time.Time
	Metadata    map[string]interface{}
}

// LimitationModel tracks known agent limitations.
type LimitationModel struct {
	limitations map[string]*Limitation
	mu          sync.RWMutex
}

// Limitation represents a known weakness or blind spot.
type Limitation struct {
	Name         string
	Severity     float64 // 0.0 (minor) to 1.0 (critical)
	Description  string
	Workaround   string // How to mitigate
	FailureCount int64
	LastFailure  time.Time
}

// UncertaintyModel estimates prediction uncertainty.
type UncertaintyModel struct {
	baseUncertainty   float64
	domainUncertainty map[string]float64
	calibrationFactor float64
	mu                sync.RWMutex
}

// PerformanceTracker tracks historical performance.
type PerformanceTracker struct {
	records    []*PerformanceRecord
	maxRecords int
	mu         sync.RWMutex
}

// PerformanceRecord represents a single performance observation.
type PerformanceRecord struct {
	TaskType         string
	Timestamp        time.Time
	Success          bool
	Quality          float64 // 0.0 to 1.0
	Duration         time.Duration
	Confidence       float64 // Predicted confidence before task
	Actual           float64 // Actual outcome
	CalibrationError float64 // |Confidence - Actual|
}

// TaskEvaluation represents the self-model's assessment of a task.
type TaskEvaluation struct {
	CanHandle         bool
	Confidence        float64
	Uncertainty       float64
	WeakestCapability string
	Limitations       []string
	Recommendation    string
}

// ============================================================================
// Self-Model Implementation
// ============================================================================

// NewSelfModel creates a new self-model for an agent.
func NewSelfModel(agentID string, tier int) *SelfModel {
	return &SelfModel{
		AgentID:            agentID,
		Tier:               tier,
		CapabilityModel:    NewCapabilityModel(),
		LimitationModel:    NewLimitationModel(),
		UncertaintyModel:   NewUncertaintyModel(),
		PerformanceTracker: NewPerformanceTracker(1000),
	}
}

// CanHandle evaluates whether this agent can handle a given task.
func (sm *SelfModel) CanHandle(task *Task) *TaskEvaluation {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// Infer required capabilities from task
	requiredCapabilities := sm.inferRequiredCapabilities(task)

	// Check each required capability
	minConfidence := 1.0
	weakestCapability := ""
	matchedCapabilities := 0

	for _, capName := range requiredCapabilities {
		cap := sm.CapabilityModel.Get(capName)
		if cap != nil {
			matchedCapabilities++
			if cap.Confidence < minConfidence {
				minConfidence = cap.Confidence
				weakestCapability = capName
			}
		} else {
			// Unknown capability - low confidence
			minConfidence = 0.3
			weakestCapability = capName
		}
	}

	// Check for relevant limitations
	relevantLimitations := sm.findRelevantLimitations(task)
	limitationNames := make([]string, len(relevantLimitations))
	for i, lim := range relevantLimitations {
		limitationNames[i] = lim.Name
		// Reduce confidence based on limitation severity
		minConfidence *= (1.0 - lim.Severity*0.5)
	}

	// Calculate uncertainty
	uncertainty := sm.UncertaintyModel.EstimateUncertainty(task.Name)

	// Determine if we can handle the task
	canHandle := minConfidence > 0.5 && len(relevantLimitations) == 0

	// Generate recommendation
	recommendation := sm.generateRecommendation(canHandle, minConfidence, relevantLimitations)

	return &TaskEvaluation{
		CanHandle:         canHandle,
		Confidence:        minConfidence,
		Uncertainty:       uncertainty,
		WeakestCapability: weakestCapability,
		Limitations:       limitationNames,
		Recommendation:    recommendation,
	}
}

// inferRequiredCapabilities infers capabilities needed for a task.
func (sm *SelfModel) inferRequiredCapabilities(task *Task) []string {
	capabilities := make([]string, 0)

	// Check task type
	if task.Name != "" {
		capabilities = append(capabilities, task.Name)
	}

	// Check task parameters for hints
	if taskType, ok := task.Parameters["type"].(string); ok {
		capabilities = append(capabilities, taskType)
	}

	// Check for specific capability requirements
	if caps, ok := task.Parameters["required_capabilities"].([]string); ok {
		capabilities = append(capabilities, caps...)
	}

	return capabilities
}

// findRelevantLimitations finds limitations that apply to a task.
func (sm *SelfModel) findRelevantLimitations(task *Task) []*Limitation {
	relevant := make([]*Limitation, 0)

	for _, lim := range sm.LimitationModel.GetAll() {
		// Check if limitation applies to this task type
		if lim.Name == task.Name || lim.Name == fmt.Sprintf("%s_limitation", task.Name) {
			relevant = append(relevant, lim)
		}
	}

	return relevant
}

// generateRecommendation creates a recommendation based on evaluation.
func (sm *SelfModel) generateRecommendation(canHandle bool, confidence float64, limitations []*Limitation) string {
	if canHandle && confidence > 0.8 {
		return "High confidence - proceed with task"
	}

	if canHandle && confidence > 0.5 {
		return "Moderate confidence - proceed with caution"
	}

	if len(limitations) > 0 {
		return fmt.Sprintf("Known limitation: %s - consider alternative agent", limitations[0].Description)
	}

	if confidence < 0.3 {
		return "Low confidence - recommend delegating to specialist agent"
	}

	return "Uncertain - request human guidance"
}

// RecordPerformance records a performance observation.
func (sm *SelfModel) RecordPerformance(record *PerformanceRecord) {
	// Update performance tracker
	sm.PerformanceTracker.Add(record)

	// Update capability model based on outcome
	if record.Success {
		sm.CapabilityModel.Reinforce(record.TaskType, record.Quality)
	} else {
		sm.CapabilityModel.Weaken(record.TaskType, 0.9)
	}

	// Update uncertainty model based on calibration
	record.CalibrationError = math.Abs(record.Confidence - record.Actual)
	sm.UncertaintyModel.UpdateCalibration(record.CalibrationError)

	// Check for new limitations
	if !record.Success && record.Confidence > 0.7 {
		// Overconfident failure - might indicate limitation
		sm.LimitationModel.RegisterPotentialLimitation(record.TaskType)
	}
}

// GetPerformanceStats returns performance statistics.
func (sm *SelfModel) GetPerformanceStats() *PerformanceStats {
	return sm.PerformanceTracker.GetStats()
}

// ============================================================================
// Capability Model Implementation
// ============================================================================

// NewCapabilityModel creates a new capability model.
func NewCapabilityModel() *CapabilityModel {
	return &CapabilityModel{
		capabilities: make(map[string]*Capability),
	}
}

// Register registers a capability with initial confidence.
func (cm *CapabilityModel) Register(name string, confidence float64) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.capabilities[name] = &Capability{
		Name:        name,
		Confidence:  confidence,
		UsageCount:  0,
		SuccessRate: confidence,
		LastUsed:    time.Now(),
		Metadata:    make(map[string]interface{}),
	}
}

// Get retrieves a capability by name.
func (cm *CapabilityModel) Get(name string) *Capability {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.capabilities[name]
}

// GetAll returns all capabilities.
func (cm *CapabilityModel) GetAll() []*Capability {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	result := make([]*Capability, 0, len(cm.capabilities))
	for _, cap := range cm.capabilities {
		result = append(result, cap)
	}
	return result
}

// Reinforce increases confidence in a capability after success.
func (cm *CapabilityModel) Reinforce(name string, quality float64) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cap, exists := cm.capabilities[name]
	if !exists {
		// Auto-register new capability
		cm.capabilities[name] = &Capability{
			Name:        name,
			Confidence:  quality,
			UsageCount:  1,
			SuccessRate: 1.0,
			LastUsed:    time.Now(),
			Metadata:    make(map[string]interface{}),
		}
		return
	}

	// Update with exponential moving average
	alpha := 0.1
	cap.Confidence = cap.Confidence*(1-alpha) + quality*alpha
	cap.UsageCount++
	cap.SuccessRate = (cap.SuccessRate*float64(cap.UsageCount-1) + 1.0) / float64(cap.UsageCount)
	cap.LastUsed = time.Now()
}

// Weaken decreases confidence in a capability after failure.
func (cm *CapabilityModel) Weaken(name string, factor float64) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cap, exists := cm.capabilities[name]
	if !exists {
		// Register as weak capability
		cm.capabilities[name] = &Capability{
			Name:        name,
			Confidence:  0.3,
			UsageCount:  1,
			SuccessRate: 0.0,
			LastUsed:    time.Now(),
			Metadata:    make(map[string]interface{}),
		}
		return
	}

	cap.Confidence *= factor
	if cap.Confidence < 0.1 {
		cap.Confidence = 0.1 // Floor
	}
	cap.UsageCount++
	cap.SuccessRate = (cap.SuccessRate * float64(cap.UsageCount-1)) / float64(cap.UsageCount)
	cap.LastUsed = time.Now()
}

// TopCapabilities returns the highest confidence capabilities.
func (cm *CapabilityModel) TopCapabilities(n int) []*Capability {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	caps := make([]*Capability, 0, len(cm.capabilities))
	for _, cap := range cm.capabilities {
		caps = append(caps, cap)
	}

	sort.Slice(caps, func(i, j int) bool {
		return caps[i].Confidence > caps[j].Confidence
	})

	if n > len(caps) {
		n = len(caps)
	}
	return caps[:n]
}

// ============================================================================
// Limitation Model Implementation
// ============================================================================

// NewLimitationModel creates a new limitation model.
func NewLimitationModel() *LimitationModel {
	return &LimitationModel{
		limitations: make(map[string]*Limitation),
	}
}

// Register registers a known limitation.
func (lm *LimitationModel) Register(name, description, workaround string, severity float64) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	lm.limitations[name] = &Limitation{
		Name:         name,
		Severity:     severity,
		Description:  description,
		Workaround:   workaround,
		FailureCount: 0,
		LastFailure:  time.Time{},
	}
}

// Get retrieves a limitation by name.
func (lm *LimitationModel) Get(name string) *Limitation {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	return lm.limitations[name]
}

// GetAll returns all limitations.
func (lm *LimitationModel) GetAll() []*Limitation {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	result := make([]*Limitation, 0, len(lm.limitations))
	for _, lim := range lm.limitations {
		result = append(result, lim)
	}
	return result
}

// RecordFailure records a failure related to a limitation.
func (lm *LimitationModel) RecordFailure(name string) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	lim, exists := lm.limitations[name]
	if exists {
		lim.FailureCount++
		lim.LastFailure = time.Now()
	}
}

// RegisterPotentialLimitation registers a potential new limitation.
func (lm *LimitationModel) RegisterPotentialLimitation(taskType string) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	name := fmt.Sprintf("%s_limitation", taskType)
	if _, exists := lm.limitations[name]; !exists {
		lm.limitations[name] = &Limitation{
			Name:         name,
			Severity:     0.3, // Start as minor
			Description:  fmt.Sprintf("Potential limitation in %s tasks (observed overconfident failure)", taskType),
			Workaround:   "Consider additional validation or alternative approach",
			FailureCount: 1,
			LastFailure:  time.Now(),
		}
	} else {
		lm.limitations[name].FailureCount++
		lm.limitations[name].LastFailure = time.Now()
		// Increase severity with repeated failures
		lm.limitations[name].Severity = math.Min(1.0, lm.limitations[name].Severity+0.1)
	}
}

// ============================================================================
// Uncertainty Model Implementation
// ============================================================================

// NewUncertaintyModel creates a new uncertainty model.
func NewUncertaintyModel() *UncertaintyModel {
	return &UncertaintyModel{
		baseUncertainty:   0.2,
		domainUncertainty: make(map[string]float64),
		calibrationFactor: 1.0,
	}
}

// EstimateUncertainty estimates uncertainty for a domain.
func (um *UncertaintyModel) EstimateUncertainty(domain string) float64 {
	um.mu.RLock()
	defer um.mu.RUnlock()

	domainUnc, exists := um.domainUncertainty[domain]
	if !exists {
		// Unknown domain - higher uncertainty
		domainUnc = um.baseUncertainty * 1.5
	}

	// Apply calibration factor
	return domainUnc * um.calibrationFactor
}

// SetDomainUncertainty sets uncertainty for a specific domain.
func (um *UncertaintyModel) SetDomainUncertainty(domain string, uncertainty float64) {
	um.mu.Lock()
	defer um.mu.Unlock()
	um.domainUncertainty[domain] = uncertainty
}

// UpdateCalibration updates calibration based on prediction error.
func (um *UncertaintyModel) UpdateCalibration(calibrationError float64) {
	um.mu.Lock()
	defer um.mu.Unlock()

	// If we're under-confident (actual > predicted), reduce calibration factor
	// If we're over-confident (actual < predicted), increase calibration factor
	// This uses exponential moving average
	alpha := 0.05
	um.calibrationFactor = um.calibrationFactor*(1-alpha) + (1.0+calibrationError)*alpha
}

// GetCalibrationFactor returns the current calibration factor.
func (um *UncertaintyModel) GetCalibrationFactor() float64 {
	um.mu.RLock()
	defer um.mu.RUnlock()
	return um.calibrationFactor
}

// ============================================================================
// Performance Tracker Implementation
// ============================================================================

// NewPerformanceTracker creates a new performance tracker.
func NewPerformanceTracker(maxRecords int) *PerformanceTracker {
	return &PerformanceTracker{
		records:    make([]*PerformanceRecord, 0),
		maxRecords: maxRecords,
	}
}

// Add adds a performance record.
func (pt *PerformanceTracker) Add(record *PerformanceRecord) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	pt.records = append(pt.records, record)

	// Trim if exceeds max
	if len(pt.records) > pt.maxRecords {
		pt.records = pt.records[1:]
	}
}

// GetStats returns performance statistics.
func (pt *PerformanceTracker) GetStats() *PerformanceStats {
	pt.mu.RLock()
	defer pt.mu.RUnlock()

	if len(pt.records) == 0 {
		return &PerformanceStats{}
	}

	stats := &PerformanceStats{
		TotalTasks: int64(len(pt.records)),
	}

	var totalQuality, totalCalibrationError float64
	var totalDuration time.Duration

	for _, record := range pt.records {
		if record.Success {
			stats.SuccessfulTasks++
		}
		totalQuality += record.Quality
		totalCalibrationError += record.CalibrationError
		totalDuration += record.Duration
	}

	stats.SuccessRate = float64(stats.SuccessfulTasks) / float64(stats.TotalTasks)
	stats.AverageQuality = totalQuality / float64(stats.TotalTasks)
	stats.AverageCalibration = totalCalibrationError / float64(stats.TotalTasks)
	stats.AverageDuration = totalDuration / time.Duration(stats.TotalTasks)

	return stats
}

// GetRecentRecords returns the most recent n records.
func (pt *PerformanceTracker) GetRecentRecords(n int) []*PerformanceRecord {
	pt.mu.RLock()
	defer pt.mu.RUnlock()

	if n > len(pt.records) {
		n = len(pt.records)
	}

	start := len(pt.records) - n
	return pt.records[start:]
}

// GetByTaskType filters records by task type.
func (pt *PerformanceTracker) GetByTaskType(taskType string) []*PerformanceRecord {
	pt.mu.RLock()
	defer pt.mu.RUnlock()

	result := make([]*PerformanceRecord, 0)
	for _, record := range pt.records {
		if record.TaskType == taskType {
			result = append(result, record)
		}
	}
	return result
}

// PerformanceStats aggregates performance statistics.
type PerformanceStats struct {
	TotalTasks         int64
	SuccessfulTasks    int64
	SuccessRate        float64
	AverageQuality     float64
	AverageCalibration float64
	AverageDuration    time.Duration
}

// ============================================================================
// Agent Self-Model Registry
// ============================================================================

// SelfModelRegistry manages self-models for multiple agents.
type SelfModelRegistry struct {
	models map[string]*SelfModel
	mu     sync.RWMutex
}

// NewSelfModelRegistry creates a new registry.
func NewSelfModelRegistry() *SelfModelRegistry {
	return &SelfModelRegistry{
		models: make(map[string]*SelfModel),
	}
}

// Register registers a self-model for an agent.
func (r *SelfModelRegistry) Register(sm *SelfModel) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.models[sm.AgentID] = sm
}

// Get retrieves a self-model by agent ID.
func (r *SelfModelRegistry) Get(agentID string) *SelfModel {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.models[agentID]
}

// GetOrCreate gets or creates a self-model for an agent.
func (r *SelfModelRegistry) GetOrCreate(agentID string, tier int) *SelfModel {
	r.mu.Lock()
	defer r.mu.Unlock()

	if sm, exists := r.models[agentID]; exists {
		return sm
	}

	sm := NewSelfModel(agentID, tier)
	r.models[agentID] = sm
	return sm
}

// GetAllAgents returns all registered agent IDs.
func (r *SelfModelRegistry) GetAllAgents() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agents := make([]string, 0, len(r.models))
	for id := range r.models {
		agents = append(agents, id)
	}
	return agents
}
