package read_file

import (
	_ "embed"
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/pprunty/magikarp/pkg/agent"
)

//go:embed tool.json
var schema []byte

type input struct {
	Path string `json:"path"`
}

func Definition() agent.ToolDefinition {
	var params map[string]agent.Parameter
	_ = json.Unmarshal(schema, &params)
	return agent.ToolDefinition{
		Name:        "read_file",
		Description: "Read the contents of a text file (max ~100 KB)",
		Parameters:  params,
		Function:    run,
	}
}

func run(ctx context.Context, inputData map[string]interface{}) (*agent.ToolResult, error) {
	var in input
	inputBytes, err := json.Marshal(inputData)
	if err != nil {
		return agent.NewToolResult("read_file", err.Error(), true), nil
	}
	
	if err := json.Unmarshal(inputBytes, &in); err != nil {
		return agent.NewToolResult("read_file", err.Error(), true), nil
	}

	if in.Path == "" {
		return agent.NewToolResult("read_file", "path is required", true), nil
	}

	if !filepath.IsLocal(in.Path) {
		return agent.NewToolResult("read_file", "path must be local", true), nil
	}

	b, err := os.ReadFile(in.Path)
	if err != nil {
		return agent.NewToolResult("read_file", err.Error(), true), nil
	}

	if len(b) > 100_000 {
		return agent.NewToolResult("read_file", "file too big", true), nil
	}

	return agent.NewToolResult("read_file", string(b), false), nil
} 