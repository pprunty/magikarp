package execute_command

import (
    _ "embed"
    "context"
    "encoding/json"
    "fmt"
    "os/exec"
    "strings"

    "github.com/pprunty/magikarp/pkg/agent"
)

//go:embed tool.json
var schema []byte

type input struct {
    Command string   `json:"command"`
    Args    []string `json:"args,omitempty"`
}

func Definition() agent.ToolDefinition {
    var sch map[string]interface{}
    _ = json.Unmarshal(schema, &sch)
    return agent.ToolDefinition{
        Name:        "execute_command",
        Description: "Execute a shell command with optional arguments",
        InputSchema: sch,
        Function:    run,
    }
}

var destructive = []string{
    "rm -rf", "rm -r", "rmdir",
    "mkfs", "dd", "shred", "truncate",
}

func run(ctx context.Context, inputData map[string]interface{}) (*agent.ToolResult, error) {
    raw, err := json.Marshal(inputData)
    if err != nil {
        return agent.NewToolResult("execute_command", err.Error(), true), nil
    }

    var in input
    if err := json.Unmarshal(raw, &in); err != nil {
        return agent.NewToolResult("execute_command", err.Error(), true), nil
    }

    if in.Command == "" {
        return agent.NewToolResult("execute_command", "command cannot be empty", true), nil
    }

    lower := strings.ToLower(in.Command + " " + strings.Join(in.Args, " "))
    for _, pat := range destructive {
        if strings.Contains(lower, pat) {
            return agent.NewToolResult("execute_command", "command rejected: destructive operation", true), nil
        }
    }

    out, err := exec.Command(in.Command, in.Args...).CombinedOutput()
    if err != nil {
        return agent.NewToolResult("execute_command",
            fmt.Sprintf("execution failed: %v\n%s", err, string(out)), true), nil
    }
    return agent.NewToolResult("execute_command", strings.TrimSpace(string(out)), false), nil
}
