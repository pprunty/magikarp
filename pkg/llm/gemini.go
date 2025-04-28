package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// GeminiClient implements the Client interface for Google's Gemini
type GeminiClient struct {
	client *genai.Client
	model  string
}

// NewGeminiClient creates a new Gemini client
func NewGeminiClient(model string) (*GeminiClient, error) {
	// Check if API key is set
	if os.Getenv("GEMINI_API_KEY") == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable is not set")
	}

	client, err := genai.NewClient(context.Background(), option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return &GeminiClient{
		client: client,
		model:  model,
	}, nil
}

// Name returns the name of the LLM
func (c *GeminiClient) Name() string {
	return c.model
}

// Chat sends a message to Gemini and returns its response
func (c *GeminiClient) Chat(ctx context.Context, messages []Message, tools []Tool) ([]Message, []ToolUse, error) {
	// Get the model
	model := c.client.GenerativeModel(c.model)

	// Convert messages to Gemini format
	geminiMessages := make([]*genai.Content, len(messages))
	for i, msg := range messages {
		geminiMessages[i] = &genai.Content{
			Parts: []genai.Part{
				genai.Text(msg.Content),
			},
			Role: msg.Role,
		}
	}

	// Start a chat session
	cs := model.StartChat()
	cs.History = geminiMessages

	// Generate response
	resp, err := cs.SendMessage(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to send message to Gemini: %w", err)
	}

	// Convert response to our format
	resultMessages := make([]Message, 0)
	var toolUses []ToolUse

	for _, candidate := range resp.Candidates {
		if candidate.Content != nil {
			for _, part := range candidate.Content.Parts {
				if text, ok := part.(genai.Text); ok {
					resultMessages = append(resultMessages, Message{
						Role:    candidate.Content.Role,
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
						toolUses = append(toolUses, ToolUse{
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

// SendToolResult sends a tool result back to Gemini and returns its response
func (c *GeminiClient) SendToolResult(ctx context.Context, messages []Message, toolResults []ToolResult) ([]Message, []ToolUse, error) {
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