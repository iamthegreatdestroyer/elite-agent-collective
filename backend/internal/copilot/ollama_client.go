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

type ollamaChatRequest struct {
	Model    string          `json:"model"`
	Messages []ollamaChatMsg `json:"messages"`
	Stream   bool            `json:"stream"`
}

type ollamaChatMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaChatResponse struct {
	Message ollamaChatMsg `json:"message"`
}

func NewOllamaClient() *OllamaClient {
	baseURL := os.Getenv("OLLAMA_URL")
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	model := os.Getenv("OLLAMA_MODEL")
	if model == "" {
		model = "phi4-mini"
	}
	return &OllamaClient{
		httpClient: &http.Client{Timeout: 120 * time.Second},
		baseURL:    baseURL,
		model:      model,
	}
}

func (c *OllamaClient) Enabled() bool {
	resp, err := c.httpClient.Get(c.baseURL + "/api/tags")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}

// BaseURL returns the configured Ollama endpoint (from OLLAMA_URL, or the
// http://localhost:11434 default).
func (c *OllamaClient) BaseURL() string {
	return c.baseURL
}

// Model returns the configured chat model name (from OLLAMA_MODEL, or the
// phi4-mini default).
func (c *OllamaClient) Model() string {
	return c.model
}

func (c *OllamaClient) Forward(ctx context.Context, systemPrompt string, req *models.CopilotRequest) (*models.CopilotResponse, error) {
	msgs := make([]ollamaChatMsg, 0, len(req.Messages)+1)
	msgs = append(msgs, ollamaChatMsg{Role: "system", Content: systemPrompt})
	for _, m := range req.Messages {
		msgs = append(msgs, ollamaChatMsg{Role: m.Role, Content: m.Content})
	}

	payload := ollamaChatRequest{
		Model:    c.model,
		Messages: msgs,
		Stream:   false,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal error: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("request error: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("ollama request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama returned %d", resp.StatusCode)
	}

	var ollamaResp ollamaChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

	log.Printf("OLLAMA_FALLBACK: generated %d chars via %s", len(ollamaResp.Message.Content), c.model)

	return &models.CopilotResponse{
		Choices: []models.Choice{
			{
				Message: models.Message{
					Role:    "assistant",
					Content: ollamaResp.Message.Content,
				},
				FinishReason: "stop",
			},
		},
	}, nil
}
