package llm

import (
    "context"
    "fmt"
    "log/slog"
    "os"
    "regexp"
    "time"
    "strings"

    "github.com/anthropics/anthropic-sdk-go"
    "github.com/pprunty/magikarp/pkg/agent"
)

/* ─────────────────────────────────── */

type AnthropicClient struct {
    client *anthropic.Client
    model  string
    logger *slog.Logger
}

/* ───────────────────────────────────
   schema builder
─────────────────────────────────── */

var propName = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,64}$`)

func (c *AnthropicClient) buildSchema(td agent.ToolDefinition) anthropic.ToolInputSchemaParam {
    // full schema already supplied?
    if s := td.InputSchema; len(s) != 0 {
       props, _ := s["properties"].(map[string]any)
       param := anthropic.ToolInputSchemaParam{Properties: props}
       param.WithExtraFields(s) // mutates
       return param
    }

    // else derive from td.Parameters
    props := map[string]any{}
    var req []string
    for n, p := range td.Parameters {
       if !propName.MatchString(n) {
          continue
       }
       props[n] = map[string]any{"type": p.Type, "description": p.Description}
       if p.Required {
          req = append(req, n)
       }
    }
    param := anthropic.ToolInputSchemaParam{Properties: props}
    param.WithExtraFields(map[string]any{
       "type":                 "object",
       "additionalProperties": false,
       "required":             req,
    })
    return param
}

/* ───────────────────────────────────
   ctor
─────────────────────────────────── */

func NewAnthropicClient(model string) (*AnthropicClient, error) {
    if os.Getenv("ANTHROPIC_API_KEY") == "" {
       return nil, fmt.Errorf("ANTHROPIC_API_KEY not set")
    }
    c := anthropic.NewClient()
    log := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
    return &AnthropicClient{client: &c, model: model, logger: log}, nil
}

func (c *AnthropicClient) Name() string { return c.model }

/* ───────────────────────────────────
   Chat - now with system prompt support
─────────────────────────────────── */

func (c *AnthropicClient) Chat(ctx context.Context, msgs []Message, tools []Tool, systemPrompt string) ([]Message, []ToolUse, error) {
    // Convert our messages to Anthropic's format
    anthropicMsgs := make([]anthropic.MessageParam, len(msgs))
    for i, m := range msgs {
        switch m.Role {
        case "user":
            anthropicMsgs[i] = anthropic.NewUserMessage(anthropic.NewTextBlock(m.Content))
        case "assistant":
            anthropicMsgs[i] = anthropic.NewAssistantMessage(anthropic.NewTextBlock(m.Content))
        case "tool":
            // For tool results, just create a user message that describes the result
            // This avoids ContentBlockParam compatibility issues
            resultMessage := fmt.Sprintf("Tool result for %s: %s", m.ToolID, m.Content)
            anthropicMsgs[i] = anthropic.NewUserMessage(anthropic.NewTextBlock(resultMessage))
        }
    }

    // Convert tools to Anthropic format
    anthropicTools := make([]anthropic.ToolUnionParam, len(tools))
    for i, t := range tools {
        td := agent.ToolDefinition{
            Name:        t.Name,
            Description: t.Description,
            InputSchema: t.InputSchema,
        }
        anthropicTools[i] = anthropic.ToolUnionParam{
            OfTool: &anthropic.ToolParam{
                Name:        t.Name,
                Description: anthropic.String(t.Description),
                InputSchema: c.buildSchema(td),
            },
        }
    }

    // Create params for the API call
    params := anthropic.MessageNewParams{
        Model:     anthropic.Model(c.model),
        MaxTokens: int64(1024),
        Messages:  anthropicMsgs,
    }

    // Add system prompt if provided
    if systemPrompt != "" {
        params.System = []anthropic.TextBlockParam{{Type: "text", Text: systemPrompt}}

        // Log the system prompt being sent
        c.logger.Info("sending system prompt to Claude",
            "model", c.model,
            "prompt_length", len(systemPrompt),
            "system_prompt", systemPrompt)
    } else {
        c.logger.Debug("no system prompt provided for Claude")
    }

    // Log tools information
    if len(anthropicTools) > 0 {
        toolNames := make([]string, len(anthropicTools))
        for i, tool := range anthropicTools {
            toolNames[i] = tool.OfTool.Name
        }
        c.logger.Debug("sending tools to Claude",
            "count", len(anthropicTools),
            "tools", strings.Join(toolNames, ", "))

        params.Tools = anthropicTools
    }

    // Call the Anthropic API
    start := time.Now()
    c.logger.Debug("sending request to Anthropic API", "message_count", len(anthropicMsgs))
    resp, err := c.client.Messages.New(ctx, params)
    latency := time.Since(start).Milliseconds()
    c.logger.Debug("Anthropic API response received", "latency_ms", latency)

    if err != nil {
        c.logger.Error("Anthropic API error", "error", err)
        return nil, nil, err
    }

    // Convert Anthropic response to our API-agnostic format
    var outMsgs []Message
    var toolCalls []ToolUse

    for _, block := range resp.Content {
        switch block.Type {
        case "text":
            outMsgs = append(outMsgs, Message{
                Role:    "assistant",
                Content: block.Text,
            })
        case "tool_use":
            toolCalls = append(toolCalls, ToolUse{
                ID:    block.ID,
                Name:  block.Name,
                Input: block.Input,
            })
        }
    }

    // Log the response summary
    c.logger.Debug("processed Claude response",
        "text_blocks", len(outMsgs),
        "tool_calls", len(toolCalls))

    return outMsgs, toolCalls, nil
}

/* ───────────────────────────────────
   SendToolResult – just reuse Chat
─────────────────────────────────── */

func (c *AnthropicClient) SendToolResult(ctx context.Context, msgs []Message, toolResults []ToolResult) ([]Message, []ToolUse, error) {
    // Just use the Chat method since we're already handling tool results via Messages
    // Pass empty system prompt since we want to maintain the previously set system prompt
    return c.Chat(ctx, msgs, nil, "")
}