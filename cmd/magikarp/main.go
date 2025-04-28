package main

import (
    "context"
    "encoding/json"
    "fmt"
    "os"
    "strings"
    "log/slog"

    "github.com/joho/godotenv"
    "github.com/manifoldco/promptui"
    "github.com/pprunty/magikarp/pkg/agent"
    "github.com/pprunty/magikarp/pkg/agents"
    "github.com/pprunty/magikarp/pkg/config"
    "github.com/pprunty/magikarp/pkg/history"
    "github.com/pprunty/magikarp/pkg/loader"
    "github.com/pprunty/magikarp/pkg/llm"
    _ "github.com/pprunty/magikarp/pkg/plugins/all"
)

// LogLevel controls the verbosity of logging
type LogLevel string

const (
    LogLevelDebug LogLevel = "debug"
    LogLevelInfo  LogLevel = "info"
    LogLevelWarn  LogLevel = "warn"
    LogLevelError LogLevel = "error"
)

type chatAgent struct {
    client          llm.Client
    getUserMessage  func() (string, bool)
    tools           []agent.ToolDefinition
    showToolResults bool
    logger          *slog.Logger
    systemPrompt    string // Added system prompt field
}

func newChatAgent(client llm.Client, getUserMessage func() (string, bool), tools []agent.ToolDefinition, logger *slog.Logger, systemPrompt string) *chatAgent {
    return &chatAgent{
        client:          client,
        getUserMessage:  getUserMessage,
        tools:           tools,
        showToolResults: false,
        logger:          logger,
        systemPrompt:    systemPrompt, // Store the system prompt
    }
}

func (a *chatAgent) Run(ctx context.Context) error {
    conversation := []llm.Message{}

    a.logger.Info("starting chat session", "provider", a.client.Name())
    fmt.Printf("Chat with %s (ctrl-C to quit)\n", a.client.Name())
    fmt.Println("Tip: type 'show tools' to toggle tool result visibility.")

    // Use readUserInput flag to control conversation flow
    readUserInput := true
    for {
        if readUserInput {
            fmt.Print("\u001b[94mYou\u001b[0m: ")
            userInput, ok := a.getUserMessage()
            if !ok {
                a.logger.Info("user terminated chat session")
                break
            }
            if strings.TrimSpace(strings.ToLower(userInput)) == "show tools" {
                a.showToolResults = !a.showToolResults
                if a.showToolResults {
                    fmt.Println("\u001b[93mAI\u001b[0m: tool results will now be shown.")
                } else {
                    fmt.Println("\u001b[93mAI\u001b[0m: tool results will now be hidden.")
                }
                continue
            }
            conversation = append(conversation, llm.Message{Role: "user", Content: userInput})
        }

        // Convert tools to LLM format
        llmTools := make([]llm.Tool, len(a.tools))
        for i, t := range a.tools {
            llmTools[i] = llm.Tool{
                Name:        t.Name,
                Description: t.Description,
                InputSchema: t.InputSchema,
            }
        }

        // Get response from the LLM - now passing system prompt
        assistantMsgs, toolCalls, err := a.client.Chat(ctx, conversation, llmTools, a.systemPrompt)
        if err != nil {
            return err
        }

        // Add assistant's response to conversation
        conversation = append(conversation, assistantMsgs...)

        // Display assistant's response
        for _, m := range assistantMsgs {
            if m.Content != "" {
                fmt.Printf("\u001b[93mAI\u001b[0m: %s\n", m.Content)
            }
        }

        // If no tool calls, get new user input
        if len(toolCalls) == 0 {
            readUserInput = true
            continue
        }

        // Process tool calls and execute them
        var toolResults []llm.Message
        for _, call := range toolCalls {
            // Execute the tool
            result := a.executeTool(call.ID, call.Name, call.Input)

            // Show tool result if enabled
            if a.showToolResults {
                color := "92"
                if result.IsError {
                    color = "91"
                }
                fmt.Printf("\u001b[%smtool result\u001b[0m: %s\n", color, prettyJSON(result.Content))
            }

            // Create a tool result message
            toolResults = append(toolResults, llm.Message{
                Role:    "tool",
                Content: result.Content,
                ToolID:  call.ID,
            })
        }

        // Add all tool results to the conversation
        conversation = append(conversation, toolResults...)

        // Continue without user input
        readUserInput = false
    }

    return nil
}

func (a *chatAgent) executeTool(id, name string, input json.RawMessage) llm.ToolResult {
    for _, t := range a.tools {
        if t.Name == name {
            var inputData map[string]interface{}
            if err := json.Unmarshal(input, &inputData); err != nil {
                a.logger.Error("failed to unmarshal tool input", "error", err, "toolName", name)
                return llm.ToolResult{ID: id, Content: err.Error(), IsError: true}
            }

            // Format the input arguments for display
            argsStr := "{}"
            if len(inputData) > 0 {
                argsBytes, _ := json.Marshal(inputData)
                argsStr = string(argsBytes)
            }

            // Print tool execution
            fmt.Printf("\u001b[96m%s(%s)\u001b[0m\n", name, argsStr)

            // Execute the tool
            result, err := t.Function(context.Background(), inputData)
            if err != nil {
                a.logger.Error("tool execution failed", "error", err, "toolName", name)
                return llm.ToolResult{ID: id, Content: err.Error(), IsError: true}
            }

            // Return the result
            return llm.ToolResult{
                ID:      id,
                Content: result.Content,
                IsError: result.IsError,
            }
        }
    }

    a.logger.Warn("tool not found", "toolName", name)
    return llm.ToolResult{ID: id, Content: "tool not found", IsError: true}
}


// Modified to ensure tool_id is properly set
func convertAgentToolResultToLLMToolResult(result agent.ToolResult, id string) llm.ToolResult {
    return llm.ToolResult{
        ID:      id, // Ensure we use the ID from the tool_use request
        Content: result.Content,
        IsError: result.IsError,
    }
}

func prettyJSON(raw string) string {
    var data interface{}
    if json.Unmarshal([]byte(raw), &data) != nil {
        return raw // not JSON
    }
    b, err := json.MarshalIndent(data, "", "  ")
    if err != nil {
        return raw
    }
    return string(b)
}

// setupLogger creates a structured logger with the specified log level
func setupLogger(level LogLevel) *slog.Logger {
    var logLevel slog.Level
    switch level {
    case LogLevelDebug:
        logLevel = slog.LevelDebug
    case LogLevelInfo:
        logLevel = slog.LevelInfo
    case LogLevelWarn:
        logLevel = slog.LevelWarn
    case LogLevelError:
        logLevel = slog.LevelError
    default:
        logLevel = slog.LevelInfo
    }

    opts := &slog.HandlerOptions{
        Level: logLevel,
    }

    handler := slog.NewJSONHandler(os.Stderr, opts)
    return slog.New(handler)
}

// ---- CLI -------------------------------------------------------------

type ProviderOption struct {
    ID, Name string
    Required bool
}

type ModelOption struct {
    Name, Description string
}

func main() {
    // Set up logging
    logLevel := LogLevel(os.Getenv("MAGIKARP_LOG_LEVEL"))
    if logLevel == "" {
        logLevel = LogLevelInfo
    }
    logger := setupLogger(logLevel)
    logger.Info("starting magikarp", "version", "0.1.0", "logLevel", logLevel)

    // Load .env file
    if err := godotenv.Load(); err != nil {
        logger.Warn("failed to load .env file", "error", err)
        fmt.Fprintf(os.Stderr, "warning: .env file not found: %v\n", err)
    }

    // --- config ----------------------------------------------------------------
    logger.Debug("loading configuration")
    cfg, err := config.LoadConfig("config.json")
    if err != nil {
        logger.Error("failed to load config", "error", err)
        fmt.Fprintf(os.Stderr, "config error: %v\n", err)
        os.Exit(1)
    }

    // --- provider selection ----------------------------------------------------
    logger.Debug("building provider options")
    providerOptions := buildProviderOptions(cfg)
    providerIdx := selectPrompt("Select LLM provider (← to go back)", providerItems(providerOptions))
    selectedProvider := providerOptions[providerIdx]
    logger.Info("selected provider", "id", selectedProvider.ID, "name", selectedProvider.Name)

    providerCfg, _ := cfg.GetProvider(selectedProvider.ID)

    // --- model selection -------------------------------------------------------
    modelIdx := selectPrompt(fmt.Sprintf("Select %s model (← to go back)", providerCfg.Name),
        modelItems(providerCfg.Models))
    selectedModel := providerCfg.Models[modelIdx]
    logger.Info("selected model", "name", selectedModel.Name)

    logger.Debug("creating LLM client")
    client, err := createClient(selectedProvider.ID, selectedModel.Name)
    if err != nil {
        logger.Error("failed to create client", "error", err)
        fmt.Fprintf(os.Stderr, "%v\n", err)
        os.Exit(1)
    }

    // --- agent selection -------------------------------------------------------
    logger.Debug("loading agent definitions")
    agentDefs, err := agents.All()
    if err != nil {
        logger.Error("failed to load agents", "error", err)
        fmt.Fprintf(os.Stderr, "failed to load agents: %v\n", err)
        os.Exit(1)
    }

    agentIdx := selectPrompt("Select Agent (auto default)", agentItems(agentDefs))
    chosenAgent := agentDefs[agentIdx]
    logger.Info("selected agent", "name", chosenAgent.Name)

    // Get system prompt from the chosen agent
    systemPrompt := chosenAgent.SystemPrompt
    if systemPrompt != "" {
        logger.Info("using agent-specific system prompt", "agentName", chosenAgent.Name)
    } else {
        // Default system prompt if none specified
        systemPrompt = "You are a helpful AI assistant with access to various tools. Use your tools when appropriate to assist the user."
        logger.Info("using default system prompt", "agentName", chosenAgent.Name)
    }

    // Build tool list based on agent definition
    logger.Debug("loading plugins")
    pluginMap := map[string]agent.Plugin{}
    for _, p := range loader.All() {
        pluginMap[p.Name()] = p
        logger.Debug("loaded plugin", "name", p.Name())
    }

    var tools []agent.ToolDefinition
    if chosenAgent.Name == "auto" {
        logger.Info("using all available tools")
        for _, p := range pluginMap {
            tools = append(tools, p.Tools()...)
        }
    } else {
        logger.Info("using tools from selected agent", "agentName", chosenAgent.Name)
        for _, pluginName := range chosenAgent.Plugins {
            if p, ok := pluginMap[pluginName]; ok {
                tools = append(tools, p.Tools()...)
                logger.Debug("added tools from plugin", "plugin", pluginName, "toolCount", len(p.Tools()))
            } else {
                logger.Warn("plugin referenced by agent not found", "plugin", pluginName, "agent", chosenAgent.Name)
                fmt.Fprintf(os.Stderr, "warning: plugin %s referenced by agent %s not found\n",
                    pluginName, chosenAgent.Name)
            }
        }
    }
    logger.Info("configured tools", "count", len(tools))

    // --- history --------------------------------------------------------------
    logger.Debug("initializing history manager")
    hist, err := history.NewHistoryManager()
    if err != nil {
        logger.Error("failed to create history manager", "error", err)
        fmt.Fprintf(os.Stderr, "history error: %v\n", err)
        os.Exit(1)
    }
    defer hist.Close()

    getUserMessage := func() (string, bool) {
        line, err := hist.ReadLine()
        if err != nil {
            logger.Debug("error reading line from history", "error", err)
            return "", false
        }
        return strings.TrimSpace(line), true
    }

    logger.Info("starting chat agent")
    // Pass system prompt to the chat agent
    ca := newChatAgent(client, getUserMessage, tools, logger, systemPrompt)
    if err := ca.Run(context.Background()); err != nil {
        logger.Error("chat agent failed", "error", err)
        fmt.Fprintf(os.Stderr, "error: %v\n", err)
        os.Exit(1)
    }
    logger.Info("chat agent terminated normally")
}

// ---------------- helper prompts & utils -------------------------------------

func providerItems(opts []ProviderOption) []string {
    items := make([]string, len(opts))
    for i, o := range opts {
        items[i] = o.Name
    }
    return items
}

func modelItems(models []config.Model) []string {
    items := make([]string, len(models))
    for i, m := range models {
        items[i] = fmt.Sprintf("%s – %s", m.Name, m.Description)
    }
    return items
}

func agentItems(defs []agents.Definition) []string {
    items := make([]string, len(defs))
    for i, d := range defs {
        items[i] = fmt.Sprintf("%s – %s", d.Name, d.Description)
    }
    return items
}

func selectPrompt(label string, items []string) int {
    p := promptui.Select{Label: label, Items: items}
    idx, _, err := p.Run()
    if err != nil {
        fmt.Printf("prompt failed: %v\n", err)
        os.Exit(1)
    }
    return idx
}

func buildProviderOptions(cfg *config.Config) []ProviderOption {
    order := []string{"auto", "anthropic", "openai", "gemini", "ollama"}
    opts := make([]ProviderOption, 0, len(order))
    for _, id := range order {
        if prov, ok := cfg.Providers[id]; ok {
            opts = append(opts, ProviderOption{ID: id, Name: prov.Name, Required: prov.Required})
        }
    }
    return opts
}

func createClient(providerID, modelName string) (llm.Client, error) {
    switch providerID {
    case "auto":
        return llm.NewAutoClient([]string{modelName})
    case "anthropic":
        return llm.NewAnthropicClient(modelName)
    case "openai":
        return llm.NewOpenAIClient(modelName)
    case "gemini":
        return llm.NewGeminiClient(modelName)
    case "ollama":
        return llm.NewOllamaClient(modelName)
    default:
        return nil, fmt.Errorf("unknown provider %s", providerID)
    }
}