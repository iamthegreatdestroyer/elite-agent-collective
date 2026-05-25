package copilot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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

// Forward prepends systemPrompt as a system message, then posts the full conversation
// to the upstream Copilot API. The GitHub bearer token is retrieved from ctx.
// Returns (nil, nil) when upstream is disabled or the call fails — callers fall back
// to the template response in that case.
func (c *UpstreamClient) Forward(ctx context.Context, systemPrompt string, req *models.CopilotRequest) (*models.CopilotResponse, error) {
	if !c.Enabled() {
		return nil, nil
	}

	// Build message list: system prompt first, then conversation history.
	messages := make([]models.Message, 0, len(req.Messages)+1)
	messages = append(messages, models.Message{Role: "system", Content: systemPrompt})
	messages = append(messages, req.Messages...)

	payload := upstreamRequest{
		Messages: messages,
		Model:    req.Model,
		Stream:   false, // upstream call is always non-streaming; we wrap the result
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
	if token := auth.GetGitHubToken(ctx); token != "" {
		httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
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
