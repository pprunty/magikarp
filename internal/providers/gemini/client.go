package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/google/generative-ai-go/genai"
	"github.com/pprunty/magikarp/internal/providers"
	"google.golang.org/api/option"
)

// GeminiClient implements the Provider interface for Google's Gemini
type GeminiClient struct {
	client *genai.Client
	apiKey string
	models []string
}

// New creates a new Gemini provider
func New(apiKey string, models []string) (*GeminiClient, error) {
	client, err := genai.NewClient(context.Background(), option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return &GeminiClient{
		client: client,
		apiKey: apiKey,
		models: models,
	}, nil
}

// NewGeminiClient creates a new Gemini client (legacy)
func NewGeminiClient(model string, configPath string) (*GeminiClient, error) {
	// Check if API key is set
	if os.Getenv("GEMINI_API_KEY") == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable is not set")
	}

	client, err := New(os.Getenv("GEMINI_API_KEY"), []string{model})
	return client, err
}

// Name returns the name of the provider
func (c *GeminiClient) Name() string {
	return "gemini"
}

// Chat sends a message to Gemini and returns its response
func (c *GeminiClient) Chat(ctx context.Context, messages []providers.ChatMessage, tools []providers.Tool) ([]providers.ChatMessage, []providers.ToolUse, error) {
	// Use first available model
	modelName := "gemini-pro"
	if len(c.models) > 0 {
		modelName = c.models[0]
	}

	// Get the model
	model := c.client.GenerativeModel(modelName)

	// Convert messages to Gemini format
	geminiMessages := make([]*genai.Content, 0)
	systemPrompt := ""
	for _, msg := range messages {
		if msg.Role == providers.RoleSystem {
			systemPrompt = msg.Content
			continue
		}

		role := "user"
		if msg.Role == providers.RoleAssistant {
			role = "model"
		}

		geminiMessages = append(geminiMessages, &genai.Content{
			Parts: []genai.Part{
				genai.Text(msg.Content),
			},
			Role: role,
		})
	}

	// Attach system instruction if provided
	if systemPrompt != "" {
		model.SystemInstruction = &genai.Content{Parts: []genai.Part{genai.Text(systemPrompt)}, Role: "system"}
	}

	// Start a chat session
	cs := model.StartChat()
	if len(geminiMessages) > 1 {
		cs.History = geminiMessages[:len(geminiMessages)-1]
	}

	// Generate response with the last message
	lastMsg := geminiMessages[len(geminiMessages)-1]
	resp, err := cs.SendMessage(ctx, lastMsg.Parts...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to send message to Gemini: %w", err)
	}

	// Convert response to our format
	resultMessages := make([]providers.ChatMessage, 0)
	var toolUses []providers.ToolUse

	for _, candidate := range resp.Candidates {
		if candidate.Content != nil {
			for _, part := range candidate.Content.Parts {
				if text, ok := part.(genai.Text); ok {
					resultMessages = append(resultMessages, providers.ChatMessage{
						Role:    providers.RoleAssistant,
						Content: string(text),
					})
				}
			}
		}

		// Handle function calls (Gemini uses a custom JSON format)
		if candidate.Content != nil {
			for _, part := range candidate.Content.Parts {
				if text, ok := part.(genai.Text); ok {
					var functionCall struct {
						Name      string          `json:"name"`
						Arguments json.RawMessage `json:"arguments"`
					}
					if err := json.Unmarshal([]byte(text), &functionCall); err == nil && functionCall.Name != "" {
						toolUses = append(toolUses, providers.ToolUse{
							Name:  functionCall.Name,
							Input: functionCall.Arguments,
						})
					}
				}
			}
		}
	}

	return resultMessages, toolUses, nil
}

// StreamChat sends a message to Gemini and returns a streaming response
func (c *GeminiClient) StreamChat(ctx context.Context, model string, messages []providers.ChatMessage, temperature float64) (<-chan string, error) {
	// Get the model
	geminiModel := c.client.GenerativeModel(model)
	temp32 := float32(temperature)
	geminiModel.Temperature = &temp32

	// Convert messages to Gemini format
	geminiMessages := make([]*genai.Content, 0)
	systemPrompt := ""
	for _, msg := range messages {
		if msg.Role == providers.RoleSystem {
			systemPrompt = msg.Content
			continue
		}

		role := "user"
		if msg.Role == providers.RoleAssistant {
			role = "model"
		}

		geminiMessages = append(geminiMessages, &genai.Content{
			Parts: []genai.Part{
				genai.Text(msg.Content),
			},
			Role: role,
		})
	}

	// attach system prompt to model
	if systemPrompt != "" {
		geminiModel.SystemInstruction = &genai.Content{Parts: []genai.Part{genai.Text(systemPrompt)}, Role: "system"}
	}

	// Create channel for streaming response
	responseChan := make(chan string, 100)

	go func() {
		defer close(responseChan)

		// Start a chat session
		cs := geminiModel.StartChat()
		if len(geminiMessages) > 1 {
			cs.History = geminiMessages[:len(geminiMessages)-1] // All but the last message
		}

		// Stream response with the last message
		lastMsg := geminiMessages[len(geminiMessages)-1]
		iter := cs.SendMessageStream(ctx, lastMsg.Parts...)

		for {
			resp, err := iter.Next()
			if err != nil {
				if err.Error() == "no more items in iterator" {
					return
				}
				responseChan <- fmt.Sprintf("Error: %v", err)
				return
			}

			for _, candidate := range resp.Candidates {
				if candidate.Content != nil {
					for _, part := range candidate.Content.Parts {
						if text, ok := part.(genai.Text); ok {
							responseChan <- string(text)
						}
					}
				}
			}
		}
	}()

	return responseChan, nil
}

// SendToolResult sends a tool result back to Gemini and returns its response
func (c *GeminiClient) SendToolResult(ctx context.Context, messages []providers.ChatMessage, toolResults []providers.ToolResult) ([]providers.ChatMessage, []providers.ToolUse, error) {
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
