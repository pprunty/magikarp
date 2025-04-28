package execute_command

import (
    _ "embed"
    "context"
    "encoding/json"
    "fmt"
    "os/exec"
    "strings"
    "time"

    "github.com/pprunty/magikarp/pkg/agent"
)

//go:embed tool.json
var schema []byte

// Input represents the parameters for the execute_command tool
type input struct {
    Command string   `json:"command"`
    Args    []string `json:"args,omitempty"`
    Timeout int      `json:"timeout,omitempty"`
    WorkDir string   `json:"work_dir,omitempty"`
}

// Definition returns the tool definition for the execute_command tool
func Definition() agent.ToolDefinition {
    var sch map[string]interface{}
    if err := json.Unmarshal(schema, &sch); err != nil {
        // Handle error properly in a production environment
        fmt.Printf("Error unmarshaling schema: %v\n", err)
    }
    return agent.ToolDefinition{
        Name:        "execute_command",
        Description: "Execute a shell command with optional arguments",
        InputSchema: sch,
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
    "|", "||", "&&", ";", "$(", "`",  // Command chaining
}

// run executes the command and returns the result
func run(ctx context.Context, inputData map[string]interface{}) (*agent.ToolResult, error) {
    // Convert generic input data to our structured input type
    raw, err := json.Marshal(inputData)
    if err != nil {
        return agent.NewToolResult("execute_command", fmt.Sprintf("Error processing input parameters: %v", err), true), nil
    }

    var in input
    if err := json.Unmarshal(raw, &in); err != nil {
        return agent.NewToolResult("execute_command", fmt.Sprintf("Error parsing input parameters: %v", err), true), nil
    }

    // Validate input
    if in.Command == "" {
        return agent.NewToolResult("execute_command", "Command parameter cannot be empty", true), nil
    }

    // Set default timeout if not specified
    timeout := 30 // Default timeout in seconds
    if in.Timeout > 0 && in.Timeout < 300 { // Cap at 5 minutes
        timeout = in.Timeout
    }

    // Security check: block potentially dangerous commands
    commandLine := strings.ToLower(in.Command + " " + strings.Join(in.Args, " "))
    for _, dangerous := range dangerousCommands {
        if strings.Contains(commandLine, dangerous) {
            return agent.NewToolResult(
                "execute_command",
                fmt.Sprintf("Command rejected for security reasons: contains '%s'", dangerous),
                true,
            ), nil
        }
    }

    // Create a context with timeout
    execCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
    defer cancel()

    // Create command with the provided context
    cmd := exec.CommandContext(execCtx, in.Command, in.Args...)

    // Set working directory if specified
    if in.WorkDir != "" {
        cmd.Dir = in.WorkDir
    }

    // Execute the command and capture output
    out, err := cmd.CombinedOutput()

    // Check for timeout
    if execCtx.Err() == context.DeadlineExceeded {
        return agent.NewToolResult(
            "execute_command",
            fmt.Sprintf("Command execution timed out after %d seconds", timeout),
            true,
        ), nil
    }

    // Handle command execution errors
    if err != nil {
        exitErr, ok := err.(*exec.ExitError)
        if ok {
            return agent.NewToolResult(
                "execute_command",
                fmt.Sprintf("Command exited with status %d\n%s", exitErr.ExitCode(), string(out)),
                true,
            ), nil
        }
        return agent.NewToolResult(
            "execute_command",
            fmt.Sprintf("Execution failed: %v\n%s", err, string(out)),
            true,
        ), nil
    }

    // Success case
    return agent.NewToolResult("execute_command", strings.TrimSpace(string(out)), false), nil
}