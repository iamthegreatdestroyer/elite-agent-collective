// Package copilot provides types matching the GitHub Copilot Extension API.
package copilot

import (
	"fmt"
	"time"
)

// Request represents the incoming request from GitHub Copilot.
// This matches the GitHub Copilot Extension API specification.
type Request struct {
	// Messages contains the conversation history.
	Messages []Message `json:"messages"`
	// Model is the model being used (e.g., "gpt-4").
	Model string `json:"model,omitempty"`
	// Stream indicates if the response should be streamed via SSE.
	Stream bool `json:"stream,omitempty"`
	// Temperature controls randomness in the response.
	Temperature float64 `json:"temperature,omitempty"`
	// TopP controls diversity via nucleus sampling.
	TopP float64 `json:"top_p,omitempty"`
	// MaxTokens limits the response length.
	MaxTokens int `json:"max_tokens,omitempty"`
	// User is the user identifier.
	User string `json:"user,omitempty"`
}

// Message represents a single message in the conversation.
type Message struct {
	// Role is the role of the message sender (system, user, assistant, tool).
	Role string `json:"role"`
	// Content is the text content of the message.
	Content string `json:"content"`
	// Name is an optional name for the participant.
	Name string `json:"name,omitempty"`
	// ToolCalls contains any tool calls requested by the assistant.
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
	// ToolCallID is the ID of the tool call this message is responding to.
	ToolCallID string `json:"tool_call_id,omitempty"`
}

// ToolCall represents a tool call in the conversation.
type ToolCall struct {
	// ID is the unique identifier for this tool call.
	ID string `json:"id"`
	// Type is the type of tool call (e.g., "function").
	Type string `json:"type"`
	// Function contains the function call details.
	Function FunctionCall `json:"function"`
}

// FunctionCall represents a function call within a tool call.
type FunctionCall struct {
	// Name is the name of the function to call.
	Name string `json:"name"`
	// Arguments is a JSON string of the function arguments.
	Arguments string `json:"arguments"`
}

// Response represents the response to GitHub Copilot.
type Response struct {
	// ID is a unique identifier for this response.
	ID string `json:"id,omitempty"`
	// Object is the object type (e.g., "chat.completion").
	Object string `json:"object,omitempty"`
	// Created is the Unix timestamp of when the response was created.
	Created int64 `json:"created,omitempty"`
	// Model is the model used to generate the response.
	Model string `json:"model,omitempty"`
	// Choices contains the response choices.
	Choices []Choice `json:"choices"`
	// Usage contains token usage statistics.
	Usage *Usage `json:"usage,omitempty"`
}

// Choice represents a single response choice.
type Choice struct {
	// Index is the index of this choice.
	Index int `json:"index"`
	// Message is the response message.
	Message Message `json:"message"`
	// FinishReason indicates why the response ended.
	FinishReason string `json:"finish_reason"`
}

// Usage contains token usage statistics.
type Usage struct {
	// PromptTokens is the number of tokens in the prompt.
	PromptTokens int `json:"prompt_tokens"`
	// CompletionTokens is the number of tokens in the completion.
	CompletionTokens int `json:"completion_tokens"`
	// TotalTokens is the total number of tokens used.
	TotalTokens int `json:"total_tokens"`
}

// StreamResponse represents a streaming response chunk.
type StreamResponse struct {
	// ID is a unique identifier for this response.
	ID string `json:"id,omitempty"`
	// Object is the object type (e.g., "chat.completion.chunk").
	Object string `json:"object,omitempty"`
	// Created is the Unix timestamp of when the response was created.
	Created int64 `json:"created,omitempty"`
	// Model is the model used to generate the response.
	Model string `json:"model,omitempty"`
	// Choices contains the streamed choice deltas.
	Choices []StreamChoice `json:"choices"`
}

// StreamChoice represents a single streaming choice.
type StreamChoice struct {
	// Index is the index of this choice.
	Index int `json:"index"`
	// Delta contains the incremental content.
	Delta StreamDelta `json:"delta"`
	// FinishReason indicates why the response ended (only set on final chunk).
	FinishReason string `json:"finish_reason,omitempty"`
}

// StreamDelta represents the incremental content in a stream.
type StreamDelta struct {
	// Role is set on the first chunk to indicate the role.
	Role string `json:"role,omitempty"`
	// Content is the incremental text content.
	Content string `json:"content,omitempty"`
	// ToolCalls contains incremental tool call data.
	ToolCalls []ToolCallDelta `json:"tool_calls,omitempty"`
}

// ToolCallDelta represents incremental tool call data in a stream.
type ToolCallDelta struct {
	// Index is the index of the tool call being updated.
	Index int `json:"index"`
	// ID is the tool call ID (set on first chunk for this tool call).
	ID string `json:"id,omitempty"`
	// Type is the tool call type (set on first chunk).
	Type string `json:"type,omitempty"`
	// Function contains the incremental function data.
	Function *FunctionCallDelta `json:"function,omitempty"`
}

// FunctionCallDelta represents incremental function call data.
type FunctionCallDelta struct {
	// Name is the function name (set on first chunk).
	Name string `json:"name,omitempty"`
	// Arguments is the incremental arguments data.
	Arguments string `json:"arguments,omitempty"`
}

// WebhookPayload represents the full webhook payload from GitHub.
type WebhookPayload struct {
	// Action is the action that triggered the webhook.
	Action string `json:"action,omitempty"`
	// Installation contains installation information.
	Installation *Installation `json:"installation,omitempty"`
	// Sender contains information about the user who triggered the event.
	Sender *User `json:"sender,omitempty"`
	// Repository contains repository information.
	Repository *Repository `json:"repository,omitempty"`
	// Messages contains the conversation for Copilot requests.
	Messages []Message `json:"messages,omitempty"`
}

// Installation represents a GitHub App installation.
type Installation struct {
	// ID is the installation ID.
	ID int64 `json:"id"`
	// Account is the account the app is installed on.
	Account *User `json:"account,omitempty"`
}

// User represents a GitHub user.
type User struct {
	// ID is the user ID.
	ID int64 `json:"id"`
	// Login is the username.
	Login string `json:"login"`
	// Type is the type of account (User, Organization, Bot).
	Type string `json:"type,omitempty"`
}

// Repository represents a GitHub repository.
type Repository struct {
	// ID is the repository ID.
	ID int64 `json:"id"`
	// Name is the repository name.
	Name string `json:"name"`
	// FullName is the full repository name (owner/repo).
	FullName string `json:"full_name"`
	// Private indicates if the repository is private.
	Private bool `json:"private"`
}

// NewMessage creates a new message with the given role and content.
func NewMessage(role, content string) Message {
	return Message{
		Role:    role,
		Content: content,
	}
}

// NewAssistantMessage creates a new assistant message.
func NewAssistantMessage(content string) Message {
	return NewMessage("assistant", content)
}

// NewUserMessage creates a new user message.
func NewUserMessage(content string) Message {
	return NewMessage("user", content)
}

// NewSystemMessage creates a new system message.
func NewSystemMessage(content string) Message {
	return NewMessage("system", content)
}

// generateResponseID creates a unique response ID.
func generateResponseID() string {
	return fmt.Sprintf("chatcmpl-%d", time.Now().UnixNano())
}

// NewResponse creates a new response with a single choice.
// Includes ID, Created timestamp, and Model fields as per GitHub Copilot API spec.
func NewResponse(content string) *Response {
	return &Response{
		ID:      generateResponseID(),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   "elite-agent-collective",
		Choices: []Choice{
			{
				Index: 0,
				Message: Message{
					Role:    "assistant",
					Content: content,
				},
				FinishReason: "stop",
			},
		},
	}
}
