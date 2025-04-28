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
	var in struct {
		Path    string `json:"path"`    // dir or file
		Pattern string `json:"pattern"` // substring to match
	}
	if b, err := json.Marshal(inputData); err == nil {
		_ = json.Unmarshal(b, &in)
	}
	if in.Path == "" || in.Pattern == "" {
		return agent.NewToolResult("search_text",
			"`path` and `pattern` are required", true), nil
	}
	if !filepath.IsLocal(in.Path) {
		return agent.NewToolResult("search_text", "path must be local", true), nil
	}

	var matches []string
	info, err := os.Stat(in.Path)
	if err != nil {
		return agent.NewToolResult("search_text", err.Error(), true), nil
	}

	searchFile := func(path string, name string) error {
		// ① match against the *filename* itself
		if strings.Contains(name, in.Pattern) {
			matches = append(matches, path)
			// still fall through to content check for extra hits
		}
		// ② match inside file contents (best-effort, ignore binary read errors)
		if content, err := os.ReadFile(path); err == nil &&
			strings.Contains(string(content), in.Pattern) {
			matches = append(matches, path)
		}
		return nil
	}

	if info.IsDir() {
		err = filepath.WalkDir(in.Path, func(p string, d os.DirEntry, e error) error {
			if e != nil {
				return e
			}
			select { // honour cancellation
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
			if d.IsDir() {
				return nil
			}
			return searchFile(p, d.Name())
		})
	} else {
		err = searchFile(in.Path, info.Name())
	}
	if err != nil {
		return agent.NewToolResult("search_text", err.Error(), true), nil
	}

	out, _ := json.Marshal(matches)
	return agent.NewToolResult("search_text", string(out), false), nil
}