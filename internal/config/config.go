package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Name      string                `yaml:"name"`
	System    string                `yaml:"system"`
	Providers map[string]Provider   `yaml:"providers"`
	Speech    Speech                `yaml:"speech"`
}

// Provider represents an LLM provider configuration
type Provider struct {
	Models      []string `yaml:"models"`
	Temperature float64  `yaml:"temperature"`
	Key         string   `yaml:"key"`
}

// Speech represents the speech-to-text configuration
type Speech struct {
	ModelPath  string `yaml:"model_path"`
	SampleRate int    `yaml:"sample_rate"`
	Keyword    string `yaml:"keyword"`
	Enabled    bool   `yaml:"enabled"`
}

// LoadConfig loads configuration from the specified file path
func LoadConfig(configPath string) (*Config, error) {
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

	// Expand environment variables in API keys
	for name, provider := range config.Providers {
		provider.Key = os.ExpandEnv(provider.Key)
		config.Providers[name] = provider
	}

	// Expand environment variables in speech model path
	config.Speech.ModelPath = os.ExpandEnv(config.Speech.ModelPath)

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
		if provider.Key == "" {
			return fmt.Errorf("provider %s must have an API key", name)
		}
	}

	// Validate speech configuration if enabled
	if c.Speech.Enabled {
		if c.Speech.ModelPath == "" {
			return fmt.Errorf("speech model path is required when speech is enabled")
		}
		if c.Speech.SampleRate <= 0 {
			return fmt.Errorf("speech sample rate must be positive")
		}
		if c.Speech.Keyword == "" {
			return fmt.Errorf("speech keyword is required when speech is enabled")
		}
	}

	return nil
}