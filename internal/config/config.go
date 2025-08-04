package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Name   string `yaml:"name"`
	System string `yaml:"system"`
	// DefaultModel is the model that should be selected when Magikarp starts.
	// If empty, the first registered model will be used instead.
	DefaultModel string `yaml:"default_model"`
	// DefaultTemperature is the global default temperature for all providers.
	// Individual providers can override this by specifying their own temperature.
	DefaultTemperature float64 `yaml:"default_temperature"`
	// Tools groups all tool related configuration (enabled/visibility)
	Tools     ToolsConfig         `yaml:"tools"`
	Providers map[string]Provider `yaml:"providers"`
}

// Provider represents an LLM provider configuration
type Provider struct {
	Models      []string `yaml:"models"`
	Temperature float64  `yaml:"temperature"`
	Key         string   `yaml:"key"`
}

// ToolsConfig represents configuration for tool usage and UI output.
type ToolsConfig struct {
	Enabled bool `yaml:"enabled"`
	Output  bool `yaml:"output"`
}

// LoadConfig loads configuration from the specified file path
func LoadConfig(configPath string) (*Config, error) {
	// Try to load .env file from multiple locations
	envPaths := []string{".env", "./.env", filepath.Join(".", ".env")}
	envLoaded := false

	for _, envPath := range envPaths {
		if err := godotenv.Load(envPath); err == nil {
			envLoaded = true
			break
		}
	}

	// Also try to load from home directory
	if !envLoaded {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			homeEnvPath := filepath.Join(homeDir, ".magikarp.env")
			godotenv.Load(homeEnvPath) // Ignore error
		}
	}

	// If no config path specified, look for config.yaml in current directory
	if configPath == "" {
		configPath = "config.yaml"
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Expand environment variables in system prompt
	config.System = os.ExpandEnv(config.System)

	// Expand environment variables in API keys
	for name, provider := range config.Providers {
		originalKey := provider.Key
		provider.Key = os.ExpandEnv(provider.Key)
		config.Providers[name] = provider

		// Debug: check if environment variable expansion worked
		if originalKey != provider.Key && provider.Key != "" {
			// Environment variable was successfully expanded
		} else if originalKey == provider.Key && originalKey != "" && originalKey[0] == '$' {
			// Environment variable was not expanded, might be missing
			// This is expected behavior - the validation happens elsewhere
		}
	}

	return &config, nil
}

// GetDefaultConfigPath returns the default configuration file path
func GetDefaultConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "config.yaml"
	}
	return filepath.Join(homeDir, ".magikarp.yaml")
}

// ValidateConfig validates the configuration
func (c *Config) ValidateConfig() error {
	if c.Name == "" {
		return fmt.Errorf("config name is required")
	}

	if len(c.Providers) == 0 {
		return fmt.Errorf("at least one provider must be configured")
	}

	for name, provider := range c.Providers {
		if len(provider.Models) == 0 {
			return fmt.Errorf("provider %s must have at least one model", name)
		}
		// Don't require API keys in validation - they'll be checked during provider initialization
	}

	if c.DefaultModel != "" {
		// Ensure the default model has a registered provider entry.
		found := false
		for _, provider := range c.Providers {
			for _, m := range provider.Models {
				if m == c.DefaultModel {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			return fmt.Errorf("default_model %s does not exist in any provider model list", c.DefaultModel)
		}
	}

	return nil
}

// GetEffectiveTemperature returns the temperature to use for a given provider.
// If the provider has a specific temperature set, it uses that; otherwise, it uses the global default.
func (c *Config) GetEffectiveTemperature(providerName string) float64 {
	if provider, ok := c.Providers[providerName]; ok && provider.Temperature != 0 {
		return provider.Temperature
	}
	return c.DefaultTemperature
}

