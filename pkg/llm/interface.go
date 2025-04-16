package llm

import (
	"context"
	"encoding/json"
)

// Message represents a message in a conversation
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Tool represents a tool that can be used by the LLM
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

// ToolUse represents a tool use request from the LLM
type ToolUse struct {
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input"`
}

// ToolResult represents a tool result to be sent back to the LLM
type ToolResult struct {
	ID      string `json:"id"`
	Content string `json:"content"`
	IsError bool   `json:"is_error"`
}

// Client is the interface that all LLM clients must implement
type Client interface {
	// Name returns the name of the LLM
	Name() string
	// Chat sends a message to the LLM and returns its response
	Chat(ctx context.Context, messages []Message, tools []Tool) ([]Message, []ToolUse, error)
	// SendToolResult sends a tool result back to the LLM and returns its response
	SendToolResult(ctx context.Context, messages []Message, toolResults []ToolResult) ([]Message, []ToolUse, error)
} 