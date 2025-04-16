package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// OllamaClient implements the Client interface for Ollama
type OllamaClient struct {
	baseURL string
	model   string
	configs *ToolConfigs
}

// NewOllamaClient creates a new Ollama client
func NewOllamaClient(model string, configPath string) (*OllamaClient, error) {
	configs, err := LoadToolConfigs(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load tool configs: %w", err)
	}

	return &OllamaClient{
		baseURL: "http://localhost:11434",
		model:   model,
		configs: configs,
	}, nil
}

// Name returns the name of the LLM
func (c *OllamaClient) Name() string {
	return c.model
}

// Chat sends a message to Ollama and returns its response
func (c *OllamaClient) Chat(ctx context.Context, messages []Message, tools []Tool) ([]Message, []ToolUse, error) {
	// Generate system message from tool configs
	systemMessage := map[string]string{
		"role":    "system",
		"content": GenerateSystemPrompt(c.configs),
	}

	// Convert messages to Ollama format
	ollamaMessages := make([]map[string]string, len(messages)+1)
	ollamaMessages[0] = systemMessage
	for i, msg := range messages {
		ollamaMessages[i+1] = map[string]string{
			"role":    msg.Role,
			"content": msg.Content,
		}
	}

	// Create request body
	reqBody := map[string]interface{}{
		"model":    c.model,
		"messages": ollamaMessages,
		"stream":   false,
	}

	// Only include tools if the last message might need them
	lastMessage := messages[len(messages)-1].Content
	needsTools := false
	if len(tools) > 0 {
		// Check if the message might need tools using trigger keywords
		lowerMsg := strings.ToLower(lastMessage)
		for _, tool := range c.configs.Tools {
			for _, keyword := range tool.TriggerKeywords {
				if strings.Contains(lowerMsg, keyword) {
					needsTools = true
					break
				}
			}
			if needsTools {
				break
			}
		}
	}

	if needsTools {
		ollamaTools := make([]map[string]interface{}, len(tools))
		for i, tool := range tools {
			ollamaTools[i] = map[string]interface{}{
				"type": "function",
				"function": map[string]interface{}{
					"name":        tool.Name,
					"description": tool.Description,
					"parameters":  tool.InputSchema,
				},
			}
		}
		reqBody["tools"] = ollamaTools
	}

	// Marshal request body
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/chat", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var response struct {
		Message struct {
			Role      string `json:"role"`
			Content   string `json:"content"`
			ToolCalls []struct {
				ID       string `json:"id"`
				Type     string `json:"type"`
				Function struct {
					Name      string          `json:"name"`
					Arguments json.RawMessage `json:"arguments"`
				} `json:"function"`
			} `json:"tool_calls"`
		} `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert tool calls to ToolUse objects
	var toolUses []ToolUse
	for _, toolCall := range response.Message.ToolCalls {
		// Skip empty tool calls
		if toolCall.Function.Name == "" {
			continue
		}
		toolUses = append(toolUses, ToolUse{
			ID:    toolCall.ID,
			Name:  toolCall.Function.Name,
			Input: toolCall.Function.Arguments,
		})
	}

	// Convert response to our format
	resultMessages := []Message{
		{
			Role:    response.Message.Role,
			Content: response.Message.Content,
		},
	}

	return resultMessages, toolUses, nil
}

// SendToolResult sends a tool result back to Ollama and returns its response
func (c *OllamaClient) SendToolResult(ctx context.Context, messages []Message, toolResults []ToolResult) ([]Message, []ToolUse, error) {
	// Create a proper tool result format for better understanding
	var toolResultsSummary string
	var lastUserMessage string

	// Find the last user message for context
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" {
			lastUserMessage = messages[i].Content
			break
		}
	}

	// Format tool results with context
	for i, result := range toolResults {
		formattedResult := fmt.Sprintf("Tool result from %s: %s", result.ID, result.Content)
		if result.IsError {
			formattedResult = fmt.Sprintf("Error from tool %s: %s", result.ID, result.Content)
		}
		toolResults[i].Content = formattedResult
		toolResultsSummary += formattedResult + "\n"
	}
	
	// Add tool results to messages
	for _, result := range toolResults {
		messages = append(messages, Message{
			Role:    "tool",
			Content: result.Content,
		})
	}

	// Create a detailed prompt that provides context about the tool results
	analysisPrompt := fmt.Sprintf(`You are a helpful AI assistant. The user asked: "%s"

Here are the results from the tools I used to help answer their question:

%s

Please provide a helpful response that:
1. Directly answers the user's question using the tool results
2. If the results are from a file or directory listing, summarize the key contents
3. If there were any errors, explain what went wrong and how to fix it
4. Suggest any next steps or additional information that might be helpful

Keep your response focused on answering the user's specific question using the tool results provided.`, lastUserMessage, toolResultsSummary)

	// Continue the conversation with the detailed prompt
	messages = append(messages, Message{
		Role:    "user",
		Content: analysisPrompt,
	})

	// Continue the conversation
	return c.Chat(ctx, messages, nil)
} 