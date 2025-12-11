// Package copilot provides utilities for handling GitHub Copilot requests and responses.
package copilot

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/pkg/models"
)

// ParseRequest parses a Copilot request from an HTTP request body.
func ParseRequest(r *http.Request) (*models.CopilotRequest, error) {
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	var req models.CopilotRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, err
	}

	return &req, nil
}

// GetLastUserMessage extracts the last user message from a Copilot request.
func GetLastUserMessage(req *models.CopilotRequest) string {
	for i := len(req.Messages) - 1; i >= 0; i-- {
		if req.Messages[i].Role == "user" {
			return req.Messages[i].Content
		}
	}
	return ""
}
