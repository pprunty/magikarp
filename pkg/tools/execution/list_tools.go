package tool_execution

import (
	"encoding/json"
	"fmt"

	"github.com/pprunty/magikarp/internal/registry"
)

// ListToolsOutput represents the output from the list_tools tool
type ListToolsOutput struct {
	Success bool         `json:"success"`
	Message string       `json:"message"`
	Tools   []ToolInfo   `json:"tools,omitempty"`
}

// ToolInfo represents information about a tool
type ToolInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func init() {
	inputs := map[string]interface{}{
		"type": "object",
	}

	outputs := map[string]interface{}{
		"properties": map[string]interface{}{
			"success": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether the operation was successful.",
			},
			"message": map[string]interface{}{
				"type":        "string",
				"description": "A message describing the result of the operation.",
			},
			"tools": map[string]interface{}{
				"type":        "array",
				"description": "The list of available tools if successful.",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name": map[string]interface{}{
							"type":        "string",
							"description": "The name of the tool.",
						},
						"description": map[string]interface{}{
							"type":        "string",
							"description": "The description of the tool.",
						},
					},
				},
			},
		},
		"required": []string{"success", "message"},
		"type":     "object",
	}

	registry.RegisterGoTool(
		"list_tools",
		"List all available tools and their descriptions",
		inputs,
		outputs,
		func(args json.RawMessage) (json.RawMessage, error) {
			// Get all registered tools
			toolNames := registry.ListTools()
			
			var toolInfos []ToolInfo
			for _, name := range toolNames {
				tool, exists := registry.GetTool(name)
				if exists {
					toolInfos = append(toolInfos, ToolInfo{
						Name:        tool.Name,
						Description: tool.Description,
					})
				}
			}
			
			if len(toolInfos) == 0 {
				return nil, fmt.Errorf("no tools found")
			}
			
			output := ListToolsOutput{
				Success: true,
				Message: "Tools listed successfully",
				Tools:   toolInfos,
			}
			return json.Marshal(output)
		},
	)
} 