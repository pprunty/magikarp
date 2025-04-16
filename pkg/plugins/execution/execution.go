package execution

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/pprunty/magikarp/pkg/agent"
)

// ExecutionPlugin implements the Plugin interface for code execution
type ExecutionPlugin struct{}

// New creates a new ExecutionPlugin instance
func New() *ExecutionPlugin {
	return &ExecutionPlugin{}
}

// Name returns the name of the plugin
func (p *ExecutionPlugin) Name() string {
	return "execution"
}

// Description returns a description of what the plugin does
func (p *ExecutionPlugin) Description() string {
	return "Provides tools for executing code and commands"
}

// Tools returns the tools provided by this plugin
func (p *ExecutionPlugin) Tools() []agent.ToolDefinition {
	return []agent.ToolDefinition{
		executeCommandTool(),
		listToolsTool(),
	}
}

func executeCommandTool() agent.ToolDefinition {
	return agent.ToolDefinition{
		Name:        "execute_command",
		Description: "Execute a shell command and return its output. Use this to run code or system commands.",
		InputSchema: agent.GenerateSchema[ExecuteCommandInput](),
		Function:    executeCommand,
	}
}

func listToolsTool() agent.ToolDefinition {
	return agent.ToolDefinition{
		Name:        "list_tools",
		Description: "List all available tools and their descriptions",
		InputSchema: agent.GenerateSchema[struct{}](),
		Function:    listTools,
	}
}

type ExecuteCommandInput struct {
	Command string   `json:"command" jsonschema_description:"The command to execute"`
	Args    []string `json:"args,omitempty" jsonschema_description:"Optional arguments to pass to the command"`
}

func executeCommand(input []byte) (string, error) {
	executeInput := ExecuteCommandInput{}
	err := json.Unmarshal(input, &executeInput)
	if err != nil {
		return "", err
	}

	cmd := exec.Command(executeInput.Command, executeInput.Args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command failed: %w\nOutput: %s", err, output)
	}

	return string(output), nil
}

func listTools(input []byte) (string, error) {
	plugins := []struct {
		Name        string
		Description string
		Tools       []struct {
			Name        string
			Description string
		}
	}{
		{
			Name:        "filesystem",
			Description: "Provides tools for file system operations like reading, listing, and editing files",
			Tools: []struct {
				Name        string
				Description string
			}{
				{
					Name:        "read_file",
					Description: "Read the contents of a given relative file path",
				},
				{
					Name:        "list_files",
					Description: "List files and directories at a given path",
				},
				{
					Name:        "edit_file",
					Description: "Make edits to a text file by replacing text",
				},
			},
		},
		{
			Name:        "execution",
			Description: "Provides tools for executing code and commands",
			Tools: []struct {
				Name        string
				Description string
			}{
				{
					Name:        "execute_command",
					Description: "Execute a shell command and return its output",
				},
				{
					Name:        "list_tools",
					Description: "List all available tools and their descriptions",
				},
			},
		},
		{
			Name:        "text",
			Description: "Provides tools for text manipulation and transformation",
			Tools: []struct {
				Name        string
				Description string
			}{
				{
					Name:        "rot13",
					Description: "Apply ROT13 transformation to text",
				},
			},
		},
	}

	result, err := json.MarshalIndent(plugins, "", "  ")
	if err != nil {
		return "", err
	}

	return string(result), nil
} 