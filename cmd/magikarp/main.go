package main

import (
    "context"
    "encoding/json"
    "fmt"
    "os"
    "strings"

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

// chatAgent orchestrates dialogue between the user, the LLM and tool plugins.
type chatAgent struct {
    client          llm.Client
    getUserMessage  func() (string, bool)
    tools           []agent.ToolDefinition
    showToolResults bool
}

func newChatAgent(client llm.Client, getUserMessage func() (string, bool), tools []agent.ToolDefinition) *chatAgent {
    return &chatAgent{
        client:          client,
        getUserMessage:  getUserMessage,
        tools:           tools,
        showToolResults: false,
    }
}

func (a *chatAgent) Run(ctx context.Context) error {
    conversation := []llm.Message{}

    fmt.Printf("Chat with %s (ctrl‑C to quit)\n", a.client.Name())
    fmt.Println("Tip: type 'show tools' to toggle tool result visibility.")

    readUserInput := true
    for {
        if readUserInput {
            fmt.Print("\u001b[94mYou\u001b[0m: ")
            userInput, ok := a.getUserMessage()
            if !ok {
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

            conversation = append(conversation, llm.Message{
                Role:    "user",
                Content: userInput,
            })
        }

        // prepare LLM tool descriptions
        llmTools := make([]llm.Tool, len(a.tools))
        for i, t := range a.tools {
            llmTools[i] = convertToolDefinitionToTool(t)
        }

        // ask the model for a reply + optional tool calls
        assistantMsgs, toolCalls, err := a.client.Chat(ctx, conversation, llmTools)
        if err != nil {
            return err
        }
        conversation = append(conversation, assistantMsgs...)

        for _, m := range assistantMsgs {
            if m.Content != "" {
                fmt.Printf("\u001b[93mAI\u001b[0m: %s\n", m.Content)
            }
        }

        if len(toolCalls) == 0 {
            readUserInput = true
            continue
        }

        // execute requested tools
        toolResults := make([]llm.ToolResult, 0, len(toolCalls))
        for _, call := range toolCalls {
            res := a.executeTool(call.ID, call.Name, call.Input)
            toolResults = append(toolResults, res)

            if a.showToolResults {
                pretty := prettyJSON(res.Content)
                color := "92" // green
                if res.IsError {
                    color = "91"
                }
                fmt.Printf("\u001b[%smtool result\u001b[0m: %s\n", color, pretty)
            }
        }

        // send results back to the LLM
        assistantMsgs, _, err = a.client.SendToolResult(ctx, conversation, toolResults)
        if err != nil {
            return err
        }
        conversation = append(conversation, assistantMsgs...)

        for _, m := range assistantMsgs {
            if m.Content != "" {
                fmt.Printf("\u001b[93mAI\u001b[0m: %s\n", m.Content)
            }
        }
        readUserInput = true
    }
    return nil
}

func (a *chatAgent) executeTool(id, name string, input json.RawMessage) llm.ToolResult {
    for _, t := range a.tools {
        if t.Name == name {
            var inputData map[string]interface{}
            if err := json.Unmarshal(input, &inputData); err != nil {
                return llm.ToolResult{ID: id, Content: err.Error(), IsError: true}
            }
            
            // Format the input arguments for display
            argsStr := "{}"
            if len(inputData) > 0 {
                argsBytes, _ := json.Marshal(inputData)
                argsStr = string(argsBytes)
            }
            
            fmt.Printf("\u001b[96m%s(%s)\u001b[0m\n", name, argsStr)
            
            result, err := t.Function(context.Background(), inputData)
            if err != nil {
                return llm.ToolResult{ID: id, Content: err.Error(), IsError: true}
            }
            return convertAgentToolResultToLLMToolResult(*result)
        }
    }
    return llm.ToolResult{ID: id, Content: "tool not found", IsError: true}
}

// Helper functions to convert between agent and llm types
func convertToolDefinitionToTool(def agent.ToolDefinition) llm.Tool {
    return llm.Tool{
        Name:        def.Name,
        Description: def.Description,
        InputSchema: def.InputSchema,
    }
}

func convertAgentToolResultToLLMToolResult(result agent.ToolResult) llm.ToolResult {
    return llm.ToolResult{
        ID:      result.ID,
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

// ---- CLI -------------------------------------------------------------

type ProviderOption struct {
    ID, Name string
    Required bool
}

type ModelOption struct {
    Name, Description string
}

func main() {
    // Load .env file
    if err := godotenv.Load(); err != nil {
        fmt.Fprintf(os.Stderr, "warning: .env file not found: %v\n", err)
    }

    // --- config ----------------------------------------------------------------
    cfg, err := config.LoadConfig("config.json")
    if err != nil {
        fmt.Fprintf(os.Stderr, "config error: %v\n", err)
        os.Exit(1)
    }

    // --- provider selection ----------------------------------------------------
    providerOptions := buildProviderOptions(cfg)
    providerIdx := selectPrompt("Select LLM provider (← to go back)", providerItems(providerOptions))
    selectedProvider := providerOptions[providerIdx]
    providerCfg, _ := cfg.GetProvider(selectedProvider.ID)

    // --- model selection -------------------------------------------------------
    modelIdx := selectPrompt(fmt.Sprintf("Select %s model (← to go back)", providerCfg.Name),
        modelItems(providerCfg.Models))
    selectedModel := providerCfg.Models[modelIdx]

    client, err := createClient(selectedProvider.ID, selectedModel.Name)
    if err != nil {
        fmt.Fprintf(os.Stderr, "%v\n", err)
        os.Exit(1)
    }

    // --- agent selection -------------------------------------------------------
    agentDefs, err := agents.All()
    if err != nil {
        fmt.Fprintf(os.Stderr, "failed to load agents: %v\n", err)
        os.Exit(1)
    }

    agentIdx := selectPrompt("Select Agent (auto default)", agentItems(agentDefs))
    chosenAgent := agentDefs[agentIdx]

    // Build tool list based on agent definition
    pluginMap := map[string]agent.Plugin{}
    for _, p := range loader.All() {
        pluginMap[p.Name()] = p
    }
    var tools []agent.ToolDefinition
    if chosenAgent.Name == "auto" {
        for _, p := range pluginMap {
            tools = append(tools, p.Tools()...)
        }
    } else {
        for _, pluginName := range chosenAgent.Plugins {
            if p, ok := pluginMap[pluginName]; ok {
                tools = append(tools, p.Tools()...)
            } else {
                fmt.Fprintf(os.Stderr, "warning: plugin %s referenced by agent %s not found\n",
                    pluginName, chosenAgent.Name)
            }
        }
    }

    // --- history --------------------------------------------------------------
    hist, err := history.NewHistoryManager()
    if err != nil {
        fmt.Fprintf(os.Stderr, "history error: %v\n", err)
        os.Exit(1)
    }
    defer hist.Close()

    getUserMessage := func() (string, bool) {
        line, err := hist.ReadLine()
        if err != nil {
            return "", false
        }
        return strings.TrimSpace(line), true
    }

    ca := newChatAgent(client, getUserMessage, tools)
    if err := ca.Run(context.Background()); err != nil {
        fmt.Fprintf(os.Stderr, "error: %v\n", err)
        os.Exit(1)
    }
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