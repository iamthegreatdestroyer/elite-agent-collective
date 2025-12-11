// Package copilot provides utilities for handling GitHub Copilot requests and responses.
package copilot

import (
	"encoding/json"
	"fmt"
	"log"
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
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Error encoding error response: %v", err)
	}
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

// SSEWriter provides Server-Sent Events streaming support for Copilot responses.
type SSEWriter struct {
	w       http.ResponseWriter
	flusher http.Flusher
}

// NewSSEWriter creates a new SSE writer for streaming responses.
// Returns nil if the ResponseWriter doesn't support flushing.
func NewSSEWriter(w http.ResponseWriter) *SSEWriter {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil
	}
	return &SSEWriter{w: w, flusher: flusher}
}

// Init initializes the SSE response headers.
func (s *SSEWriter) Init() {
	s.w.Header().Set("Content-Type", "text/event-stream")
	s.w.Header().Set("Cache-Control", "no-cache")
	s.w.Header().Set("Connection", "keep-alive")
	s.w.Header().Set("X-Accel-Buffering", "no")
}

// StreamChunk represents a single chunk in a streaming response.
type StreamChunk struct {
	Choices []StreamChoice `json:"choices"`
}

// StreamChoice represents a choice in a streaming response.
type StreamChoice struct {
	Index        int         `json:"index"`
	Delta        StreamDelta `json:"delta"`
	FinishReason string      `json:"finish_reason,omitempty"`
}

// StreamDelta represents the delta content in a streaming response.
type StreamDelta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

// WriteChunk writes a content chunk to the SSE stream.
func (s *SSEWriter) WriteChunk(content string) error {
	chunk := StreamChunk{
		Choices: []StreamChoice{
			{
				Index: 0,
				Delta: StreamDelta{
					Content: content,
				},
			},
		},
	}
	return s.writeData(chunk)
}

// WriteRole writes the role for the first chunk.
func (s *SSEWriter) WriteRole(role string) error {
	chunk := StreamChunk{
		Choices: []StreamChoice{
			{
				Index: 0,
				Delta: StreamDelta{
					Role: role,
				},
			},
		},
	}
	return s.writeData(chunk)
}

// WriteEnd writes the final chunk to close the stream.
func (s *SSEWriter) WriteEnd() error {
	chunk := StreamChunk{
		Choices: []StreamChoice{
			{
				Index:        0,
				Delta:        StreamDelta{},
				FinishReason: "stop",
			},
		},
	}
	if err := s.writeData(chunk); err != nil {
		return err
	}
	// Write the final [DONE] marker
	_, err := fmt.Fprintf(s.w, "data: [DONE]\n\n")
	if err != nil {
		return err
	}
	s.flusher.Flush()
	return nil
}

// writeData marshals and writes data to the SSE stream.
func (s *SSEWriter) writeData(data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(s.w, "data: %s\n\n", jsonData)
	if err != nil {
		return err
	}
	s.flusher.Flush()
	return nil
}

// WriteStreamingResponse writes a complete response as a stream.
// This is useful for backwards compatibility when streaming is requested
// but only a complete response is available.
// If the ResponseWriter doesn't support streaming (no Flusher interface),
// it falls back to a regular JSON response.
func WriteStreamingResponse(w http.ResponseWriter, content string) error {
	sse := NewSSEWriter(w)
	if sse == nil {
		// Fall back to regular response if streaming not supported
		// Log this fallback so it's visible in debugging
		log.Printf("SSE streaming not supported, falling back to JSON response")
		return WriteResponse(w, NewResponse(content))
	}

	sse.Init()

	// Write role
	if err := sse.WriteRole("assistant"); err != nil {
		return err
	}

	// Write content as a single chunk
	if err := sse.WriteChunk(content); err != nil {
		return err
	}

	// Write end
	return sse.WriteEnd()
}
