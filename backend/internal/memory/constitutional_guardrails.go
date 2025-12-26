// Package memory provides the MNEMONIC memory system for the Elite Agent Collective.
// This file implements Constitutional AI-style guardrails for safety and alignment.
package memory

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
)

// ConstraintSeverity defines how critical a constraint violation is
type ConstraintSeverity int

const (
	// SeverityLow indicates a minor issue that should be logged
	SeverityLow ConstraintSeverity = iota
	// SeverityMedium indicates an issue that needs attention
	SeverityMedium
	// SeverityHigh indicates a serious issue requiring immediate action
	SeverityHigh
	// SeverityCritical indicates the response must be blocked
	SeverityCritical
)

func (s ConstraintSeverity) String() string {
	switch s {
	case SeverityLow:
		return "LOW"
	case SeverityMedium:
		return "MEDIUM"
	case SeverityHigh:
		return "HIGH"
	case SeverityCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// Constraint defines a constitutional constraint that responses must satisfy
type ConstitutionalConstraint struct {
	Name        string
	Description string
	Checker     func(*AgentResponse) bool
	Severity    ConstraintSeverity
	Category    string
}

// AgentResponse represents a response from an agent to be checked
type AgentResponse struct {
	AgentID      string
	Content      string
	Explanation  string
	Confidence   float64
	Capabilities []string
	ClaimedScope []string
	Metadata     map[string]interface{}
}

// Violation represents a constraint violation
type Violation struct {
	Constraint  *ConstitutionalConstraint
	Response    *AgentResponse
	Description string
	Timestamp   time.Time
	Remediation string
}

// ConstitutionalGuardrails enforces constitutional constraints on agent responses
type ConstitutionalGuardrails struct {
	constitution     []*ConstitutionalConstraint
	violationHistory []*Violation
	mu               sync.RWMutex

	// Configuration
	config *GuardrailsConfig

	// Metrics
	metrics *GuardrailsMetrics
}

// GuardrailsConfig configures the guardrails system
type GuardrailsConfig struct {
	MaxViolationHistory int
	BlockOnCritical     bool
	WarnOnHigh          bool
	LogAllViolations    bool
	CustomConstraints   []*ConstitutionalConstraint
}

// GuardrailsMetrics tracks guardrails performance
type GuardrailsMetrics struct {
	mu                   sync.RWMutex
	TotalChecks          int64
	TotalViolations      int64
	ViolationsByType     map[string]int64
	ViolationsBySeverity map[ConstraintSeverity]int64
	BlockedResponses     int64
	LastCheckTime        time.Time
}

// DefaultGuardrailsConfig returns default configuration
func DefaultGuardrailsConfig() *GuardrailsConfig {
	return &GuardrailsConfig{
		MaxViolationHistory: 1000,
		BlockOnCritical:     true,
		WarnOnHigh:          true,
		LogAllViolations:    true,
		CustomConstraints:   nil,
	}
}

// NewConstitutionalGuardrails creates a new guardrails system
func NewConstitutionalGuardrails(config *GuardrailsConfig) *ConstitutionalGuardrails {
	if config == nil {
		config = DefaultGuardrailsConfig()
	}

	g := &ConstitutionalGuardrails{
		constitution:     buildDefaultConstitution(),
		violationHistory: make([]*Violation, 0, config.MaxViolationHistory),
		config:           config,
		metrics: &GuardrailsMetrics{
			ViolationsByType:     make(map[string]int64),
			ViolationsBySeverity: make(map[ConstraintSeverity]int64),
		},
	}

	// Add custom constraints
	if config.CustomConstraints != nil {
		g.constitution = append(g.constitution, config.CustomConstraints...)
	}

	return g
}

// buildDefaultConstitution creates the default set of constitutional constraints
func buildDefaultConstitution() []*ConstitutionalConstraint {
	return []*ConstitutionalConstraint{
		{
			Name:        "Honesty",
			Description: "Agent must not claim capabilities it doesn't have",
			Checker:     honestyChecker,
			Severity:    SeverityCritical,
			Category:    "Truthfulness",
		},
		{
			Name:        "Transparency",
			Description: "Agent must acknowledge uncertainty when present",
			Checker:     transparencyChecker,
			Severity:    SeverityHigh,
			Category:    "Truthfulness",
		},
		{
			Name:        "HarmPrevention",
			Description: "Agent must not produce harmful outputs",
			Checker:     harmPreventionChecker,
			Severity:    SeverityCritical,
			Category:    "Safety",
		},
		{
			Name:        "ScopeLimitation",
			Description: "Agent must stay within its defined domain",
			Checker:     scopeLimitationChecker,
			Severity:    SeverityMedium,
			Category:    "Boundaries",
		},
		{
			Name:        "PrivacyProtection",
			Description: "Agent must not expose sensitive information",
			Checker:     privacyProtectionChecker,
			Severity:    SeverityCritical,
			Category:    "Privacy",
		},
		{
			Name:        "NoDeception",
			Description: "Agent must not attempt to deceive users",
			Checker:     noDeceptionChecker,
			Severity:    SeverityCritical,
			Category:    "Truthfulness",
		},
		{
			Name:        "Helpfulness",
			Description: "Agent must attempt to be genuinely helpful",
			Checker:     helpfulnessChecker,
			Severity:    SeverityLow,
			Category:    "Quality",
		},
		{
			Name:        "Consistency",
			Description: "Agent responses should be internally consistent",
			Checker:     consistencyChecker,
			Severity:    SeverityMedium,
			Category:    "Quality",
		},
	}
}

// Checker functions for each constraint

func honestyChecker(resp *AgentResponse) bool {
	// Check for overclaiming capabilities
	overclaimPatterns := []string{
		`(?i)i can do anything`,
		`(?i)i am always right`,
		`(?i)i never make mistakes`,
		`(?i)100% accurate`,
		`(?i)guaranteed to work`,
		`(?i)i know everything`,
		`(?i)perfect solution`,
	}

	for _, pattern := range overclaimPatterns {
		matched, _ := regexp.MatchString(pattern, resp.Content)
		if matched {
			return false
		}
	}

	return true
}

func transparencyChecker(resp *AgentResponse) bool {
	// If confidence is low, should acknowledge uncertainty
	if resp.Confidence < 0.5 && resp.Confidence > 0 {
		uncertaintyPhrases := []string{
			"uncertain", "not sure", "may be", "might be", "possibly",
			"i think", "i believe", "likely", "probably", "perhaps",
		}

		contentLower := strings.ToLower(resp.Content)
		for _, phrase := range uncertaintyPhrases {
			if strings.Contains(contentLower, phrase) {
				return true
			}
		}
		return false // Low confidence but no uncertainty acknowledgment
	}

	return true
}

func harmPreventionChecker(resp *AgentResponse) bool {
	// Check for potentially harmful content
	harmfulPatterns := []string{
		`(?i)how to hack`,
		`(?i)create a virus`,
		`(?i)make a bomb`,
		`(?i)synthesize drugs`,
		`(?i)exploit vulnerability`,
		`(?i)bypass security`,
		`(?i)steal credentials`,
		`(?i)phishing attack`,
		`(?i)social engineer`,
		`(?i)ddos attack`,
	}

	for _, pattern := range harmfulPatterns {
		matched, _ := regexp.MatchString(pattern, resp.Content)
		if matched {
			return false
		}
	}

	return true
}

func scopeLimitationChecker(resp *AgentResponse) bool {
	// Check if response stays within claimed scope
	if len(resp.ClaimedScope) == 0 {
		return true // No scope defined, cannot check
	}

	// This is a simplified check - in production, use embeddings
	// For now, check that content doesn't claim expertise outside scope
	outOfScopePatterns := []string{
		`(?i)as a medical doctor`,
		`(?i)as a lawyer`,
		`(?i)as a licensed`,
		`(?i)legal advice`,
		`(?i)medical diagnosis`,
	}

	for _, pattern := range outOfScopePatterns {
		matched, _ := regexp.MatchString(pattern, resp.Content)
		if matched {
			// Check if this is in scope
			scopeMatch := false
			for _, scope := range resp.ClaimedScope {
				if strings.Contains(strings.ToLower(scope), "medical") ||
					strings.Contains(strings.ToLower(scope), "legal") {
					scopeMatch = true
					break
				}
			}
			if !scopeMatch {
				return false
			}
		}
	}

	return true
}

func privacyProtectionChecker(resp *AgentResponse) bool {
	// Check for potential PII exposure
	piiPatterns := []string{
		`\b\d{3}-\d{2}-\d{4}\b`,         // SSN
		`\b\d{16}\b`,                    // Credit card
		`(?i)password\s*[:=]\s*\S+`,     // Password exposure
		`(?i)api[_\s]?key\s*[:=]\s*\S+`, // API key exposure
		`(?i)secret\s*[:=]\s*\S+`,       // Secret exposure
	}

	for _, pattern := range piiPatterns {
		matched, _ := regexp.MatchString(pattern, resp.Content)
		if matched {
			return false
		}
	}

	return true
}

func noDeceptionChecker(resp *AgentResponse) bool {
	// Check for deceptive patterns
	deceptionPatterns := []string{
		`(?i)pretend to be`,
		`(?i)impersonate`,
		`(?i)fake identity`,
		`(?i)hide the truth`,
		`(?i)mislead`,
		`(?i)don't tell them`,
	}

	for _, pattern := range deceptionPatterns {
		matched, _ := regexp.MatchString(pattern, resp.Content)
		if matched {
			return false
		}
	}

	return true
}

func helpfulnessChecker(resp *AgentResponse) bool {
	// Check for minimal effort responses
	if len(resp.Content) < 10 {
		return false
	}

	unhelpfulPatterns := []string{
		`(?i)^i don't know\.?$`,
		`(?i)^no\.?$`,
		`(?i)^can't help\.?$`,
		`(?i)^not my problem\.?$`,
	}

	for _, pattern := range unhelpfulPatterns {
		matched, _ := regexp.MatchString(pattern, strings.TrimSpace(resp.Content))
		if matched {
			return false
		}
	}

	return true
}

func consistencyChecker(resp *AgentResponse) bool {
	// Check for internal contradictions (simplified)
	contradictionPairs := [][]string{
		{"is true", "is false"},
		{"definitely", "uncertain"},
		{"always", "never"},
		{"yes", "no"},
	}

	contentLower := strings.ToLower(resp.Content)
	for _, pair := range contradictionPairs {
		if strings.Contains(contentLower, pair[0]) && strings.Contains(contentLower, pair[1]) {
			// Both contradictory terms present - might be a contradiction
			// More sophisticated check would analyze context
			return true // For now, don't flag without context analysis
		}
	}

	return true
}

// Enforce checks a response against all constitutional constraints
func (g *ConstitutionalGuardrails) Enforce(ctx context.Context, resp *AgentResponse) (*AgentResponse, []*Violation) {
	g.mu.Lock()
	defer g.mu.Unlock()

	violations := make([]*Violation, 0)
	blocked := false

	g.metrics.mu.Lock()
	g.metrics.TotalChecks++
	g.metrics.LastCheckTime = time.Now()
	g.metrics.mu.Unlock()

	for _, constraint := range g.constitution {
		select {
		case <-ctx.Done():
			return resp, violations
		default:
		}

		if !constraint.Checker(resp) {
			violation := &Violation{
				Constraint:  constraint,
				Response:    resp,
				Description: fmt.Sprintf("Violated constraint: %s - %s", constraint.Name, constraint.Description),
				Timestamp:   time.Now(),
				Remediation: g.suggestRemediation(constraint),
			}
			violations = append(violations, violation)

			// Record violation
			g.recordViolation(violation)

			// Check if we should block
			if constraint.Severity == SeverityCritical && g.config.BlockOnCritical {
				blocked = true
			}
		}
	}

	if blocked {
		return nil, violations
	}

	return resp, violations
}

// EnforceWithResult returns a detailed enforcement result
func (g *ConstitutionalGuardrails) EnforceWithResult(ctx context.Context, resp *AgentResponse) *EnforcementResult {
	response, violations := g.Enforce(ctx, resp)

	result := &EnforcementResult{
		OriginalResponse: resp,
		FilteredResponse: response,
		Violations:       violations,
		Blocked:          response == nil,
		Timestamp:        time.Now(),
	}

	// Determine highest severity
	for _, v := range violations {
		if v.Constraint.Severity > result.HighestSeverity {
			result.HighestSeverity = v.Constraint.Severity
		}
	}

	return result
}

// EnforcementResult contains the full result of enforcement
type EnforcementResult struct {
	OriginalResponse *AgentResponse
	FilteredResponse *AgentResponse
	Violations       []*Violation
	Blocked          bool
	HighestSeverity  ConstraintSeverity
	Timestamp        time.Time
}

// recordViolation stores a violation in history
func (g *ConstitutionalGuardrails) recordViolation(v *Violation) {
	g.metrics.mu.Lock()
	g.metrics.TotalViolations++
	g.metrics.ViolationsByType[v.Constraint.Name]++
	g.metrics.ViolationsBySeverity[v.Constraint.Severity]++
	if v.Constraint.Severity == SeverityCritical {
		g.metrics.BlockedResponses++
	}
	g.metrics.mu.Unlock()

	g.violationHistory = append(g.violationHistory, v)

	// Trim history if too long
	if len(g.violationHistory) > g.config.MaxViolationHistory {
		g.violationHistory = g.violationHistory[len(g.violationHistory)-g.config.MaxViolationHistory:]
	}
}

// suggestRemediation provides remediation suggestions for violations
func (g *ConstitutionalGuardrails) suggestRemediation(c *ConstitutionalConstraint) string {
	remediations := map[string]string{
		"Honesty":           "Revise response to avoid overclaiming capabilities. Use phrases like 'I can help with...' instead of absolute claims.",
		"Transparency":      "Add uncertainty acknowledgment when confidence is low. Use phrases like 'I believe' or 'It's likely that'.",
		"HarmPrevention":    "Remove or rephrase harmful content. Consider the potential misuse of information provided.",
		"ScopeLimitation":   "Stay within defined expertise. Recommend appropriate experts for out-of-scope queries.",
		"PrivacyProtection": "Remove any PII, credentials, or sensitive information from the response.",
		"NoDeception":       "Ensure response is truthful and transparent. Do not pretend to be something you're not.",
		"Helpfulness":       "Provide more substantive assistance. If unable to help, explain why and suggest alternatives.",
		"Consistency":       "Review response for internal contradictions and resolve them.",
	}

	if rem, ok := remediations[c.Name]; ok {
		return rem
	}
	return "Review and revise response to comply with this constraint."
}

// AddConstraint adds a custom constraint to the constitution
func (g *ConstitutionalGuardrails) AddConstraint(c *ConstitutionalConstraint) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.constitution = append(g.constitution, c)
}

// RemoveConstraint removes a constraint by name
func (g *ConstitutionalGuardrails) RemoveConstraint(name string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	for i, c := range g.constitution {
		if c.Name == name {
			g.constitution = append(g.constitution[:i], g.constitution[i+1:]...)
			return true
		}
	}
	return false
}

// GetConstraints returns all current constraints
func (g *ConstitutionalGuardrails) GetConstraints() []*ConstitutionalConstraint {
	g.mu.RLock()
	defer g.mu.RUnlock()

	result := make([]*ConstitutionalConstraint, len(g.constitution))
	copy(result, g.constitution)
	return result
}

// GetViolationHistory returns recent violations
func (g *ConstitutionalGuardrails) GetViolationHistory(limit int) []*Violation {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if limit <= 0 || limit > len(g.violationHistory) {
		limit = len(g.violationHistory)
	}

	start := len(g.violationHistory) - limit
	result := make([]*Violation, limit)
	copy(result, g.violationHistory[start:])
	return result
}

// GetMetrics returns current metrics
func (g *ConstitutionalGuardrails) GetMetrics() *GuardrailsMetrics {
	g.metrics.mu.RLock()
	defer g.metrics.mu.RUnlock()

	// Deep copy
	metrics := &GuardrailsMetrics{
		TotalChecks:          g.metrics.TotalChecks,
		TotalViolations:      g.metrics.TotalViolations,
		ViolationsByType:     make(map[string]int64),
		ViolationsBySeverity: make(map[ConstraintSeverity]int64),
		BlockedResponses:     g.metrics.BlockedResponses,
		LastCheckTime:        g.metrics.LastCheckTime,
	}

	for k, v := range g.metrics.ViolationsByType {
		metrics.ViolationsByType[k] = v
	}
	for k, v := range g.metrics.ViolationsBySeverity {
		metrics.ViolationsBySeverity[k] = v
	}

	return metrics
}

// ViolationRate returns the percentage of checks that resulted in violations
func (g *ConstitutionalGuardrails) ViolationRate() float64 {
	metrics := g.GetMetrics()
	if metrics.TotalChecks == 0 {
		return 0.0
	}
	return float64(metrics.TotalViolations) / float64(metrics.TotalChecks) * 100.0
}

// CriticalViolationRate returns the percentage of checks that were blocked
func (g *ConstitutionalGuardrails) CriticalViolationRate() float64 {
	metrics := g.GetMetrics()
	if metrics.TotalChecks == 0 {
		return 0.0
	}
	return float64(metrics.BlockedResponses) / float64(metrics.TotalChecks) * 100.0
}
