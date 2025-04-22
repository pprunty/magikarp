package tool_filesystem

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pprunty/magikarp/internal/registry"
)

// ListFilesInput represents the input for the list_files tool
type ListFilesInput struct {
	Path string `json:"path,omitempty"`
}

// ListFilesOutput represents the output from the list_files tool
type ListFilesOutput struct {
	Success bool     `json:"success"`
	Message string   `json:"message"`
	Files   []string `json:"files,omitempty"`
}

func init() {
	inputs := map[string]interface{}{
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Optional relative path to list files from. Defaults to current directory if not provided.",
			},
		},
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
			"files": map[string]interface{}{
				"type":        "array",
				"description": "The list of files and directories if successful.",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
		},
		"required": []string{"success", "message"},
		"type":     "object",
	}

	registry.RegisterGoTool(
		"list_files",
		"List files and directories at a given path. If no path is provided, lists files in the current directory.",
		inputs,
		outputs,
		func(args json.RawMessage) (json.RawMessage, error) {
			var input ListFilesInput
			if err := json.Unmarshal(args, &input); err != nil {
				return nil, fmt.Errorf("failed to unmarshal input: %w", err)
			}

			dir := "."
			if input.Path != "" {
				dir = input.Path
			}

			var files []string
			err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				relPath, err := filepath.Rel(dir, path)
				if err != nil {
					return err
				}

				if relPath != "." {
					if info.IsDir() {
						files = append(files, relPath+"/")
					} else {
						files = append(files, relPath)
					}
				}
				return nil
			})

			if err != nil {
				output := ListFilesOutput{
					Success: false,
					Message: fmt.Sprintf("Failed to list files: %v", err),
				}
				return json.Marshal(output)
			}

			output := ListFilesOutput{
				Success: true,
				Message: "Files listed successfully",
				Files:   files,
			}
			return json.Marshal(output)
		},
	)
} 