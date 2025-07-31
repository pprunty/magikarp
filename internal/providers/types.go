package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// Role constants for chat messages
const (
	RoleSystem    = "system"
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleTool      = "tool"
)

// ChatMessage represents a message in a conversation
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Tool represents a tool that can be used by the LLM
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

// ToolUse represents a tool use request from the LLM
type ToolUse struct {
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input"`
}

// ToolResult represents a tool result to be sent back to the LLM
type ToolResult struct {
	ID      string `json:"id"`
	Content string `json:"content"`
	IsError bool   `json:"is_error"`
}

// ToolDefinition represents a tool definition for the agent system
type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"`
	Function    func(ctx context.Context, input map[string]interface{}) (*ToolResult, error)
}

// NewToolResult creates a new tool result
func NewToolResult(toolName, content string, isError bool) *ToolResult {
	return &ToolResult{
		ID:      "", // Can be set by caller if needed
		Content: content,
		IsError: isError,
	}
}

// Provider is the interface that all LLM providers must implement
type Provider interface {
	// Name returns the name of the provider
	Name() string

	// Chat sends a message to the LLM and returns its response
	Chat(ctx context.Context, messages []ChatMessage, tools []Tool) ([]ChatMessage, []ToolUse, error)

	// StreamChat sends a message to the LLM and returns a streaming response
	StreamChat(ctx context.Context, model string, messages []ChatMessage, temperature float64) (<-chan string, error)

	// SendToolResult sends a tool result back to the LLM and returns its response
	SendToolResult(ctx context.Context, messages []ChatMessage, toolResults []ToolResult) ([]ChatMessage, []ToolUse, error)
}

// Legacy Message type for backward compatibility - will be removed
type Message = ChatMessage

// Legacy Client interface for backward compatibility - will be removed
type Client = Provider

// ChatAgent handles multi-turn conversations with tool support
type ChatAgent struct {
	client          Provider
	getUserMessage  func() (string, bool)
	tools           []ToolDefinition
	showToolResults bool
	systemPrompt    string
	temperature     float64
	conversation    []ChatMessage
}

// NewChatAgent creates a new chat agent
func NewChatAgent(client Provider, getUserMessage func() (string, bool), tools []ToolDefinition, systemPrompt string, temperature float64) *ChatAgent {
	return &ChatAgent{
		client:          client,
		getUserMessage:  getUserMessage,
		tools:           tools,
		showToolResults: false,
		systemPrompt:    systemPrompt,
		temperature:     temperature,
		conversation:    []ChatMessage{},
	}
}

// Run starts the chat session loop
func (a *ChatAgent) Run(ctx context.Context) error {
	fmt.Printf("Chat with %s (ctrl-C to quit)\n", a.client.Name())
	fmt.Println("Tip: type 'show tools' to toggle tool result visibility.")

	// Use readUserInput flag to control conversation flow
	readUserInput := true
	for {
		if readUserInput {
			fmt.Print("\u001b[94mYou\u001b[0m: ")
			userInput, ok := a.getUserMessage()
			if !ok {
				break
			}
			if strings.TrimSpace(strings.ToLower(userInput)) == "show tools" {
				a.showToolResults = !a.showToolResults
				if a.showToolResults {
					fmt.Println("\u001b[93mAI\u001b[0m: tool results will now be shown.")
				} else {
					fmt.Println("\u001b[93mAI\u001b[0m: tool results will now be hidden.")
				}
				continue
			}
			a.conversation = append(a.conversation, ChatMessage{Role: RoleUser, Content: userInput})
		}

		// Convert tools to provider format
		providerTools := make([]Tool, len(a.tools))
		for i, t := range a.tools {
			providerTools[i] = Tool{
				Name:        t.Name,
				Description: t.Description,
				InputSchema: t.InputSchema,
			}
		}

		// Build messages with system prompt
		messages := []ChatMessage{
			{Role: RoleSystem, Content: a.systemPrompt},
		}
		messages = append(messages, a.conversation...)

		// Get response from the LLM
		assistantMsgs, toolCalls, err := a.client.Chat(ctx, messages, providerTools)
		if err != nil {
			return err
		}

		// Add assistant's response to conversation
		a.conversation = append(a.conversation, assistantMsgs...)

		// Display assistant's response
		for _, m := range assistantMsgs {
			if m.Content != "" {
				fmt.Printf("\u001b[93mAI\u001b[0m: %s\n", m.Content)
			}
		}

		// If no tool calls, get new user input
		if len(toolCalls) == 0 {
			readUserInput = true
			continue
		}

		// Process tool calls and execute them
		var toolResults []ChatMessage
		for _, call := range toolCalls {
			// Execute the tool
			result := a.executeTool(call.ID, call.Name, call.Input)

			// Show tool result if enabled
			if a.showToolResults {
				color := "92"
				if result.IsError {
					color = "91"
				}
				fmt.Printf("\u001b[%smtool result\u001b[0m: %s\n", color, a.prettyJSON(result.Content))
			}

			// Create a tool result message
			toolResults = append(toolResults, ChatMessage{
				Role:    RoleTool,
				Content: result.Content,
			})
		}

		// Add all tool results to the conversation
		a.conversation = append(a.conversation, toolResults...)

		// Continue without user input
		readUserInput = false
	}

	return nil
}

// executeTool executes a tool call and returns the result
func (a *ChatAgent) executeTool(toolID, toolName string, input json.RawMessage) *ToolResult {
	// Find the tool in our tool definitions
	for _, tool := range a.tools {
		if tool.Name == toolName {
			if tool.Function != nil {
				// Parse input into a map
				var inputMap map[string]interface{}
				if err := json.Unmarshal(input, &inputMap); err != nil {
					return &ToolResult{
						ID:      toolID,
						Content: fmt.Sprintf("Invalid input for %s: %v", toolName, err),
						IsError: true,
					}
				}

				// Execute the tool function
				result, err := tool.Function(context.Background(), inputMap)
				if err != nil {
					return &ToolResult{
						ID:      toolID,
						Content: fmt.Sprintf("Tool execution error: %v", err),
						IsError: true,
					}
				}
				return result
			}
		}
	}

	return &ToolResult{
		ID:      toolID,
		Content: fmt.Sprintf("Unknown tool: %s", toolName),
		IsError: true,
	}
}

// prettyJSON formats JSON for display
func (a *ChatAgent) prettyJSON(content string) string {
	var obj interface{}
	if err := json.Unmarshal([]byte(content), &obj); err != nil {
		return content // Return as-is if not JSON
	}

	pretty, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return content
	}

	return string(pretty)
}
