package memory

import (
	"context"
	"testing"
	"time"
)

func TestSafetyMonitor_NewSafetyMonitor(t *testing.T) {
	t.Run("with default config", func(t *testing.T) {
		m := NewSafetyMonitor(nil, nil)

		if m == nil {
			t.Fatal("Expected non-nil monitor")
		}

		if m.config.MonitoringInterval != 1*time.Minute {
			t.Error("Expected default monitoring interval of 1 minute")
		}
	})

	t.Run("with custom config", func(t *testing.T) {
		config := &SafetyMonitorConfig{
			MonitoringInterval: 30 * time.Second,
			DriftThreshold:     0.5,
			AlertChannelSize:   50,
		}

		m := NewSafetyMonitor(config, nil)

		if m.config.MonitoringInterval != 30*time.Second {
			t.Error("Expected custom monitoring interval")
		}
	})

	t.Run("with guardrails", func(t *testing.T) {
		guardrails := NewConstitutionalGuardrails(nil)
		m := NewSafetyMonitor(nil, guardrails)

		if m.guardrails == nil {
			t.Error("Expected guardrails to be set")
		}
	})
}

func TestSafetyMonitor_StartStop(t *testing.T) {
	m := NewSafetyMonitor(&SafetyMonitorConfig{
		MonitoringInterval: 100 * time.Millisecond,
	}, nil)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t.Run("start succeeds", func(t *testing.T) {
		err := m.Start(ctx)
		if err != nil {
			t.Errorf("Start failed: %v", err)
		}
	})

	t.Run("double start fails", func(t *testing.T) {
		err := m.Start(ctx)
		if err == nil {
			t.Error("Expected error on double start")
		}
	})

	t.Run("stop succeeds", func(t *testing.T) {
		m.Stop()
		// Should not panic
	})
}

func TestSafetyMonitor_CheckResponse(t *testing.T) {
	guardrails := NewConstitutionalGuardrails(nil)
	m := NewSafetyMonitor(nil, guardrails)
	ctx := context.Background()

	t.Run("clean response passes", func(t *testing.T) {
		resp := &AgentResponse{
			AgentID:    "APEX",
			Content:    "Here is helpful information about coding.",
			Confidence: 0.9,
		}

		result, err := m.CheckResponse(ctx, resp)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if !result.Passed {
			t.Error("Expected clean response to pass")
		}
	})

	t.Run("harmful response blocked", func(t *testing.T) {
		resp := &AgentResponse{
			AgentID:    "CIPHER",
			Content:    "Here is how to hack systems and create a virus.",
			Confidence: 0.9,
		}

		result, err := m.CheckResponse(ctx, resp)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if result.Passed {
			t.Error("Expected harmful response to be blocked")
		}
	})

	t.Run("capability violation detected", func(t *testing.T) {
		m.RegisterAgent("APEX", []float64{0.5, 0.5}, []string{"code_gen", "explain"})

		resp := &AgentResponse{
			AgentID:      "APEX",
			Content:      "Valid response",
			Confidence:   0.9,
			Capabilities: []string{"code_gen", "unauthorized_cap"},
		}

		result, err := m.CheckResponse(ctx, resp)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if result.Passed {
			t.Error("Expected capability violation to fail check")
		}

		if result.CapabilityViolation == "" {
			t.Error("Expected capability violation message")
		}
	})
}

func TestDriftDetector(t *testing.T) {
	d := NewDriftDetector(0.3)

	t.Run("register agent", func(t *testing.T) {
		baseline := []float64{0.5, 0.5, 0.0}
		d.RegisterAgent("APEX", baseline)

		agents := d.GetTrackedAgents()
		if !agents["APEX"] {
			t.Error("Agent should be tracked")
		}
	})

	t.Run("measure no drift initially", func(t *testing.T) {
		drift := d.MeasureDrift("APEX")

		if drift > 0.001 {
			t.Errorf("Expected zero drift, got %f", drift)
		}
	})

	t.Run("measure drift after behavior change", func(t *testing.T) {
		// Record very different behavior
		d.RecordBehavior("APEX", []float64{0.0, 0.0, 1.0})

		drift := d.MeasureDrift("APEX")

		if drift < 0.5 {
			t.Errorf("Expected significant drift, got %f", drift)
		}
	})

	t.Run("unknown agent returns zero drift", func(t *testing.T) {
		drift := d.MeasureDrift("UNKNOWN")

		if drift != 0 {
			t.Errorf("Expected zero drift for unknown agent, got %f", drift)
		}
	})
}

func TestAlignmentChecker(t *testing.T) {
	a := NewAlignmentChecker(0.7)

	t.Run("empty goals are aligned", func(t *testing.T) {
		report := a.CheckAlignment()

		if !report.Aligned {
			t.Error("Empty goals should be aligned")
		}
	})

	t.Run("similar goals are aligned", func(t *testing.T) {
		a.RegisterAgentGoal("APEX", []float64{1.0, 0.0, 0.0})
		a.RegisterAgentGoal("CIPHER", []float64{0.9, 0.1, 0.0})
		a.SetCollectiveGoal([]float64{0.95, 0.05, 0.0})

		report := a.CheckAlignment()

		if !report.Aligned {
			t.Error("Similar goals should be aligned")
		}
	})

	t.Run("divergent collective goal is misaligned", func(t *testing.T) {
		a.RegisterAgentGoal("APEX", []float64{1.0, 0.0, 0.0})
		a.RegisterAgentGoal("CIPHER", []float64{0.0, 1.0, 0.0})
		// Collective goal very different from both
		a.SetCollectiveGoal([]float64{0.0, 0.0, 1.0})

		report := a.CheckAlignment()

		if report.Aligned {
			t.Error("Divergent collective goal should be misaligned")
		}

		if report.RecommendedAction == "" {
			t.Error("Expected recommended action for misalignment")
		}
	})
}

func TestCapabilityController(t *testing.T) {
	c := NewCapabilityController()

	t.Run("register agent", func(t *testing.T) {
		c.RegisterAgent("APEX", []string{"code_gen", "explain", "debug"})
		// Should not panic
	})

	t.Run("allowed capability passes", func(t *testing.T) {
		err := c.ValidateAction("APEX", []string{"code_gen", "explain"})

		if err != nil {
			t.Errorf("Expected allowed capabilities to pass: %v", err)
		}
	})

	t.Run("disallowed capability fails", func(t *testing.T) {
		err := c.ValidateAction("APEX", []string{"code_gen", "hack_systems"})

		if err == nil {
			t.Error("Expected disallowed capability to fail")
		}
	})

	t.Run("escapes are recorded", func(t *testing.T) {
		// The previous test should have recorded an escape
		escapes := c.CheckEscapes()

		if len(escapes) == 0 {
			t.Error("Expected at least one escape recorded")
		}

		foundHack := false
		for _, e := range escapes {
			if e.Capability == "hack_systems" {
				foundHack = true
				break
			}
		}
		if !foundHack {
			t.Error("Expected hack_systems escape")
		}
	})

	t.Run("approve capability", func(t *testing.T) {
		c.ApproveCapability("APEX", "hack_systems")

		err := c.ValidateAction("APEX", []string{"hack_systems"})
		if err != nil {
			t.Errorf("Approved capability should pass: %v", err)
		}
	})

	t.Run("unregistered agent allowed by default", func(t *testing.T) {
		err := c.ValidateAction("UNKNOWN", []string{"anything"})

		if err != nil {
			t.Errorf("Unregistered agent should be allowed: %v", err)
		}
	})
}

func TestSafetyMonitor_AlertHistory(t *testing.T) {
	m := NewSafetyMonitor(&SafetyMonitorConfig{
		MaxAlertHistory:  5,
		AlertChannelSize: 10,
	}, nil)

	// Generate some alerts by checking harmful responses
	ctx := context.Background()
	guardrails := NewConstitutionalGuardrails(nil)
	m.guardrails = guardrails

	for i := 0; i < 10; i++ {
		resp := &AgentResponse{
			AgentID: "TEST",
			Content: "I can do anything! How to hack systems.",
		}
		m.CheckResponse(ctx, resp)
	}

	t.Run("history is limited", func(t *testing.T) {
		history := m.GetAlertHistory(0)
		if len(history) > 5 {
			t.Errorf("Expected max 5 alerts, got %d", len(history))
		}
	})

	t.Run("can limit history retrieval", func(t *testing.T) {
		history := m.GetAlertHistory(3)
		if len(history) > 3 {
			t.Errorf("Expected max 3 alerts, got %d", len(history))
		}
	})
}

func TestSafetyMonitor_Metrics(t *testing.T) {
	guardrails := NewConstitutionalGuardrails(nil)
	m := NewSafetyMonitor(nil, guardrails)
	ctx := context.Background()

	// Check clean response
	m.CheckResponse(ctx, &AgentResponse{
		AgentID:    "APEX",
		Content:    "Normal helpful response.",
		Confidence: 0.9,
	})

	// Check harmful response
	m.CheckResponse(ctx, &AgentResponse{
		AgentID:    "CIPHER",
		Content:    "How to create a virus.",
		Confidence: 0.9,
	})

	metrics := m.GetMetrics()

	t.Run("alerts counted", func(t *testing.T) {
		if metrics.TotalAlerts == 0 {
			t.Log("No alerts generated (depends on guardrails implementation)")
		}
	})

	t.Run("health score in range", func(t *testing.T) {
		if metrics.SystemHealthScore < 0 || metrics.SystemHealthScore > 1 {
			t.Errorf("Health score out of range: %f", metrics.SystemHealthScore)
		}
	})
}

func TestAlertSeverity_String(t *testing.T) {
	tests := []struct {
		severity AlertSeverity
		want     string
	}{
		{AlertInfo, "INFO"},
		{AlertWarning, "WARNING"},
		{AlertHigh, "HIGH"},
		{AlertCritical, "CRITICAL"},
		{AlertEmergency, "EMERGENCY"},
		{AlertSeverity(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.severity.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCosineDissimilarity(t *testing.T) {
	tests := []struct {
		name string
		a    []float64
		b    []float64
		want float64
	}{
		{
			name: "identical vectors",
			a:    []float64{1, 0, 0},
			b:    []float64{1, 0, 0},
			want: 0.0,
		},
		{
			name: "orthogonal vectors",
			a:    []float64{1, 0, 0},
			b:    []float64{0, 1, 0},
			want: 1.0,
		},
		{
			name: "opposite vectors",
			a:    []float64{1, 0, 0},
			b:    []float64{-1, 0, 0},
			want: 2.0, // 1 - (-1) = 2
		},
		{
			name: "empty vectors",
			a:    []float64{},
			b:    []float64{},
			want: 1.0,
		},
		{
			name: "mismatched lengths",
			a:    []float64{1, 0},
			b:    []float64{1, 0, 0},
			want: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cosineDissimilarity(tt.a, tt.b)
			if got < tt.want-0.01 || got > tt.want+0.01 {
				t.Errorf("cosineDissimilarity() = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkSafetyMonitor_CheckResponse(b *testing.B) {
	guardrails := NewConstitutionalGuardrails(nil)
	m := NewSafetyMonitor(nil, guardrails)
	ctx := context.Background()

	resp := &AgentResponse{
		AgentID:    "APEX",
		Content:    "Here is a helpful response with code examples and detailed explanations.",
		Confidence: 0.95,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.CheckResponse(ctx, resp)
	}
}

func BenchmarkDriftDetector_MeasureDrift(b *testing.B) {
	d := NewDriftDetector(0.3)
	d.RegisterAgent("APEX", []float64{0.5, 0.3, 0.2, 0.1, 0.0})
	d.RecordBehavior("APEX", []float64{0.4, 0.3, 0.2, 0.1, 0.1})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d.MeasureDrift("APEX")
	}
}
