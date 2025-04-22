package tool_execution

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/pprunty/magikarp/internal/registry"
)

// ExecuteCommandInput represents the input for the execute_command tool
type ExecuteCommandInput struct {
	Command string   `json:"command"`
	Args    []string `json:"args,omitempty"`
}

// ExecuteCommandOutput represents the output from the execute_command tool
type ExecuteCommandOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Output  string `json:"output,omitempty"`
}

func init() {
	inputs := map[string]interface{}{
		"properties": map[string]interface{}{
			"command": map[string]interface{}{
				"type":        "string",
				"description": "The command to execute",
			},
			"args": map[string]interface{}{
				"type":        "array",
				"description": "Optional arguments to pass to the command",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
		},
		"required": []string{"command"},
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
			"output": map[string]interface{}{
				"type":        "string",
				"description": "The command output if successful.",
			},
		},
		"required": []string{"success", "message"},
		"type":     "object",
	}

	registry.RegisterGoTool(
		"execute_command",
		"Execute a shell command and return its output. Use this to run code or system commands.",
		inputs,
		outputs,
		func(args json.RawMessage) (json.RawMessage, error) {
			var input ExecuteCommandInput
			if err := json.Unmarshal(args, &input); err != nil {
				return nil, fmt.Errorf("failed to unmarshal input: %w", err)
			}

			cmd := exec.Command(input.Command, input.Args...)
			output, err := cmd.CombinedOutput()
			if err != nil {
				executeOutput := ExecuteCommandOutput{
					Success: false,
					Message: fmt.Sprintf("Command failed: %v", err),
					Output:  string(output),
				}
				return json.Marshal(executeOutput)
			}

			executeOutput := ExecuteCommandOutput{
				Success: true,
				Message: "Command executed successfully",
				Output:  string(output),
			}
			return json.Marshal(executeOutput)
		},
	)
} 