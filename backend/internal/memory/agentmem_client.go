package memory

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// AgentMemClient talks to the agentmem MCP server
// (github.com/iamthegreatdestroyer/agentmem), a real 4-layer memory system
// with TF-IDF-based semantic recall. When configured, Store delegates
// Remember/Recall to it instead of the local per-user JSON files.
//
// Mapping notes: agentmem's semantic layer stores subject/predicate/value
// triples partitioned by an agent_id string, with no concept of "recorded
// by" independent of that partition key. So:
//   - agentmem's agent_id  <- this client's userID param (keeps per-user
//     isolation, same guarantee the local JSON-file store gave)
//   - agentmem's subject   <- the constant "user" (every fact recorded
//     through this bridge is about the user, not an arbitrary entity graph)
//   - agentmem's predicate <- Fact.Key
//   - agentmem's value     <- Fact.Value
//
// Fact.AgentID (which Elite agent recorded a fact) has no agentmem-side
// field to round-trip through, so facts recalled via this client leave it
// blank. FormatContext - the only place a Fact is actually rendered into a
// prompt - never read AgentID either, so nothing currently depends on it.
type AgentMemClient struct {
	httpClient *http.Client
	baseURL    string
}

const agentMemFactSubject = "user"

// NewAgentMemClient builds a client pointed at baseURL (e.g.
// "http://localhost:8100"). An empty baseURL produces a disabled client.
func NewAgentMemClient(baseURL string) *AgentMemClient {
	return &AgentMemClient{
		httpClient: &http.Client{Timeout: 5 * time.Second},
		baseURL:    baseURL,
	}
}

// Enabled reports whether this client has a server configured.
func (c *AgentMemClient) Enabled() bool {
	return c != nil && c.baseURL != ""
}

// BaseURL returns the configured agentmem MCP server URL.
func (c *AgentMemClient) BaseURL() string {
	return c.baseURL
}

type toolCallRequest struct {
	Tool string         `json:"tool"`
	Args map[string]any `json:"args"`
}

func (c *AgentMemClient) callTool(tool string, args map[string]any) (map[string]any, error) {
	body, err := json.Marshal(toolCallRequest{Tool: tool, Args: args})
	if err != nil {
		return nil, fmt.Errorf("agentmem: marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(c.baseURL+"/call", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("agentmem: call %s: %w", tool, err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("agentmem: read response for %s: %w", tool, err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("agentmem: unmarshal response for %s: %w", tool, err)
	}
	if errMsg, ok := result["error"]; ok {
		return nil, fmt.Errorf("agentmem: %s returned an error: %v", tool, errMsg)
	}
	return result, nil
}

// Remember asserts a fact via agentmem's semantic layer.
func (c *AgentMemClient) Remember(userID, agentID, key, value string) error {
	_, err := c.callTool("memory_assert_fact", map[string]any{
		"agent_id":  userID,
		"subject":   agentMemFactSubject,
		"predicate": key,
		"value":     value,
	})
	return err
}

// Recall queries agentmem's semantic memory for facts scoped to userID.
// Facts are returned newest-first to match the local JSON store's contract,
// though agentmem itself orders by similarity, not recency - callers doing
// an unfiltered "give me everything" query (query="") get whatever order
// agentmem's search returns.
func (c *AgentMemClient) Recall(userID string) []Fact {
	result, err := c.callTool("memory_recall_facts", map[string]any{
		"agent_id": userID,
		"query":    agentMemFactSubject,
		"top_k":    100,
	})
	if err != nil {
		return nil
	}

	rawFacts, _ := result["facts"].([]any)
	facts := make([]Fact, 0, len(rawFacts))
	for _, rf := range rawFacts {
		m, ok := rf.(map[string]any)
		if !ok {
			continue
		}
		content, _ := m["content"].(map[string]any)
		predicate, _ := content["predicate"].(string)
		value, _ := content["value"].(string)
		if predicate == "" {
			continue
		}
		facts = append(facts, Fact{Key: predicate, Value: value})
	}
	return facts
}

// Healthy checks agentmem's /health endpoint.
func (c *AgentMemClient) Healthy() bool {
	if !c.Enabled() {
		return false
	}
	resp, err := c.httpClient.Get(c.baseURL + "/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}
