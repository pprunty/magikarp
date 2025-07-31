package anthropic

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/pprunty/magikarp/internal/providers"
)

// Enable debug logs for Anthropic provider if MAGIKARP_DEBUG=1
var anthropicDebug = os.Getenv("MAGIKARP_DEBUG") == "1"
var debugFile *os.File

func init() {
	// Always try to create debug file to test if this init runs
	var err error
	debugFile, err = os.OpenFile("magikarp_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		// Silently continue without debug logging if file can't be opened
		anthropicDebug = false
		return
	}

	// Write a startup message
	timestamp := time.Now().Format("2006/01/02 15:04:05")
	fmt.Fprintf(debugFile, "%s [Anthropic] Init: MAGIKARP_DEBUG=%s, debug enabled=%t\n", timestamp, os.Getenv("MAGIKARP_DEBUG"), anthropicDebug)
	debugFile.Sync()
}

func debugLog(format string, args ...interface{}) {
	if anthropicDebug && debugFile != nil {
		timestamp := time.Now().Format("2006/01/02 15:04:05")
		fmt.Fprintf(debugFile, "%s [Anthropic] "+format+"\n", append([]interface{}{timestamp}, args...)...)
		debugFile.Sync() // Flush immediately so we can tail -f
	}
}

// AnthropicClient implements the Provider interface for Anthropic
type AnthropicClient struct {
	client *anthropic.Client
	apiKey string
	models []string
}

// New creates a new Anthropic provider
func New(apiKey string, models []string) *AnthropicClient {
	client := anthropic.NewClient(option.WithAPIKey(apiKey))
	return &AnthropicClient{
		client: &client,
		apiKey: apiKey,
		models: models,
	}
}

// NewAnthropicClient creates a new Anthropic client (legacy)
func NewAnthropicClient(model string, configPath string) (*AnthropicClient, error) {
	// Check if API key is set
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY environment variable is not set")
	}

	return New(os.Getenv("ANTHROPIC_API_KEY"), []string{model}), nil
}

// Name returns the name of the provider
func (c *AnthropicClient) Name() string {
	return "anthropic"
}

// Chat sends a message to Anthropic and returns its response
func (c *AnthropicClient) Chat(ctx context.Context, messages []providers.ChatMessage, tools []providers.Tool) ([]providers.ChatMessage, []providers.ToolUse, error) {
	debugLog("Chat call: model list=%v, user/assistant messages=%d, tools=%d", c.models, len(messages), len(tools))
	// Convert messages to Anthropic format
	anthropicMessages := make([]anthropic.MessageParam, 0)

	systemPrompt := ""
	for _, msg := range messages {
		if msg.Role == providers.RoleSystem {
			// Skip system messages for now - Anthropic handles them differently
			continue
		} else if msg.Role == providers.RoleUser {
			anthropicMessages = append(anthropicMessages, anthropic.NewUserMessage(anthropic.NewTextBlock(msg.Content)))
		} else if msg.Role == providers.RoleAssistant {
			anthropicMessages = append(anthropicMessages, anthropic.NewAssistantMessage(anthropic.NewTextBlock(msg.Content)))
		} else if msg.Role == providers.RoleTool {
			anthropicMessages = append(anthropicMessages, anthropic.NewUserMessage(anthropic.NewTextBlock(msg.Content)))
		}
	}

	// Convert tools to Anthropic format
	anthropicTools := make([]anthropic.ToolUnionParam, len(tools))
	for i, tool := range tools {
		// tool.InputSchema is full schema map; extract standard fields
		props := map[string]any{}
		if p, ok := tool.InputSchema["properties"].(map[string]any); ok {
			props = p
		}
		req := toStringSlice(tool.InputSchema["required"])

		schema := anthropic.ToolInputSchemaParam{
			Type:       "object", // plain string
			Properties: props,
			Required:   req,
		}
		anthropicTools[i] = anthropic.ToolUnionParam{
			OfTool: &anthropic.ToolParam{
				Name:        tool.Name,
				Description: anthropic.String(tool.Description),
				InputSchema: schema,
			},
		}
	}

	if len(c.models) == 0 {
		return nil, nil, fmt.Errorf("anthropic client has no model configured")
	}
	model := c.models[0]

	// Prepare system prompt parameter
	var systemBlocks []anthropic.TextBlockParam
	if systemPrompt != "" {
		systemBlocks = []anthropic.TextBlockParam{{Type: "text", Text: systemPrompt}}
	}

	// Send request to Anthropic
	message, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(model),
		MaxTokens: 1024,
		Messages:  anthropicMessages,
		Tools:     anthropicTools,
		System:    systemBlocks,
	})
	if err != nil {
		debugLog("Chat error: %v", err)
		return nil, nil, err
	}

	// Convert response to our format
	resultMessages := make([]providers.ChatMessage, 0)
	var toolUses []providers.ToolUse

	for _, content := range message.Content {
		switch content.Type {
		case "text":
			resultMessages = append(resultMessages, providers.ChatMessage{
				Role:    providers.RoleAssistant,
				Content: content.Text,
			})
		case "tool_use":
			toolUses = append(toolUses, providers.ToolUse{
				ID:    content.ID,
				Name:  content.Name,
				Input: content.Input,
			})
		}
	}

	return resultMessages, toolUses, nil
}

// StreamChat sends a message to Anthropic and returns a streaming response
func (c *AnthropicClient) StreamChat(ctx context.Context, model string, messages []providers.ChatMessage, temperature float64) (<-chan string, error) {
	// Convert messages to Anthropic format
	anthropicMessages := make([]anthropic.MessageParam, 0)
	systemPrompt := ""

	debugLog("StreamChat: model=%s, temperature=%f, total_messages=%d", model, temperature, len(messages))

	for _, msg := range messages {
		if msg.Role == providers.RoleSystem {
			// capture system prompt; Anthropic expects it separately
			systemPrompt = msg.Content
			continue
		} else if msg.Role == providers.RoleUser {
			anthropicMessages = append(anthropicMessages, anthropic.NewUserMessage(anthropic.NewTextBlock(msg.Content)))
		} else if msg.Role == providers.RoleAssistant {
			anthropicMessages = append(anthropicMessages, anthropic.NewAssistantMessage(anthropic.NewTextBlock(msg.Content)))
		}
	}

	// Create stream
	stream := c.client.Messages.NewStreaming(ctx, anthropic.MessageNewParams{
		Model:       anthropic.Model(model),
		MaxTokens:   1024,
		Messages:    anthropicMessages,
		System:      []anthropic.TextBlockParam{{Type: "text", Text: systemPrompt}},
		Temperature: anthropic.Float(temperature),
	})

	debugLog("StreamChat: stream created, waiting for events")

	// Create channel for streaming response
	responseChan := make(chan string, 100)

	go func() {
		defer close(responseChan)
		defer stream.Close()

		for stream.Next() {
			event := stream.Current()
			debugLog("StreamChat: received event type=%s", event.Type)
			switch event.Type {
			case "content_block_delta":
				if event.Delta.Type == "text_delta" {
					responseChan <- event.Delta.Text
				}
			case "message_stop":
				return
			}
		}

		if err := stream.Err(); err != nil {
			// Send error as final message
			debugLog("StreamChat: stream error: %v", err)
			responseChan <- fmt.Sprintf("Error: %v", err)
		}
	}()

	return responseChan, nil
}

// SendToolResult sends a tool result back to Anthropic and returns its response
func (c *AnthropicClient) SendToolResult(ctx context.Context, messages []providers.ChatMessage, toolResults []providers.ToolResult) ([]providers.ChatMessage, []providers.ToolUse, error) {
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

func toStringSlice(v any) []string {
	if v == nil {
		return nil
	}
	if raw, ok := v.([]any); ok {
		out := make([]string, 0, len(raw))
		for _, e := range raw {
			if s, ok := e.(string); ok {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}
