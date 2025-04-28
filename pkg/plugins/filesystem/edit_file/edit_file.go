package edit_file

import (
	_ "embed"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/pprunty/magikarp/pkg/agent"
)

//go:embed tool.json
var schema []byte

type input struct {
	Path        string `json:"path"`
	Search      string `json:"search,omitempty"`
	ReplaceWith string `json:"replace_with,omitempty"`
	Append      string `json:"append,omitempty"`
}

func Definition() agent.ToolDefinition {
	var params map[string]agent.Parameter
	_ = json.Unmarshal(schema, &params)
	return agent.ToolDefinition{
		Name:        "edit_file",
		Description: "Search/replace or append to a text file",
		Parameters:  params,
		Function:    run,
	}
}

func run(ctx context.Context, inputData map[string]interface{}) (*agent.ToolResult, error) {
	var in input
	inputBytes, err := json.Marshal(inputData)
	if err != nil {
		return agent.NewToolResult("edit_file", err.Error(), true), nil
	}
	
	if err := json.Unmarshal(inputBytes, &in); err != nil {
		return agent.NewToolResult("edit_file", err.Error(), true), nil
	}

	if !filepath.IsLocal(in.Path) {
		return agent.NewToolResult("edit_file", "path must be local", true), nil
	}

	data, err := os.ReadFile(in.Path)
	if err != nil {
		return agent.NewToolResult("edit_file", err.Error(), true), nil
	}
	orig := string(data)

	var changed string
	switch {
	case in.Search != "":
		changed = strings.ReplaceAll(orig, in.Search, in.ReplaceWith)
	case in.Append != "":
		changed = orig + in.Append
	default:
		return agent.NewToolResult("edit_file", "no operation provided", true), nil
	}

	if changed == orig {
		return agent.NewToolResult("edit_file", "no changes made", false), nil
	}

	if err := os.WriteFile(in.Path, []byte(changed), 0644); err != nil {
		return agent.NewToolResult("edit_file", err.Error(), true), nil
	}
	return agent.NewToolResult("edit_file", "file updated", false), nil
} 