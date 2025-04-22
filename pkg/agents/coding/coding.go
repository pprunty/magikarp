package coding

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pprunty/magikarp/pkg/agents"
	"github.com/pprunty/magikarp/pkg/models"
)

// CodingAgent implements the Agent interface for coding tasks
type CodingAgent struct {
	*agents.BaseAgent
	conversation []models.Message
}

// NewCodingAgent creates a new coding agent
func NewCodingAgent(getUserMessage func() (string, bool)) *CodingAgent {
	baseAgent := agents.NewBaseAgent("coding", "Writes and refactors source code", getUserMessage)
	return &CodingAgent{
		BaseAgent:    baseAgent,
		conversation: []models.Message{},
	}
}

// Run starts the agent's execution
func (a *CodingAgent) Run(ctx context.Context) error {
	// Set up system message for coding tasks
	systemMsg := models.Message{
		Role:    "system",
		Content: "You are a coding assistant that helps with writing and refactoring code. Use tools to interact with the filesystem and execute commands.",
	}
	a.conversation = append(a.conversation, systemMsg)

	fmt.Printf("Chat with %s (use 'ctrl-c' to quit)\n", a.client.Name())
	fmt.Println("Tip: Type 'show tools' to toggle tool result visibility")

	for {
		fmt.Print("\u001b[94mYou\u001b[0m: ")
		userInput, ok := a.getUserMessage()
		if !ok {
			break
		}

		// Handle show tools toggle command
		if userInput == "show tools" {
			a.EnableToolResults(!a.showToolResults)
			if a.showToolResults {
				fmt.Println("\u001b[93mAI\u001b[0m: Tool results will now be shown")
			} else {
				fmt.Println("\u001b[93mAI\u001b[0m: Tool results will now be hidden")
			}
			continue
		}

		userMessage := models.Message{
			Role:    "user",
			Content: userInput,
		}
		a.conversation = append(a.conversation, userMessage)

		// Convert tools to LLM format
		llmTools := make([]models.Tool, len(a.tools))
		for i, tool := range a.tools {
			llmTools[i] = models.Tool{
				Name:        tool.Name,
				Description: tool.Description,
				InputSchema: tool.InputSchema,
			}
		}

		// Send the conversation to the LLM
		messages, toolUses, err := a.client.Chat(ctx, a.conversation, llmTools)
		if err != nil {
			return err
		}
		a.conversation = append(a.conversation, messages...)

		// Print AI's initial response
		for _, content := range messages {
			if content.Content != "" {
				fmt.Printf("\u001b[93mAI\u001b[0m: %s\n", content.Content)
			}
		}

		// Execute tools if needed
		toolResults := []models.ToolResult{}
		if len(toolUses) > 0 {
			for _, toolUse := range toolUses {
				// For potentially destructive commands, add a confirmation prompt
				needsConfirmation := false
				if toolUse.Name == "execute_command" {
					var cmdInput map[string]interface{}
					if err := json.Unmarshal(toolUse.Input, &cmdInput); err == nil {
						if cmd, ok := cmdInput["command"].(string); ok {
							if isDestructiveCommand(cmd) {
								needsConfirmation = true
								fmt.Printf("\u001b[91mCAUTION\u001b[0m: About to run potentially destructive command: %s\n", cmd)
								fmt.Print("Continue? (y/n): ")
								var confirm string
								fmt.Scanln(&confirm)
								if confirm != "y" && confirm != "Y" {
									toolResults = append(toolResults, models.ToolResult{
										ID:      toolUse.ID,
										Content: "User aborted the command execution.",
										IsError: true,
									})
									continue
								}
							}
						}
					}
				}

				// Execute the tool
				result := a.ExecuteTool(ctx, toolUse.ID, toolUse.Name, toolUse.Input)
				toolResults = append(toolResults, models.ToolResult{
					ID:      result.ID,
					Content: result.Content,
					IsError: result.IsError,
				})

				// Print tool results if enabled
				if a.showToolResults {
					if result.IsError {
						fmt.Printf("\u001b[91mtool result (ERROR)\u001b[0m: \n%s\n", result.Content)
					} else {
						fmt.Printf("\u001b[92mtool result\u001b[0m: \n%s\n", result.Content)
					}
				}
			}

			// Send tool results back to the LLM
			messages, _, err = a.client.SendToolResult(ctx, a.conversation, toolResults)
			if err != nil {
				return err
			}
			a.conversation = append(a.conversation, messages...)

			// Print AI's response after tool results
			for _, content := range messages {
				if content.Content != "" {
					fmt.Printf("\u001b[93mAI\u001b[0m: %s\n", content.Content)
				}
			}
		}
	}

	return nil
}

// isDestructiveCommand checks if the command is potentially destructive
func isDestructiveCommand(cmd string) bool {
	destructivePatterns := []string{
		"rm -rf", "rm -r", "rmdir",
		"dd", "mkfs",
		"format",
		"> /dev/",
		"truncate",
		"shred",
	}

	for _, pattern := range destructivePatterns {
		if containsPattern(cmd, pattern) {
			return true
		}
	}
	return false
}

// containsPattern is a helper function to check if a string contains a pattern
func containsPattern(s, pattern string) bool {
	for i := 0; i <= len(s)-len(pattern); i++ {
		if s[i:i+len(pattern)] == pattern {
			return true
		}
	}
	return false
} 