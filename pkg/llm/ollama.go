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

// shouldUseTool checks if a message indicates the need for a specific tool
func (c *OllamaClient) shouldUseTool(message string, tool ToolConfig) bool {
	lowerMsg := strings.ToLower(message)

	// Check direct keywords
	for _, keyword := range tool.TriggerKeywords {
		if strings.Contains(lowerMsg, keyword) {
			return true
		}
	}

	// Check for tool combinations
	toolCombos := map[string][]string{
		"read_file": {"edit", "modify", "change", "update", "execute", "run"},
		"edit_file": {"after", "then", "read", "check"},
		"execute_command": {"output", "result", "after", "then"},
		"list_files": {"then", "read", "edit", "execute"},
	}

	if keywords, ok := toolCombos[tool.Name]; ok {
		for _, keyword := range keywords {
			if strings.Contains(lowerMsg, keyword) {
				return true
			}
		}
	}

	return false
}

// Chat sends a message to Ollama and returns its response
func (c *OllamaClient) Chat(ctx context.Context, messages []Message, tools []Tool) ([]Message, []ToolUse, error) {
	// Generate enhanced system message
	systemPrompt := `You are a helpful AI assistant with access to various tools. When using tools:

1. For simple tasks, use a single tool directly
2. For complex tasks that require multiple steps:
   - Explain your plan first
   - Use tools in a logical sequence
   - Process each tool's output before proceeding
   - Combine results meaningfully

Common tool combinations:
- Reading then editing files
- Reading files then executing commands
- Listing files then reading/editing them
- Executing commands then processing their output

Always explain what you're doing and why. If a task requires multiple tools, explain the sequence.`

	systemPrompt += "\n\n" + GenerateSystemPrompt(c.configs)

	// Convert messages to Ollama format
	ollamaMessages := []map[string]string{
		{
			"role":    "system",
			"content": systemPrompt,
		},
	}

	for _, msg := range messages {
		ollamaMessages = append(ollamaMessages, map[string]string{
			"role":    msg.Role,
			"content": msg.Content,
		})
	}

	// Create request body
	reqBody := map[string]interface{}{
		"model":    c.model,
		"messages": ollamaMessages,
		"stream":   false,
	}

	// Check if tools should be included
	if len(tools) > 0 && len(messages) > 0 {
		lastMessage := messages[len(messages)-1].Content
		needsTools := false

		// Check each tool and potential combinations
		for _, tool := range c.configs.Tools {
			if c.shouldUseTool(lastMessage, tool) {
				needsTools = true
				break
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
	// Add tool results to messages with context
	for _, result := range toolResults {
		messages = append(messages, Message{
			Role:    "tool",
			Content: fmt.Sprintf("Tool '%s' returned: %s", result.ID, result.Content),
		})
	}

	// Continue the conversation with all tools available
	return c.Chat(ctx, messages, nil)
} 