package alibaba

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/pprunty/magikarp/internal/providers"
	"github.com/sashabaranov/go-openai"
)

// AlibabaClient implements the Provider interface for Alibaba Qwen using OpenAI-compatible API
type AlibabaClient struct {
	client       *openai.Client
	apiKey       string
	models       []string
	temperature  float64
	systemPrompt string
}

// New creates a new Alibaba provider
func New(apiKey string, models []string, temperature float64, systemPrompt string) (*AlibabaClient, error) {
	config := openai.DefaultConfig(apiKey)
	// Use Alibaba's OpenAI-compatible endpoint
	config.BaseURL = "https://dashscope-intl.aliyuncs.com/compatible-mode/v1"
	client := openai.NewClientWithConfig(config)
	
	return &AlibabaClient{
		client:       client,
		apiKey:       apiKey,
		models:       models,
		temperature:  temperature,
		systemPrompt: systemPrompt,
	}, nil
}

// NewAlibabaClient creates a new Alibaba client (legacy)
func NewAlibabaClient(model string, configPath string) (*AlibabaClient, error) {
	// Check if API key is set
	if os.Getenv("ALIBABA_API_KEY") == "" {
		return nil, fmt.Errorf("ALIBABA_API_KEY environment variable is not set")
	}

	client, err := New(os.Getenv("ALIBABA_API_KEY"), []string{model}, 0.0, "")
	return client, err
}

// Name returns the name of the provider
func (c *AlibabaClient) Name() string {
	return "alibaba"
}

// Chat sends a message to Alibaba Qwen and returns its response
func (c *AlibabaClient) Chat(ctx context.Context, messages []providers.ChatMessage, tools []providers.Tool) ([]providers.ChatMessage, []providers.ToolUse, error) {
	if len(c.models) == 0 {
		return nil, nil, fmt.Errorf("alibaba client has no model configured")
	}
	
	// Convert messages to OpenAI format (since we're using OpenAI-compatible API)
	openaiMessages := make([]openai.ChatCompletionMessage, 0)
	
	// Add system prompt if configured
	systemPrompt := c.systemPrompt
	for _, msg := range messages {
		if msg.Role == providers.RoleSystem {
			// Use system message from conversation if provided, otherwise use config
			if msg.Content != "" {
				systemPrompt = msg.Content
			}
			continue
		} else if msg.Role == providers.RoleUser {
			openaiMessages = append(openaiMessages, openai.ChatCompletionMessage{
				Role:    "user",
				Content: msg.Content,
			})
		} else if msg.Role == providers.RoleAssistant {
			openaiMessages = append(openaiMessages, openai.ChatCompletionMessage{
				Role:    "assistant",
				Content: msg.Content,
			})
		} else if msg.Role == providers.RoleTool {
			openaiMessages = append(openaiMessages, openai.ChatCompletionMessage{
				Role:    "user",
				Content: msg.Content,
			})
		}
	}
	
	// Add system message at the beginning if we have one
	if systemPrompt != "" {
		systemMsg := openai.ChatCompletionMessage{
			Role:    "system",
			Content: systemPrompt,
		}
		openaiMessages = append([]openai.ChatCompletionMessage{systemMsg}, openaiMessages...)
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
	model := c.models[0]

	// Create chat completion request
	req := openai.ChatCompletionRequest{
		Model:       model,
		Messages:    openaiMessages,
		Tools:       openaiTools,
		Temperature: float32(c.temperature),
	}

	// Send request to Alibaba Qwen via OpenAI-compatible API
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

// StreamChat sends a message to Alibaba Qwen and returns a streaming response
func (c *AlibabaClient) StreamChat(ctx context.Context, model string, messages []providers.ChatMessage, temperature float64) (<-chan string, error) {
	// Convert messages to OpenAI format
	openaiMessages := make([]openai.ChatCompletionMessage, 0)
	systemPrompt := c.systemPrompt
	
	for _, msg := range messages {
		if msg.Role == providers.RoleSystem {
			// Use system message from conversation if provided, otherwise use config
			if msg.Content != "" {
				systemPrompt = msg.Content
			}
			continue
		} else if msg.Role == providers.RoleUser {
			openaiMessages = append(openaiMessages, openai.ChatCompletionMessage{
				Role:    "user",
				Content: msg.Content,
			})
		} else if msg.Role == providers.RoleAssistant {
			openaiMessages = append(openaiMessages, openai.ChatCompletionMessage{
				Role:    "assistant",
				Content: msg.Content,
			})
		}
	}
	
	// Add system message at the beginning if we have one
	if systemPrompt != "" {
		systemMsg := openai.ChatCompletionMessage{
			Role:    "system",
			Content: systemPrompt,
		}
		openaiMessages = append([]openai.ChatCompletionMessage{systemMsg}, openaiMessages...)
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
				if err.Error() == "EOF" {
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

// SendToolResult sends a tool result back to Alibaba Qwen and returns its response
func (c *AlibabaClient) SendToolResult(ctx context.Context, messages []providers.ChatMessage, toolResults []providers.ToolResult) ([]providers.ChatMessage, []providers.ToolUse, error) {
	// Append each tool result as a ChatMessage with RoleTool so Chat() can convert.
	augmented := make([]providers.ChatMessage, len(messages))
	copy(augmented, messages)

	for _, res := range toolResults {
		augmented = append(augmented, providers.ChatMessage{
			Role:    providers.RoleTool,
			Content: res.Content,
		})
	}

	// Continue conversation without re-sending tool definitions (nil tools).
	return c.Chat(ctx, augmented, nil)
}