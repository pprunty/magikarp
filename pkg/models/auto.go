package llm

import (
	"context"
	"fmt"
	"strings"
)

// AutoClient implements the Client interface with automatic model switching
type AutoClient struct {
	clients map[string]Client
	model   string
	history []Message
}

// NewAutoClient creates a new AutoClient with multiple model support
func NewAutoClient(models []string, configPath string) (*AutoClient, error) {
	clients := make(map[string]Client)
	
	// Initialize all available clients
	for _, model := range models {
		var client Client
		var err error
		
		// Determine which client to create based on model name
		switch {
		case strings.HasPrefix(model, "gpt-"):
			client, err = NewOpenAIClient(model, configPath)
		case strings.HasPrefix(model, "claude-"):
			client, err = NewAnthropicClient(model, configPath)
		case strings.HasPrefix(model, "gemini-"):
			client, err = NewGeminiClient(model, configPath)
		case strings.HasPrefix(model, "llama"):
			client, err = NewOllamaClient(model, configPath)
		default:
			return nil, fmt.Errorf("unknown model type: %s", model)
		}
		
		if err != nil {
			return nil, fmt.Errorf("failed to create client for %s: %w", model, err)
		}
		
		clients[model] = client
	}
	
	return &AutoClient{
		clients: clients,
		model:   models[0], // Default to first model
		history: make([]Message, 0),
	}, nil
}

// Name returns the name of the current model
func (c *AutoClient) Name() string {
	return c.model
}

// selectModel chooses the most appropriate model based on the prompt
func (c *AutoClient) selectModel(prompt string) string {
	// Convert prompt to lowercase for case-insensitive matching
	lowerPrompt := strings.ToLower(prompt)
	
	// Model selection rules
	switch {
	// Code-related tasks
	case strings.Contains(lowerPrompt, "code") || 
		 strings.Contains(lowerPrompt, "program") || 
		 strings.Contains(lowerPrompt, "function") || 
		 strings.Contains(lowerPrompt, "class") || 
		 strings.Contains(lowerPrompt, "method"):
		return "gpt-4" // Best for code understanding and generation
	
	// Creative writing
	case strings.Contains(lowerPrompt, "write") || 
		 strings.Contains(lowerPrompt, "story") || 
		 strings.Contains(lowerPrompt, "poem") || 
		 strings.Contains(lowerPrompt, "creative"):
		return "claude-3-opus" // Best for creative tasks
	
	// General knowledge and reasoning
	case strings.Contains(lowerPrompt, "explain") || 
		 strings.Contains(lowerPrompt, "why") || 
		 strings.Contains(lowerPrompt, "how") || 
		 strings.Contains(lowerPrompt, "what"):
		return "gemini-pro" // Good for general knowledge
	
	// Default to the current model
	default:
		return c.model
	}
}

// Chat sends a message to the appropriate model and returns its response
func (c *AutoClient) Chat(ctx context.Context, messages []Message, tools []Tool) ([]Message, []ToolUse, error) {
	// Update history
	c.history = append(c.history, messages...)
	
	// Get the last user message
	var lastUserMessage string
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" {
			lastUserMessage = messages[i].Content
			break
		}
	}
	
	// Select appropriate model
	selectedModel := c.selectModel(lastUserMessage)
	
	// Switch model if needed
	if selectedModel != c.model {
		fmt.Printf("Switching to model: %s\n", selectedModel)
		c.model = selectedModel
	}
	
	// Get the selected client
	client, ok := c.clients[c.model]
	if !ok {
		return nil, nil, fmt.Errorf("model not found: %s", c.model)
	}
	
	// Send message to selected model
	return client.Chat(ctx, c.history, tools)
}

// SendToolResult sends a tool result back to the current model
func (c *AutoClient) SendToolResult(ctx context.Context, messages []Message, toolResults []ToolResult) ([]Message, []ToolUse, error) {
	// Update history
	c.history = append(c.history, messages...)
	
	// Get the current client
	client, ok := c.clients[c.model]
	if !ok {
		return nil, nil, fmt.Errorf("model not found: %s", c.model)
	}
	
	// Send tool result to current model
	return client.SendToolResult(ctx, c.history, toolResults)
} 