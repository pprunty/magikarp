package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
	"github.com/pprunty/magikarp/internal/registry"
	"github.com/pprunty/magikarp/models"
	"github.com/pprunty/magikarp/utils/history"
)

var toolTemplate = `package tool_{{.Name}}

import (
	"encoding/json"
	"fmt"

	"github.com/pprunty/magikarp/internal/registry"
)

// {{.Name | title}}Input represents the input for the {{.Name}} tool
type {{.Name | title}}Input struct {
	// TODO: Define input fields
}

// {{.Name | title}}Output represents the output from the {{.Name}} tool
type {{.Name | title}}Output struct {
	Success bool   ` + "`json:\"success\"`" + `
	Message string ` + "`json:\"message\"`" + `
	// TODO: Define output fields
}

func init() {
	inputs := map[string]interface{}{
		"properties": map[string]interface{}{
			// TODO: Define input schema
		},
		"type": "object",
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
			// TODO: Define output schema
		},
		"required": []string{"success", "message"},
		"type":     "object",
	}

	registry.RegisterGoTool(
		"{{.Name}}",
		"{{.Description}}",
		inputs,
		outputs,
		func(args json.RawMessage) (json.RawMessage, error) {
			var input {{.Name | title}}Input
			if err := json.Unmarshal(args, &input); err != nil {
				return nil, fmt.Errorf("failed to unmarshal input: %w", err)
			}

			// TODO: Implement the tool functionality

			output := {{.Name | title}}Output{
				Success: true,
				Message: "Operation completed successfully",
			}
			return json.Marshal(output)
		},
	)
}
`

var pythonToolTemplate = `#!/usr/bin/env python3
import json
import sys

def main():
    # Read input JSON from stdin
    args = json.load(sys.stdin)
    
    # TODO: Implement your tool functionality here
    result = {
        "success": True,
        "message": "Hello from {{.Name}} tool"
    }
    
    # Write result as JSON to stdout
    json.dump(result, sys.stdout)

if __name__ == "__main__":
    main()
`

var toolJsonTemplate = `{
  "name": "{{.Name}}",
  "description": "{{.Description}}",
  "language": "{{.Language}}",
  "entrypoint": "{{.Entrypoint}}",
  "run": "{{.Run}}",
  "inputs": {
    "properties": {
      // TODO: Define input schema
    },
    "type": "object"
  },
  "outputs": {
    "properties": {
      "success": {
        "type": "boolean",
        "description": "Whether the operation was successful."
      },
      "message": {
        "type": "string",
        "description": "A message describing the result of the operation."
      }
      // TODO: Define output schema
    },
    "required": ["success", "message"],
    "type": "object"
  }
}
`

var agentJsonTemplate = `{
  "name": "{{.Name}}",
  "description": "{{.Description}}",
  "tools": [],
  "model": "gpt-4",
  "strategy": "react"
}
`

// Title is a template function that converts first character to uppercase
func title(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func main() {
	var root = &cobra.Command{
		Use:   "mgk",
		Short: "Magikarp CLI for managing tools and agents",
		Long:  "Magikarp CLI for creating, installing, and managing tools and agents in the magikarp framework",
	}

	root.AddCommand(
		newCreateCmd("tool"),
		newCreateCmd("agent"),
		newAddCmd("tool"),
		newAddCmd("agent"),
		newListCmd(),
		newRunCmd(),
	)

	_ = root.Execute()
}

func newCreateCmd(kind string) *cobra.Command {
	// Define flags for the create command
	var language string
	var description string
	
	cmd := &cobra.Command{
		Use:   fmt.Sprintf("create %s <n>", kind),
		Short: "Scaffold a new " + kind,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			var tgtDir string
			
			switch kind {
			case "tool":
				tgtDir = path.Join("pkg", "tools", name)
				if err := scaffoldTool(tgtDir, name, language, description); err != nil {
					return err
				}
			case "agent":
				tgtDir = path.Join("pkg", "agents", name)
				if err := scaffoldAgent(tgtDir, name, description); err != nil {
					return err
				}
			}
			
			fmt.Println("Created", tgtDir)
			return nil
		},
	}
	
	// Add flags
	if kind == "tool" {
		cmd.Flags().StringVarP(&language, "language", "l", "go", "Tool implementation language (go, python, etc.)")
	}
	cmd.Flags().StringVarP(&description, "description", "d", "", "Description of the "+kind)
	
	return cmd
}

func newAddCmd(kind string) *cobra.Command {
	return &cobra.Command{
		Use:   fmt.Sprintf("add %s <url>", kind),
		Short: "Download a " + kind + " descriptor and code",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			url := args[0]
			resp, err := http.Get(url)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)
			switch kind {
			case "tool":
				var desc registry.ToolDescriptor
				if err := json.Unmarshal(body, &desc); err != nil {
					return err
				}
				return registry.InstallTool(desc)
			case "agent":
				var desc registry.AgentDescriptor
				if err := json.Unmarshal(body, &desc); err != nil {
					return err
				}
				return registry.InstallAgent(desc)
			}
			return nil
		},
	}
}

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list [tools|agents]",
		Short: "List all available tools or agents",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Default to listing both
			listTools := true
			listAgents := true
			
			if len(args) > 0 {
				switch strings.ToLower(args[0]) {
				case "tools":
					listAgents = false
				case "agents":
					listTools = false
				default:
					return fmt.Errorf("unknown resource type: %s (expected 'tools' or 'agents')", args[0])
				}
			}
			
			// Load and discover tools/agents
			if err := registry.DiscoverAndRegisterTools("."); err != nil {
				fmt.Printf("Warning: Failed to discover all tools: %v\n", err)
			}
			
			if err := registry.DiscoverAndRegisterAgents("."); err != nil {
				fmt.Printf("Warning: Failed to discover all agents: %v\n", err)
			}
			
			// Print tools
			if listTools {
				tools := registry.ListTools()
				fmt.Println("Available tools:")
				for _, name := range tools {
					tool, _ := registry.GetTool(name)
					fmt.Printf("  - %s: %s\n", name, tool.Description)
				}
				fmt.Println()
			}
			
			// Print agents
			if listAgents {
				agents := registry.ListAgents()
				fmt.Println("Available agents:")
				for _, name := range agents {
					agent, _ := registry.GetAgent(name)
					fmt.Printf("  - %s: %s\n", name, agent.Description)
				}
			}
			
			return nil
		},
	}
}

func newRunCmd() *cobra.Command {
	var modelFlag string
	
	cmd := &cobra.Command{
		Use:   "run <agent>",
		Short: "Run an agent with the specified model",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			agentName := args[0]
			
			// Discover available agents
			if err := registry.DiscoverAndRegisterAgents("."); err != nil {
				fmt.Printf("Warning: Failed to discover all agents: %v\n", err)
			}
			
			// Get the requested agent
			agentDef, exists := registry.GetAgent(agentName)
			if !exists {
				return fmt.Errorf("agent not found: %s", agentName)
			}
			
			// Use specified model or default from agent definition
			modelName := modelFlag
			if modelName == "" {
				modelName = agentDef.Model
			}
			
			// Set up the LLM client
			// TODO: Implement dynamic client creation based on model name
			client, err := models.NewDummyClient(modelName)
			if err != nil {
				return fmt.Errorf("failed to create LLM client: %w", err)
			}
			
			// Set up the history manager
			hist, err := history.NewHistoryManager()
			if err != nil {
				return fmt.Errorf("failed to set up command history: %w", err)
			}
			defer hist.Close()
			
			// Get user message function
			getUserMessage := func() (string, bool) {
				line, err := hist.ReadLine()
				if err != nil {
					return "", false
				}
				return strings.TrimSpace(line), true
			}
			
			// TODO: Create the agent instance dynamically based on agent type
			fmt.Printf("Running agent: %s with model: %s\n", agentName, modelName)
			
			// For now, just display a placeholder
			fmt.Println("Agent functionality coming soon!")
			
			return nil
		},
	}
	
	cmd.Flags().StringVarP(&modelFlag, "model", "m", "", "Specify the LLM model to use")
	
	return cmd
}

// scaffoldTool creates a new tool directory with template files
func scaffoldTool(dir, name, language, description string) error {
	// Create directory
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	// Create template data
	data := struct {
		Name        string
		Description string
		Language    string
		Entrypoint  string
		Run         string
	}{
		Name:        name,
		Description: description,
		Language:    language,
	}
	
	// Create custom template functions
	funcMap := template.FuncMap{
		"title": title,
	}
	
	// Create the appropriate files based on language
	switch language {
	case "go":
		data.Entrypoint = fmt.Sprintf("%s.go", name)
		data.Run = ""
		
		// Create Go file
		tmpl := template.Must(template.New("tool").Funcs(funcMap).Parse(toolTemplate))
		file, err := os.Create(path.Join(dir, data.Entrypoint))
		if err != nil {
			return fmt.Errorf("failed to create tool file: %w", err)
		}
		defer file.Close()
		
		if err := tmpl.Execute(file, data); err != nil {
			return fmt.Errorf("failed to write tool template: %w", err)
		}
		
	case "python":
		data.Entrypoint = "main.py"
		data.Run = "python3 main.py"
		
		// Create Python file
		tmpl := template.Must(template.New("pythontool").Funcs(funcMap).Parse(pythonToolTemplate))
		file, err := os.Create(path.Join(dir, data.Entrypoint))
		if err != nil {
			return fmt.Errorf("failed to create tool file: %w", err)
		}
		defer file.Close()
		
		if err := tmpl.Execute(file, data); err != nil {
			return fmt.Errorf("failed to write tool template: %w", err)
		}
		
		// Make the Python file executable
		if err := os.Chmod(path.Join(dir, data.Entrypoint), 0755); err != nil {
			return fmt.Errorf("failed to make script executable: %w", err)
		}
		
	default:
		// For other languages, just create a placeholder file
		file, err := os.Create(path.Join(dir, fmt.Sprintf("implement_me.%s", language)))
		if err != nil {
			return fmt.Errorf("failed to create placeholder file: %w", err)
		}
		file.Close()
		
		data.Entrypoint = fmt.Sprintf("implement_me.%s", language)
		data.Run = fmt.Sprintf("./%s", data.Entrypoint)
	}
	
	// Always create a tool.json descriptor
	jsonTmpl := template.Must(template.New("tooljson").Funcs(funcMap).Parse(toolJsonTemplate))
	jsonFile, err := os.Create(path.Join(dir, "tool.json"))
	if err != nil {
		return fmt.Errorf("failed to create tool.json: %w", err)
	}
	defer jsonFile.Close()
	
	if err := jsonTmpl.Execute(jsonFile, data); err != nil {
		return fmt.Errorf("failed to write tool.json template: %w", err)
	}
	
	return nil
}

// scaffoldAgent creates a new agent directory with template files
func scaffoldAgent(dir, name, description string) error {
	// Create directory
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	// Create template data
	data := struct {
		Name        string
		Description string
	}{
		Name:        name,
		Description: description,
	}
	
	// Create agent.json file
	jsonTmpl := template.Must(template.New("agentjson").Parse(agentJsonTemplate))
	jsonFile, err := os.Create(path.Join(dir, "agent.json"))
	if err != nil {
		return fmt.Errorf("failed to create agent.json: %w", err)
	}
	defer jsonFile.Close()
	
	if err := jsonTmpl.Execute(jsonFile, data); err != nil {
		return fmt.Errorf("failed to write agent.json template: %w", err)
	}
	
	return nil
} 