package terminal

import (
	"os"
	"path/filepath"
	"gopkg.in/yaml.v3"
)

// Config represents the structure of config.yaml
type Config struct {
	Name     string `yaml:"name"`
	Version  string `yaml:"version"`
	// Other fields can be added as needed
}

// GetVersion returns the version from config.yaml
func GetVersion() string {
	// Try to find config.yaml in current directory or parent directories
	configPath := findConfigFile()
	if configPath == "" {
		return "unknown"
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return "unknown"
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return "unknown"
	}

	if config.Version == "" {
		return "unknown"
	}

	return config.Version
}

// findConfigFile searches for config.yaml starting from current directory
func findConfigFile() string {
	// Try current working directory first
	wd, err := os.Getwd()
	if err == nil {
		configPath := filepath.Join(wd, "config.yaml")
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}
	}

	// Try common locations
	commonPaths := []string{
		"config.yaml",
		"./config.yaml",
		"../config.yaml",
		"../../config.yaml",
	}

	for _, path := range commonPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// Search up the directory tree as fallback
	if wd != "" {
		for {
			configPath := filepath.Join(wd, "config.yaml")
			if _, err := os.Stat(configPath); err == nil {
				return configPath
			}

			parent := filepath.Dir(wd)
			if parent == wd {
				// Reached root directory
				break
			}
			wd = parent
		}
	}

	return ""
}

// GetVersionDisplay returns the formatted version string for display
func GetVersionDisplay() string {
	version := GetVersion()
	if version == "unknown" {
		return "Magikarp version unknown"
	}
	return "Magikarp " + version
}