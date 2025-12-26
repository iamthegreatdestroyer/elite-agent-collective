// Package memory provides the MNEMONIC memory system for the Elite Agent Collective.
// This file implements the Safety Monitor System for continuous safety monitoring.
package memory

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"
)

// AlertSeverity defines the severity of a safety alert
type AlertSeverity int

const (
	AlertInfo AlertSeverity = iota
	AlertWarning
	AlertHigh
	AlertCritical
	AlertEmergency
)

func (s AlertSeverity) String() string {
	switch s {
	case AlertInfo:
		return "INFO"
	case AlertWarning:
		return "WARNING"
	case AlertHigh:
		return "HIGH"
	case AlertCritical:
		return "CRITICAL"
	case AlertEmergency:
		return "EMERGENCY"
	default:
		return "UNKNOWN"
	}
}

// AlertType categorizes safety alerts
type AlertType string

const (
	AlertAlignmentDrift      AlertType = "alignment_drift"
	AlertBehaviorDrift       AlertType = "behavior_drift"
	AlertCapabilityEscape    AlertType = "capability_escape"
	AlertFitnessGaming       AlertType = "fitness_gaming"
	AlertEmergentMisalign    AlertType = "emergent_misalignment"
	AlertConstraintViolation AlertType = "constraint_violation"
	AlertAnomalousPattern    AlertType = "anomalous_pattern"
)

// SafetyAlert represents a safety-related alert
type SafetyAlert struct {
	ID          string
	Severity    AlertSeverity
	Type        AlertType
	AgentID     string
	Description string
	Evidence    map[string]interface{}
	Timestamp   time.Time
	Resolved    bool
	Resolution  string
}

// SafetyMonitor continuously monitors the system for safety issues
type SafetyMonitor struct {
	mu sync.RWMutex

	// Sub-monitors
	driftDetector        *DriftDetector
	alignmentChecker     *AlignmentChecker
	capabilityController *CapabilityController
	guardrails           *ConstitutionalGuardrails

	// Alert management
	alertChannel chan *SafetyAlert
	alertHistory []*SafetyAlert

	// Configuration
	config *SafetyMonitorConfig

	// Metrics
	metrics *SafetyMetrics

	// State
	running bool
	stopCh  chan struct{}
}

// SafetyMonitorConfig configures the safety monitor
type SafetyMonitorConfig struct {
	MonitoringInterval time.Duration
	DriftThreshold     float64
	AlignmentThreshold float64
	MaxAlertHistory    int
	AlertChannelSize   int
	EnableAutoShutdown bool
	ShutdownThreshold  int // Number of critical alerts before shutdown
}

// SafetyMetrics tracks safety monitoring metrics
type SafetyMetrics struct {
	mu                sync.RWMutex
	TotalAlerts       int64
	AlertsBySeverity  map[AlertSeverity]int64
	AlertsByType      map[AlertType]int64
	MonitoringCycles  int64
	LastCheckTime     time.Time
	SystemHealthScore float64
}

// DefaultSafetyMonitorConfig returns default configuration
func DefaultSafetyMonitorConfig() *SafetyMonitorConfig {
	return &SafetyMonitorConfig{
		MonitoringInterval: 1 * time.Minute,
		DriftThreshold:     0.3,
		AlignmentThreshold: 0.7,
		MaxAlertHistory:    1000,
		AlertChannelSize:   100,
		EnableAutoShutdown: false,
		ShutdownThreshold:  5,
	}
}

// NewSafetyMonitor creates a new safety monitor
func NewSafetyMonitor(config *SafetyMonitorConfig, guardrails *ConstitutionalGuardrails) *SafetyMonitor {
	if config == nil {
		config = DefaultSafetyMonitorConfig()
	}

	return &SafetyMonitor{
		driftDetector:        NewDriftDetector(config.DriftThreshold),
		alignmentChecker:     NewAlignmentChecker(config.AlignmentThreshold),
		capabilityController: NewCapabilityController(),
		guardrails:           guardrails,
		alertChannel:         make(chan *SafetyAlert, config.AlertChannelSize),
		alertHistory:         make([]*SafetyAlert, 0, config.MaxAlertHistory),
		config:               config,
		metrics: &SafetyMetrics{
			AlertsBySeverity:  make(map[AlertSeverity]int64),
			AlertsByType:      make(map[AlertType]int64),
			SystemHealthScore: 1.0,
		},
		stopCh: make(chan struct{}),
	}
}

// Start begins continuous monitoring
func (m *SafetyMonitor) Start(ctx context.Context) error {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return fmt.Errorf("safety monitor already running")
	}
	m.running = true
	m.mu.Unlock()

	go m.monitoringLoop(ctx)
	return nil
}

// Stop stops the monitoring
func (m *SafetyMonitor) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		close(m.stopCh)
		m.running = false
	}
}

// monitoringLoop runs the continuous monitoring
func (m *SafetyMonitor) monitoringLoop(ctx context.Context) {
	ticker := time.NewTicker(m.config.MonitoringInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.stopCh:
			return
		case <-ticker.C:
			m.runMonitoringCycle(ctx)
		}
	}
}

// runMonitoringCycle performs one monitoring cycle
func (m *SafetyMonitor) runMonitoringCycle(ctx context.Context) {
	m.metrics.mu.Lock()
	m.metrics.MonitoringCycles++
	m.metrics.LastCheckTime = time.Now()
	m.metrics.mu.Unlock()

	// Check alignment
	alignmentReport := m.alignmentChecker.CheckAlignment()
	if !alignmentReport.Aligned {
		m.emitAlert(&SafetyAlert{
			ID:          generateAlertID(),
			Severity:    AlertCritical,
			Type:        AlertEmergentMisalign,
			Description: "Collective goal drift detected",
			Evidence: map[string]interface{}{
				"report":   alignmentReport,
				"distance": alignmentReport.DriftDistance,
			},
			Timestamp: time.Now(),
		})
	}

	// Check drift for each tracked agent
	for agentID := range m.driftDetector.GetTrackedAgents() {
		drift := m.driftDetector.MeasureDrift(agentID)
		if drift > m.config.DriftThreshold {
			severity := AlertWarning
			if drift > 0.5 {
				severity = AlertHigh
			}
			if drift > 0.7 {
				severity = AlertCritical
			}

			m.emitAlert(&SafetyAlert{
				ID:          generateAlertID(),
				Severity:    severity,
				Type:        AlertBehaviorDrift,
				AgentID:     agentID,
				Description: fmt.Sprintf("Agent %s behavior drift: %.2f", agentID, drift),
				Evidence: map[string]interface{}{
					"drift_score": drift,
				},
				Timestamp: time.Now(),
			})
		}
	}

	// Check for capability escapes
	escapes := m.capabilityController.CheckEscapes()
	for _, escape := range escapes {
		m.emitAlert(&SafetyAlert{
			ID:          generateAlertID(),
			Severity:    AlertHigh,
			Type:        AlertCapabilityEscape,
			AgentID:     escape.AgentID,
			Description: fmt.Sprintf("Agent %s using unapproved capability: %s", escape.AgentID, escape.Capability),
			Evidence: map[string]interface{}{
				"capability": escape.Capability,
				"action":     escape.Action,
			},
			Timestamp: time.Now(),
		})
	}

	// Update health score
	m.updateHealthScore()
}

// emitAlert records and emits an alert
func (m *SafetyMonitor) emitAlert(alert *SafetyAlert) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Record in history
	m.alertHistory = append(m.alertHistory, alert)
	if len(m.alertHistory) > m.config.MaxAlertHistory {
		m.alertHistory = m.alertHistory[len(m.alertHistory)-m.config.MaxAlertHistory:]
	}

	// Update metrics
	m.metrics.mu.Lock()
	m.metrics.TotalAlerts++
	m.metrics.AlertsBySeverity[alert.Severity]++
	m.metrics.AlertsByType[alert.Type]++
	m.metrics.mu.Unlock()

	// Send to channel (non-blocking)
	select {
	case m.alertChannel <- alert:
	default:
		// Channel full, log but don't block
	}

	// Check for auto-shutdown
	if m.config.EnableAutoShutdown {
		criticalCount := m.metrics.AlertsBySeverity[AlertCritical] + m.metrics.AlertsBySeverity[AlertEmergency]
		if int(criticalCount) >= m.config.ShutdownThreshold {
			m.initiateEmergencyShutdown()
		}
	}
}

// initiateEmergencyShutdown triggers emergency shutdown
func (m *SafetyMonitor) initiateEmergencyShutdown() {
	// In production, this would trigger actual shutdown procedures
	m.emitAlert(&SafetyAlert{
		ID:          generateAlertID(),
		Severity:    AlertEmergency,
		Type:        AlertAnomalousPattern,
		Description: "Emergency shutdown initiated due to critical safety threshold exceeded",
		Timestamp:   time.Now(),
	})
}

// updateHealthScore calculates current system health
func (m *SafetyMonitor) updateHealthScore() {
	m.metrics.mu.Lock()
	defer m.metrics.mu.Unlock()

	// Simple health calculation - can be made more sophisticated
	totalAlerts := m.metrics.TotalAlerts
	if totalAlerts == 0 {
		m.metrics.SystemHealthScore = 1.0
		return
	}

	// Weight alerts by severity
	weightedSum := float64(0)
	weightedSum += float64(m.metrics.AlertsBySeverity[AlertInfo]) * 0.1
	weightedSum += float64(m.metrics.AlertsBySeverity[AlertWarning]) * 0.3
	weightedSum += float64(m.metrics.AlertsBySeverity[AlertHigh]) * 0.5
	weightedSum += float64(m.metrics.AlertsBySeverity[AlertCritical]) * 0.8
	weightedSum += float64(m.metrics.AlertsBySeverity[AlertEmergency]) * 1.0

	// Decay based on monitoring cycles
	cycles := float64(m.metrics.MonitoringCycles)
	if cycles > 0 {
		m.metrics.SystemHealthScore = math.Max(0, 1.0-(weightedSum/(cycles*10)))
	}
}

// GetAlerts returns the alert channel for consumers
func (m *SafetyMonitor) GetAlerts() <-chan *SafetyAlert {
	return m.alertChannel
}

// GetAlertHistory returns recent alerts
func (m *SafetyMonitor) GetAlertHistory(limit int) []*SafetyAlert {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if limit <= 0 || limit > len(m.alertHistory) {
		limit = len(m.alertHistory)
	}

	start := len(m.alertHistory) - limit
	result := make([]*SafetyAlert, limit)
	copy(result, m.alertHistory[start:])
	return result
}

// GetMetrics returns current safety metrics
func (m *SafetyMonitor) GetMetrics() *SafetyMetrics {
	m.metrics.mu.RLock()
	defer m.metrics.mu.RUnlock()

	metrics := &SafetyMetrics{
		TotalAlerts:       m.metrics.TotalAlerts,
		AlertsBySeverity:  make(map[AlertSeverity]int64),
		AlertsByType:      make(map[AlertType]int64),
		MonitoringCycles:  m.metrics.MonitoringCycles,
		LastCheckTime:     m.metrics.LastCheckTime,
		SystemHealthScore: m.metrics.SystemHealthScore,
	}

	for k, v := range m.metrics.AlertsBySeverity {
		metrics.AlertsBySeverity[k] = v
	}
	for k, v := range m.metrics.AlertsByType {
		metrics.AlertsByType[k] = v
	}

	return metrics
}

// CheckResponse runs safety checks on a response
func (m *SafetyMonitor) CheckResponse(ctx context.Context, resp *AgentResponse) (*SafetyCheckResult, error) {
	result := &SafetyCheckResult{
		Passed:    true,
		Timestamp: time.Now(),
	}

	// Run constitutional guardrails
	if m.guardrails != nil {
		enforcementResult := m.guardrails.EnforceWithResult(ctx, resp)
		if enforcementResult.Blocked {
			result.Passed = false
			result.BlockReason = "Constitutional constraint violation"
			result.Violations = enforcementResult.Violations

			m.emitAlert(&SafetyAlert{
				ID:          generateAlertID(),
				Severity:    AlertHigh,
				Type:        AlertConstraintViolation,
				AgentID:     resp.AgentID,
				Description: fmt.Sprintf("Agent %s response blocked by guardrails", resp.AgentID),
				Evidence: map[string]interface{}{
					"violations": len(enforcementResult.Violations),
				},
				Timestamp: time.Now(),
			})
		}
		result.GuardrailsResult = enforcementResult
	}

	// Check capability usage
	if capErr := m.capabilityController.ValidateAction(resp.AgentID, resp.Capabilities); capErr != nil {
		result.Passed = false
		result.CapabilityViolation = capErr.Error()
	}

	return result, nil
}

// SafetyCheckResult contains the result of a safety check
type SafetyCheckResult struct {
	Passed              bool
	BlockReason         string
	Violations          []*Violation
	GuardrailsResult    *EnforcementResult
	CapabilityViolation string
	Timestamp           time.Time
}

// RegisterAgent registers an agent for monitoring
func (m *SafetyMonitor) RegisterAgent(agentID string, baselineBehavior []float64, capabilities []string) {
	m.driftDetector.RegisterAgent(agentID, baselineBehavior)
	m.capabilityController.RegisterAgent(agentID, capabilities)
}

// RecordAgentBehavior records behavior for drift detection
func (m *SafetyMonitor) RecordAgentBehavior(agentID string, behavior []float64) {
	m.driftDetector.RecordBehavior(agentID, behavior)
}

// DriftDetector detects behavioral drift in agents
type DriftDetector struct {
	mu            sync.RWMutex
	threshold     float64
	baselines     map[string][]float64
	currentStates map[string][]float64
	history       map[string][][]float64
	maxHistory    int
}

// NewDriftDetector creates a new drift detector
func NewDriftDetector(threshold float64) *DriftDetector {
	return &DriftDetector{
		threshold:     threshold,
		baselines:     make(map[string][]float64),
		currentStates: make(map[string][]float64),
		history:       make(map[string][][]float64),
		maxHistory:    100,
	}
}

// RegisterAgent registers an agent with its baseline behavior
func (d *DriftDetector) RegisterAgent(agentID string, baseline []float64) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.baselines[agentID] = baseline
	d.currentStates[agentID] = baseline
	d.history[agentID] = make([][]float64, 0, d.maxHistory)
}

// RecordBehavior records a new behavior observation
func (d *DriftDetector) RecordBehavior(agentID string, behavior []float64) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.currentStates[agentID] = behavior
	d.history[agentID] = append(d.history[agentID], behavior)

	if len(d.history[agentID]) > d.maxHistory {
		d.history[agentID] = d.history[agentID][1:]
	}
}

// MeasureDrift measures behavioral drift from baseline
func (d *DriftDetector) MeasureDrift(agentID string) float64 {
	d.mu.RLock()
	defer d.mu.RUnlock()

	baseline, ok := d.baselines[agentID]
	if !ok {
		return 0
	}

	current, ok := d.currentStates[agentID]
	if !ok {
		return 0
	}

	return cosineDissimilarity(baseline, current)
}

// GetTrackedAgents returns all tracked agent IDs
func (d *DriftDetector) GetTrackedAgents() map[string]bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	result := make(map[string]bool)
	for id := range d.baselines {
		result[id] = true
	}
	return result
}

// cosineDissimilarity computes 1 - cosine similarity
func cosineDissimilarity(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 1.0 // Maximum dissimilarity for incompatible vectors
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 1.0
	}

	similarity := dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
	return 1.0 - similarity
}

// AlignmentChecker checks for emergent misalignment
type AlignmentChecker struct {
	mu              sync.RWMutex
	threshold       float64
	individualGoals map[string][]float64
	collectiveGoal  []float64
}

// NewAlignmentChecker creates a new alignment checker
func NewAlignmentChecker(threshold float64) *AlignmentChecker {
	return &AlignmentChecker{
		threshold:       threshold,
		individualGoals: make(map[string][]float64),
	}
}

// RegisterAgentGoal registers an agent's goal vector
func (a *AlignmentChecker) RegisterAgentGoal(agentID string, goal []float64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.individualGoals[agentID] = goal
}

// SetCollectiveGoal sets the observed collective goal
func (a *AlignmentChecker) SetCollectiveGoal(goal []float64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.collectiveGoal = goal
}

// AlignmentReport contains alignment check results
type AlignmentReport struct {
	Aligned           bool
	DriftDistance     float64
	CollectiveGoal    []float64
	RecommendedAction string
	CheckTime         time.Time
}

// CheckAlignment checks if collective goal is aligned with individual goals
func (a *AlignmentChecker) CheckAlignment() *AlignmentReport {
	a.mu.RLock()
	defer a.mu.RUnlock()

	report := &AlignmentReport{
		Aligned:        true,
		CollectiveGoal: a.collectiveGoal,
		CheckTime:      time.Now(),
	}

	if len(a.collectiveGoal) == 0 || len(a.individualGoals) == 0 {
		return report
	}

	// Calculate distance from collective goal to convex hull of individual goals
	minDistance := math.MaxFloat64
	for _, goal := range a.individualGoals {
		dist := cosineDissimilarity(a.collectiveGoal, goal)
		if dist < minDistance {
			minDistance = dist
		}
	}

	report.DriftDistance = minDistance

	if minDistance > (1.0 - a.threshold) {
		report.Aligned = false
		report.RecommendedAction = "Review and potentially reset collective memory"
	}

	return report
}

// CapabilityController controls what capabilities agents can use
type CapabilityController struct {
	mu                   sync.RWMutex
	allowedCapabilities  map[string]map[string]bool // agent -> capability -> allowed
	emergentCapabilities map[string][]string
	pendingApprovals     map[string]bool // capability -> pending
	escapes              []*CapabilityEscape
}

// CapabilityEscape represents an unapproved capability use
type CapabilityEscape struct {
	AgentID    string
	Capability string
	Action     string
	Timestamp  time.Time
}

// NewCapabilityController creates a new capability controller
func NewCapabilityController() *CapabilityController {
	return &CapabilityController{
		allowedCapabilities:  make(map[string]map[string]bool),
		emergentCapabilities: make(map[string][]string),
		pendingApprovals:     make(map[string]bool),
		escapes:              make([]*CapabilityEscape, 0),
	}
}

// RegisterAgent registers an agent with its allowed capabilities
func (c *CapabilityController) RegisterAgent(agentID string, capabilities []string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.allowedCapabilities[agentID] = make(map[string]bool)
	for _, cap := range capabilities {
		c.allowedCapabilities[agentID][cap] = true
	}
}

// ValidateAction validates if an agent can use certain capabilities
func (c *CapabilityController) ValidateAction(agentID string, capabilities []string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	allowed, ok := c.allowedCapabilities[agentID]
	if !ok {
		return nil // Agent not registered, allow by default
	}

	for _, cap := range capabilities {
		if !allowed[cap] {
			// Check if it's pending approval
			if c.pendingApprovals[cap] {
				continue // Allow pending capabilities
			}

			// Record escape
			c.escapes = append(c.escapes, &CapabilityEscape{
				AgentID:    agentID,
				Capability: cap,
				Timestamp:  time.Now(),
			})

			// Register as emergent
			c.emergentCapabilities[agentID] = append(c.emergentCapabilities[agentID], cap)
			c.pendingApprovals[cap] = true

			return fmt.Errorf("capability %s not approved for agent %s", cap, agentID)
		}
	}

	return nil
}

// ApproveCapability approves an emergent capability
func (c *CapabilityController) ApproveCapability(agentID, capability string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.allowedCapabilities[agentID] == nil {
		c.allowedCapabilities[agentID] = make(map[string]bool)
	}
	c.allowedCapabilities[agentID][capability] = true
	delete(c.pendingApprovals, capability)
}

// CheckEscapes returns recent capability escapes
func (c *CapabilityController) CheckEscapes() []*CapabilityEscape {
	c.mu.Lock()
	defer c.mu.Unlock()

	result := make([]*CapabilityEscape, len(c.escapes))
	copy(result, c.escapes)
	c.escapes = c.escapes[:0] // Clear after returning
	return result
}

// GetPendingApprovals returns capabilities pending approval
func (c *CapabilityController) GetPendingApprovals() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]string, 0, len(c.pendingApprovals))
	for cap := range c.pendingApprovals {
		result = append(result, cap)
	}
	return result
}

// Helper to generate alert IDs
var alertIDCounter int64
var alertIDMu sync.Mutex

func generateAlertID() string {
	alertIDMu.Lock()
	defer alertIDMu.Unlock()
	alertIDCounter++
	return fmt.Sprintf("alert-%d-%d", time.Now().UnixNano(), alertIDCounter)
}
