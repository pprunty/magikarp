package llm

import (
	"context"
	"fmt"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
)

// AnthropicClient implements the Client interface for Anthropic
type AnthropicClient struct {
	client *anthropic.Client
}

// NewAnthropicClient creates a new Anthropic client
func NewAnthropicClient(configPath string) (*AnthropicClient, error) {
	// Check if API key is set
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY environment variable is not set")
	}

	client := anthropic.NewClient()
	return &AnthropicClient{
		client: &client,
	}, nil
}

// Chat sends a message to Anthropic and returns its response
func (c *AnthropicClient) Chat(ctx context.Context, messages []Message, tools []Tool) ([]Message, []ToolUse, error) {
	// Convert messages to Anthropic format
	anthropicMessages := make([]anthropic.MessageParam, len(messages))
	for i, msg := range messages {
		if msg.Role == "user" {
			anthropicMessages[i] = anthropic.NewUserMessage(anthropic.NewTextBlock(msg.Content))
		} else if msg.Role == "assistant" {
			anthropicMessages[i] = anthropic.NewAssistantMessage(anthropic.NewTextBlock(msg.Content))
		} else if msg.Role == "tool" {
			anthropicMessages[i] = anthropic.NewUserMessage(anthropic.NewTextBlock(msg.Content))
		}
	}

	// Convert tools to Anthropic format
	anthropicTools := make([]anthropic.ToolUnionParam, len(tools))
	for i, tool := range tools {
		schema := anthropic.ToolInputSchemaParam{
			Properties: tool.InputSchema,
		}
		anthropicTools[i] = anthropic.ToolUnionParam{
			OfTool: &anthropic.ToolParam{
				Name:        tool.Name,
				Description: anthropic.String(tool.Description),
				InputSchema: schema,
			},
		}
	}

	// Send request to Anthropic
	message, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaude3_7SonnetLatest,
		MaxTokens: int64(1024),
		Messages:  anthropicMessages,
		Tools:     anthropicTools,
	})
	if err != nil {
		return nil, nil, err
	}

	// Convert response to our format
	resultMessages := make([]Message, 0)
	var toolUses []ToolUse

	for _, content := range message.Content {
		switch content.Type {
		case "text":
			resultMessages = append(resultMessages, Message{
				Role:    "assistant",
				Content: content.Text,
			})
		case "tool_use":
			toolUses = append(toolUses, ToolUse{
				ID:    content.ID,
				Name:  content.Name,
				Input: content.Input,
			})
		}
	}

	return resultMessages, toolUses, nil
}

// SendToolResult sends a tool result back to Anthropic and returns its response
func (c *AnthropicClient) SendToolResult(ctx context.Context, messages []Message, toolResults []ToolResult) ([]Message, []ToolUse, error) {
	// Convert messages to Anthropic format
	anthropicMessages := make([]anthropic.MessageParam, len(messages))
	for i, msg := range messages {
		if msg.Role == "user" {
			anthropicMessages[i] = anthropic.NewUserMessage(anthropic.NewTextBlock(msg.Content))
		} else if msg.Role == "assistant" {
			anthropicMessages[i] = anthropic.NewAssistantMessage(anthropic.NewTextBlock(msg.Content))
		} else if msg.Role == "tool" {
			anthropicMessages[i] = anthropic.NewUserMessage(anthropic.NewTextBlock(msg.Content))
		}
	}

	// Convert tool results to Anthropic format
	toolResultBlocks := make([]anthropic.ContentBlockParamUnion, len(toolResults))
	for i, result := range toolResults {
		toolResultBlocks[i] = anthropic.NewToolResultBlock(result.ID, result.Content, result.IsError)
	}

	// Add tool results to messages
	anthropicMessages = append(anthropicMessages, anthropic.NewUserMessage(toolResultBlocks...))

	// Continue the conversation
	return c.Chat(ctx, messages, nil)
}

// Name returns the name of the LLM
func (c *AnthropicClient) Name() string {
	return "Claude"
}