package web

import (
    _ "embed"
    "context"
    "encoding/json"
    "github.com/pprunty/magikarp/pkg/agent"
)

//go:embed tool.json
var schema []byte

type input struct {
    // Add your input fields here
}

func Definition() agent.ToolDefinition {
    var params map[string]agent.Parameter
    _ = json.Unmarshal(schema, &params)
    return agent.ToolDefinition{
        Name:        "web",
        Description: "Tool description",
        Parameters:  params,
        Function:    run,
    }
}

func run(ctx context.Context, inputData map[string]interface{}) (*agent.ToolResult, error) {
    var in input
    inputBytes, err := json.Marshal(inputData)
    if err != nil {
        return agent.NewToolResult("web", err.Error(), true), nil
    }
    
    if err := json.Unmarshal(inputBytes, &in); err != nil {
        return agent.NewToolResult("web", err.Error(), true), nil
    }

    // TODO: Implement tool logic here

    return agent.NewToolResult("web", "not implemented", true), nil
}