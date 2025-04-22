package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Model represents a single model configuration
type Model struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Provider represents a provider configuration
type Provider struct {
	Name     string  `json:"name"`
	Models   []Model `json:"models"`
	Required bool    `json:"required"`
}

// Config represents the entire configuration
type Config struct {
	Providers map[string]Provider `json:"providers"`
}

// LoadConfig loads the configuration from a file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// GetProvider returns a provider configuration by ID
func (c *Config) GetProvider(id string) (*Provider, error) {
	provider, ok := c.Providers[id]
	if !ok {
		return nil, fmt.Errorf("provider %s not found", id)
	}
	return &provider, nil
}

// GetModel returns a model configuration by provider ID and model name
func (c *Config) GetModel(providerID, modelName string) (*Model, error) {
	provider, err := c.GetProvider(providerID)
	if err != nil {
		return nil, err
	}

	for _, model := range provider.Models {
		if model.Name == modelName {
			return &model, nil
		}
	}

	return nil, fmt.Errorf("model %s not found for provider %s", modelName, providerID)
} 