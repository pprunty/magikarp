package terminal

import (
	"os"
	"strings"
	"gopkg.in/yaml.v3"
)

// SlashCommand represents a slash command with its name and description
type SlashCommand struct {
	Name        string
	Description string
}

// GetAvailableCommands returns the list of available slash commands in alphabetical order
func GetAvailableCommands() []SlashCommand {
	return []SlashCommand{
		{Name: "/exit", Description: "Exit Magikarp"},
		{Name: "/help", Description: "Show help information"},
		{Name: "/model", Description: "Switch between AI models"},
		{Name: "/speech", Description: "Toggle speech mode on/off"},
	}
}

// ConfigYAML represents the structure of config.yaml for model loading
type ConfigYAML struct {
	Providers map[string]struct {
		Models []string `yaml:"models"`
	} `yaml:"providers"`
}

// GetAvailableModels returns the list of available AI models from config.yaml
func GetAvailableModels() []string {
	// Try to find and read config.yaml
	configPath := findConfigFile()
	if configPath == "" {
		// Fallback to hardcoded models if config not found
		return []string{"claude-3-5-sonnet-20240620", "gpt-4o", "gemini-pro"}
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		// Fallback on read error
		return []string{"claude-3-5-sonnet-20240620", "gpt-4o", "gemini-pro"}
	}

	var config ConfigYAML
	if err := yaml.Unmarshal(data, &config); err != nil {
		// Fallback on parse error
		return []string{"claude-3-5-sonnet-20240620", "gpt-4o", "gemini-pro"}
	}

	// Collect all models from all providers
	var allModels []string
	for _, provider := range config.Providers {
		allModels = append(allModels, provider.Models...)
	}

	if len(allModels) == 0 {
		// Fallback if no models found
		return []string{"claude-3-5-sonnet-20240620", "gpt-4o", "gemini-pro"}
	}

	return allModels
}

// FilterCommands filters slash commands based on the input text
func FilterCommands(input string) []SlashCommand {
	if input == "/" || input == "" {
		return GetAvailableCommands()
	}
	
	// Remove the leading "/" for filtering
	filterText := strings.ToLower(strings.TrimPrefix(input, "/"))
	allCommands := GetAvailableCommands()
	var filtered []SlashCommand
	
	for _, cmd := range allCommands {
		// Check if command name (without /) contains the filter text
		cmdName := strings.ToLower(strings.TrimPrefix(cmd.Name, "/"))
		cmdDesc := strings.ToLower(cmd.Description)
		
		if strings.Contains(cmdName, filterText) || strings.Contains(cmdDesc, filterText) {
			filtered = append(filtered, cmd)
		}
	}
	
	return filtered
}

// GetModelDisplayName returns the full model name for display
func GetModelDisplayName(modelName string) string {
	// Since we're now using actual model names, just return the model name
	return modelName
}