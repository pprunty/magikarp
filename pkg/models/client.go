package models

import "context"

// Message represents a message in a conversation
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Tool represents a tool definition that can be used by an LLM
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

// ToolUse represents a tool usage request from an LLM
type ToolUse struct {
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Input []byte          `json:"input"`
}

// ToolResult represents the result of a tool execution
type ToolResult struct {
	ID      string `json:"id"`
	Content string `json:"content"`
	IsError bool   `json:"is_error"`
}

// Client defines the interface for all LLM clients
type Client interface {
	// Name returns the name of the LLM model
	Name() string

	// Chat sends a conversation to the LLM and returns the response and any tool use requests
	Chat(ctx context.Context, messages []Message, tools []Tool) ([]Message, []ToolUse, error)

	// SendToolResult sends the result of a tool execution to the LLM and returns the updated response
	SendToolResult(ctx context.Context, messages []Message, results []ToolResult) ([]Message, []ToolUse, error)
} 