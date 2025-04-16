package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/pprunty/magikarp/pkg/agent"
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

func main() {
	// Try to create Anthropic client first
	var client llm.Client
	var err error
	
	client, err = llm.NewAnthropicClient("tools.json")
	if err != nil {
		fmt.Printf("Note: %v\n", err)
		fmt.Println("Falling back to Ollama with llama3.2 model...")
		fmt.Println("Make sure Ollama server is running with 'make ollama' in a separate terminal")
		// Create Ollama client with model as fallback
		client, err = llm.NewOllamaClient("llama3.2", "tools.json")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating Ollama client: %v\n", err)
			client, err = llm.NewOllamaClient("llama3.2", "tools.json")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error creating fallback Ollama client: %v\n", err)
				os.Exit(1)
			}
		}
	}

	scanner := bufio.NewScanner(os.Stdin)
	getUserMessage := func() (string, bool) {
		if !scanner.Scan() {
			return "", false
		}
		return scanner.Text(), true
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