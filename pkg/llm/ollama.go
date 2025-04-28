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
}

// NewOllamaClient creates a new Ollama client
func NewOllamaClient(model string) (*OllamaClient, error) {
	return &OllamaClient{
		baseURL: "http://localhost:11434",
		model:   model,
	}, nil
}

// Name returns the name of the LLM
func (c *OllamaClient) Name() string {
	return c.model
}

// Chat sends a message to Ollama and returns its response
func (c *OllamaClient) Chat(ctx context.Context, messages []Message, tools []Tool) ([]Message, []ToolUse, error) {
	// Generate enhanced system message
	systemPrompt := `You are a helpful AI assistant with access to system tools. Follow these rules EXACTLY:

1. NEVER make up or hallucinate information
2. ONLY state what you can verify from tool results
3. If you're not sure about something, say so
4. If a tool returns an error, acknowledge it and explain what happened

When using tools:
1. Tell the user what you're going to do
2. Use the appropriate tool
3. Wait for the result
4. ONLY describe what was in the result

CRITICAL RULES FOR TOOL RESULTS:
- For read_file: ONLY summarize the EXACT content that was read
- For list_files: ONLY list the EXACT files that were found
- For execute_command: ONLY explain the EXACT output received
- For edit_file: ALWAYS read first, then make targeted edits

If you receive a tool error:
1. Acknowledge the error
2. Explain what happened
3. Suggest what to do next

ABSOLUTELY NO HALLUCINATIONS:
- Never make up file contents
- Never assume what files contain
- Never add information that wasn't in the tool result
- Never describe files you haven't read
- Never make assumptions about command outputs
- If you're not sure about something, say so`

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

	// If tools are provided, always make them available
	if len(tools) > 0 {
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
		// Create appropriate context for the tool result
		var contextMsg string
		
		// Extract the actual content from the tool result
		content := result.Content
		if strings.Contains(content, "File read successfully") {
			// Parse the JSON response to extract the actual file content
			var toolResult struct {
				Success bool        `json:"success"`
				Message string     `json:"message"`
				Data    string     `json:"data"`
			}
			if err := json.Unmarshal([]byte(content), &toolResult); err == nil && toolResult.Success {
				content = toolResult.Data
			} else {
				content = "Error: Failed to read file content"
			}
		} else if strings.Contains(content, "Files listed successfully") {
			var toolResult struct {
				Success bool        `json:"success"`
				Message string     `json:"message"`
				Data    []string   `json:"data"`
			}
			if err := json.Unmarshal([]byte(content), &toolResult); err == nil && toolResult.Success {
				content = strings.Join(toolResult.Data, "\n")
			}
		}
		
		// Determine the type of tool result and format accordingly
		switch {
		case strings.Contains(result.Content, "File read successfully"):
			// For read_file, include a prompt to analyze the file contents
			contextMsg = fmt.Sprintf("Here is the EXACT content of the file. Do NOT add any information that is not present here:\n\n%s\n\nProvide a direct summary of ONLY what is shown above. Do not make ANY assumptions about content not shown.", content)
		case strings.Contains(result.Content, "Files listed successfully"):
			// For list_files, include a prompt to analyze the directory contents
			contextMsg = fmt.Sprintf("Here are the EXACT files found:\n\n%s\n\nList ONLY the files shown above. Do not make ANY assumptions about other files.", content)
		case strings.Contains(result.Content, "Command executed successfully"):
			// For execute_command, include a prompt to explain the command output
			contextMsg = fmt.Sprintf("Here is the EXACT command output:\n\n%s\n\nExplain ONLY what is shown in the output above. Do not make ANY assumptions about other output.", content)
		default:
			// Default format for other tools
			contextMsg = fmt.Sprintf("Tool result: %s\n\nRespond ONLY based on this result. Do not make ANY assumptions.", content)
		}
		
		// Add the context message
		messages = append(messages, Message{
			Role:    "user",
			Content: contextMsg,
		})
	}

	// Continue the conversation with all tools available
	return c.Chat(ctx, messages, nil)
} 