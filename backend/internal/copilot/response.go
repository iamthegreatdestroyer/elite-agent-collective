// Package copilot provides utilities for handling GitHub Copilot requests and responses.
package copilot

import (
	"encoding/json"
	"net/http"

	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/pkg/models"
)

// WriteResponse writes a Copilot response to the HTTP response writer.
func WriteResponse(w http.ResponseWriter, resp *models.CopilotResponse) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(resp)
}

// WriteError writes an error response to the HTTP response writer.
func WriteError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	resp := &models.CopilotResponse{
		Choices: []models.Choice{
			{
				Message: models.Message{
					Role:    "assistant",
					Content: message,
				},
				FinishReason: "stop",
			},
		},
	}
	json.NewEncoder(w).Encode(resp)
}

// NewResponse creates a new Copilot response with the given content.
func NewResponse(content string) *models.CopilotResponse {
	return &models.CopilotResponse{
		Choices: []models.Choice{
			{
				Message: models.Message{
					Role:    "assistant",
					Content: content,
				},
				FinishReason: "stop",
			},
		},
	}
}
