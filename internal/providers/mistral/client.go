package mistral

import (
	"context"
	"fmt"
	"os"

	"github.com/gage-technologies/mistral-go"
	"github.com/pprunty/magikarp/internal/providers"
)

// MistralClient implements the Provider interface for Mistral AI
type MistralClient struct {
	client       *mistral.MistralClient
	apiKey       string
	models       []string
	temperature  float64
	systemPrompt string
}

// New creates a new Mistral provider
func New(apiKey string, models []string, temperature float64, systemPrompt string) (*MistralClient, error) {
	client := mistral.NewMistralClientDefault(apiKey)
	
	return &MistralClient{
		client:       client,
		apiKey:       apiKey,
		models:       models,
		temperature:  temperature,
		systemPrompt: systemPrompt,
	}, nil
}

// NewMistralClient creates a new Mistral client (legacy)
func NewMistralClient(model string, configPath string) (*MistralClient, error) {
	// Check if API key is set
	if os.Getenv("MISTRAL_API_KEY") == "" {
		return nil, fmt.Errorf("MISTRAL_API_KEY environment variable is not set")
	}

	client, err := New(os.Getenv("MISTRAL_API_KEY"), []string{model}, 0.0, "")
	return client, err
}

// Name returns the name of the provider
func (c *MistralClient) Name() string {
	return "mistral"
}

// Chat sends a message to Mistral and returns its response
func (c *MistralClient) Chat(ctx context.Context, messages []providers.ChatMessage, tools []providers.Tool) ([]providers.ChatMessage, []providers.ToolUse, error) {
	// Use first available model
	modelName := "mistral-large-latest"
	if len(c.models) > 0 {
		modelName = c.models[0]
	}

	// Convert messages to Mistral format
	mistralMessages := make([]mistral.ChatMessage, 0)
	
	for _, msg := range messages {
		if msg.Role == providers.RoleSystem {
			mistralMessages = append(mistralMessages, mistral.ChatMessage{
				Role:    mistral.RoleSystem,
				Content: msg.Content,
			})
		} else if msg.Role == providers.RoleUser {
			mistralMessages = append(mistralMessages, mistral.ChatMessage{
				Role:    mistral.RoleUser,
				Content: msg.Content,
			})
		} else if msg.Role == providers.RoleAssistant {
			mistralMessages = append(mistralMessages, mistral.ChatMessage{
				Role:    mistral.RoleAssistant,
				Content: msg.Content,
			})
		} else if msg.Role == providers.RoleTool {
			mistralMessages = append(mistralMessages, mistral.ChatMessage{
				Role:    mistral.RoleUser,
				Content: msg.Content,
			})
		}
	}

	// Add system message at the beginning if we have one from config and no system message in conversation
	if c.systemPrompt != "" {
		hasSystemMessage := false
		for _, msg := range messages {
			if msg.Role == providers.RoleSystem {
				hasSystemMessage = true
				break
			}
		}
		if !hasSystemMessage {
			systemMsg := mistral.ChatMessage{
				Role:    mistral.RoleSystem,
				Content: c.systemPrompt,
			}
			mistralMessages = append([]mistral.ChatMessage{systemMsg}, mistralMessages...)
		}
	}

	// Send request to Mistral using the API
	chatRes, err := c.client.Chat(modelName, mistralMessages, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create chat completion: %w", err)
	}

	// Convert response to our format
	resultMessages := make([]providers.ChatMessage, 0)
	var toolUses []providers.ToolUse

	for _, choice := range chatRes.Choices {
		if choice.Message.Content != "" {
			resultMessages = append(resultMessages, providers.ChatMessage{
				Role:    providers.RoleAssistant,
				Content: choice.Message.Content,
			})
		}

		// Handle tool calls (if supported by this version of the SDK)
		// Note: Tool calling might not be available in all versions
	}

	return resultMessages, toolUses, nil
}

// StreamChat sends a message to Mistral and returns a streaming response
func (c *MistralClient) StreamChat(ctx context.Context, model string, messages []providers.ChatMessage, temperature float64) (<-chan string, error) {
	// Convert messages to Mistral format
	mistralMessages := make([]mistral.ChatMessage, 0)
	
	for _, msg := range messages {
		if msg.Role == providers.RoleSystem {
			mistralMessages = append(mistralMessages, mistral.ChatMessage{
				Role:    mistral.RoleSystem,
				Content: msg.Content,
			})
		} else if msg.Role == providers.RoleUser {
			mistralMessages = append(mistralMessages, mistral.ChatMessage{
				Role:    mistral.RoleUser,
				Content: msg.Content,
			})
		} else if msg.Role == providers.RoleAssistant {
			mistralMessages = append(mistralMessages, mistral.ChatMessage{
				Role:    mistral.RoleAssistant,
				Content: msg.Content,
			})
		}
	}
	
	// Add system message at the beginning if we have one from config and no system message in conversation
	if c.systemPrompt != "" {
		hasSystemMessage := false
		for _, msg := range messages {
			if msg.Role == providers.RoleSystem {
				hasSystemMessage = true
				break
			}
		}
		if !hasSystemMessage {
			systemMsg := mistral.ChatMessage{
				Role:    mistral.RoleSystem,
				Content: c.systemPrompt,
			}
			mistralMessages = append([]mistral.ChatMessage{systemMsg}, mistralMessages...)
		}
	}

	// Create streaming channel
	responseChan := make(chan string, 100)

	go func() {
		defer close(responseChan)

		// Use the ChatStream method
		chatResChan, err := c.client.ChatStream(model, mistralMessages, nil)
		if err != nil {
			responseChan <- fmt.Sprintf("Error: %v", err)
			return
		}

		for chatResChunk := range chatResChan {
			if chatResChunk.Error != nil {
				responseChan <- fmt.Sprintf("Error: %v", chatResChunk.Error)
				break
			}
			
			for _, choice := range chatResChunk.Choices {
				if choice.Delta.Content != "" {
					responseChan <- choice.Delta.Content
				}
			}
		}
	}()

	return responseChan, nil
}

// SendToolResult sends a tool result back to Mistral and returns its response
func (c *MistralClient) SendToolResult(ctx context.Context, messages []providers.ChatMessage, toolResults []providers.ToolResult) ([]providers.ChatMessage, []providers.ToolUse, error) {
	// Add tool results to messages
	augmented := make([]providers.ChatMessage, len(messages))
	copy(augmented, messages)

	for _, result := range toolResults {
		augmented = append(augmented, providers.ChatMessage{
			Role:    providers.RoleTool,
			Content: result.Content,
		})
	}

	// Continue the conversation with all tools available
	return c.Chat(ctx, augmented, nil)
}