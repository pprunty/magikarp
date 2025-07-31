package control_state

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	cfg "github.com/pprunty/magikarp/internal/config"
	"github.com/pprunty/magikarp/internal/orchestration"
	"github.com/pprunty/magikarp/internal/providers"
	"github.com/pprunty/magikarp/internal/terminal"
	"github.com/pprunty/magikarp/internal/tools"
)

//go:embed tool.json
var schema []byte

// Input parameters for the tool.
type input struct {
	// action can be: toggle_tools, toggle_speech, switch_model
	Action string `json:"action"`
	// value is optional – e.g. "on"/"off" for toggles, or provider/model name
	Value string `json:"value,omitempty"`
}

func Definition() providers.ToolDefinition {
	var sch map[string]interface{}
	_ = json.Unmarshal(schema, &sch)

	return providers.ToolDefinition{
		Name:        sch["name"].(string),
		Description: sch["description"].(string),
		InputSchema: sch["input_schema"].(map[string]interface{}),
		Function:    run,
	}
}

func run(ctx context.Context, data map[string]interface{}) (*providers.ToolResult, error) {
	raw, _ := json.Marshal(data)
	var in input
	if err := json.Unmarshal(raw, &in); err != nil {
		return providers.NewToolResult("control_state", fmt.Sprintf("invalid input: %v", err), true), nil
	}

	action := strings.ToLower(strings.TrimSpace(in.Action))
	switch action {
	case "toggle_tools":
		desired := strings.ToLower(in.Value)
		enable := desired == "on" || desired == "enable" || desired == "true" || desired == "1"
		current := terminal.GetToolsEnabled()
		if current == enable {
			return providers.NewToolResult("control_state", fmt.Sprintf("Tools already %v", stateStr(current)), false), nil
		}
		terminal.ToggleTools()

		// Build list of user-visible tools (exclude core)
		var names []string
		for _, t := range tools.GetAllTools() {
			// skip core tools to avoid cluttering
			if t.Name == "control_state" || t.Name == "list_tools" || t.Name == "get_model_version" {
				continue
			}
			names = append(names, t.Name)
		}
		return providers.NewToolResult("control_state", fmt.Sprintf("Tools turned %v. Available tools: %s", stateStr(enable), strings.Join(names, ", ")), false), nil

	case "toggle_speech":
		desired := strings.ToLower(in.Value)
		enable := desired == "on" || desired == "enable" || desired == "true" || desired == "1"
		current := terminal.SpeechModeEnabled()
		if current == enable {
			return providers.NewToolResult("control_state", fmt.Sprintf("Speech-to-text already %v", stateStr(current)), false), nil
		}
		terminal.SetSpeechModeEnabled(enable)
		return providers.NewToolResult("control_state", fmt.Sprintf("Speech-to-text turned %v", stateStr(enable)), false), nil

	case "switch_model":
		// value can be a full model id or provider alias (openai/anthropic/gemini)
		target := strings.TrimSpace(in.Value)
		if target == "" {
			return providers.NewToolResult("control_state", "value must specify model or provider", true), nil
		}
		// Check if target is exact model name already registered
		if _, err := orchestration.ProviderFor(target); err == nil {
			terminal.SetCurrentModel(target)
			return providers.NewToolResult("control_state", fmt.Sprintf("Switched to model %s", target), false), nil
		}
		// Otherwise treat as provider alias and pick first model from config
		conf, err := cfg.LoadConfig("config.yaml")
		if err != nil {
			return providers.NewToolResult("control_state", fmt.Sprintf("failed to load config: %v", err), true), nil
		}
		pCfg, ok := conf.Providers[strings.ToLower(target)]
		if !ok || len(pCfg.Models) == 0 {
			return providers.NewToolResult("control_state", fmt.Sprintf("unknown provider or no models for %s", target), true), nil
		}
		chosen := pCfg.Models[0]
		terminal.SetCurrentModel(chosen)
		return providers.NewToolResult("control_state", fmt.Sprintf("Switched to provider %s (model %s)", target, chosen), false), nil
	default:
		return providers.NewToolResult("control_state", "unknown action", true), nil
	}
}

func stateStr(b bool) string {
	if b {
		return "on"
	}
	return "off"
}
