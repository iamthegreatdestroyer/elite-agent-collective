// Package memory provides the MNEMONIC system for the Elite Agent Collective.
// This file implements the unified Cognitive Framework for Phase 0 Foundation.
//
// The Cognitive Framework provides:
// - CognitiveComponent interface for all cognitive subsystems
// - Unified request/response types for inter-component communication
// - Integration hooks for ReMem loop and other systems
// - Performance metrics collection

package memory

import (
	"context"
	"time"
)

// ============================================================================
// Cognitive Component Interface - Core Abstraction for Phase 0
// ============================================================================

// CognitiveComponent is the base interface for all cognitive system components.
// Every component in the cognitive architecture must implement this interface.
// This is the key abstraction that enables modular cognitive processing in Phase 0.
//
// Design rationale:
// - Simple interface (5 methods) enables many implementations
// - Initialize/Shutdown enable resource management
// - Process handles the actual work
// - GetMetrics enables monitoring and adaptation
// - GetName enables component identification and logging
type CognitiveComponent interface {
	// Initialize sets up the component with the provided configuration
	// Config structure is component-specific
	Initialize(config interface{}) error

	// Process executes the component's core functionality
	// The component receives a request with full cognitive context,
	// processes it using the component's specific logic,
	// and returns a result with decision trace and metrics
	Process(ctx context.Context, request *CognitiveProcessRequest) (*CognitiveProcessResult, error)

	// Shutdown gracefully shuts down the component
	// Should release any resources and wait for in-flight requests to complete
	Shutdown() error

	// GetMetrics returns current performance and operational metrics
	// Called periodically for monitoring and adaptation
	GetMetrics() CognitiveMetrics

	// GetName returns the component's identifier
	// Used for logging, monitoring, and component registration
	GetName() string
}

// ============================================================================
// Cognitive Processing Request - Input to Components
// ============================================================================

// CognitiveProcessRequest bundles all inputs needed for cognitive processing.
// This is the standard input type for all CognitiveComponent.Process() calls.
//
// The request carries:
// - Identification and timing information
// - References to cognitive state (working memory, goal stack)
// - Domain context and constraints
// - Integration hooks to other systems
type CognitiveProcessRequest struct {
	// Request identification and timing
	RequestID string     // Unique request ID for tracing
	AgentID   string     // Which agent is making this request
	Timestamp time.Time  // When the request was created
	Deadline  *time.Time // Optional deadline for processing

	// Cognitive state references - these reference existing systems, not copies
	WorkingMemory *CognitiveWorkingMemory // Reference to the working memory system
	GoalStack     *GoalStack              // Reference to the goal stack
	CurrentGoal   *Goal                   // The primary goal being processed

	// Domain context and constraints
	AvailableAgents   []string                   // List of agent IDs available
	ActiveConstraints []ConstitutionalConstraint // Safety constraints to check
	ApprovedActions   []string                   // Pre-approved actions
	DomainContext     map[string]interface{}     // Domain-specific metadata

	// Performance tracking - used to build operation chain
	StartTime      time.Time
	OperationChain []string // Components already executed in this request

	// Integration hooks - references to other systems
	ReMem         interface{} // Reference to ReMem loop (for REFLECT phase integration)
	AgentRegistry interface{} // Reference to agent registry
	MemorySystem  interface{} // Reference to MNEMONIC memory system
	SafetyMonitor interface{} // Reference to safety monitor
}

// ============================================================================
// Cognitive Processing Result - Output from Components
// ============================================================================

// CognitiveProcessResult bundles all outputs from cognitive processing.
// This is the standard output type for all CognitiveComponent.Process() calls.
//
// The result carries:
// - Status indicating success/impasse/error
// - Selected agents or actions
// - Decision trace for interpretability
// - Execution metrics
// - Safety validation results
type CognitiveProcessResult struct {
	// Core result status and output
	Status         ProcessingStatus
	Output         interface{} // Component-specific output
	SelectedAgents []string    // IDs of agents chosen to execute
	DecisionTrace  *DecisionTrace

	// Decision quality metrics
	Confidence          float64 // 0-1, confidence in this decision
	Alternatives        []DecisionOption
	Explanation         string // Human-readable explanation of decision
	RequiresHumanReview bool

	// Execution metrics for performance monitoring
	ProcessingTime     time.Duration
	ComponentsInvolved []string
	ExecutionSteps     []*ExecutionStep

	// Error handling
	Error        error
	ErrorDetails map[string]interface{}

	// Safety validation
	SafetyCheckResults   []SafetyValidation
	AllConstraintsPassed bool

	// Follow-up actions (e.g., impasse detected, needs decomposition)
	RequiresImpasse bool   // If true, impasse detection was triggered
	ImpasseType     string // Type of impasse: resource, knowledge, conflicting-goals, etc.
	NextActions     []string
}

// ProcessingStatus indicates the outcome of cognitive processing
type ProcessingStatus int

const (
	ProcessSuccess ProcessingStatus = iota
	ProcessPartialSuccess
	ProcessImpasse
	ProcessError
	ProcessBlocked
	ProcessRequiresReview
)

func (ps ProcessingStatus) String() string {
	switch ps {
	case ProcessSuccess:
		return "success"
	case ProcessPartialSuccess:
		return "partial_success"
	case ProcessImpasse:
		return "impasse"
	case ProcessError:
		return "error"
	case ProcessBlocked:
		return "blocked"
	case ProcessRequiresReview:
		return "requires_review"
	default:
		return "unknown"
	}
}

// ============================================================================
// Decision Trace - For Interpretability and Debugging
// ============================================================================

// DecisionTrace represents the reasoning steps taken to reach a decision
// Used for interpretability, debugging, and learning from decisions
type DecisionTrace struct {
	Steps                  []*DecisionStep        // Sequence of reasoning steps
	InitialState           map[string]interface{} // Starting cognitive state
	FinalState             map[string]interface{} // Ending cognitive state
	AssumptionsMade        []string               // Explicit assumptions
	FallaciesDetected      []string               // Potential logical fallacies
	ImplicationsConsidered []string               // Implications of the decision
}

// DecisionStep represents a single step in the decision reasoning process
type DecisionStep struct {
	Index       int
	Description string
	Input       interface{}
	Output      interface{}
	Logic       string  // The inference rule or heuristic used
	Confidence  float64 // 0-1, confidence in this step
	Timestamp   time.Time
	Duration    time.Duration
}

// DecisionOption represents an alternative option not chosen
// Used for explaining why other options were rejected
type DecisionOption struct {
	ID              string   // Option identifier
	Description     string   // What this option would do
	Agents          []string // Agents needed for this option
	Confidence      float64  // Estimated success rate
	Reasoning       string   // Why this option was considered
	Advantages      []string // Pros of this option
	Disadvantages   []string // Cons of this option
	RejectionReason string   // Why it wasn't chosen
}

// ============================================================================
// Execution Tracking - For Monitoring and Debugging
// ============================================================================

// ExecutionStep represents one component's execution in the processing chain
// Used to track which components were invoked and their performance
type ExecutionStep struct {
	ComponentName string
	StepNumber    int
	Input         interface{}
	Output        interface{}
	Duration      time.Duration
	Status        string
	Error         error
}

// SafetyValidation represents the result of a single safety constraint check
// Each constraint is checked and the result is recorded
type SafetyValidation struct {
	ConstraintName string             // Name of the constraint checked
	Passed         bool               // Whether the constraint was satisfied
	Severity       ConstraintSeverity // How critical was this constraint
	Message        string             // Human-readable result message
	ViolationType  string             // Type of violation (if any)
	Evidence       interface{}        // Evidence supporting this result
	Timestamp      time.Time
}

// ============================================================================
// Performance Metrics - For Monitoring and Adaptation
// ============================================================================

// CognitiveMetrics holds performance and operational metrics for a component
// Used for monitoring, adaptation, and identifying bottlenecks
type CognitiveMetrics struct {
	ComponentName      string
	TotalRequests      int64         // Total requests processed
	SuccessfulRequests int64         // Requests that succeeded
	FailedRequests     int64         // Requests that failed
	AverageLatency     time.Duration // Average processing time
	PeakLatency        time.Duration // Maximum processing time observed
	MinLatency         time.Duration // Minimum processing time observed
	ErrorRate          float64       // Fraction of requests that failed
	LastUpdated        time.Time
	CustomMetrics      map[string]interface{} // Component-specific metrics
}

// ============================================================================
// Cognitive Component Registry - For Component Management
// ============================================================================

// CognitiveComponentRegistry manages the registration and lookup of cognitive components
// Used during Phase 0 to register components, and during execution to find components by name
type CognitiveComponentRegistry struct {
	components map[string]CognitiveComponent
}

// NewCognitiveComponentRegistry creates a new component registry
func NewCognitiveComponentRegistry() *CognitiveComponentRegistry {
	return &CognitiveComponentRegistry{
		components: make(map[string]CognitiveComponent),
	}
}

// Register adds a component to the registry
func (r *CognitiveComponentRegistry) Register(name string, component CognitiveComponent) error {
	if _, exists := r.components[name]; exists {
		return ErrComponentAlreadyRegistered
	}
	r.components[name] = component
	return nil
}

// Get retrieves a component from the registry by name
func (r *CognitiveComponentRegistry) Get(name string) (CognitiveComponent, error) {
	component, exists := r.components[name]
	if !exists {
		return nil, ErrComponentNotFound
	}
	return component, nil
}

// List returns all registered component names
func (r *CognitiveComponentRegistry) List() []string {
	names := make([]string, 0, len(r.components))
	for name := range r.components {
		names = append(names, name)
	}
	return names
}

// Count returns the number of registered components
func (r *CognitiveComponentRegistry) Count() int {
	return len(r.components)
}

// ============================================================================
// Cognitive Processing Chain - For Orchestration
// ============================================================================

// CognitiveProcessingChain orchestrates multiple cognitive components in sequence
// Used to execute a pipeline of components, passing output of one to input of the next
type CognitiveProcessingChain struct {
	components []CognitiveComponent
	chain      []string // Names of components in order
}

// NewCognitiveProcessingChain creates a new processing chain
func NewCognitiveProcessingChain(components []CognitiveComponent, names []string) *CognitiveProcessingChain {
	return &CognitiveProcessingChain{
		components: components,
		chain:      names,
	}
}

// Execute runs components in sequence through the cognitive chain
// Returns early if any component returns an error or impasse
func (chain *CognitiveProcessingChain) Execute(ctx context.Context, request *CognitiveProcessRequest) (*CognitiveProcessResult, error) {
	result := &CognitiveProcessResult{
		Status:             ProcessSuccess,
		ComponentsInvolved: chain.chain,
		ExecutionSteps:     make([]*ExecutionStep, 0),
		SafetyCheckResults: make([]SafetyValidation, 0),
	}

	// Track operation chain for debugging
	request.OperationChain = append(request.OperationChain, "chain_start")

	for i, component := range chain.components {
		select {
		case <-ctx.Done():
			result.Status = ProcessError
			result.Error = ctx.Err()
			return result, ctx.Err()
		default:
		}

		stepStart := time.Now()

		// Execute component
		stepResult, err := component.Process(ctx, request)
		stepDuration := time.Since(stepStart)

		// Record execution step
		step := &ExecutionStep{
			ComponentName: component.GetName(),
			StepNumber:    i,
			Input:         request,
			Output:        stepResult,
			Duration:      stepDuration,
			Status:        "success",
		}

		if err != nil {
			step.Status = "error"
			step.Error = err
			result.Error = err
			result.Status = ProcessError
			result.ExecutionSteps = append(result.ExecutionSteps, step)
			return result, err
		}

		result.ExecutionSteps = append(result.ExecutionSteps, step)

		// Check if component encountered an impasse
		if stepResult.RequiresImpasse {
			result.RequiresImpasse = true
			result.ImpasseType = stepResult.ImpasseType
			result.Status = ProcessImpasse
			result.Output = stepResult.Output
			return result, nil
		}

		// Update tracking
		request.OperationChain = append(request.OperationChain, component.GetName())

		// Pass result output as next component's context (if both are interfaces)
		// In actual use, components would update relevant state in working memory/goal stack
	}

	result.Status = ProcessSuccess
	return result, nil
}

// GetComponentNames returns the list of component names in this chain
func (chain *CognitiveProcessingChain) GetComponentNames() []string {
	return chain.chain
}

// GetComponentCount returns the number of components in this chain
func (chain *CognitiveProcessingChain) GetComponentCount() int {
	return len(chain.components)
}

// ============================================================================
// Error Types - For Proper Error Handling
// ============================================================================

var (
	// ErrComponentNotFound indicates a component was not found in the registry
	ErrComponentNotFound = NewCognitiveError("component_not_found", "cognitive component not found in registry")

	// ErrComponentAlreadyRegistered indicates a component with that name already exists
	ErrComponentAlreadyRegistered = NewCognitiveError("component_exists", "cognitive component already registered with that name")

	// ErrInvalidProcessRequest indicates the request is invalid
	ErrInvalidProcessRequest = NewCognitiveError("invalid_request", "invalid or missing cognitive process request")

	// ErrProcessingFailed indicates processing encountered a fatal error
	ErrProcessingFailed = NewCognitiveError("processing_failed", "cognitive processing failed")

	// ErrComponentInitializationFailed indicates initialization failed
	ErrComponentInitializationFailed = NewCognitiveError("init_failed", "cognitive component initialization failed")
)

// CognitiveError represents an error in the cognitive system
type CognitiveError struct {
	Code    string
	Message string
}

// NewCognitiveError creates a new cognitive error
func NewCognitiveError(code, message string) *CognitiveError {
	return &CognitiveError{
		Code:    code,
		Message: message,
	}
}

// Error implements the error interface
func (ce *CognitiveError) Error() string {
	return ce.Message
}

// String provides string representation with code
func (ce *CognitiveError) String() string {
	return "CognitiveError[" + ce.Code + "]: " + ce.Message
}
