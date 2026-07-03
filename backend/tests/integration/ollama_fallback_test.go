//go:build integration
// +build integration

package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/copilot"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/pkg/models"
)

// templateMarkers are phrases that only appear in the canned templateResponse
// text (see backend/internal/agents/handlers/{base,apex}.go). If a response
// contains one of these, it came from the template fallback, not a real
// upstream or Ollama completion.
var templateMarkers = []string{
	"I'm ready to assist you with my expertise",
	"Let me analyze your request...",
}

// ollamaTestModel is deliberately smaller/faster than the production default
// (phi4-mini) purely so this test completes in a reasonable time on
// CPU-only hardware. It does not change the production default: main.go and
// ollama_client.go's NewOllamaClient() still default to phi4-mini unless an
// operator sets OLLAMA_MODEL themselves.
const ollamaTestModel = "smollm2:1.7b"

func init() {
	// Set before TestMain (in integration_test.go) constructs the shared
	// testOllamaClient via copilot.NewOllamaClient(), so the server-side
	// registry actually uses the faster model too. Go guarantees all
	// init() funcs in a package run before TestMain / any test.
	if os.Getenv("OLLAMA_MODEL") == "" {
		os.Setenv("OLLAMA_MODEL", ollamaTestModel)
	}
}

// requireOllama skips the test unless a real local Ollama server is
// reachable. This test exercises the actual second-tier fallback wiring
// end-to-end (Handle -> OllamaClient.Forward -> POST /api/chat), so it
// needs a real server, not a mock, to prove the wiring genuinely works.
func requireOllama(t *testing.T) {
	t.Helper()
	oc := copilot.NewOllamaClient()
	if !oc.Enabled() {
		t.Skipf("skipping: no reachable Ollama server at %s (start Ollama or set OLLAMA_URL)", oc.BaseURL())
	}
}

// TestOllamaFallback_APEX verifies that when upstream Copilot is unavailable
// (testRegistry is built with a nil upstream client), APEX's Handle() falls
// through to the real local Ollama model and returns genuine model output —
// not the canned template response. This is the one true end-to-end proof
// that the Ollama fallback wiring works against a live Ollama server.
//
// A second, fast, deterministic proof that the *other* Handle()
// implementation (BaseAgent, used by every non-APEX agent) wires up the same
// way lives in backend/internal/agents/handlers/base_test.go, using a mock
// HTTP server instead of a live model — that keeps this slow, hardware-timing
// -sensitive real-model test to a single case instead of chaining multiple
// slow real-inference round trips in one run.
//
// Note: on CPU-only hardware a single short chat completion can legitimately
// take 30-100+ seconds (confirmed manually on a 2-core box: phi4-mini ~100s,
// smollm2:1.7b ~55-100s depending on system prompt length and load). Run
// with a generous -timeout (e.g. `go test -tags integration -timeout 300s`).
func TestOllamaFallback_APEX(t *testing.T) {
	requireOllama(t)

	reqBody := models.CopilotRequest{
		Messages: []models.Message{
			{Role: "user", Content: "In one short sentence, what is a hash table?"},
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
		t.Fatalf("failed to invoke APEX: %v", err)
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
	if content == "" {
		t.Fatal("expected non-empty response content")
	}

	for _, marker := range templateMarkers {
		if strings.Contains(content, marker) {
			t.Fatalf("response came from the template fallback (matched %q), expected a real Ollama completion; got: %s", marker, content)
		}
	}

	t.Logf("APEX Ollama fallback response: %s", content)
}
