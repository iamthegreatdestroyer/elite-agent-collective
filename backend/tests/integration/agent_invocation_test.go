//go:build integration
// +build integration

package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/pkg/models"
)

// allAgentCodenames contains all 40 agent codenames for comprehensive testing.
var allAgentCodenames = []string{
	// Tier 1: Foundational Agents
	"APEX", "CIPHER", "ARCHITECT", "AXIOM", "VELOCITY",
	// Tier 2: Specialist Agents
	"QUANTUM", "TENSOR", "FORTRESS", "NEURAL", "CRYPTO",
	"FLUX", "PRISM", "SYNAPSE", "CORE", "HELIX",
	"VANGUARD", "ECLIPSE",
	// Tier 3: Innovator Agents
	"NEXUS", "GENESIS",
	// Tier 4: Meta Agents
	"OMNISCIENT",
	// Tier 5: Domain Specialists
	"ATLAS", "FORGE", "SENTRY", "VERTEX", "STREAM",
	// Tier 6: Emerging Tech Specialists
	"PHOTON", "LATTICE", "MORPH", "PHANTOM", "ORBIT",
	// Tier 7: Human-Centric Specialists
	"CANVAS", "LINGUA", "SCRIBE", "MENTOR", "BRIDGE",
	// Tier 8: Enterprise & Compliance Specialists
	"AEGIS", "LEDGER", "PULSE", "ARBITER", "ORACLE",
}

// TestAllAgentsInvocable tests that all 40 agents can be invoked via the /agents/{codename}/invoke endpoint.
func TestAllAgentsInvocable(t *testing.T) {
	for _, codename := range allAgentCodenames {
		t.Run(codename, func(t *testing.T) {
			reqBody := models.CopilotRequest{
				Messages: []models.Message{
					{Role: "user", Content: "Help me with a task"},
				},
				Model:  "gpt-4",
				Stream: false,
			}
			body, err := json.Marshal(reqBody)
			if err != nil {
				t.Fatalf("failed to marshal request: %v", err)
			}

			url := getTestServerURL() + "/agents/" + codename + "/invoke"
			resp, err := http.Post(url, "application/json", bytes.NewReader(body))
			if err != nil {
				t.Fatalf("failed to make request to %s: %v", codename, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("expected status 200 for %s, got %d", codename, resp.StatusCode)
			}

			var copilotResp models.CopilotResponse
			if err := json.NewDecoder(resp.Body).Decode(&copilotResp); err != nil {
				t.Fatalf("failed to decode response for %s: %v", codename, err)
			}

			if len(copilotResp.Choices) == 0 {
				t.Fatalf("expected at least one choice in response for %s", codename)
			}

			if copilotResp.Choices[0].Message.Content == "" {
				t.Errorf("expected non-empty content for %s", codename)
			}
		})
	}
}

// TestAgentResponseContainsIdentity tests that each agent's response contains its identity.
func TestAgentResponseContainsIdentity(t *testing.T) {
	testCases := []struct {
		codename    string
		shouldMatch []string
	}{
		{"APEX", []string{"APEX", "Computer Science"}},
		{"CIPHER", []string{"CIPHER", "Cryptography"}},
		{"ARCHITECT", []string{"ARCHITECT", "Architecture"}},
		{"QUANTUM", []string{"QUANTUM", "Quantum"}},
		{"TENSOR", []string{"TENSOR", "Machine Learning"}},
		{"FORTRESS", []string{"FORTRESS", "Security"}},
	}

	for _, tc := range testCases {
		t.Run(tc.codename, func(t *testing.T) {
			reqBody := models.CopilotRequest{
				Messages: []models.Message{
					{Role: "user", Content: "Introduce yourself"},
				},
				Model:  "gpt-4",
				Stream: false,
			}
			body, err := json.Marshal(reqBody)
			if err != nil {
				t.Fatalf("failed to marshal request: %v", err)
			}

			url := getTestServerURL() + "/agents/" + tc.codename + "/invoke"
			resp, err := http.Post(url, "application/json", bytes.NewReader(body))
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
			for _, match := range tc.shouldMatch {
				if !strings.Contains(content, match) {
					t.Errorf("expected response to contain '%s' for agent %s, got: %s", match, tc.codename, content)
				}
			}
		})
	}
}

// TestAgentCollaboration tests invoking multiple agents in sequence.
func TestAgentCollaboration(t *testing.T) {
	// Invoke APEX first
	apexReq := models.CopilotRequest{
		Messages: []models.Message{
			{Role: "user", Content: "Design an algorithm"},
		},
		Model:  "gpt-4",
		Stream: false,
	}
	apexBody, _ := json.Marshal(apexReq)

	apexResp, err := http.Post(getTestServerURL()+"/agents/APEX/invoke", "application/json", bytes.NewReader(apexBody))
	if err != nil {
		t.Fatalf("failed to invoke APEX: %v", err)
	}
	defer apexResp.Body.Close()

	if apexResp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 for APEX, got %d", apexResp.StatusCode)
	}

	// Then invoke ARCHITECT
	archReq := models.CopilotRequest{
		Messages: []models.Message{
			{Role: "user", Content: "Design a system architecture"},
		},
		Model:  "gpt-4",
		Stream: false,
	}
	archBody, _ := json.Marshal(archReq)

	archResp, err := http.Post(getTestServerURL()+"/agents/ARCHITECT/invoke", "application/json", bytes.NewReader(archBody))
	if err != nil {
		t.Fatalf("failed to invoke ARCHITECT: %v", err)
	}
	defer archResp.Body.Close()

	if archResp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 for ARCHITECT, got %d", archResp.StatusCode)
	}

	// Verify both responses
	var apexCopilotResp models.CopilotResponse
	if err := json.NewDecoder(apexResp.Body).Decode(&apexCopilotResp); err != nil {
		t.Fatalf("failed to decode APEX response: %v", err)
	}

	var archCopilotResp models.CopilotResponse
	if err := json.NewDecoder(archResp.Body).Decode(&archCopilotResp); err != nil {
		t.Fatalf("failed to decode ARCHITECT response: %v", err)
	}

	if !strings.Contains(apexCopilotResp.Choices[0].Message.Content, "APEX") {
		t.Error("expected APEX response to mention APEX")
	}

	if !strings.Contains(archCopilotResp.Choices[0].Message.Content, "ARCHITECT") {
		t.Error("expected ARCHITECT response to mention ARCHITECT")
	}
}

// TestAgentContextAwareness tests that agents process context from the request.
func TestAgentContextAwareness(t *testing.T) {
	// Send a request with specific context
	reqBody := models.CopilotRequest{
		Messages: []models.Message{
			{Role: "system", Content: "You are helping with Go development."},
			{Role: "user", Content: "I need help with sorting in Go"},
		},
		Model:  "gpt-4",
		Stream: false,
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	resp, err := http.Post(getTestServerURL()+"/agents/APEX/invoke", "application/json", bytes.NewReader(body))
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

	// Verify the response acknowledges the user's request
	content := copilotResp.Choices[0].Message.Content
	if !strings.Contains(content, "sorting") {
		t.Errorf("expected response to acknowledge 'sorting', got: %s", content)
	}
}

// TestAgentInvocationWithCopilotEndpoint tests all agents via the /copilot endpoint.
func TestAgentInvocationWithCopilotEndpoint(t *testing.T) {
	testAgents := []string{"APEX", "CIPHER", "ARCHITECT", "FLUX", "TENSOR"}

	for _, codename := range testAgents {
		t.Run(codename, func(t *testing.T) {
			reqBody := models.CopilotRequest{
				Messages: []models.Message{
					{Role: "user", Content: "@" + codename + " help me with a task"},
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
				t.Fatalf("failed to make request for %s: %v", codename, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("expected status 200 for %s, got %d", codename, resp.StatusCode)
			}

			var copilotResp models.CopilotResponse
			if err := json.NewDecoder(resp.Body).Decode(&copilotResp); err != nil {
				t.Fatalf("failed to decode response for %s: %v", codename, err)
			}

			content := copilotResp.Choices[0].Message.Content
			if !strings.Contains(content, codename) {
				t.Errorf("expected response from %s to mention agent name, got: %s", codename, content)
			}
		})
	}
}

// TestAllAgentsCount verifies that exactly 40 agents are registered.
func TestAllAgentsCount(t *testing.T) {
	resp, err := http.Get(getTestServerURL() + "/agents")
	if err != nil {
		t.Fatalf("failed to get agents list: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var agents []models.Agent
	if err := json.NewDecoder(resp.Body).Decode(&agents); err != nil {
		t.Fatalf("failed to decode agents: %v", err)
	}

	if len(agents) != 40 {
		t.Errorf("expected 40 agents, got %d", len(agents))
	}
}
