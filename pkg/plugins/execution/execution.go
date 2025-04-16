package execution

import (
	"fmt"
	"os/exec"

	"github.com/pprunty/magikarp/pkg/agent"
)

// ExecutionPlugin implements the Plugin interface for code execution
type ExecutionPlugin struct {
	*agent.BasePlugin
}

// New creates a new ExecutionPlugin instance
func New() *ExecutionPlugin {
	plugin := &ExecutionPlugin{
		BasePlugin: agent.NewBasePlugin("execution", "Provides tools for executing code and commands"),
	}

	// Add tools during initialization
	plugin.AddTool("execute_command", "Execute a shell command and return its output. Use this to run code or system commands.",
		agent.GenerateSchema[ExecuteCommandInput](),
		plugin.executeCommand)

	plugin.AddTool("list_tools", "List all available tools and their descriptions",
		agent.GenerateSchema[struct{}](),
		plugin.listTools)

	return plugin
}

// Initialize is called when the plugin is loaded
func (p *ExecutionPlugin) Initialize() error {
	// Any initialization code can go here
	return nil
}

// Cleanup is called when the plugin is unloaded
func (p *ExecutionPlugin) Cleanup() error {
	// Any cleanup code can go here
	return nil
}

type ExecuteCommandInput struct {
	Command string   `json:"command" jsonschema_description:"The command to execute"`
	Args    []string `json:"args,omitempty" jsonschema_description:"Optional arguments to pass to the command"`
}

func (p *ExecutionPlugin) executeCommand(input []byte) (string, error) {
	var toolInput agent.ToolInput
	toolInput.Data = input

	var executeInput ExecuteCommandInput
	if err := toolInput.UnmarshalInput(&executeInput); err != nil {
		return "", err
	}

	cmd := exec.Command(executeInput.Command, executeInput.Args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return agent.NewToolResult(false, fmt.Sprintf("Command failed: %v\nOutput: %s", err, output), nil).ToJSON()
	}

	return agent.NewToolResult(true, "Command executed successfully", string(output)).ToJSON()
}

func (p *ExecutionPlugin) listTools(input []byte) (string, error) {
	var toolInput agent.ToolInput
	toolInput.Data = input

	// Unmarshal empty struct to validate input
	var empty struct{}
	if err := toolInput.UnmarshalInput(&empty); err != nil {
		return "", err
	}

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

	return agent.NewToolResult(true, "Tools listed successfully", plugins).ToJSON()
} 