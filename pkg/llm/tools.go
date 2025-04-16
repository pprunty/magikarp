package llm

import (
	"encoding/json"
	"fmt"
	"os"
)

// ToolConfig represents the configuration for a tool
type ToolConfig struct {
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	Category        string                 `json:"category"`
	TriggerKeywords []string               `json:"trigger_keywords"`
	InputSchema     map[string]interface{} `json:"input_schema"`
}

// ToolConfigs represents a collection of tool configurations
type ToolConfigs struct {
	Tools []ToolConfig `json:"tools"`
}

// LoadToolConfigs loads tool configurations from a JSON file
func LoadToolConfigs(configPath string) (*ToolConfigs, error) {
	// If no path is provided, use default path
	if configPath == "" {
		configPath = "tools.json"
	}

	// Read the config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read tool config file: %w", err)
	}

	var configs ToolConfigs
	if err := json.Unmarshal(data, &configs); err != nil {
		return nil, fmt.Errorf("failed to parse tool config file: %w", err)
	}

	return &configs, nil
}

// GenerateSystemPrompt generates a system prompt from tool configurations
func GenerateSystemPrompt(configs *ToolConfigs) string {
	// Group tools by category
	categories := make(map[string][]ToolConfig)
	for _, tool := range configs.Tools {
		categories[tool.Category] = append(categories[tool.Category], tool)
	}

	// Build the system prompt
	prompt := "You are a helpful, smart, and precise AI assistant with access to specific tools that let you interact with the user's system. You should use these tools whenever a request requires system access or specific information.\n\n"
	prompt += "THE TOOLS AVAILABLE TO YOU:\n\n"

	// Add tools by category
	categoryIndex := 1
	for category, tools := range categories {
		prompt += fmt.Sprintf("%d. %s Tools:\n", categoryIndex, category)
		categoryIndex++
		for _, tool := range tools {
			prompt += fmt.Sprintf("   - %s: %s\n", tool.Name, tool.Description)
		}
		prompt += "\n"
	}

	// Add detailed usage instructions with examples
	prompt += `WHEN TO USE TOOLS:
1. Use tools when the user asks for information that requires access to files or system commands
2. Use tools when asked to perform actions like reading, creating, or modifying files
3. Use tools when asked to execute commands or scripts
4. Use tools when asked to explore or navigate the filesystem

HOW TO USE TOOLS CORRECTLY:
1. FIRST: Tell the user what you plan to do (e.g., "I'll check what's in this directory for you")
2. THEN: Use the appropriate tool with the correct parameters
3. WAIT for the tool result to come back
4. FINALLY: Clearly explain the results to the user in a helpful way

CONCRETE EXAMPLES:

Example 1 - Reading a file:
User: "What's in the README.md file?"
Assistant: "I'll read the README.md file for you."
(Assistant uses read_file tool with {"path": "README.md"})
Assistant: "The README.md file contains... [summary of content]"

Example 2 - Listing directory contents:
User: "What files are in this project?"
Assistant: "I'll list the files in the current directory."
(Assistant uses list_files tool with {})
Assistant: "Here's what I found in the directory: [explanation of important files/folders]"

Example 3 - Running a command:
User: "Can you check the Go version?"
Assistant: "I'll check the Go version installed on your system."
(Assistant uses execute_command tool with {"command": "go", "args": ["version"]})
Assistant: "Your system is running Go version X.Y.Z"

Example 4 - Editing a file:
User: "Can you improve the introduction in README.md?"
Assistant: "I'll first read the current README.md to see its content."
(Assistant uses read_file tool with {"path": "README.md"})
Assistant: "I'll now update the introduction to make it more concise and informative."
(Assistant uses edit_file tool with appropriate parameters)
Assistant: "I've improved the introduction paragraph. It now reads: [new content]"

IMPORTANT: When using the execute_command tool, make sure to check that the command exists and is safe to run. If a command fails, suggest alternatives or troubleshooting steps.

IMPORTANT: Always use tools for actions that require system access. Don't just describe what to do - actually use the tools!

For simple questions or discussions that don't require system access, just respond naturally.`

	return prompt
} 