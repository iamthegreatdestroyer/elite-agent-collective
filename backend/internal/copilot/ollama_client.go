package copilot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/pkg/models"
)

type OllamaClient struct {
	httpClient *http.Client
	baseURL    string
	model      string
}

// chatRequest is the OpenAI-compatible /v1/chat/completions request body. The
// Ryzanstein gateway (baseURL, default http://localhost:8000) accepts this
// shape and routes it through the Token Recycler cache + per-model quality
// tiers — unlike the Ollama-native /api/chat path, which bypasses both.
type chatRequest struct {
	Model     string    `json:"model"`
	Messages  []chatMsg `json:"messages"`
	Stream    bool      `json:"stream"`
	MaxTokens int       `json:"max_tokens,omitempty"`
}

type chatMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// chatResponse is the OpenAI-compatible /v1/chat/completions response body.
type chatResponse struct {
	Choices []struct {
		Message      chatMsg `json:"message"`
		FinishReason string  `json:"finish_reason"`
	} `json:"choices"`
}

func NewOllamaClient() *OllamaClient {
	baseURL := os.Getenv("OLLAMA_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8000"
	}
	model := os.Getenv("OLLAMA_MODEL")
	if model == "" {
		// gemma3:4b-it-qat is the gateway-advertised quality tier: warm,
		// policy-clean (not Meta), and small enough for the CPU-only box.
		model = "gemma3:4b-it-qat"
	}
	return &OllamaClient{
		httpClient: &http.Client{Timeout: 120 * time.Second},
		baseURL:    baseURL,
		model:      model,
	}
}

func (c *OllamaClient) Enabled() bool {
	resp, err := c.httpClient.Get(c.baseURL + "/v1/models")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}

// BaseURL returns the configured gateway endpoint (from OLLAMA_URL, or the
// http://localhost:8000 default).
func (c *OllamaClient) BaseURL() string {
	return c.baseURL
}

// Model returns the configured chat model name (from OLLAMA_MODEL, or the
// gemma3:4b-it-qat default).
func (c *OllamaClient) Model() string {
	return c.model
}

// Warmup fires a tiny completion so the gateway pre-loads the configured model
// into memory at boot. On the CPU-only box a cold model load can take ~40-50s,
// which would otherwise time out the first real Copilot request. Best-effort:
// errors are logged and swallowed, never fatal.
func (c *OllamaClient) Warmup(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()

	payload := chatRequest{
		Model:     c.model,
		Messages:  []chatMsg{{Role: "user", Content: "ping"}},
		Stream:    false,
		MaxTokens: 1,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		log.Printf("OLLAMA_WARMUP: pre-load call failed for %s: %v", c.model, err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Printf("OLLAMA_WARMUP: pre-load returned %d for %s", resp.StatusCode, c.model)
		return
	}
	log.Printf("OLLAMA_WARMUP: model %s pre-loaded via %s", c.model, c.baseURL)
}

func (c *OllamaClient) Forward(ctx context.Context, systemPrompt string, req *models.CopilotRequest) (*models.CopilotResponse, error) {
	msgs := make([]chatMsg, 0, len(req.Messages)+1)
	msgs = append(msgs, chatMsg{Role: "system", Content: systemPrompt})
	for _, m := range req.Messages {
		msgs = append(msgs, chatMsg{Role: m.Role, Content: m.Content})
	}

	payload := chatRequest{
		Model:    c.model,
		Messages: msgs,
		Stream:   false,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal error: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("request error: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("gateway request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gateway returned %d", resp.StatusCode)
	}

	var chatResp chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}
	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("gateway returned no choices")
	}

	content := chatResp.Choices[0].Message.Content
	finish := chatResp.Choices[0].FinishReason
	if finish == "" {
		finish = "stop"
	}
	log.Printf("OLLAMA_FALLBACK: generated %d chars via %s", len(content), c.model)

	return &models.CopilotResponse{
		Choices: []models.Choice{
			{
				Message: models.Message{
					Role:    "assistant",
					Content: content,
				},
				FinishReason: finish,
			},
		},
	}, nil
}
