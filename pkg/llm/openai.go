package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/sashabaranov/go-openai"
)

// OpenAIClient implements the Client interface for OpenAI
type OpenAIClient struct {
	client *openai.Client
	model  string
}

// NewOpenAIClient creates a new OpenAI client
func NewOpenAIClient(model string) (*OpenAIClient, error) {
	// Check if API key is set
	if os.Getenv("OPENAI_API_KEY") == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable is not set")
	}

	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	return &OpenAIClient{
		client: client,
		model:  model,
	}, nil
}

// Name returns the name of the LLM
func (c *OpenAIClient) Name() string {
	return c.model
}

// Chat sends a message to OpenAI and returns its response
func (c *OpenAIClient) Chat(ctx context.Context, messages []Message, tools []Tool) ([]Message, []ToolUse, error) {
	// Convert messages to OpenAI format
	openaiMessages := make([]openai.ChatCompletionMessage, len(messages))
	for i, msg := range messages {
		openaiMessages[i] = openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// Convert tools to OpenAI format
	var openaiTools []openai.Tool
	if len(tools) > 0 {
		openaiTools = make([]openai.Tool, len(tools))
		for i, tool := range tools {
			openaiTools[i] = openai.Tool{
				Type: "function",
				Function: openai.FunctionDefinition{
					Name:        tool.Name,
					Description: tool.Description,
					Parameters:  tool.InputSchema,
				},
			}
		}
	}

	// Create chat completion request
	req := openai.ChatCompletionRequest{
		Model:    c.model,
		Messages: openaiMessages,
		Tools:    openaiTools,
	}

	// Send request to OpenAI
	resp, err := c.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create chat completion: %w", err)
	}

	// Convert response to our format
	resultMessages := make([]Message, 0)
	var toolUses []ToolUse

	for _, choice := range resp.Choices {
		if choice.Message.Content != "" {
			resultMessages = append(resultMessages, Message{
				Role:    choice.Message.Role,
				Content: choice.Message.Content,
			})
		}

		// Handle tool calls
		for _, toolCall := range choice.Message.ToolCalls {
			if toolCall.Function.Name == "" {
				continue
			}

			toolUses = append(toolUses, ToolUse{
				ID:    toolCall.ID,
				Name:  toolCall.Function.Name,
				Input: json.RawMessage(toolCall.Function.Arguments),
			})
		}
	}

	return resultMessages, toolUses, nil
}

// SendToolResult sends a tool result back to OpenAI and returns its response
func (c *OpenAIClient) SendToolResult(ctx context.Context, messages []Message, toolResults []ToolResult) ([]Message, []ToolUse, error) {
	// Add tool results to messages
	for _, result := range toolResults {
		messages = append(messages, Message{
			Role:    "tool",
			Content: result.Content,
		})
	}

	// Continue the conversation with all tools available
	return c.Chat(ctx, messages, nil)
} 