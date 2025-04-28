package list_tools

import (
	_ "embed"
	"context"
	"encoding/json"
	"github.com/pprunty/magikarp/pkg/agent"
	"github.com/pprunty/magikarp/pkg/loader"
)

//go:embed tool.json
var schema []byte

func Definition() agent.ToolDefinition {
	var sch map[string]interface{}
	_ = json.Unmarshal(schema, &sch)
	return agent.ToolDefinition{
		Name:        "list_tools",
		Description: "List all available tools and their descriptions",
		InputSchema: sch,
		Function:    run,
	}
}

func run(ctx context.Context, input map[string]interface{}) (*agent.ToolResult, error) {
	// Get all registered plugins
	plugins := loader.All()
	
	// Convert plugins to the expected format
	result := make([]struct {
		Name  string
		Tools []struct {
			Name        string
			Description string
		}
	}, len(plugins))
	
	for i, p := range plugins {
		result[i].Name = p.Name()
		
		// Get tools for this plugin
		tools := p.Tools()
		result[i].Tools = make([]struct {
			Name        string
			Description string
		}, len(tools))
		
		for j, t := range tools {
			result[i].Tools[j].Name = t.Name
			result[i].Tools[j].Description = t.Description
		}
	}

	// Marshal to JSON
	jsonResult, err := json.Marshal(result)
	if err != nil {
		return agent.NewToolResult("list_tools", err.Error(), true), nil
	}

	return agent.NewToolResult("list_tools", string(jsonResult), false), nil
} 