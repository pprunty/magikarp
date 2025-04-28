package agent

import (
	"context"
	"encoding/json"
)

// Plugin defines the interface for a tool plugin
type Plugin interface {
	Name() string
	Tools() []ToolDefinition
}

// Parameter defines a parameter for a tool
type Parameter struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

// ToolFunc is the function signature for tool implementations
type ToolFunc func(ctx context.Context, input map[string]interface{}) (*ToolResult, error)

// ToolDefinition defines a tool that can be used by the agent
type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Function    ToolFunc              `json:"-"`
	Parameters  map[string]Parameter   `json:"parameters"`
	InputSchema map[string]any        `json:"input_schema,omitempty"`
	Fn          func(context.Context, json.RawMessage) (string, error) `json:"-"`
}

// // ToolUse represents a tool use request from the LLM
// type ToolUse struct {
// 	ID    string          `json:"id"`
// 	Name  string          `json:"name"`
// 	Input json.RawMessage `json:"input"`
// }
//
// // ToolResult represents a tool result to be sent back to the LLM
// type ToolResult struct {
// 	ID      string `json:"id"`
// 	Content string `json:"content"`
// 	IsError bool   `json:"is_error"`
// }

// ToPluginFormat converts the ToolDefinition to a format suitable for plugins
func (td *ToolDefinition) ToPluginFormat() *ToolDefinition {
	if td.Fn != nil {
		return &ToolDefinition{
			Name:        td.Name,
			Description: td.Description,
			InputSchema: td.InputSchema,
			Fn:          td.Fn,
		}
	}
	return td
}