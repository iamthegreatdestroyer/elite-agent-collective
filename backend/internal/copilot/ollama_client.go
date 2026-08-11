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
		// 50s is the EFFECTIVE per-request model budget: Forward inherits the caller's
		// context and sets no deadline of its own, so this client timeout is the innermost
		// one that actually fires on the request path. It sits well under the server's 90s
		// WriteTimeout so the error is still writable. Was 120s — TWICE the server's budget,
		// so the client kept waiting long after the connection had been abandoned.
		httpClient: &http.Client{Timeout: 50 * time.Second},
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
	// Bounds the WARM-UP probe only — Forward does not set a deadline of its own, so this
	// does not govern real requests (their budget is the 50s httpClient timeout above).
	// 45s ~= 880 tokens at the measured 19.5 tok/s on sigma-infer (2026-08-10), far beyond
	// any real answer, so a warmup that exceeds it means the model is not serving.
	ctx, cancel := context.WithTimeout(ctx, 45*time.Second)
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
		Model: c.model,
		// MaxTokens is set EXPLICITLY because the struct tag is `omitempty`: leaving it zero
		// drops the field and the gateway applies its own default, which was measured
		// truncating answers at ~1,300 characters with finish_reason="length" on 3 of 4 real
		// questions. The truncation was not the model choosing to stop; it was a limit nobody
		// had chosen.
		//
		// WHY 512 AND NOT MORE — the budget arithmetic, measured 2026-08-11 on this gateway:
		//   generation runs at ~18 tok/s   (768 tok -> 42s, 512 -> 28s, 384 -> 21s)
		//   the client timeout above is 50s, and prefill (persona + retrieval context) eats
		//   several more seconds before the first token.
		// 512 finishes in ~28s with comfortable headroom. 768 measured 42s, leaving ~8s on a
		// host that has produced 60s outliers — that trades truncated answers for timed-out
		// ones, which is a worse failure because it returns nothing at all.
		//
		// Raising this REQUIRES raising the client timeout and the server WriteTimeout with it,
		// in that order (innermost first), or the deadline stack inverts again.
		MaxTokens: 512,
		Messages:  msgs,
		Stream:    false,
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
