//go:build integration
// +build integration

package integration

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/pkg/models"
)

// TestSSEStreamingResponse tests Server-Sent Events streaming responses.
func TestSSEStreamingResponse(t *testing.T) {
	t.Run("streaming response format", func(t *testing.T) {
		reqBody := models.CopilotRequest{
			Messages: []models.Message{
				{Role: "user", Content: "@APEX explain sorting"},
			},
			Model:  "gpt-4",
			Stream: true,
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

		// Check content type for SSE
		contentType := resp.Header.Get("Content-Type")
		if !strings.Contains(contentType, "text/event-stream") {
			t.Errorf("expected text/event-stream content type, got: %s", contentType)
		}

		// Read and validate SSE format
		reader := bufio.NewReader(resp.Body)
		hasData := false
		hasDone := false

		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					break
				}
				t.Fatalf("error reading response: %v", err)
			}

			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			if strings.HasPrefix(line, "data: ") {
				data := strings.TrimPrefix(line, "data: ")
				if data == "[DONE]" {
					hasDone = true
					break
				}
				hasData = true

				// Validate JSON structure
				var chunk map[string]interface{}
				if err := json.Unmarshal([]byte(data), &chunk); err != nil {
					t.Errorf("invalid JSON in SSE data: %v", err)
				}

				// Validate it has choices
				if _, ok := chunk["choices"]; !ok {
					t.Error("SSE chunk missing 'choices' field")
				}
			}
		}

		if !hasData {
			t.Error("expected at least one data line in SSE response")
		}
		if !hasDone {
			t.Error("expected [DONE] marker at end of SSE response")
		}
	})

	t.Run("non-streaming response when stream=false", func(t *testing.T) {
		reqBody := models.CopilotRequest{
			Messages: []models.Message{
				{Role: "user", Content: "@APEX help"},
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

		// Should be regular JSON, not SSE
		contentType := resp.Header.Get("Content-Type")
		if strings.Contains(contentType, "text/event-stream") {
			t.Error("expected non-streaming response for stream=false")
		}

		// Should be valid JSON
		var copilotResp models.CopilotResponse
		if err := json.NewDecoder(resp.Body).Decode(&copilotResp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
	})

	t.Run("streaming via agent invoke endpoint", func(t *testing.T) {
		reqBody := models.CopilotRequest{
			Messages: []models.Message{
				{Role: "user", Content: "help me"},
			},
			Model:  "gpt-4",
			Stream: true,
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

		// Should be SSE format
		contentType := resp.Header.Get("Content-Type")
		if !strings.Contains(contentType, "text/event-stream") {
			t.Errorf("expected text/event-stream for streaming, got: %s", contentType)
		}
	})
}

// TestStreamingResponseContent validates the content of streaming responses.
func TestStreamingResponseContent(t *testing.T) {
	reqBody := models.CopilotRequest{
		Messages: []models.Message{
			{Role: "user", Content: "@CIPHER explain encryption"},
		},
		Model:  "gpt-4",
		Stream: true,
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

	// Read entire response
	reader := bufio.NewReader(resp.Body)
	var contentParts []string

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			t.Fatalf("error reading response: %v", err)
		}

		line = strings.TrimSpace(line)
		if line == "" || !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var chunk struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
		}

		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}

		if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
			contentParts = append(contentParts, chunk.Choices[0].Delta.Content)
		}
	}

	// Combine all content chunks
	fullContent := strings.Join(contentParts, "")

	// Verify agent responded
	if !strings.Contains(fullContent, "CIPHER") {
		t.Errorf("expected CIPHER in response content, got: %s", fullContent)
	}
}
