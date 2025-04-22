package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

// ToolDefinition represents a registered tool that can be executed
type ToolDefinition struct {
	Name        string                  `json:"name"`
	Description string                  `json:"description"`
	Language    string                  `json:"language"`
	InputSchema map[string]interface{}  `json:"inputs"`
	OutputSchema map[string]interface{} `json:"outputs"`
	Runner      Runner                  `json:"-"` // Not serialized
}

// AgentDefinition represents a registered agent configuration
type AgentDefinition struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Tools       []string `json:"tools"`
	Model       string   `json:"model"`
	Strategy    string   `json:"strategy"`
}

// Runner interface defines how a tool is executed
type Runner interface {
	// Run executes the tool with JSON-encoded args; returns JSON result.
	Run(ctx context.Context, args json.RawMessage) (json.RawMessage, error)
}

var (
	toolsMu     sync.RWMutex
	tools       = make(map[string]ToolDefinition)
	
	agentsMu    sync.RWMutex
	agents      = make(map[string]AgentDefinition)
)

// RegisterTool registers a tool with the given name and definition
func RegisterTool(name string, def ToolDefinition) error {
	toolsMu.Lock()
	defer toolsMu.Unlock()
	
	if _, exists := tools[name]; exists {
		return fmt.Errorf("tool already registered: %s", name)
	}
	
	tools[name] = def
	return nil
}

// GetTool returns a tool definition by name
func GetTool(name string) (ToolDefinition, bool) {
	toolsMu.RLock()
	defer toolsMu.RUnlock()
	
	def, exists := tools[name]
	return def, exists
}

// ListTools returns all registered tool names
func ListTools() []string {
	toolsMu.RLock()
	defer toolsMu.RUnlock()
	
	var names []string
	for name := range tools {
		names = append(names, name)
	}
	return names
}

// RegisterAgent registers an agent with the given name and definition
func RegisterAgent(name string, def AgentDefinition) error {
	agentsMu.Lock()
	defer agentsMu.Unlock()
	
	if _, exists := agents[name]; exists {
		return fmt.Errorf("agent already registered: %s", name)
	}
	
	agents[name] = def
	return nil
}

// GetAgent returns an agent definition by name
func GetAgent(name string) (AgentDefinition, bool) {
	agentsMu.RLock()
	defer agentsMu.RUnlock()
	
	def, exists := agents[name]
	return def, exists
}

// ListAgents returns all registered agent names
func ListAgents() []string {
	agentsMu.RLock()
	defer agentsMu.RUnlock()
	
	var names []string
	for name := range agents {
		names = append(names, name)
	}
	return names
}

// RunTool executes a registered tool with the given arguments
func RunTool(ctx context.Context, name string, args interface{}) (json.RawMessage, error) {
	toolsMu.RLock()
	tool, exists := tools[name]
	toolsMu.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("tool not found: %s", name)
	}
	
	if tool.Runner == nil {
		return nil, fmt.Errorf("tool has no runner: %s", name)
	}
	
	// Convert args to JSON
	argsJSON, err := json.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal arguments: %w", err)
	}
	
	return tool.Runner.Run(ctx, argsJSON)
} 