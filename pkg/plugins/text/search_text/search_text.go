package search_text

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
	Path    string `json:"path"`
	Pattern string `json:"pattern"`
}

func Definition() agent.ToolDefinition {
	var sch map[string]interface{}
	_ = json.Unmarshal(schema, &sch)
	return agent.ToolDefinition{
		Name:        "search_text",
		Description: "Search for text in files or directories",
		InputSchema: sch,
		Function:    run,
	}
}

func run(ctx context.Context, inputData map[string]interface{}) (*agent.ToolResult, error) {
	var in input
	inputBytes, err := json.Marshal(inputData)
	if err != nil {
		return agent.NewToolResult("search_text", err.Error(), true), nil
	}

	if err := json.Unmarshal(inputBytes, &in); err != nil {
		return agent.NewToolResult("search_text", err.Error(), true), nil
	}

	if in.Path == "" || in.Pattern == "" {
		return agent.NewToolResult("search_text", "path and pattern are required", true), nil
	}

	if !filepath.IsLocal(in.Path) {
		return agent.NewToolResult("search_text", "path must be local", true), nil
	}

	info, err := os.Stat(in.Path)
	if err != nil {
		return agent.NewToolResult("search_text", err.Error(), true), nil
	}

	var matches []string
	if info.IsDir() {
		err = filepath.Walk(in.Path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				content, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				if strings.Contains(string(content), in.Pattern) {
					matches = append(matches, path)
				}
			}
			return nil
		})
	} else {
		content, err := os.ReadFile(in.Path)
		if err != nil {
			return agent.NewToolResult("search_text", err.Error(), true), nil
		}
		if strings.Contains(string(content), in.Pattern) {
			matches = append(matches, in.Path)
		}
	}

	if err != nil {
		return agent.NewToolResult("search_text", err.Error(), true), nil
	}

	result, err := json.Marshal(matches)
	if err != nil {
		return agent.NewToolResult("search_text", err.Error(), true), nil
	}

	return agent.NewToolResult("search_text", string(result), false), nil
} 