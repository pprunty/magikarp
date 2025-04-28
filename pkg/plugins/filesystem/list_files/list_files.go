package list_files

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
	Path string `json:"path,omitempty"`
}

func Definition() agent.ToolDefinition {
	var sch map[string]interface{}
	_ = json.Unmarshal(schema, &sch)
	return agent.ToolDefinition{
		Name:        "list_files",
		Description: "List files and directories at a given path (defaults to current directory)",
		InputSchema: sch,
		Function:    run,
	}
}

func run(ctx context.Context, inputData map[string]interface{}) (*agent.ToolResult, error) {
	var in input
	inputBytes, err := json.Marshal(inputData)
	if err != nil {
		return agent.NewToolResult("list_files", err.Error(), true), nil
	}
	
	if err := json.Unmarshal(inputBytes, &in); err != nil {
		return agent.NewToolResult("list_files", err.Error(), true), nil
	}

	// If path is not provided, use current directory
	if in.Path == "" {
		in.Path = "."
	}

	if !filepath.IsLocal(in.Path) {
		return agent.NewToolResult("list_files", "path must be local", true), nil
	}

	entries, err := os.ReadDir(in.Path)
	if err != nil {
		return agent.NewToolResult("list_files", err.Error(), true), nil
	}

	var files []string
	for _, entry := range entries {
		files = append(files, entry.Name())
	}

	result, err := json.Marshal(files)
	if err != nil {
		return agent.NewToolResult("list_files", err.Error(), true), nil
	}

	return agent.NewToolResult("list_files", string(result), false), nil
} 