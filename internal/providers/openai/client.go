package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/pprunty/magikarp/internal/providers"
	"github.com/sashabaranov/go-openai"
)

// Enable debug logs for OpenAI provider if MAGIKARP_DEBUG=1
var openaiDebug = os.Getenv("MAGIKARP_DEBUG") == "1"
var debugFile *os.File

func init() {
	// Always try to create debug file to test if this init runs
	var err error
	debugFile, err = os.OpenFile("magikarp_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		// Silently continue without debug logging if file can't be opened
		openaiDebug = false
		return
	}

	// Write a startup message
	timestamp := time.Now().Format("2006/01/02 15:04:05")
	fmt.Fprintf(debugFile, "%s [OpenAI] Init: MAGIKARP_DEBUG=%s, debug enabled=%t\n", timestamp, os.Getenv("MAGIKARP_DEBUG"), openaiDebug)
	debugFile.Sync()
}

func debugLog(format string, args ...interface{}) {
	if openaiDebug && debugFile != nil {
		timestamp := time.Now().Format("2006/01/02 15:04:05")
		fmt.Fprintf(debugFile, "%s [OpenAI] "+format+"\n", append([]interface{}{timestamp}, args...)...)
		debugFile.Sync() // Flush immediately so we can tail -f
	}
}

// OpenAIClient implements the Provider interface for OpenAI
type OpenAIClient struct {
	client       *openai.Client
	apiKey       string
	models       []string
	temperature  float64
	systemPrompt string
}

// New creates a new OpenAI provider
func New(apiKey string, models []string, temperature float64, systemPrompt string) *OpenAIClient {
	client := openai.NewClient(apiKey)
	return &OpenAIClient{
		client:       client,
		apiKey:       apiKey,
		models:       models,
		temperature:  temperature,
		systemPrompt: systemPrompt,
	}
}

// NewOpenAIClient creates a new OpenAI client (legacy)
func NewOpenAIClient(model string, configPath string) (*OpenAIClient, error) {
	// Check if API key is set
	if os.Getenv("OPENAI_API_KEY") == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable is not set")
	}

	return New(os.Getenv("OPENAI_API_KEY"), []string{model}, 0.0, ""), nil
}

// Name returns the name of the provider
func (c *OpenAIClient) Name() string {
	return "openai"
}

// Chat sends a message to OpenAI and returns its response
func (c *OpenAIClient) Chat(ctx context.Context, messages []providers.ChatMessage, tools []providers.Tool) ([]providers.ChatMessage, []providers.ToolUse, error) {
	debugLog("Chat call: model list=%v, user/assistant messages=%d, tools=%d", c.models, len(messages), len(tools))
	
	if len(c.models) == 0 {
		return nil, nil, fmt.Errorf("openai client has no model configured")
	}
	
	// Convert messages to OpenAI format
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
		Model:    model,
		Messages: openaiMessages,
		Tools:    openaiTools,
	}

	// Only set temperature for non-o* models (o1, o3 series have fixed parameters)
	if !isOSeriesModel(model) {
		req.Temperature = float32(c.temperature)
	}

	// Send request to OpenAI
	resp, err := c.client.CreateChatCompletion(ctx, req)
	if err != nil {
		debugLog("Chat error: %v", err)
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
	debugLog("StreamChat: model=%s, temperature=%f, total_messages=%d", model, temperature, len(messages))
	
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
		Model:    model,
		Messages: openaiMessages,
		Stream:   true,
	}

	// Only set temperature for non-o* models (o1, o3 series have fixed parameters)
	if !isOSeriesModel(model) {
		req.Temperature = float32(temperature)
	}

	// Create stream
	stream, err := c.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat completion stream: %w", err)
	}
	
	debugLog("StreamChat: stream created, waiting for events")

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
				debugLog("StreamChat: stream error: %v", err)
				responseChan <- fmt.Sprintf("Error: %v", err)
				return
			}

			if len(response.Choices) > 0 {
				delta := response.Choices[0].Delta
				if delta.Content != "" {
					debugLog("StreamChat: received content delta")
					responseChan <- delta.Content
				}
			}
		}
	}()

	return responseChan, nil
}

// SendToolResult sends a tool result back to OpenAI and returns its response
func (c *OpenAIClient) SendToolResult(ctx context.Context, messages []providers.ChatMessage, toolResults []providers.ToolResult) ([]providers.ChatMessage, []providers.ToolUse, error) {
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

// isOSeriesModel checks if the model is from the o-series (o1, o3) which have fixed parameters
func isOSeriesModel(model string) bool {
	model = strings.ToLower(model)
	return strings.HasPrefix(model, "o1") || strings.HasPrefix(model, "o3")
}
