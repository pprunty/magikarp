package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/pprunty/magikarp/pkg/agent"
	"github.com/pprunty/magikarp/pkg/config"
	"github.com/pprunty/magikarp/pkg/history"
	"github.com/pprunty/magikarp/pkg/llm"
	"github.com/pprunty/magikarp/pkg/plugins/execution"
	"github.com/pprunty/magikarp/pkg/plugins/filesystem"
	"github.com/pprunty/magikarp/pkg/plugins/text"
)

type Agent struct {
	client         llm.Client
	getUserMessage func() (string, bool)
	tools          []agent.ToolDefinition
	showToolResults bool
}

func NewAgent(client llm.Client, getUserMessage func() (string, bool), tools []agent.ToolDefinition) *Agent {
	return &Agent{
		client:         client,
		getUserMessage: getUserMessage,
		tools:          tools,
		showToolResults: false, // Default to not showing tool results
	}
}

func (a *Agent) Run(ctx context.Context) error {
	conversation := []llm.Message{}

	fmt.Printf("Chat with %s (use 'ctrl-c' to quit)\n", a.client.Name())
	fmt.Println("Tip: Type 'show tools' to toggle tool result visibility")

	readUserInput := true
	for {
		if readUserInput {
			fmt.Print("\u001b[94mYou\u001b[0m: ")
			userInput, ok := a.getUserMessage()
			if !ok {
				break
			}

			// Handle show tools toggle command
			if strings.TrimSpace(strings.ToLower(userInput)) == "show tools" {
				a.showToolResults = !a.showToolResults
				if a.showToolResults {
					fmt.Println("\u001b[93mAI\u001b[0m: Tool results will now be shown")
				} else {
					fmt.Println("\u001b[93mAI\u001b[0m: Tool results will now be hidden")
				}
				continue
			}

			userMessage := llm.Message{
				Role:    "user",
				Content: userInput,
			}
			conversation = append(conversation, userMessage)
		}

		// Convert tools to LLM format
		llmTools := make([]llm.Tool, len(a.tools))
		for i, tool := range a.tools {
			llmTools[i] = llm.Tool{
				Name:        tool.Name,
				Description: tool.Description,
				InputSchema: tool.InputSchema,
			}
		}

		messages, toolUses, err := a.client.Chat(ctx, conversation, llmTools)
		if err != nil {
			return err
		}
		conversation = append(conversation, messages...)

		// Print AI's initial response
		for _, content := range messages {
			if content.Content != "" {
				fmt.Printf("\u001b[93mAI\u001b[0m: %s\n", content.Content)
			}
		}

		// Initialize toolResults outside the if block
		toolResults := []llm.ToolResult{}
		
		// Print tool results only if showToolResults is true
		if a.showToolResults {
			firstToolUsed := false
			for _, toolUse := range toolUses {
				if firstToolUsed {
					fmt.Println("\u001b[36m--------------------\u001b[0m")
				}
				firstToolUsed = true
				
				// Verify if this is an execute_command tool and display caution for destructive commands
				if toolUse.Name == "execute_command" {
					var cmdInput struct {
						Command string   `json:"command"`
						Args    []string `json:"args,omitempty"`
					}
					if err := json.Unmarshal(toolUse.Input, &cmdInput); err == nil {
						cmd := cmdInput.Command
						if hasDestructivePattern(cmd) {
							fmt.Printf("\u001b[91mCAUTION\u001b[0m: About to run potentially destructive command: %s\n", cmd)
							fmt.Print("Continue? (y/n): ")
							var confirm string
							fmt.Scanln(&confirm)
							if confirm != "y" && confirm != "Y" {
								toolResults = append(toolResults, llm.ToolResult{
									ID:      toolUse.ID,
									Content: "User aborted the command execution.",
									IsError: true,
								})
								continue
							}
						}
					}
				}
				
				// Execute the tool and get the result
				result := a.executeTool(toolUse.ID, toolUse.Name, toolUse.Input)
				toolResults = append(toolResults, result)
				
				// Print the tool result to the user with pretty formatting
				formattedResult := prettyPrintJSON(result.Content)
				if result.IsError {
					fmt.Printf("\u001b[91mtool result (ERROR)\u001b[0m: \n%s\n", formattedResult)
				} else {
					fmt.Printf("\u001b[92mtool result\u001b[0m: \n%s\n", formattedResult)
				}
			}
		} else {
			// Execute tools without showing results
			for _, toolUse := range toolUses {
				if toolUse.Name == "execute_command" {
					var cmdInput struct {
						Command string   `json:"command"`
						Args    []string `json:"args,omitempty"`
					}
					if err := json.Unmarshal(toolUse.Input, &cmdInput); err == nil {
						cmd := cmdInput.Command
						if hasDestructivePattern(cmd) {
							fmt.Printf("\u001b[91mCAUTION\u001b[0m: About to run potentially destructive command: %s\n", cmd)
							fmt.Print("Continue? (y/n): ")
							var confirm string
							fmt.Scanln(&confirm)
							if confirm != "y" && confirm != "Y" {
								toolResults = append(toolResults, llm.ToolResult{
									ID:      toolUse.ID,
									Content: "User aborted the command execution.",
									IsError: true,
								})
								continue
							}
						}
					}
				}
				result := a.executeTool(toolUse.ID, toolUse.Name, toolUse.Input)
				toolResults = append(toolResults, result)
			}
		}

		if len(toolResults) == 0 {
			readUserInput = true
			continue
		}

		readUserInput = false
		messages, _, err = a.client.SendToolResult(ctx, conversation, toolResults)
		if err != nil {
			return err
		}
		conversation = append(conversation, messages...)

		// Print AI's response after tool results
		for _, content := range messages {
			if content.Content != "" {
				fmt.Printf("\u001b[93mAI\u001b[0m: %s\n", content.Content)
			}
		}
	}

	return nil
}

func (a *Agent) executeTool(id, name string, input json.RawMessage) llm.ToolResult {
	var toolDef agent.ToolDefinition
	var found bool
	for _, tool := range a.tools {
		if tool.Name == name {
			toolDef = tool
			found = true
			break
		}
	}
	if !found {
		return llm.ToolResult{
			ID:      id,
			Content: "tool not found",
			IsError: true,
		}
	}

	fmt.Printf("\u001b[92mtool\u001b[0m: %s(%s)\n", name, input)
	response, err := toolDef.Function(input)
	if err != nil {
		return llm.ToolResult{
			ID:      id,
			Content: err.Error(),
			IsError: true,
		}
	}
	return llm.ToolResult{
		ID:      id,
		Content: response,
		IsError: false,
	}
}

// Checks if a command contains potentially destructive patterns
func hasDestructivePattern(cmd string) bool {
	destructivePatterns := []string{
		"rm -rf", "rm -r", "rmdir", 
		"dd", "mkfs", 
		"format", 
		"> /dev/", 
		"truncate",
		"shred",
	}
	
	cmd = strings.ToLower(cmd)
	for _, pattern := range destructivePatterns {
		if strings.Contains(cmd, pattern) {
			return true
		}
	}
	return false
}

// Pretty prints JSON content if the content is valid JSON
func prettyPrintJSON(content string) string {
	// Try to unmarshal as JSON
	var data interface{}
	if err := json.Unmarshal([]byte(content), &data); err != nil {
		// Not JSON, return as is
		return content
	}
	
	// Format the JSON with indentation
	pretty, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		// If formatting fails, return original
		return content
	}
	
	return string(pretty)
}

// ProviderOption represents a selectable provider option
type ProviderOption struct {
	ID       string
	Name     string
	Required bool
}

// ModelOption represents a selectable model option
type ModelOption struct {
	Name        string
	Description string
}

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Create provider options
	var providerOptions []ProviderOption
	
	// Add auto first
	if autoProvider, ok := cfg.Providers["auto"]; ok {
		providerOptions = append(providerOptions, ProviderOption{
			ID:       "auto",
			Name:     autoProvider.Name,
			Required: autoProvider.Required,
		})
	}
	
	// Add other providers in order (excluding auto and ollama)
	providerOrder := []string{"anthropic", "openai", "gemini"}
	for _, id := range providerOrder {
		if provider, ok := cfg.Providers[id]; ok {
			providerOptions = append(providerOptions, ProviderOption{
				ID:       id,
				Name:     provider.Name,
				Required: provider.Required,
			})
		}
	}
	
	// Add ollama last
	if ollamaProvider, ok := cfg.Providers["ollama"]; ok {
		providerOptions = append(providerOptions, ProviderOption{
			ID:       "ollama",
			Name:     ollamaProvider.Name,
			Required: ollamaProvider.Required,
		})
	}

	// Create provider selection prompt
	providerPrompt := promptui.Select{
		Label: "Select LLM Provider (use backspace to go back)",
		Items: providerOptions,
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . }}",
			Active:   "→ {{ .Name | cyan }}",
			Inactive: "  {{ .Name | white }}",
			Selected: "✓ {{ .Name | green }}",
			Details: `
--------- Provider Details ----------
{{ "Name:" | faint }}	{{ .Name }}
{{ "Required:" | faint }}	{{ if .Required }}Yes{{ else }}No{{ end }}`,
		},
		Keys: &promptui.SelectKeys{
			Prev:     promptui.Key{Code: promptui.KeyPrev, Display: "↑"},
			Next:     promptui.Key{Code: promptui.KeyNext, Display: "↓"},
			PageUp:   promptui.Key{Code: promptui.KeyBackspace, Display: "←"},
			PageDown: promptui.Key{Code: promptui.KeyNext, Display: "→"},
			Search:   promptui.Key{Code: '/', Display: "/"},
		},
	}

	// Run provider selection
	providerIndex, _, err := providerPrompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}

	selectedProvider := providerOptions[providerIndex]
	provider, err := cfg.GetProvider(selectedProvider.ID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting provider: %v\n", err)
		os.Exit(1)
	}

	// Skip model selection for auto provider
	var client llm.Client
	var result any

	if selectedProvider.ID == "auto" {
		// Get all available models from all providers
		var allModels []string
		for _, provider := range cfg.Providers {
			for _, model := range provider.Models {
				allModels = append(allModels, model.Name)
			}
		}
		result, err = llm.NewAutoClient(allModels, "tools.json")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating auto client: %v\n", err)
			os.Exit(1)
		}
		client = result.(llm.Client)
	} else {
		// Create model options
		var modelOptions []ModelOption
		for _, model := range provider.Models {
			modelOptions = append(modelOptions, ModelOption{
				Name:        model.Name,
				Description: model.Description,
			})
		}

		// Create model selection prompt
		modelPrompt := promptui.Select{
			Label: fmt.Sprintf("Select %s Model (use backspace to go back)", provider.Name),
			Items: modelOptions,
			Templates: &promptui.SelectTemplates{
				Label:    "{{ . }}",
				Active:   "→ {{ .Name | cyan }} ({{ .Description | yellow }})",
				Inactive: "  {{ .Name | white }} ({{ .Description | yellow }})",
				Selected: "✓ {{ .Name | green }} ({{ .Description | yellow }})",
			},
			Keys: &promptui.SelectKeys{
				Prev:     promptui.Key{Code: promptui.KeyPrev, Display: "↑"},
				Next:     promptui.Key{Code: promptui.KeyNext, Display: "↓"},
				PageUp:   promptui.Key{Code: promptui.KeyBackspace, Display: "←"},
				PageDown: promptui.Key{Code: promptui.KeyNext, Display: "→"},
				Search:   promptui.Key{Code: '/', Display: "/"},
			},
		}

		// Run model selection
		modelIndex, _, err := modelPrompt.Run()
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}

		selectedModel := modelOptions[modelIndex]
		fmt.Printf("You selected %s (%s)\n", provider.Name, selectedModel.Name)

		// Create the selected LLM client
		switch selectedProvider.ID {
		case "anthropic":
			result, err = llm.NewAnthropicClient(selectedModel.Name, "tools.json")
		case "openai":
			result, err = llm.NewOpenAIClient(selectedModel.Name, "tools.json")
		case "gemini":
			result, err = llm.NewGeminiClient(selectedModel.Name, "tools.json")
		case "ollama":
			result, err = llm.NewOllamaClient(selectedModel.Name, "tools.json")
		default:
			fmt.Fprintf(os.Stderr, "Unknown provider: %s\n", selectedProvider.ID)
			os.Exit(1)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating %s client: %v\n", provider.Name, err)
			os.Exit(1)
		}

		// Type assert the result to llm.Client
		client = result.(llm.Client)
	}

	// Setup history manager
	hist, err := history.NewHistoryManager()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error setting up command history: %v\n", err)
		os.Exit(1)
	}
	defer hist.Close()

	getUserMessage := func() (string, bool) {
		line, err := hist.ReadLine()
		if err != nil {
			return "", false
		}
		return strings.TrimSpace(line), true
	}

	plugins := []agent.Plugin{
		execution.New(),
		filesystem.New(),
		text.New(),
	}

	var tools []agent.ToolDefinition
	for _, plugin := range plugins {
		tools = append(tools, plugin.Tools()...)
	}

	agent := NewAgent(client, getUserMessage, tools)
	if err := agent.Run(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
} 