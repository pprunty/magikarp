package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/pprunty/magikarp/internal/providers"
	"github.com/sashabaranov/go-openai"
)

// OpenAIClient implements the Provider interface for OpenAI
type OpenAIClient struct {
	client *openai.Client
	apiKey string
	models []string
}

// New creates a new OpenAI provider
func New(apiKey string, models []string) *OpenAIClient {
	client := openai.NewClient(apiKey)
	return &OpenAIClient{
		client: client,
		apiKey: apiKey,
		models: models,
	}
}

// NewOpenAIClient creates a new OpenAI client (legacy)
func NewOpenAIClient(model string, configPath string) (*OpenAIClient, error) {
	// Check if API key is set
	if os.Getenv("OPENAI_API_KEY") == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable is not set")
	}

	return New(os.Getenv("OPENAI_API_KEY"), []string{model}), nil
}

// Name returns the name of the provider
func (c *OpenAIClient) Name() string {
	return "openai"
}

// Chat sends a message to OpenAI and returns its response
func (c *OpenAIClient) Chat(ctx context.Context, messages []providers.ChatMessage, tools []providers.Tool) ([]providers.ChatMessage, []providers.ToolUse, error) {
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
				Function: &openai.FunctionDefinition{
					Name:        tool.Name,
					Description: tool.Description,
					Parameters:  tool.InputSchema,
				},
			}
		}
	}

	// Use first available model
	model := "gpt-4o-mini"
	if len(c.models) > 0 {
		model = c.models[0]
	}

	// Create chat completion request
	req := openai.ChatCompletionRequest{
		Model:    model,
		Messages: openaiMessages,
		Tools:    openaiTools,
	}

	// Send request to OpenAI
	resp, err := c.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create chat completion: %w", err)
	}

	// Convert response to our format
	resultMessages := make([]providers.ChatMessage, 0)
	var toolUses []providers.ToolUse

	for _, choice := range resp.Choices {
		if choice.Message.Content != "" {
			resultMessages = append(resultMessages, providers.ChatMessage{
				Role:    choice.Message.Role,
				Content: choice.Message.Content,
			})
		}

		// Handle tool calls
		for _, toolCall := range choice.Message.ToolCalls {
			if toolCall.Function.Name == "" {
				continue
			}

			toolUses = append(toolUses, providers.ToolUse{
				ID:    toolCall.ID,
				Name:  toolCall.Function.Name,
				Input: json.RawMessage(toolCall.Function.Arguments),
			})
		}
	}

	return resultMessages, toolUses, nil
}

// StreamChat sends a message to OpenAI and returns a streaming response
func (c *OpenAIClient) StreamChat(ctx context.Context, model string, messages []providers.ChatMessage, temperature float64) (<-chan string, error) {
	// Convert messages to OpenAI format
	openaiMessages := make([]openai.ChatCompletionMessage, len(messages))
	for i, msg := range messages {
		openaiMessages[i] = openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// Create streaming chat completion request
	req := openai.ChatCompletionRequest{
		Model:       model,
		Messages:    openaiMessages,
		Temperature: float32(temperature),
		Stream:      true,
	}

	// Create stream
	stream, err := c.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat completion stream: %w", err)
	}

	// Create channel for streaming response
	responseChan := make(chan string, 100)

	go func() {
		defer close(responseChan)
		defer stream.Close()

		for {
			response, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					return
				}
				responseChan <- fmt.Sprintf("Error: %v", err)
				return
			}

			if len(response.Choices) > 0 {
				delta := response.Choices[0].Delta
				if delta.Content != "" {
					responseChan <- delta.Content
				}
			}
		}
	}()

	return responseChan, nil
}

// SendToolResult sends a tool result back to OpenAI and returns its response
func (c *OpenAIClient) SendToolResult(ctx context.Context, messages []providers.ChatMessage, toolResults []providers.ToolResult) ([]providers.ChatMessage, []providers.ToolUse, error) {
	// Add tool results to messages
	for _, result := range toolResults {
		messages = append(messages, providers.ChatMessage{
			Role:    providers.RoleTool,
			Content: result.Content,
		})
	}

	// Continue the conversation with all tools available
	return c.Chat(ctx, messages, nil)
}
