package agent

import (
	"context"
	"encoding/json"
)

// ChatMessage is what we pass to / from an LLM.
type ChatMessage struct {
	Role    string `json:"role"`   // user | assistant | system | tool
	Content string `json:"content"`
}

// ToolCall is emitted by the LLM when it wants to use a tool.
type ToolCall struct {
	ID       string          `json:"id"`
	Name     string          `json:"name"`
	RawInput json.RawMessage `json:"arguments"`
}

// ToolResult is returned to the LLM after we run a tool.
type ToolResult struct {
	ID      string `json:"id"`
	Content string `json:"content"`
	IsError bool   `json:"is_error"`
}

// ToJSON converts the ToolResult to JSON
func (tr *ToolResult) ToJSON() string {
	b, err := json.Marshal(tr)
	if err != nil {
		return `{"error": "failed to marshal tool result"}`
	}
	return string(b)
}

// NewToolResult creates a new tool result
func NewToolResult(id string, content string, isError bool) *ToolResult {
	return &ToolResult{
		ID:      id,
		Content: content,
		IsError: isError,
	}
}

// LLM client interface (implemented elsewhere)
type LLM interface {
	Name() string
	Chat(ctx context.Context,
		convo []ChatMessage,
		tools []ToolDefinition,
	) (assistant []ChatMessage, toolCalls []ToolCall, err error)

	SendToolResults(ctx context.Context,
		convo []ChatMessage,
		results []ToolResult,
	) (assistant []ChatMessage, err error)
} 