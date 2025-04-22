package tool_filesystem

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/pprunty/magikarp/internal/registry"
)

// ReadFileInput represents the input for the read_file tool
type ReadFileInput struct {
	Path string `json:"path"`
}

// ReadFileOutput represents the output from the read_file tool
type ReadFileOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Content string `json:"content,omitempty"`
}

func init() {
	inputs := map[string]interface{}{
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "The relative path of a file in the working directory.",
			},
		},
		"required": []string{"path"},
		"type":     "object",
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
			"content": map[string]interface{}{
				"type":        "string",
				"description": "The content of the file if successful.",
			},
		},
		"required": []string{"success", "message"},
		"type":     "object",
	}

	registry.RegisterGoTool(
		"read_file",
		"Read the contents of a given relative file path. Use this when you want to see what's inside a file.",
		inputs,
		outputs,
		func(args json.RawMessage) (json.RawMessage, error) {
			var input ReadFileInput
			if err := json.Unmarshal(args, &input); err != nil {
				return nil, fmt.Errorf("failed to unmarshal input: %w", err)
			}

			content, err := os.ReadFile(input.Path)
			if err != nil {
				output := ReadFileOutput{
					Success: false,
					Message: fmt.Sprintf("Failed to read file: %v", err),
				}
				return json.Marshal(output)
			}

			output := ReadFileOutput{
				Success: true,
				Message: "File read successfully",
				Content: string(content),
			}
			return json.Marshal(output)
		},
	)
} 