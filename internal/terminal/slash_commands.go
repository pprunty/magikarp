package terminal

import (
	"os"
	"strings"

	cfg "github.com/pprunty/magikarp/internal/config"
	"github.com/pprunty/magikarp/internal/orchestration"
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
		{Name: "/tools", Description: "Toggle tools on/off"},
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
	// Load configuration once
	configPath := findConfigFile()
	var conf *cfg.Config
	if path := configPath; path != "" {
		data, err := os.ReadFile(path)
		if err == nil {
			if err2 := yaml.Unmarshal(data, &conf); err2 == nil {
				// Expand env vars using existing helper
				// but easier: load via cfg.LoadConfig which handles expansion
			}
		}
	}

	// Simpler – just load via cfg.LoadConfig API which already does env expansion/validation
	c, err := cfg.LoadConfig(configPath)
	if err != nil {
		// fallback to default list
		return []string{"claude-3-5-sonnet-20240620", "gpt-4o", "gemini-pro"}
	}

	// Initialise registry; ignore errors – we only care about available models
	_ = orchestration.Init(c)

	// Collect models that have a registered provider
	models := orchestration.Models()

	if len(models) == 0 {
		return []string{"claude-3-5-sonnet-20240620", "gpt-4o", "gemini-pro"}
	}
	return models
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
