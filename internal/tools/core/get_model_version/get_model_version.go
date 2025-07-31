package get_model_version

import (
	"context"
	_ "embed"
	"encoding/json"

	"github.com/pprunty/magikarp/internal/providers"
	"github.com/pprunty/magikarp/internal/terminal"
)

//go:embed tool.json
var raw []byte

// Definition returns the providers.ToolDefinition for get_model_version.
func Definition() providers.ToolDefinition {
	var meta map[string]interface{}
	_ = json.Unmarshal(raw, &meta)
	schema := meta["input_schema"].(map[string]interface{})
	return providers.ToolDefinition{
		Name:        meta["name"].(string),
		Description: meta["description"].(string),
		InputSchema: schema,
		Function:    run,
	}
}

func run(ctx context.Context, _ map[string]interface{}) (*providers.ToolResult, error) {
	model := terminal.CurrentModel()
	if model == "" {
		model = "unknown"
	}
	return providers.NewToolResult("get_model_version", model, false), nil
}
