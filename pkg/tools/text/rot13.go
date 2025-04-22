package tool_text

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pprunty/magikarp/internal/registry"
)

// Rot13Input represents the input for the rot13 tool
type Rot13Input struct {
	Text string `json:"text"`
}

// Rot13Output represents the output from the rot13 tool
type Rot13Output struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Result  string `json:"result,omitempty"`
}

func init() {
	inputs := map[string]interface{}{
		"properties": map[string]interface{}{
			"text": map[string]interface{}{
				"type":        "string",
				"description": "The text to transform",
			},
		},
		"required": []string{"text"},
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
			"result": map[string]interface{}{
				"type":        "string",
				"description": "The transformed text if successful.",
			},
		},
		"required": []string{"success", "message"},
		"type":     "object",
	}

	registry.RegisterGoTool(
		"rot13",
		"Apply ROT13 transformation to text",
		inputs,
		outputs,
		func(args json.RawMessage) (json.RawMessage, error) {
			var input Rot13Input
			if err := json.Unmarshal(args, &input); err != nil {
				return nil, fmt.Errorf("failed to unmarshal input: %w", err)
			}

			transformed := rot13Transform(input.Text)
			
			output := Rot13Output{
				Success: true,
				Message: "Text transformed successfully",
				Result:  transformed,
			}
			return json.Marshal(output)
		},
	)
}

// rot13Transform applies the ROT13 transformation to the given text
func rot13Transform(text string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case r >= 'A' && r <= 'Z':
			return 'A' + (r-'A'+13)%26
		case r >= 'a' && r <= 'z':
			return 'a' + (r-'a'+13)%26
		default:
			return r
		}
	}, text)
} 