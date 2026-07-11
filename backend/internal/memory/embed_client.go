package memory

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// EmbedClient calls the Ryzanstein gateway's embeddings endpoint (/api/embed)
// to turn text into a dense vector, used by the Store to make memory recall
// semantic. An empty baseURL yields a disabled client, and every method is
// fail-open: callers fall back to non-semantic (recency) behavior on any error.
type EmbedClient struct {
	httpClient *http.Client
	baseURL    string
	model      string
}

// NewEmbedClient builds a client pointed at baseURL (e.g. http://localhost:8000).
// An empty baseURL produces a disabled client. An empty model defaults to
// nomic-embed-text (policy-clean, 768-dim).
func NewEmbedClient(baseURL, model string) *EmbedClient {
	if model == "" {
		model = "nomic-embed-text"
	}
	return &EmbedClient{
		httpClient: &http.Client{Timeout: 5 * time.Second},
		baseURL:    baseURL,
		model:      model,
	}
}

// Enabled reports whether this client has an endpoint configured. Nil-safe so
// the Store can call it on an unset embedder.
func (c *EmbedClient) Enabled() bool {
	return c != nil && c.baseURL != ""
}

// BaseURL returns the configured gateway endpoint (or "" when disabled).
func (c *EmbedClient) BaseURL() string {
	if c == nil {
		return ""
	}
	return c.baseURL
}

// Embed returns the embedding vector for text via POST /api/embed. It returns
// an error (fail-open contract) when the client is disabled, on any
// transport/decode error, on a non-200 status, or when the gateway returns no
// embedding.
func (c *EmbedClient) Embed(text string) ([]float32, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("embed: client disabled")
	}
	body, err := json.Marshal(map[string]string{"model": c.model, "input": text})
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Post(c.baseURL+"/api/embed", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embed: gateway returned %d", resp.StatusCode)
	}
	var out struct {
		Embeddings [][]float32 `json:"embeddings"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if len(out.Embeddings) == 0 || len(out.Embeddings[0]) == 0 {
		return nil, fmt.Errorf("embed: empty embedding")
	}
	return out.Embeddings[0], nil
}
