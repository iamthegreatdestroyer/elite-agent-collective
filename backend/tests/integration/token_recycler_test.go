//go:build integration
// +build integration

package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/copilot"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/pkg/models"
)

// TestSummarizeHistory_BelowThreshold verifies that SummarizeHistory is a no-op
// when the message count is at or below the threshold.
func TestSummarizeHistory_BelowThreshold(t *testing.T) {
	t.Setenv("COPILOT_UPSTREAM_URL", "")
	t.Setenv("TOKEN_RECYCLER_THRESHOLD", "12")

	client := copilot.NewUpstreamClient()

	messages := make([]models.Message, 8)
	for i := range messages {
		messages[i] = models.Message{Role: "user", Content: fmt.Sprintf("message %d", i)}
	}

	result := client.SummarizeHistory(context.Background(), messages, "")
	if len(result) != len(messages) {
		t.Fatalf("expected %d messages unchanged, got %d", len(messages), len(result))
	}
	for i, m := range result {
		if m.Content != messages[i].Content {
			t.Fatalf("message %d content changed: got %q, want %q", i, m.Content, messages[i].Content)
		}
	}
}

// TestSummarizeHistory_AboveThreshold verifies that SummarizeHistory replaces older turns
// with a summary message when the count exceeds the threshold.
func TestSummarizeHistory_AboveThreshold(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := models.CopilotResponse{
			Choices: []models.Choice{
				{
					Message:      models.Message{Role: "assistant", Content: "Compressed summary of older turns."},
					FinishReason: "stop",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp) //nolint:errcheck
	}))
	defer mockServer.Close()

	t.Setenv("COPILOT_UPSTREAM_URL", mockServer.URL)
	t.Setenv("TOKEN_RECYCLER_THRESHOLD", "12")
	t.Setenv("TOKEN_RECYCLER_PRESERVE", "4")

	client := copilot.NewUpstreamClient()

	// 14 messages exceeds the default threshold of 12.
	messages := make([]models.Message, 14)
	for i := range messages {
		messages[i] = models.Message{Role: "user", Content: fmt.Sprintf("turn %d", i)}
	}

	result := client.SummarizeHistory(context.Background(), messages, "test-token")

	// Expect: 1 system summary + 4 preserved recent turns = 5.
	want := 1 + 4
	if len(result) != want {
		t.Fatalf("expected %d messages after compression, got %d", want, len(result))
	}
	if result[0].Role != "system" {
		t.Fatalf("expected result[0] role=system, got %q", result[0].Role)
	}
	if result[0].Content != "Compressed summary of older turns." {
		t.Fatalf("unexpected summary content: %q", result[0].Content)
	}
	// Verify the preserved turns are the last 4 originals (indices 10–13).
	for i := 0; i < 4; i++ {
		wantMsg := messages[10+i]
		gotMsg := result[1+i]
		if gotMsg.Content != wantMsg.Content {
			t.Fatalf("preserved turn %d: got %q, want %q", i, gotMsg.Content, wantMsg.Content)
		}
	}
}
