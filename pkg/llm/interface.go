package llm

import (
    "context"
    "encoding/json"
)

// Message represents a single message in the conversation
type Message struct {
    Role    string // "user", "assistant", or "tool"
    Content string
    ToolID  string // Used to associate tool results with tool calls
}

// Tool represents a tool definition passed to Claude
type Tool struct {
    Name        string
    Description string
    InputSchema map[string]any
}

// ToolUse represents a tool call from Claude
type ToolUse struct {
    ID    string
    Name  string
    Input json.RawMessage
}

// ToolResult represents the result of executing a tool
type ToolResult struct {
    ID      string
    Content string
    IsError bool
}

// Client interface defines methods for interacting with LLM providers
type Client interface {
    Name() string
    Chat(ctx context.Context, messages []Message, tools []Tool, systemPrompt string) ([]Message, []ToolUse, error)
    SendToolResult(ctx context.Context, messages []Message, results []ToolResult) ([]Message, []ToolUse, error)
}