package registry

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// ToolDescriptor represents a tool descriptor loaded from a JSON file
type ToolDescriptor struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Language    string                 `json:"language"`
	Entrypoint  string                 `json:"entrypoint"`
	Build       string                 `json:"build,omitempty"`
	Run         string                 `json:"run"`
	Inputs      map[string]interface{} `json:"inputs"`
	Outputs     map[string]interface{} `json:"outputs"`
}

// AgentDescriptor represents an agent descriptor loaded from a JSON file
type AgentDescriptor struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Tools       []string `json:"tools"`
	Model       string   `json:"model"`
	Strategy    string   `json:"strategy"`
}

// LoadToolFromFile loads a tool descriptor from a JSON file
func LoadToolFromFile(path string) (*ToolDescriptor, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read tool descriptor: %w", err)
	}

	var desc ToolDescriptor
	if err := json.Unmarshal(data, &desc); err != nil {
		return nil, fmt.Errorf("failed to parse tool descriptor: %w", err)
	}

	return &desc, nil
}

// LoadAgentFromFile loads an agent descriptor from a JSON file
func LoadAgentFromFile(path string) (*AgentDescriptor, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read agent descriptor: %w", err)
	}

	var desc AgentDescriptor
	if err := json.Unmarshal(data, &desc); err != nil {
		return nil, fmt.Errorf("failed to parse agent descriptor: %w", err)
	}

	return &desc, nil
}

// RegisterToolFromFile loads and registers a tool from a JSON file
func RegisterToolFromFile(path string) error {
	desc, err := LoadToolFromFile(path)
	if err != nil {
		return err
	}

	// Create a runner based on the language
	var runner Runner
	switch desc.Language {
	case "go":
		// For Go tools, we expect the tool to be registered directly
		return fmt.Errorf("go tools should be registered directly, not through JSON: %s", desc.Name)
	default:
		// For other languages, create a process runner
		runner = NewProcessRunner(desc)
	}

	// Register the tool
	return RegisterTool(desc.Name, ToolDefinition{
		Name:        desc.Name,
		Description: desc.Description,
		Language:    desc.Language,
		InputSchema: desc.Inputs,
		OutputSchema: desc.Outputs,
		Runner:      runner,
	})
}

// RegisterAgentFromFile loads and registers an agent from a JSON file
func RegisterAgentFromFile(path string) error {
	desc, err := LoadAgentFromFile(path)
	if err != nil {
		return err
	}

	// Register the agent
	return RegisterAgent(desc.Name, AgentDefinition{
		Name:        desc.Name,
		Description: desc.Description,
		Tools:       desc.Tools,
		Model:       desc.Model,
		Strategy:    desc.Strategy,
	})
}

// DiscoverAndRegisterTools finds and registers all tools in the tools directory
func DiscoverAndRegisterTools(baseDir string) error {
	toolsDir := filepath.Join(baseDir, "pkg", "tools")

	// Walk through all subdirectories in the tools directory
	err := filepath.Walk(toolsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Look for tool.json files
		if !info.IsDir() && info.Name() == "tool.json" {
			if err := RegisterToolFromFile(path); err != nil {
				fmt.Printf("Warning: Failed to register tool from %s: %v\n", path, err)
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to discover tools: %w", err)
	}

	return nil
}

// DiscoverAndRegisterAgents finds and registers all agents in the agents directory
func DiscoverAndRegisterAgents(baseDir string) error {
	agentsDir := filepath.Join(baseDir, "pkg", "agents")

	// Walk through all subdirectories in the agents directory
	err := filepath.Walk(agentsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Look for agent.json files
		if !info.IsDir() && info.Name() == "agent.json" {
			if err := RegisterAgentFromFile(path); err != nil {
				fmt.Printf("Warning: Failed to register agent from %s: %v\n", path, err)
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to discover agents: %w", err)
	}

	return nil
}

// InstallTool installs a tool from a descriptor
func InstallTool(desc ToolDescriptor) error {
	// Create the tool directory
	toolDir := filepath.Join("pkg", "tools", desc.Name)
	if err := os.MkdirAll(toolDir, 0755); err != nil {
		return fmt.Errorf("failed to create tool directory: %w", err)
	}

	// Write the descriptor to a file
	descBytes, err := json.MarshalIndent(desc, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tool descriptor: %w", err)
	}

	if err := os.WriteFile(filepath.Join(toolDir, "tool.json"), descBytes, 0644); err != nil {
		return fmt.Errorf("failed to write tool descriptor: %w", err)
	}

	// TODO: Fetch tool code if URL is provided

	return nil
}

// InstallAgent installs an agent from a descriptor
func InstallAgent(desc AgentDescriptor) error {
	// Create the agent directory
	agentDir := filepath.Join("pkg", "agents", desc.Name)
	if err := os.MkdirAll(agentDir, 0755); err != nil {
		return fmt.Errorf("failed to create agent directory: %w", err)
	}

	// Write the descriptor to a file
	descBytes, err := json.MarshalIndent(desc, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal agent descriptor: %w", err)
	}

	if err := os.WriteFile(filepath.Join(agentDir, "agent.json"), descBytes, 0644); err != nil {
		return fmt.Errorf("failed to write agent descriptor: %w", err)
	}

	// TODO: Fetch agent code if URL is provided

	return nil
} 