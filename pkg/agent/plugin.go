package agent

import (
	"encoding/json"
	"fmt"
)

// PluginBase provides common functionality for all plugins
type PluginBase struct {
	name        string
	description string
	tools       []ToolDefinition
}

// NewPluginBase creates a new PluginBase instance
func NewPluginBase(name, description string) *PluginBase {
	return &PluginBase{
		name:        name,
		description: description,
		tools:       make([]ToolDefinition, 0),
	}
}

// Name returns the name of the plugin
func (p *PluginBase) Name() string {
	return p.name
}

// Description returns a description of what the plugin does
func (p *PluginBase) Description() string {
	return p.description
}

// Tools returns the tools provided by this plugin
func (p *PluginBase) Tools() []ToolDefinition {
	return p.tools
}

// AddTool adds a new tool to the plugin
func (p *PluginBase) AddTool(name, description string, inputSchema map[string]interface{}, function func(input []byte) (string, error)) {
	p.tools = append(p.tools, ToolDefinition{
		Name:        name,
		Description: description,
		InputSchema: inputSchema,
		Function:    function,
	})
}

// Plugin is the interface that all plugins must implement
type Plugin interface {
	// Name returns the name of the plugin
	Name() string
	// Description returns a description of what the plugin does
	Description() string
	// Tools returns the tools provided by this plugin
	Tools() []ToolDefinition
	// Initialize is called when the plugin is loaded
	Initialize() error
	// Cleanup is called when the plugin is unloaded
	Cleanup() error
}

// BasePlugin provides a base implementation of the Plugin interface
type BasePlugin struct {
	*PluginBase
}

// NewBasePlugin creates a new BasePlugin instance
func NewBasePlugin(name, description string) *BasePlugin {
	return &BasePlugin{
		PluginBase: NewPluginBase(name, description),
	}
}

// Initialize is called when the plugin is loaded
func (p *BasePlugin) Initialize() error {
	return nil
}

// Cleanup is called when the plugin is unloaded
func (p *BasePlugin) Cleanup() error {
	return nil
}

// ToolInput is a helper struct for unmarshaling tool input
type ToolInput struct {
	Data json.RawMessage
}

// UnmarshalInput unmarshals the tool input into the provided struct
func (t *ToolInput) UnmarshalInput(v interface{}) error {
	if err := json.Unmarshal(t.Data, v); err != nil {
		return fmt.Errorf("failed to unmarshal tool input: %w", err)
	}
	return nil
}

// ToolResult is a helper struct for creating tool results
type ToolResult struct {
	Success bool
	Message string
	Data    interface{}
}

// NewToolResult creates a new ToolResult
func NewToolResult(success bool, message string, data interface{}) ToolResult {
	return ToolResult{
		Success: success,
		Message: message,
		Data:    data,
	}
}

// ToJSON converts the tool result to JSON
func (t ToolResult) ToJSON() (string, error) {
	data := map[string]interface{}{
		"success": t.Success,
		"message": t.Message,
		"data":    t.Data,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal tool result: %w", err)
	}
	return string(jsonData), nil
} 