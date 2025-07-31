package list_tools

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/pprunty/magikarp/internal/providers"
	"github.com/pprunty/magikarp/internal/tools"
)

//go:embed tool.json
var rawSchema []byte

// Definition defines the list_tools tool.
func Definition() providers.ToolDefinition {
	var meta map[string]interface{}
	_ = json.Unmarshal(rawSchema, &meta)
	schema := meta["input_schema"].(map[string]interface{})

	return providers.ToolDefinition{
		Name:        meta["name"].(string),
		Description: meta["description"].(string),
		InputSchema: schema,
		Function:    run,
	}
}

// run returns a list of all registered tools.
func run(ctx context.Context, _ map[string]interface{}) (*providers.ToolResult, error) {
	all := tools.GetAllTools()
	var out string
	for _, t := range all {
		out += fmt.Sprintf("- %s: %s\n", t.Name, t.Description)
	}
	if out == "" {
		out = "No tools registered"
	}
	return providers.NewToolResult("list_tools", out, false), nil
}
