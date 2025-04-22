package models

import (
	"context"
	"fmt"
)

// DummyClient implements a dummy LLM client for testing
type DummyClient struct {
	modelName string
}

// NewDummyClient creates a new dummy client
func NewDummyClient(modelName string) (*DummyClient, error) {
	return &DummyClient{
		modelName: modelName,
	}, nil
}

// Name returns the name of the LLM model
func (c *DummyClient) Name() string {
	return c.modelName
}

// Chat sends a conversation to the LLM and returns the response and any tool use requests
func (c *DummyClient) Chat(ctx context.Context, messages []Message, tools []Tool) ([]Message, []ToolUse, error) {
	// Just echo the last user message and pretend to use a tool
	var lastUserMsg string
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" {
			lastUserMsg = messages[i].Content
			break
		}
	}

	if lastUserMsg == "" {
		lastUserMsg = "No user message found"
	}

	response := Message{
		Role:    "assistant",
		Content: fmt.Sprintf("I'll help you with: %s", lastUserMsg),
	}

	var toolUses []ToolUse
	if len(tools) > 0 {
		// Dummy tool use
		toolUses = append(toolUses, ToolUse{
			ID:    "dummy-tool-1",
			Name:  tools[0].Name,
			Input: []byte(`{}`),
		})
	}

	return []Message{response}, toolUses, nil
}

// SendToolResult sends the result of a tool execution to the LLM and returns the updated response
func (c *DummyClient) SendToolResult(ctx context.Context, messages []Message, results []ToolResult) ([]Message, []ToolUse, error) {
	// Just acknowledge the tool results
	response := Message{
		Role:    "assistant",
		Content: "I've processed the tool results. Here's what I found...",
	}

	return []Message{response}, nil, nil
} 