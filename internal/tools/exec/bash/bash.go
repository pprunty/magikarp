package bash

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/pprunty/magikarp/internal/providers"
)

//go:embed tool.json
var schema []byte

// Input represents the parameters for the execute_command tool
type input struct {
	Script  string `json:"script"`
	Timeout int    `json:"timeout,omitempty"`
	WorkDir string `json:"work_dir,omitempty"`
}

// Definition returns the tool definition for the execute_command tool
func Definition() providers.ToolDefinition {
	var sch map[string]interface{}
	if err := json.Unmarshal(schema, &sch); err != nil {
		// Handle error properly in a production environment
		fmt.Printf("Error unmarshaling schema: %v\n", err)
	}
	return providers.ToolDefinition{
		Name:        sch["name"].(string),
		Description: sch["description"].(string),
		InputSchema: sch["input_schema"].(map[string]interface{}),
		Function:    run,
	}
}

// List of potentially dangerous commands or command patterns that should be blocked
var dangerousCommands = []string{
	// File system destructive operations
	"rm -rf", "rm -r", "rmdir", "mkfs", "dd", "shred", "truncate",

	// System control commands
	"shutdown", "reboot", "halt", "poweroff",

	// Network-altering commands
	"iptables", "ip6tables", "ufw",

	// User management
	"passwd", "useradd", "userdel", "groupadd", "groupdel",

	// Privilege escalation
	"sudo", "su", "doas",

	// Other potentially harmful commands
	"> /dev/null", ">/dev/null", "> /dev/", ">/dev/",
	"> /proc/", ">/proc/", "> /sys/", ">/sys/",
	"|", "||", "&&", ";", "$(", "`", // Command chaining
}

// run executes the command and returns the result
func run(ctx context.Context, inputData map[string]interface{}) (*providers.ToolResult, error) {
	// Convert generic input data to our structured input type
	raw, err := json.Marshal(inputData)
	if err != nil {
		return providers.NewToolResult("bash", fmt.Sprintf("Error processing input parameters: %v", err), true), nil
	}

	var in input
	if err := json.Unmarshal(raw, &in); err != nil {
		return providers.NewToolResult("bash", fmt.Sprintf("Error parsing input parameters: %v", err), true), nil
	}

	// Validate input
	if strings.TrimSpace(in.Script) == "" {
		return providers.NewToolResult("bash", "script parameter cannot be empty", true), nil
	}

	// Set default timeout if not specified
	timeout := 30                           // Default timeout in seconds
	if in.Timeout > 0 && in.Timeout < 300 { // Cap at 5 minutes
		timeout = in.Timeout
	}

	// Security check: block potentially dangerous commands
	commandLine := strings.ToLower(in.Script)
	for _, dangerous := range dangerousCommands {
		if strings.Contains(commandLine, dangerous) {
			return providers.NewToolResult(
				"bash",
				fmt.Sprintf("Command rejected for security reasons: contains '%s'", dangerous),
				true,
			), nil
		}
	}

	// Create a context with timeout
	execCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	// Create command with the provided context (bash -c "script")
	cmd := exec.CommandContext(execCtx, "bash", "-c", in.Script)

	// Set working directory if specified
	if in.WorkDir != "" {
		cmd.Dir = in.WorkDir
	}

	// Execute the command and capture output
	out, err := cmd.CombinedOutput()

	// Check for timeout
	if execCtx.Err() == context.DeadlineExceeded {
		return providers.NewToolResult(
			"bash",
			fmt.Sprintf("Command execution timed out after %d seconds", timeout),
			true,
		), nil
	}

	// Handle command execution errors
	if err != nil {
		exitErr, ok := err.(*exec.ExitError)
		if ok {
			return providers.NewToolResult(
				"bash",
				fmt.Sprintf("Command exited with status %d\n%s", exitErr.ExitCode(), string(out)),
				true,
			), nil
		}
		return providers.NewToolResult(
			"bash",
			fmt.Sprintf("Execution failed: %v\n%s", err, string(out)),
			true,
		), nil
	}

	// Success case
	return providers.NewToolResult("bash", strings.TrimSpace(string(out)), false), nil
}
