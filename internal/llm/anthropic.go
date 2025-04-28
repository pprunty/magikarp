package llm

import (
	"context"
	"fmt"
	"os"
	"regexp"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/pprunty/magikarp/pkg/agent"
)

// AnthropicClient implements the Client interface for Anthropic
type AnthropicClient struct {
	client *anthropic.Client
	model  string
}

var propertyNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,64}$`)

// buildSchema converts a ToolDefinition to Anthropic's schema format
func buildSchema(td agent.ToolDefinition) anthropic.ToolInputSchemaParam {
	// Start with a proper JSON Schema draft 2020-12 structure
	schemaObj := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{},
		"additionalProperties": false,
	}
	
	// Add required array if needed
	requiredFields := []string{}
	
	// Process parameters from the tool definition
	if len(td.Parameters) > 0 {
		properties := map[string]interface{}{}
		
		for name, param := range td.Parameters {
			// Skip invalid property names
			if !propertyNameRegex.MatchString(name) {
				continue
			}
			
			// Create valid JSON Schema property
			prop := map[string]interface{}{
				"type": param.Type,
				"description": param.Description,
			}
			
			properties[name] = prop
			
			if param.Required {
				requiredFields = append(requiredFields, name)
			}
		}
		
		schemaObj["properties"] = properties
	} else if td.InputSchema != nil {
		// Use the existing input schema if available
		if props, ok := td.InputSchema["properties"].(map[string]interface{}); ok {
			schemaObj["properties"] = props
		}
		
		if req, ok := td.InputSchema["required"].([]string); ok && len(req) > 0 {
			requiredFields = req
		}
	}
	
	// Only add required if there are required fields
	if len(requiredFields) > 0 {
		schemaObj["required"] = requiredFields
	}
	
	// Return the final schema
	return anthropic.ToolInputSchemaParam{
		Properties: schemaObj,
	}
}

// NewAnthropicClient creates a new Anthropic client
func NewAnthropicClient(model string, configPath string) (*AnthropicClient, error) {
	// Check if API key is set
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY environment variable is not set")
	}

	client := anthropic.NewClient()
	return &AnthropicClient{
		client: &client,
		model:  model,
	}, nil
}

// Name returns the name of the LLM
func (c *AnthropicClient) Name() string {
	return c.model
}

// Chat sends a message to Anthropic and returns its response
func (c *AnthropicClient) Chat(ctx context.Context, messages []Message, tools []agent.ToolDefinition) ([]Message, []ToolUse, error) {
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
		anthropicTools[i] = anthropic.ToolUnionParam{
			OfTool: &anthropic.ToolParam{
				Name:        tool.Name,
				Description: anthropic.String(tool.Description),
				InputSchema: buildSchema(tool),
			},
		}
	}

	// Send request to Anthropic
	message, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(c.model),
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
func (c *AnthropicClient) SendToolResult(ctx context.Context, messages []Message, toolResults []agent.ToolResult) ([]Message, []ToolUse, error) {
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