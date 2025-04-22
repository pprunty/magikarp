package agents

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pprunty/magikarp/pkg/internal/registry"
	"github.com/pprunty/magikarp/pkg/models"
)

// Tool represents a tool that can be used by an agent
type Tool struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	InputSchema  map[string]interface{} `json:"input_schema"`
}

// ToolResult represents the result of a tool execution
type ToolResult struct {
	ID       string `json:"id"`
	Content  string `json:"content"`
	IsError  bool   `json:"is_error"`
}

// Agent defines the interface for all agent implementations
type Agent interface {
	// Run starts the agent's execution
	Run(ctx context.Context) error
	
	// Name returns the agent's name
	Name() string
	
	// Description returns the agent's description
	Description() string
	
	// SetClient sets the LLM client for the agent
	SetClient(client models.Client)
	
	// AddTool adds a tool to the agent
	AddTool(tool Tool)
	
	// EnableToolResults controls whether tool results are shown to the user
	EnableToolResults(enable bool)
}

// BaseAgent implements the common functionality for all agents
type BaseAgent struct {
	name            string
	description     string
	client          models.Client
	tools           []Tool
	showToolResults bool
	getUserMessage  func() (string, bool)
}

// NewBaseAgent creates a new base agent with the given name and description
func NewBaseAgent(name, description string, getUserMessage func() (string, bool)) *BaseAgent {
	return &BaseAgent{
		name:            name,
		description:     description,
		tools:           []Tool{},
		showToolResults: false,
		getUserMessage:  getUserMessage,
	}
}

// Name returns the agent's name
func (a *BaseAgent) Name() string {
	return a.name
}

// Description returns the agent's description
func (a *BaseAgent) Description() string {
	return a.description
}

// SetClient sets the LLM client for the agent
func (a *BaseAgent) SetClient(client models.Client) {
	a.client = client
}

// AddTool adds a tool to the agent
func (a *BaseAgent) AddTool(tool Tool) {
	a.tools = append(a.tools, tool)
}

// EnableToolResults controls whether tool results are shown to the user
func (a *BaseAgent) EnableToolResults(enable bool) {
	a.showToolResults = enable
}

// LoadToolsFromRegistry loads all tools from the registry
func (a *BaseAgent) LoadToolsFromRegistry(toolNames []string) error {
	for _, name := range toolNames {
		toolDef, exists := registry.GetTool(name)
		if !exists {
			return fmt.Errorf("tool not found: %s", name)
		}
		
		a.AddTool(Tool{
			Name:        toolDef.Name,
			Description: toolDef.Description,
			InputSchema: toolDef.InputSchema,
		})
	}
	
	return nil
}

// ExecuteTool executes a tool with the given name and arguments
func (a *BaseAgent) ExecuteTool(ctx context.Context, id, name string, input json.RawMessage) ToolResult {
	// Try to find the tool in the registry
	toolDef, exists := registry.GetTool(name)
	if !exists {
		return ToolResult{
			ID:      id,
			Content: fmt.Sprintf("tool not found: %s", name),
			IsError: true,
		}
	}
	
	// Execute the tool
	fmt.Printf("\u001b[92mtool\u001b[0m: %s(%s)\n", name, string(input))
	response, err := registry.RunTool(ctx, name, input)
	if err != nil {
		return ToolResult{
			ID:      id,
			Content: err.Error(),
			IsError: true,
		}
	}
	
	return ToolResult{
		ID:      id,
		Content: string(response),
		IsError: false,
	}
}

// Run is a placeholder that should be overridden by specific agent implementations
func (a *BaseAgent) Run(ctx context.Context) error {
	return fmt.Errorf("Run method not implemented for base agent")
} 