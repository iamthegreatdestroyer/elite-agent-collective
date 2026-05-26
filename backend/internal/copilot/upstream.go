package copilot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/auth"
	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/pkg/models"
)

// UpstreamClient forwards requests to the GitHub Copilot upstream API.
type UpstreamClient struct {
	httpClient *http.Client
	apiURL     string
}

// NewUpstreamClient creates a new upstream client.
// apiURL is read from COPILOT_UPSTREAM_URL; if empty, upstream forwarding is disabled.
func NewUpstreamClient() *UpstreamClient {
	return &UpstreamClient{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		apiURL:     os.Getenv("COPILOT_UPSTREAM_URL"),
	}
}

// Enabled reports whether upstream forwarding is configured.
func (c *UpstreamClient) Enabled() bool {
	return c.apiURL != ""
}

// upstreamRequest mirrors CopilotRequest but is used for the outbound payload.
type upstreamRequest struct {
	Messages []models.Message `json:"messages"`
	Model    string           `json:"model"`
	Stream   bool             `json:"stream"`
}

const summarizePromptHeader = `Compress the following conversation history into a concise summary that preserves:
1. All technical decisions and their rationale
2. Code structures, file paths, and variable names mentioned
3. Agent personas active (@APEX, @CIPHER, etc.)
4. Any open questions or pending tasks
5. Key constraints established

Return ONLY the summary text, no preamble. Max 200 words.

`

// envInt reads an integer env var, returning def for missing or invalid values.
func envInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return def
	}
	return n
}

// SummarizeHistory compresses old conversation turns into a summary when the
// message history exceeds a token threshold. It replaces all but the last
// N messages with a single system message containing a compressed summary.
//
// Threshold: compress when len(messages) > TOKEN_RECYCLER_THRESHOLD (default 12)
// Preserve: always keep the last TOKEN_RECYCLER_PRESERVE turns verbatim (default 4)
// Summary: call the upstream model once with a summarization prompt to
// compress older turns, inject as system message at position 0
//
// If COPILOT_UPSTREAM_URL is not set or the summarization call fails,
// return the original messages unchanged (graceful degradation).
func (c *UpstreamClient) SummarizeHistory(ctx context.Context, messages []models.Message, githubToken string) []models.Message {
	threshold := envInt("TOKEN_RECYCLER_THRESHOLD", 12)
	preserve := envInt("TOKEN_RECYCLER_PRESERVE", 4)

	if len(messages) <= threshold || !c.Enabled() {
		return messages
	}

	cutoff := len(messages) - preserve
	if cutoff < 1 {
		return messages
	}

	older := messages[:cutoff]
	recent := messages[cutoff:]

	var sb strings.Builder
	for _, m := range older {
		sb.WriteString(fmt.Sprintf("[%s]: %s\n", m.Role, m.Content))
	}

	payload := upstreamRequest{
		Messages: []models.Message{
			{Role: "user", Content: summarizePromptHeader + sb.String()},
		},
		Model:  "gpt-4",
		Stream: false,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return messages
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.apiURL, bytes.NewReader(body))
	if err != nil {
		return messages
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	if githubToken != "" {
		httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", githubToken))
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return messages
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return messages
	}

	var copilotResp models.CopilotResponse
	if err := json.NewDecoder(resp.Body).Decode(&copilotResp); err != nil {
		return messages
	}

	if len(copilotResp.Choices) == 0 || copilotResp.Choices[0].Message.Content == "" {
		return messages
	}

	summary := copilotResp.Choices[0].Message.Content
	log.Printf("TOKEN_RECYCLER: compressed %d turns to summary (%d turns preserved)", len(older), len(recent))

	result := make([]models.Message, 0, 1+len(recent))
	result = append(result, models.Message{Role: "system", Content: summary})
	result = append(result, recent...)
	return result
}

// Forward prepends systemPrompt as a system message, then posts the full conversation
// to the upstream Copilot API. The GitHub bearer token is retrieved from ctx.
// Returns (nil, nil) when upstream is disabled or the call fails — callers fall back
// to the template response in that case.
func (c *UpstreamClient) Forward(ctx context.Context, systemPrompt string, req *models.CopilotRequest) (*models.CopilotResponse, error) {
	if !c.Enabled() {
		return nil, nil
	}

	githubToken := auth.GetGitHubToken(ctx)

	// Compress history before forwarding.
	compressed := c.SummarizeHistory(ctx, req.Messages, githubToken)

	// Build message list: system prompt first, then conversation history.
	messages := make([]models.Message, 0, len(compressed)+1)
	messages = append(messages, models.Message{Role: "system", Content: systemPrompt})
	messages = append(messages, compressed...)

	payload := upstreamRequest{
		Messages: messages,
		Model:    req.Model,
		Stream:   false,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, nil
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.apiURL, bytes.NewReader(body))
	if err != nil {
		return nil, nil
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	// Forward the GitHub token from context — never log it.
	if githubToken != "" {
		httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", githubToken))
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil
	}

	var copilotResp models.CopilotResponse
	if err := json.NewDecoder(resp.Body).Decode(&copilotResp); err != nil {
		return nil, nil
	}

	return &copilotResp, nil
}
