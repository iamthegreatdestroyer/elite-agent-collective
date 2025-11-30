//go:build integration
// +build integration

package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/auth"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/pkg/models"
)

// TestSignatureVerification tests GitHub webhook signature verification.
func TestSignatureVerification(t *testing.T) {
	// Note: The test server doesn't have signature verification enabled
	// This test verifies the signature module works correctly via unit tests
	// Integration tests for signature verification require a configured webhook secret

	secret := "test-webhook-secret"
	body := []byte(`{"messages":[{"role":"user","content":"@APEX help"}],"model":"gpt-4"}`)

	t.Run("compute and validate signature", func(t *testing.T) {
		signature := auth.ComputeSignature(secret, body)
		
		// Verify the signature format
		if !strings.HasPrefix(signature, "sha256=") {
			t.Errorf("expected signature to start with sha256=, got: %s", signature)
		}

		// Validate the signature
		err := auth.ValidateSignature(secret, signature, body)
		if err != nil {
			t.Errorf("expected valid signature, got error: %v", err)
		}
	})

	t.Run("reject invalid signature", func(t *testing.T) {
		wrongSignature := auth.ComputeSignature("wrong-secret", body)
		err := auth.ValidateSignature(secret, wrongSignature, body)
		if err == nil {
			t.Error("expected error for invalid signature")
		}
	})
}

// TestMultiAgentCollaboration tests invoking multiple agents together.
func TestMultiAgentCollaboration(t *testing.T) {
	t.Run("two agents collaboration", func(t *testing.T) {
		reqBody := models.CopilotRequest{
			Messages: []models.Message{
				{Role: "user", Content: "@APEX @ARCHITECT design a system"},
			},
			Model:  "gpt-4",
			Stream: false,
		}
		body, err := json.Marshal(reqBody)
		if err != nil {
			t.Fatalf("failed to marshal request: %v", err)
		}

		resp, err := http.Post(getTestServerURL()+"/copilot", "application/json", bytes.NewReader(body))
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status 200, got %d", resp.StatusCode)
		}

		var copilotResp models.CopilotResponse
		if err := json.NewDecoder(resp.Body).Decode(&copilotResp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(copilotResp.Choices) == 0 {
			t.Fatal("expected at least one choice in response")
		}

		content := copilotResp.Choices[0].Message.Content
		// Multi-agent response should contain collaboration header
		if !strings.Contains(content, "Multi-Agent") || !strings.Contains(content, "APEX") {
			t.Errorf("expected multi-agent response mentioning APEX, got: %s", content)
		}
	})

	t.Run("three agents collaboration", func(t *testing.T) {
		reqBody := models.CopilotRequest{
			Messages: []models.Message{
				{Role: "user", Content: "@APEX @CIPHER @FORTRESS security review"},
			},
			Model:  "gpt-4",
			Stream: false,
		}
		body, err := json.Marshal(reqBody)
		if err != nil {
			t.Fatalf("failed to marshal request: %v", err)
		}

		resp, err := http.Post(getTestServerURL()+"/copilot", "application/json", bytes.NewReader(body))
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status 200, got %d", resp.StatusCode)
		}

		var copilotResp models.CopilotResponse
		if err := json.NewDecoder(resp.Body).Decode(&copilotResp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		content := copilotResp.Choices[0].Message.Content
		// Should mention multiple agents
		if !strings.Contains(content, "APEX") || !strings.Contains(content, "CIPHER") {
			t.Errorf("expected response from multiple agents, got: %s", content)
		}
	})

	t.Run("duplicate agents deduplicated", func(t *testing.T) {
		reqBody := models.CopilotRequest{
			Messages: []models.Message{
				{Role: "user", Content: "@APEX @APEX @APEX help me"},
			},
			Model:  "gpt-4",
			Stream: false,
		}
		body, err := json.Marshal(reqBody)
		if err != nil {
			t.Fatalf("failed to marshal request: %v", err)
		}

		resp, err := http.Post(getTestServerURL()+"/copilot", "application/json", bytes.NewReader(body))
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status 200, got %d", resp.StatusCode)
		}

		var copilotResp models.CopilotResponse
		if err := json.NewDecoder(resp.Body).Decode(&copilotResp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		content := copilotResp.Choices[0].Message.Content
		// With duplicates removed, should be single agent response (not multi-agent)
		if strings.Contains(content, "Multi-Agent") {
			t.Errorf("expected single agent response (duplicates deduplicated), got multi-agent: %s", content)
		}
	})
}

// TestAgentCodenameExtraction tests various agent mention formats.
func TestAgentCodenameExtraction(t *testing.T) {
	testCases := []struct {
		name     string
		message  string
		expected string
	}{
		{"uppercase at start", "@APEX help me", "APEX"},
		{"lowercase at start", "@apex help me", "APEX"},
		{"mixed case at start", "@Apex help me", "APEX"},
		{"middle of message", "Please @TENSOR analyze this", "TENSOR"},
		{"multiple agents", "@APEX @ARCHITECT design", "APEX"}, // First agent
		{"no at sign", "help me APEX", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reqBody := models.CopilotRequest{
				Messages: []models.Message{
					{Role: "user", Content: tc.message},
				},
				Model:  "gpt-4",
				Stream: false,
			}
			body, err := json.Marshal(reqBody)
			if err != nil {
				t.Fatalf("failed to marshal request: %v", err)
			}

			resp, err := http.Post(getTestServerURL()+"/copilot", "application/json", bytes.NewReader(body))
			if err != nil {
				t.Fatalf("failed to make request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("expected status 200, got %d", resp.StatusCode)
			}

			var copilotResp models.CopilotResponse
			if err := json.NewDecoder(resp.Body).Decode(&copilotResp); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			content := copilotResp.Choices[0].Message.Content
			expectedAgent := tc.expected
			if expectedAgent == "" {
				expectedAgent = "APEX" // Default agent
			}
			if !strings.Contains(content, expectedAgent) {
				t.Errorf("expected response from %s, got: %s", expectedAgent, content)
			}
		})
	}
}

// TestCopilotSpecificTests tests Copilot-specific test scenarios.
func TestCopilotSpecificTests(t *testing.T) {
	t.Run("response format matches copilot spec", func(t *testing.T) {
		reqBody := models.CopilotRequest{
			Messages: []models.Message{
				{Role: "user", Content: "@APEX help"},
			},
			Model:  "gpt-4",
			Stream: false,
		}
		body, _ := json.Marshal(reqBody)

		resp, err := http.Post(getTestServerURL()+"/copilot", "application/json", bytes.NewReader(body))
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		// Check content type
		contentType := resp.Header.Get("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			t.Errorf("expected application/json content type, got: %s", contentType)
		}

		// Decode and validate structure
		var copilotResp models.CopilotResponse
		if err := json.NewDecoder(resp.Body).Decode(&copilotResp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		// Validate required fields
		if len(copilotResp.Choices) == 0 {
			t.Fatal("response must have at least one choice")
		}

		choice := copilotResp.Choices[0]
		if choice.Message.Role != "assistant" {
			t.Errorf("expected role 'assistant', got: %s", choice.Message.Role)
		}

		if choice.FinishReason != "stop" {
			t.Errorf("expected finish_reason 'stop', got: %s", choice.FinishReason)
		}

		if choice.Message.Content == "" {
			t.Error("expected non-empty content")
		}
	})
}
