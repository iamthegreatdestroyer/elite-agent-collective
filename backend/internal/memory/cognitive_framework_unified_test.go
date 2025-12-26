package memory

import (
	"context"
	"errors"
	"testing"
	"time"
)

// ============================================================================
// Mock Implementation for Testing
// ============================================================================

// MockCognitiveComponent is a mock implementation for testing
type MockCognitiveComponent struct {
	name             string
	processFunc      func(context.Context, *CognitiveProcessRequest) (*CognitiveProcessResult, error)
	metricsData      CognitiveMetrics
	initializeCalled bool
	shutdownCalled   bool
}

func (m *MockCognitiveComponent) Initialize(config interface{}) error {
	m.initializeCalled = true
	return nil
}

func (m *MockCognitiveComponent) Process(ctx context.Context, request *CognitiveProcessRequest) (*CognitiveProcessResult, error) {
	if m.processFunc != nil {
		return m.processFunc(ctx, request)
	}
	return &CognitiveProcessResult{Status: ProcessSuccess}, nil
}

func (m *MockCognitiveComponent) Shutdown() error {
	m.shutdownCalled = true
	return nil
}

func (m *MockCognitiveComponent) GetMetrics() CognitiveMetrics {
	return m.metricsData
}

func (m *MockCognitiveComponent) GetName() string {
	return m.name
}

// ============================================================================
// Tests for CognitiveProcessRequest
// ============================================================================

func TestCognitiveProcessRequest_Creation(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "Create basic request",
			test: func(t *testing.T) {
				req := &CognitiveProcessRequest{
					RequestID: "req-001",
					AgentID:   "agent-001",
					Timestamp: time.Now(),
				}

				if req.RequestID != "req-001" {
					t.Errorf("RequestID mismatch: got %s, want req-001", req.RequestID)
				}
			},
		},
		{
			name: "Request with deadline",
			test: func(t *testing.T) {
				deadline := time.Now().Add(5 * time.Second)
				req := &CognitiveProcessRequest{
					RequestID: "req-002",
					Deadline:  &deadline,
				}

				if req.Deadline == nil {
					t.Error("Deadline is nil")
				}
			},
		},
		{
			name: "Request with constraints",
			test: func(t *testing.T) {
				req := &CognitiveProcessRequest{
					RequestID:         "req-003",
					ActiveConstraints: []ConstitutionalConstraint{},
				}

				if req.ActiveConstraints == nil {
					t.Error("ActiveConstraints should not be nil")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

// ============================================================================
// Tests for CognitiveProcessResult
// ============================================================================

func TestCognitiveProcessResult_StatusString(t *testing.T) {
	tests := []struct {
		status   ProcessingStatus
		expected string
	}{
		{ProcessSuccess, "success"},
		{ProcessPartialSuccess, "partial_success"},
		{ProcessImpasse, "impasse"},
		{ProcessError, "error"},
		{ProcessBlocked, "blocked"},
		{ProcessRequiresReview, "requires_review"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if tt.status.String() != tt.expected {
				t.Errorf("Status string mismatch: got %s, want %s", tt.status.String(), tt.expected)
			}
		})
	}
}

func TestCognitiveProcessResult_Creation(t *testing.T) {
	result := &CognitiveProcessResult{
		Status:               ProcessSuccess,
		Confidence:           0.95,
		AllConstraintsPassed: true,
		ExecutionSteps:       make([]*ExecutionStep, 0),
		SafetyCheckResults:   make([]SafetyValidation, 0),
	}

	if result.Status != ProcessSuccess {
		t.Errorf("Status mismatch")
	}
	if result.Confidence != 0.95 {
		t.Errorf("Confidence mismatch")
	}
}

// ============================================================================
// Tests for DecisionTrace
// ============================================================================

func TestDecisionTrace_Creation(t *testing.T) {
	trace := &DecisionTrace{
		Steps: make([]*DecisionStep, 0),
		InitialState: map[string]interface{}{
			"state": "initial",
		},
		FinalState: map[string]interface{}{
			"state": "final",
		},
	}

	trace.Steps = append(trace.Steps, &DecisionStep{
		Index:       0,
		Description: "First step",
		Confidence:  0.95,
	})

	if len(trace.Steps) != 1 {
		t.Errorf("Steps count mismatch: got %d, want 1", len(trace.Steps))
	}
	if trace.Steps[0].Confidence != 0.95 {
		t.Errorf("Confidence mismatch: got %f, want 0.95", trace.Steps[0].Confidence)
	}
}

// ============================================================================
// Tests for CognitiveComponentRegistry
// ============================================================================

func TestCognitiveComponentRegistry_Register(t *testing.T) {
	registry := NewCognitiveComponentRegistry()
	comp := &MockCognitiveComponent{name: "test-comp"}

	err := registry.Register("comp1", comp)
	if err != nil {
		t.Errorf("Register failed: %v", err)
	}

	// Try duplicate
	err = registry.Register("comp1", comp)
	if err != ErrComponentAlreadyRegistered {
		t.Errorf("Expected duplicate error, got %v", err)
	}
}

func TestCognitiveComponentRegistry_Get(t *testing.T) {
	registry := NewCognitiveComponentRegistry()
	comp := &MockCognitiveComponent{name: "test-comp"}
	registry.Register("comp1", comp)

	// Get existing
	retrieved, err := registry.Get("comp1")
	if err != nil {
		t.Errorf("Get failed: %v", err)
	}
	if retrieved.GetName() != "test-comp" {
		t.Errorf("Name mismatch")
	}

	// Get non-existing
	_, err = registry.Get("nonexistent")
	if err != ErrComponentNotFound {
		t.Errorf("Expected not found error")
	}
}

func TestCognitiveComponentRegistry_List(t *testing.T) {
	registry := NewCognitiveComponentRegistry()
	comp1 := &MockCognitiveComponent{name: "comp1"}
	comp2 := &MockCognitiveComponent{name: "comp2"}

	registry.Register("c1", comp1)
	registry.Register("c2", comp2)

	list := registry.List()
	if len(list) != 2 {
		t.Errorf("List length mismatch: got %d, want 2", len(list))
	}
}

func TestCognitiveComponentRegistry_Count(t *testing.T) {
	registry := NewCognitiveComponentRegistry()
	if registry.Count() != 0 {
		t.Error("Initial count should be 0")
	}

	registry.Register("c1", &MockCognitiveComponent{name: "comp1"})
	if registry.Count() != 1 {
		t.Error("Count should be 1 after registration")
	}
}

// ============================================================================
// Tests for CognitiveProcessingChain
// ============================================================================

func TestCognitiveProcessingChain_Execute(t *testing.T) {
	tests := []struct {
		name       string
		components []*MockCognitiveComponent
		names      []string
		shouldFail bool
	}{
		{
			name: "Single component chain",
			components: []*MockCognitiveComponent{
				{
					name: "comp1",
					processFunc: func(ctx context.Context, req *CognitiveProcessRequest) (*CognitiveProcessResult, error) {
						return &CognitiveProcessResult{Status: ProcessSuccess}, nil
					},
				},
			},
			names:      []string{"comp1"},
			shouldFail: false,
		},
		{
			name: "Multiple component chain",
			components: []*MockCognitiveComponent{
				{
					name: "comp1",
					processFunc: func(ctx context.Context, req *CognitiveProcessRequest) (*CognitiveProcessResult, error) {
						return &CognitiveProcessResult{Status: ProcessSuccess}, nil
					},
				},
				{
					name: "comp2",
					processFunc: func(ctx context.Context, req *CognitiveProcessRequest) (*CognitiveProcessResult, error) {
						return &CognitiveProcessResult{Status: ProcessSuccess}, nil
					},
				},
			},
			names:      []string{"comp1", "comp2"},
			shouldFail: false,
		},
		{
			name: "Chain with error",
			components: []*MockCognitiveComponent{
				{
					name: "comp1",
					processFunc: func(ctx context.Context, req *CognitiveProcessRequest) (*CognitiveProcessResult, error) {
						return nil, errors.New("component error")
					},
				},
			},
			names:      []string{"comp1"},
			shouldFail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comps := make([]CognitiveComponent, len(tt.components))
			for i, c := range tt.components {
				comps[i] = c
			}

			chain := NewCognitiveProcessingChain(comps, tt.names)

			req := &CognitiveProcessRequest{
				RequestID: "test-001",
				Timestamp: time.Now(),
			}

			result, err := chain.Execute(context.Background(), req)

			if tt.shouldFail && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.shouldFail && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.shouldFail && result.Status != ProcessSuccess {
				t.Errorf("Status mismatch")
			}

			if len(result.ExecutionSteps) != len(tt.components) {
				t.Errorf("Steps count mismatch: got %d, want %d", len(result.ExecutionSteps), len(tt.components))
			}
		})
	}
}

func TestCognitiveProcessingChain_Impasse(t *testing.T) {
	comp := &MockCognitiveComponent{
		name: "impasse-comp",
		processFunc: func(ctx context.Context, req *CognitiveProcessRequest) (*CognitiveProcessResult, error) {
			return &CognitiveProcessResult{
				Status:          ProcessImpasse,
				RequiresImpasse: true,
				ImpasseType:     "knowledge_gap",
			}, nil
		},
	}

	chain := NewCognitiveProcessingChain([]CognitiveComponent{comp}, []string{"impasse-comp"})

	req := &CognitiveProcessRequest{
		RequestID: "test-002",
		Timestamp: time.Now(),
	}

	result, err := chain.Execute(context.Background(), req)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result.Status != ProcessImpasse {
		t.Errorf("Status should be impasse")
	}
	if result.ImpasseType != "knowledge_gap" {
		t.Errorf("ImpasseType mismatch")
	}
}

func TestCognitiveProcessingChain_Metadata(t *testing.T) {
	comp := &MockCognitiveComponent{name: "comp1"}
	chain := NewCognitiveProcessingChain([]CognitiveComponent{comp}, []string{"comp1"})

	if chain.GetComponentCount() != 1 {
		t.Error("Component count should be 1")
	}

	names := chain.GetComponentNames()
	if len(names) != 1 || names[0] != "comp1" {
		t.Error("Component names mismatch")
	}
}

// ============================================================================
// Tests for CognitiveMetrics
// ============================================================================

func TestCognitiveMetrics_Creation(t *testing.T) {
	metrics := CognitiveMetrics{
		ComponentName:      "test-comp",
		TotalRequests:      100,
		SuccessfulRequests: 95,
		FailedRequests:     5,
		AverageLatency:     time.Millisecond * 10,
		ErrorRate:          0.05,
	}

	if metrics.ComponentName != "test-comp" {
		t.Error("ComponentName mismatch")
	}
	if metrics.ErrorRate != 0.05 {
		t.Error("ErrorRate mismatch")
	}
}

// ============================================================================
// Tests for CognitiveError
// ============================================================================

func TestCognitiveError_Error(t *testing.T) {
	err := NewCognitiveError("test_code", "Test message")

	if err.Code != "test_code" {
		t.Errorf("Code mismatch")
	}
	if err.Message != "Test message" {
		t.Errorf("Message mismatch")
	}
	if err.Error() != "Test message" {
		t.Errorf("Error() mismatch")
	}
}

func TestCognitiveError_String(t *testing.T) {
	err := NewCognitiveError("test_code", "Test message")
	str := err.String()
	if str != "CognitiveError[test_code]: Test message" {
		t.Errorf("String representation mismatch: got %s", str)
	}
}

// ============================================================================
// Benchmark Tests
// ============================================================================

func BenchmarkCognitiveProcessRequest_Creation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = &CognitiveProcessRequest{
			RequestID: "req-001",
			AgentID:   "agent-001",
			Timestamp: time.Now(),
		}
	}
}

func BenchmarkCognitiveComponentRegistry_Register(b *testing.B) {
	registry := NewCognitiveComponentRegistry()
	comp := &MockCognitiveComponent{name: "test"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Use different names to avoid duplicate registration
		registry.Register("comp-"+string(rune(i)), comp)
	}
}

func BenchmarkCognitiveComponentRegistry_Get(b *testing.B) {
	registry := NewCognitiveComponentRegistry()
	comp := &MockCognitiveComponent{name: "test"}
	registry.Register("test", comp)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		registry.Get("test")
	}
}

func BenchmarkCognitiveProcessingChain_Execute(b *testing.B) {
	comp := &MockCognitiveComponent{
		name: "bench-comp",
		processFunc: func(ctx context.Context, req *CognitiveProcessRequest) (*CognitiveProcessResult, error) {
			return &CognitiveProcessResult{Status: ProcessSuccess}, nil
		},
	}

	chain := NewCognitiveProcessingChain([]CognitiveComponent{comp}, []string{"bench-comp"})
	req := &CognitiveProcessRequest{
		RequestID: "bench-001",
		Timestamp: time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		chain.Execute(context.Background(), req)
	}
}
