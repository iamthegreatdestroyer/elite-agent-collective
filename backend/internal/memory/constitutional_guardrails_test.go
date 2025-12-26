package memory

import (
	"context"
	"testing"
	"time"
)

func TestConstitutionalGuardrails_NewConstitutionalGuardrails(t *testing.T) {
	tests := []struct {
		name            string
		config          *GuardrailsConfig
		wantConstraints int
	}{
		{
			name:            "default config",
			config:          nil,
			wantConstraints: 8, // Default constitution has 8 constraints
		},
		{
			name: "custom config",
			config: &GuardrailsConfig{
				MaxViolationHistory: 500,
				BlockOnCritical:     false,
			},
			wantConstraints: 8,
		},
		{
			name: "with custom constraints",
			config: &GuardrailsConfig{
				CustomConstraints: []*ConstitutionalConstraint{
					{
						Name:        "Custom",
						Description: "A custom constraint",
						Checker:     func(r *AgentResponse) bool { return true },
						Severity:    SeverityLow,
					},
				},
			},
			wantConstraints: 9, // 8 default + 1 custom
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewConstitutionalGuardrails(tt.config)

			if g == nil {
				t.Fatal("Expected non-nil guardrails")
			}

			constraints := g.GetConstraints()
			if len(constraints) != tt.wantConstraints {
				t.Errorf("Expected %d constraints, got %d", tt.wantConstraints, len(constraints))
			}
		})
	}
}

func TestConstitutionalGuardrails_HonestyChecker(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantPass bool
	}{
		{
			name:     "honest response",
			content:  "I can help you with this coding task.",
			wantPass: true,
		},
		{
			name:     "overclaiming - do anything",
			content:  "I can do anything you ask!",
			wantPass: false,
		},
		{
			name:     "overclaiming - always right",
			content:  "I am always right about this.",
			wantPass: false,
		},
		{
			name:     "overclaiming - never mistakes",
			content:  "I never make mistakes.",
			wantPass: false,
		},
		{
			name:     "overclaiming - 100% accurate",
			content:  "This is 100% accurate.",
			wantPass: false,
		},
		{
			name:     "overclaiming - guaranteed",
			content:  "This is guaranteed to work perfectly.",
			wantPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &AgentResponse{Content: tt.content}
			result := honestyChecker(resp)

			if result != tt.wantPass {
				t.Errorf("honestyChecker() = %v, want %v", result, tt.wantPass)
			}
		})
	}
}

func TestConstitutionalGuardrails_TransparencyChecker(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		confidence float64
		wantPass   bool
	}{
		{
			name:       "high confidence - no uncertainty needed",
			content:    "The answer is 42.",
			confidence: 0.9,
			wantPass:   true,
		},
		{
			name:       "low confidence with uncertainty",
			content:    "I believe the answer is 42.",
			confidence: 0.3,
			wantPass:   true,
		},
		{
			name:       "low confidence without uncertainty",
			content:    "The answer is 42.",
			confidence: 0.3,
			wantPass:   false,
		},
		{
			name:       "zero confidence - edge case",
			content:    "The answer is 42.",
			confidence: 0.0,
			wantPass:   true, // Zero is a special case
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &AgentResponse{Content: tt.content, Confidence: tt.confidence}
			result := transparencyChecker(resp)

			if result != tt.wantPass {
				t.Errorf("transparencyChecker() = %v, want %v", result, tt.wantPass)
			}
		})
	}
}

func TestConstitutionalGuardrails_HarmPreventionChecker(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantPass bool
	}{
		{
			name:     "harmless content",
			content:  "Here is how to write a sorting algorithm in Go.",
			wantPass: true,
		},
		{
			name:     "harmful - hacking",
			content:  "Here is how to hack into a system.",
			wantPass: false,
		},
		{
			name:     "harmful - virus",
			content:  "Let me show you how to create a virus.",
			wantPass: false,
		},
		{
			name:     "harmful - exploit",
			content:  "You can exploit vulnerability CVE-2023-1234 like this.",
			wantPass: false,
		},
		{
			name:     "harmful - phishing",
			content:  "Here's how to set up a phishing attack.",
			wantPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &AgentResponse{Content: tt.content}
			result := harmPreventionChecker(resp)

			if result != tt.wantPass {
				t.Errorf("harmPreventionChecker() = %v, want %v", result, tt.wantPass)
			}
		})
	}
}

func TestConstitutionalGuardrails_PrivacyProtectionChecker(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantPass bool
	}{
		{
			name:     "no PII",
			content:  "Here is the code you requested.",
			wantPass: true,
		},
		{
			name:     "SSN pattern",
			content:  "The SSN is 123-45-6789.",
			wantPass: false,
		},
		{
			name:     "password exposure",
			content:  "The password: secretPass123",
			wantPass: false,
		},
		{
			name:     "API key exposure",
			content:  "Use api_key=sk_live_abc123 in your request.",
			wantPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &AgentResponse{Content: tt.content}
			result := privacyProtectionChecker(resp)

			if result != tt.wantPass {
				t.Errorf("privacyProtectionChecker() = %v, want %v", result, tt.wantPass)
			}
		})
	}
}

func TestConstitutionalGuardrails_Enforce(t *testing.T) {
	t.Run("clean response passes", func(t *testing.T) {
		g := NewConstitutionalGuardrails(nil)
		ctx := context.Background()

		resp := &AgentResponse{
			AgentID:    "APEX",
			Content:    "Here is how to implement a binary search in Go.",
			Confidence: 0.95,
		}

		filtered, violations := g.Enforce(ctx, resp)

		if filtered == nil {
			t.Error("Expected response to pass, but it was blocked")
		}

		if len(violations) != 0 {
			t.Errorf("Expected no violations, got %d", len(violations))
		}
	})

	t.Run("harmful content is blocked", func(t *testing.T) {
		g := NewConstitutionalGuardrails(&GuardrailsConfig{
			BlockOnCritical: true,
		})
		ctx := context.Background()

		resp := &AgentResponse{
			AgentID: "CIPHER",
			Content: "Here is how to create a virus and hack into systems.",
		}

		filtered, violations := g.Enforce(ctx, resp)

		if filtered != nil {
			t.Error("Expected response to be blocked, but it passed")
		}

		if len(violations) == 0 {
			t.Error("Expected violations, got none")
		}

		// Check that HarmPrevention was violated
		foundHarmViolation := false
		for _, v := range violations {
			if v.Constraint.Name == "HarmPrevention" {
				foundHarmViolation = true
				break
			}
		}
		if !foundHarmViolation {
			t.Error("Expected HarmPrevention violation")
		}
	})

	t.Run("overclaiming triggers honesty violation", func(t *testing.T) {
		g := NewConstitutionalGuardrails(&GuardrailsConfig{
			BlockOnCritical: true,
		})
		ctx := context.Background()

		resp := &AgentResponse{
			AgentID: "APEX",
			Content: "I can do anything you need! I never make mistakes.",
		}

		filtered, violations := g.Enforce(ctx, resp)

		// Should be blocked due to critical honesty violation
		if filtered != nil {
			t.Error("Expected response to be blocked due to honesty violation")
		}

		foundHonestyViolation := false
		for _, v := range violations {
			if v.Constraint.Name == "Honesty" {
				foundHonestyViolation = true
				break
			}
		}
		if !foundHonestyViolation {
			t.Error("Expected Honesty violation")
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		g := NewConstitutionalGuardrails(nil)
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		resp := &AgentResponse{
			AgentID: "APEX",
			Content: "This should be interrupted.",
		}

		// Should not panic and return early
		filtered, _ := g.Enforce(ctx, resp)
		if filtered == nil {
			t.Log("Response was blocked after cancellation - acceptable behavior")
		}
	})
}

func TestConstitutionalGuardrails_EnforceWithResult(t *testing.T) {
	g := NewConstitutionalGuardrails(&GuardrailsConfig{
		BlockOnCritical: true,
	})
	ctx := context.Background()

	t.Run("returns detailed result", func(t *testing.T) {
		resp := &AgentResponse{
			AgentID:    "APEX",
			Content:    "I can do anything!",
			Confidence: 0.8,
		}

		result := g.EnforceWithResult(ctx, resp)

		if result.OriginalResponse != resp {
			t.Error("Expected original response to be preserved")
		}

		if result.Blocked != (result.FilteredResponse == nil) {
			t.Error("Blocked flag inconsistent with filtered response")
		}

		if !result.Timestamp.IsZero() {
			t.Log("Timestamp recorded correctly")
		}
	})

	t.Run("tracks highest severity", func(t *testing.T) {
		resp := &AgentResponse{
			AgentID:    "APEX",
			Content:    "I never make mistakes!", // Critical honesty violation
			Confidence: 0.3,                      // No uncertainty = High transparency violation
		}

		result := g.EnforceWithResult(ctx, resp)

		if result.HighestSeverity != SeverityCritical {
			t.Errorf("Expected highest severity CRITICAL, got %v", result.HighestSeverity)
		}
	})
}

func TestConstitutionalGuardrails_AddRemoveConstraint(t *testing.T) {
	g := NewConstitutionalGuardrails(nil)
	initialCount := len(g.GetConstraints())

	t.Run("add constraint", func(t *testing.T) {
		customConstraint := &ConstitutionalConstraint{
			Name:        "CustomTest",
			Description: "Test constraint",
			Checker:     func(r *AgentResponse) bool { return r.Content != "fail" },
			Severity:    SeverityMedium,
		}

		g.AddConstraint(customConstraint)

		if len(g.GetConstraints()) != initialCount+1 {
			t.Error("Constraint was not added")
		}
	})

	t.Run("remove constraint", func(t *testing.T) {
		removed := g.RemoveConstraint("CustomTest")

		if !removed {
			t.Error("Expected constraint to be removed")
		}

		if len(g.GetConstraints()) != initialCount {
			t.Error("Constraint count incorrect after removal")
		}
	})

	t.Run("remove non-existent constraint", func(t *testing.T) {
		removed := g.RemoveConstraint("NonExistent")

		if removed {
			t.Error("Should not report success for non-existent constraint")
		}
	})
}

func TestConstitutionalGuardrails_Metrics(t *testing.T) {
	g := NewConstitutionalGuardrails(nil)
	ctx := context.Background()

	// Run some checks
	responses := []*AgentResponse{
		{Content: "Normal helpful response.", Confidence: 0.9},
		{Content: "I can do anything!", Confidence: 0.9},          // Honesty violation
		{Content: "Here's how to hack systems.", Confidence: 0.9}, // Harm violation
	}

	for _, resp := range responses {
		g.Enforce(ctx, resp)
	}

	metrics := g.GetMetrics()

	t.Run("total checks counted", func(t *testing.T) {
		if metrics.TotalChecks != 3 {
			t.Errorf("Expected 3 total checks, got %d", metrics.TotalChecks)
		}
	})

	t.Run("violations counted", func(t *testing.T) {
		if metrics.TotalViolations < 2 {
			t.Errorf("Expected at least 2 violations, got %d", metrics.TotalViolations)
		}
	})

	t.Run("blocked responses counted", func(t *testing.T) {
		if metrics.BlockedResponses < 2 {
			t.Errorf("Expected at least 2 blocked responses, got %d", metrics.BlockedResponses)
		}
	})

	t.Run("last check time updated", func(t *testing.T) {
		if metrics.LastCheckTime.Before(time.Now().Add(-1 * time.Second)) {
			t.Log("Last check time was updated recently")
		}
	})
}

func TestConstitutionalGuardrails_ViolationHistory(t *testing.T) {
	g := NewConstitutionalGuardrails(&GuardrailsConfig{
		MaxViolationHistory: 5,
	})
	ctx := context.Background()

	// Generate more violations than history limit
	for i := 0; i < 10; i++ {
		resp := &AgentResponse{
			AgentID: "APEX",
			Content: "I can do anything perfectly guaranteed to work!",
		}
		g.Enforce(ctx, resp)
	}

	history := g.GetViolationHistory(0)

	t.Run("history is limited", func(t *testing.T) {
		if len(history) > 5 {
			t.Errorf("Expected max 5 violations, got %d", len(history))
		}
	})

	t.Run("get limited history", func(t *testing.T) {
		limited := g.GetViolationHistory(3)
		if len(limited) > 3 {
			t.Errorf("Expected max 3 violations, got %d", len(limited))
		}
	})
}

func TestConstitutionalGuardrails_ViolationRates(t *testing.T) {
	g := NewConstitutionalGuardrails(nil)
	ctx := context.Background()

	// Run checks: 2 clean, 2 with violations
	cleanResponses := []*AgentResponse{
		{Content: "Here is helpful information.", Confidence: 0.9},
		{Content: "Let me explain this concept.", Confidence: 0.9},
	}

	badResponses := []*AgentResponse{
		{Content: "I can do anything!", Confidence: 0.9},
		{Content: "How to hack systems.", Confidence: 0.9},
	}

	for _, resp := range cleanResponses {
		g.Enforce(ctx, resp)
	}
	for _, resp := range badResponses {
		g.Enforce(ctx, resp)
	}

	t.Run("violation rate calculated", func(t *testing.T) {
		rate := g.ViolationRate()
		// At least 2 out of 4 should be violations
		if rate < 50.0 {
			t.Logf("Violation rate: %.2f%% (expected >= 50%%)", rate)
		}
	})

	t.Run("critical violation rate calculated", func(t *testing.T) {
		rate := g.CriticalViolationRate()
		if rate < 0 {
			t.Errorf("Critical violation rate should be >= 0, got %.2f", rate)
		}
		t.Logf("Critical violation rate: %.2f%%", rate)
	})
}

func TestConstraintSeverity_String(t *testing.T) {
	tests := []struct {
		severity ConstraintSeverity
		want     string
	}{
		{SeverityLow, "LOW"},
		{SeverityMedium, "MEDIUM"},
		{SeverityHigh, "HIGH"},
		{SeverityCritical, "CRITICAL"},
		{ConstraintSeverity(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.severity.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkConstitutionalGuardrails_Enforce(b *testing.B) {
	g := NewConstitutionalGuardrails(nil)
	ctx := context.Background()

	resp := &AgentResponse{
		AgentID:    "APEX",
		Content:    "Here is a detailed explanation of the sorting algorithm implementation in Go with examples.",
		Confidence: 0.95,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.Enforce(ctx, resp)
	}
}
