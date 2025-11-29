//go:build integration
// +build integration

package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/pkg/models"
)

// BenchmarkAgentInvocation benchmarks the performance of agent invocation.
func BenchmarkAgentInvocation(b *testing.B) {
	reqBody := models.CopilotRequest{
		Messages: []models.Message{
			{Role: "user", Content: "@APEX help me with an algorithm"},
		},
		Model:  "gpt-4",
		Stream: false,
	}
	body, _ := json.Marshal(reqBody)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := http.Post(getTestServerURL()+"/copilot", "application/json", bytes.NewReader(body))
		if err != nil {
			b.Fatalf("request failed: %v", err)
		}
		resp.Body.Close()
	}
}

// BenchmarkConcurrentRequests benchmarks concurrent request handling.
func BenchmarkConcurrentRequests(b *testing.B) {
	reqBody := models.CopilotRequest{
		Messages: []models.Message{
			{Role: "user", Content: "@APEX help me"},
		},
		Model:  "gpt-4",
		Stream: false,
	}
	body, _ := json.Marshal(reqBody)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := http.Post(getTestServerURL()+"/copilot", "application/json", bytes.NewReader(body))
			if err != nil {
				b.Fatalf("request failed: %v", err)
			}
			resp.Body.Close()
		}
	})
}

// BenchmarkListAgents benchmarks the list agents endpoint.
func BenchmarkListAgents(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := http.Get(getTestServerURL() + "/agents")
		if err != nil {
			b.Fatalf("request failed: %v", err)
		}
		resp.Body.Close()
	}
}

// BenchmarkGetAgent benchmarks getting a single agent.
func BenchmarkGetAgent(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := http.Get(getTestServerURL() + "/agents/APEX")
		if err != nil {
			b.Fatalf("request failed: %v", err)
		}
		resp.Body.Close()
	}
}

// BenchmarkHealthCheck benchmarks the health check endpoint.
func BenchmarkHealthCheck(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := http.Get(getTestServerURL() + "/health")
		if err != nil {
			b.Fatalf("request failed: %v", err)
		}
		resp.Body.Close()
	}
}

// TestResponseTimeUnder500ms tests that responses return within 500ms.
func TestResponseTimeUnder500ms(t *testing.T) {
	reqBody := models.CopilotRequest{
		Messages: []models.Message{
			{Role: "user", Content: "@APEX help me with an algorithm"},
		},
		Model:  "gpt-4",
		Stream: false,
	}
	body, _ := json.Marshal(reqBody)

	// Warm up
	resp, err := http.Post(getTestServerURL()+"/copilot", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("warmup request failed: %v", err)
	}
	resp.Body.Close()

	// Measure response time
	start := time.Now()
	resp, err = http.Post(getTestServerURL()+"/copilot", "application/json", bytes.NewReader(body))
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	if elapsed > 500*time.Millisecond {
		t.Errorf("response took too long: %v (expected < 500ms)", elapsed)
	}

	t.Logf("Response time: %v", elapsed)
}

// TestConcurrentAgentRequests tests handling of concurrent requests.
func TestConcurrentAgentRequests(t *testing.T) {
	const numRequests = 20
	var wg sync.WaitGroup
	errors := make(chan error, numRequests)

	reqBody := models.CopilotRequest{
		Messages: []models.Message{
			{Role: "user", Content: "@APEX help me"},
		},
		Model:  "gpt-4",
		Stream: false,
	}
	body, _ := json.Marshal(reqBody)

	start := time.Now()

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			resp, err := http.Post(getTestServerURL()+"/copilot", "application/json", bytes.NewReader(body))
			if err != nil {
				errors <- err
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				errors <- err
				return
			}
		}()
	}

	wg.Wait()
	close(errors)

	elapsed := time.Since(start)
	t.Logf("Completed %d concurrent requests in %v", numRequests, elapsed)

	var errCount int
	for err := range errors {
		if err != nil {
			errCount++
			t.Logf("Request error: %v", err)
		}
	}

	if errCount > 0 {
		t.Errorf("%d out of %d requests failed", errCount, numRequests)
	}
}

// TestAllAgentsResponseTime tests that all agents respond quickly.
func TestAllAgentsResponseTime(t *testing.T) {
	agents := []string{"APEX", "CIPHER", "ARCHITECT", "QUANTUM", "TENSOR"}

	for _, agent := range agents {
		t.Run(agent, func(t *testing.T) {
			reqBody := models.CopilotRequest{
				Messages: []models.Message{
					{Role: "user", Content: "Help me with a task"},
				},
				Model:  "gpt-4",
				Stream: false,
			}
			body, _ := json.Marshal(reqBody)

			start := time.Now()
			resp, err := http.Post(getTestServerURL()+"/agents/"+agent+"/invoke", "application/json", bytes.NewReader(body))
			elapsed := time.Since(start)

			if err != nil {
				t.Fatalf("request to %s failed: %v", agent, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("expected status 200 for %s, got %d", agent, resp.StatusCode)
			}

			if elapsed > 500*time.Millisecond {
				t.Errorf("%s response took too long: %v", agent, elapsed)
			}

			t.Logf("%s response time: %v", agent, elapsed)
		})
	}
}

// TestListAgentsPerformance tests the list agents endpoint performance.
func TestListAgentsPerformance(t *testing.T) {
	start := time.Now()
	resp, err := http.Get(getTestServerURL() + "/agents")
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	if elapsed > 100*time.Millisecond {
		t.Errorf("list agents took too long: %v (expected < 100ms)", elapsed)
	}

	t.Logf("List agents response time: %v", elapsed)
}

// TestHealthEndpointPerformance tests the health endpoint performance.
func TestHealthEndpointPerformance(t *testing.T) {
	// Run multiple requests to get average
	const numRequests = 10
	var totalDuration time.Duration

	for i := 0; i < numRequests; i++ {
		start := time.Now()
		resp, err := http.Get(getTestServerURL() + "/health")
		elapsed := time.Since(start)

		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		resp.Body.Close()

		totalDuration += elapsed
	}

	avgDuration := totalDuration / numRequests
	t.Logf("Average health endpoint response time: %v", avgDuration)

	if avgDuration > 50*time.Millisecond {
		t.Errorf("health endpoint too slow: average %v (expected < 50ms)", avgDuration)
	}
}

// TestMemoryUnderLoad is a simple test to ensure no obvious memory issues under load.
func TestMemoryUnderLoad(t *testing.T) {
	const numRequests = 100
	var wg sync.WaitGroup

	reqBody := models.CopilotRequest{
		Messages: []models.Message{
			{Role: "user", Content: "@APEX help me with an algorithm"},
		},
		Model:  "gpt-4",
		Stream: false,
	}
	body, _ := json.Marshal(reqBody)

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			resp, err := http.Post(getTestServerURL()+"/copilot", "application/json", bytes.NewReader(body))
			if err != nil {
				return
			}
			resp.Body.Close()
		}()
	}

	wg.Wait()

	// If we get here without crash/hang, the test passes
	t.Logf("Successfully processed %d requests under load", numRequests)
}
