package cli
import(
    "fmt"
    "os"
    "path/filepath"
    "github.com/urfave/cli/v2"
)

func NewCreateToolCmd()*cli.Command{
    return &cli.Command{
        Name:"create-tool",
        Usage:"Create new tool skeleton",
        Flags: []cli.Flag{
            &cli.StringFlag{Name:"plugin",Usage:"parent plugin"},
            &cli.StringFlag{Name:"name",Usage:"tool name"},
            &cli.StringFlag{Name:"description",Usage:"tool description",Value:"Tool description"},
        },
        Action: func(c *cli.Context) error{
            plugin:=c.String("plugin")
            name:=c.String("name")
            if plugin==""||name==""{
                return fmt.Errorf("plugin and name required")
            }
            
            // Create tool directory
            dir:=filepath.Join("pkg/plugins",plugin,name)
            if err:=os.MkdirAll(dir,0755); err!=nil {return err}
            
            // Create Go file
            gofile:=filepath.Join(dir,name+".go")
            content:=fmt.Sprintf(`package %s

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
        Name:        "%s",
        Description: "%s",
        Parameters:  params,
        Function:    run,
    }
}

func run(ctx context.Context, inputData map[string]interface{}) (*agent.ToolResult, error) {
    var in input
    inputBytes, err := json.Marshal(inputData)
    if err != nil {
        return agent.NewToolResult("%s", err.Error(), true), nil
    }
    
    if err := json.Unmarshal(inputBytes, &in); err != nil {
        return agent.NewToolResult("%s", err.Error(), true), nil
    }

    // TODO: Implement tool logic here

    return agent.NewToolResult("%s", "not implemented", true), nil
}`, name, name, c.String("description"), name, name, name)
            
            if err:=os.WriteFile(gofile,[]byte(content),0644); err!=nil {return err}
            
            // Create tool.json with proper schema
            jsonFile:=filepath.Join(dir,"tool.json")
            jsonContent:=fmt.Sprintf(`{
  "name": "%s",
  "description": "%s",
  "input_schema": {
    "type": "object",
    "properties": {
      "example_field": {
        "type": "string",
        "description": "Example field description"
      }
    },
    "required": [],
    "additionalProperties": false
  }
}`, name, c.String("description"))
            
            return os.WriteFile(jsonFile,[]byte(jsonContent),0644)
        },
    }
}
