package tool_filesystem

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/pprunty/magikarp/internal/registry"
)

// EditFileInput represents the input for the edit_file tool
type EditFileInput struct {
	Path   string `json:"path"`
	OldStr string `json:"old_str"`
	NewStr string `json:"new_str"`
}

// EditFileOutput represents the output from the edit_file tool
type EditFileOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func init() {
	inputs := map[string]interface{}{
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "The path to the file",
			},
			"old_str": map[string]interface{}{
				"type":        "string",
				"description": "Text to search for - must match exactly and must only have one match exactly",
			},
			"new_str": map[string]interface{}{
				"type":        "string",
				"description": "Text to replace old_str with",
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
		},
		"required": []string{"success", "message"},
		"type":     "object",
	}

	registry.RegisterGoTool(
		"edit_file",
		"Make edits to a text file. Replaces 'old_str' with 'new_str' in the given file.",
		inputs,
		outputs,
		func(args json.RawMessage) (json.RawMessage, error) {
			var input EditFileInput
			if err := json.Unmarshal(args, &input); err != nil {
				return nil, fmt.Errorf("failed to unmarshal input: %w", err)
			}

			if input.Path == "" || input.OldStr == input.NewStr {
				output := EditFileOutput{
					Success: false,
					Message: "Invalid input parameters",
				}
				return json.Marshal(output)
			}

			content, err := os.ReadFile(input.Path)
			if err != nil {
				if os.IsNotExist(err) && input.OldStr == "" {
					result, err := createNewFile(input.Path, input.NewStr)
					if err != nil {
						output := EditFileOutput{
							Success: false,
							Message: fmt.Sprintf("Failed to create file: %v", err),
						}
						return json.Marshal(output)
					}
					
					output := EditFileOutput{
						Success: true,
						Message: result,
					}
					return json.Marshal(output)
				}
				
				output := EditFileOutput{
					Success: false,
					Message: fmt.Sprintf("Failed to read file: %v", err),
				}
				return json.Marshal(output)
			}

			oldContent := string(content)
			newContent := strings.Replace(oldContent, input.OldStr, input.NewStr, -1)

			if oldContent == newContent && input.OldStr != "" {
				output := EditFileOutput{
					Success: false,
					Message: "old_str not found in file",
				}
				return json.Marshal(output)
			}

			err = os.WriteFile(input.Path, []byte(newContent), 0644)
			if err != nil {
				output := EditFileOutput{
					Success: false,
					Message: fmt.Sprintf("Failed to write file: %v", err),
				}
				return json.Marshal(output)
			}

			output := EditFileOutput{
				Success: true,
				Message: "File edited successfully",
			}
			return json.Marshal(output)
		},
	)
}

// createNewFile creates a new file with the given content
func createNewFile(filePath, content string) (string, error) {
	dir := path.Dir(filePath)
	if dir != "." {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return "", fmt.Errorf("failed to create directory: %w", err)
		}
	}

	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}

	return fmt.Sprintf("Successfully created file %s", filePath), nil
} 