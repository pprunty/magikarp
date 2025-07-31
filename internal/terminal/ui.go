package terminal

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	cfg "github.com/pprunty/magikarp/internal/config"
	"github.com/pprunty/magikarp/internal/orchestration"
)

// Debug logging for UI
var uiDebug = os.Getenv("MAGIKARP_DEBUG") == "1"
var uiDebugFile *os.File

// Global config for runtime modifications
var globalConfig *cfg.Config

// ToggleTools toggles the tools enabled/disabled state in the global config
func ToggleTools() {
	if globalConfig != nil {
		globalConfig.Tools.Enabled = !globalConfig.Tools.Enabled
	}
}

// GetToolsEnabled returns whether tools are currently enabled
func GetToolsEnabled() bool {
	if globalConfig != nil {
		return globalConfig.Tools.Enabled
	}
	return false
}

// GetToolsOutputEnabled returns whether tool output should be shown in the UI
func GetToolsOutputEnabled() bool {
	if globalConfig != nil {
		return globalConfig.Tools.Output
	}
	return false
}

func init() {
	if uiDebug {
		var err error
		uiDebugFile, err = os.OpenFile("magikarp_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err == nil {
			timestamp := time.Now().Format("2006/01/02 15:04:05")
			fmt.Fprintf(uiDebugFile, "%s [UI] Init: debug enabled\n", timestamp)
			uiDebugFile.Sync()
		}
	}
}

func uiDebugLog(format string, args ...interface{}) {
	if uiDebug && uiDebugFile != nil {
		timestamp := time.Now().Format("2006/01/02 15:04:05")
		fmt.Fprintf(uiDebugFile, "%s [UI] "+format+"\n", append([]interface{}{timestamp}, args...)...)
		uiDebugFile.Sync()
	}
}

// StartUI initializes and runs the Bubble Tea program
func StartUI() error {
	// Show welcome box with version and start directly with default model (first configured)
	fmt.Print(renderWelcomeBoxWithVersion() + "\n\n")

	// Load configuration
	cfgPath := "config.yaml"
	conf, err := cfg.LoadConfig(cfgPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Validate configuration (ensures default_model exists in provider list)
	if err := conf.ValidateConfig(); err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	// Set global config for runtime modifications
	globalConfig = conf

	// Initialise provider registry
	if err := orchestration.Init(conf); err != nil {
		return fmt.Errorf("initialising providers: %w", err)
	}

	var defaultModel string
	if conf.DefaultModel != "" {
		if _, err := orchestration.ProviderFor(conf.DefaultModel); err == nil {
			defaultModel = conf.DefaultModel
		} else {
			// Fallback to first available model if the configured one is not registered
			defaultModel, err = orchestration.FirstModel()
			if err != nil {
				return err // bubble up – UI can't continue without provider
			}
		}
	} else {
		var err error
		defaultModel, err = orchestration.FirstModel()
		if err != nil {
			return err // bubble up – UI can't continue without provider
		}
	}

	return startChatInput(defaultModel, conf)
}

// startChatInput launches the text input screen for the selected provider
func startChatInput(provider string, conf *cfg.Config) error {
	// Don't clear screen - let welcome box persist

	inputModel := NewInputModel(provider)

	for {
		p := tea.NewProgram(inputModel)

		finalModel, err := p.Run()
		if err != nil {
			return fmt.Errorf("failed to start chat input: %w", err)
		}

		// Check what happened with the input model
		if m, ok := finalModel.(InputModel); ok {
			if m.ShouldTriggerHelp() {
				// Show help screen
				if err := showHelpScreen(); err != nil {
					return fmt.Errorf("failed to show help screen: %w", err)
				}
				// Reset the help trigger and continue with chat
				inputModel = m
				inputModel.triggerHelpScreen = false
				continue
			} else if m.ShouldTriggerModelSelect() {
				// Show model selection screen
				selectedModel, err := showModelSelectScreen()
				if err != nil {
					return fmt.Errorf("failed to show model selection screen: %w", err)
				}
				// Reset the model selection trigger and continue with chat
				inputModel = m
				inputModel.triggerModelSelect = false
				// Update provider if a model was selected
				if selectedModel != "" {
					inputModel.provider = selectedModel
					provider = selectedModel
				}
				continue
			} else if m.quitting {
				// User wants to quit the session
				break
			} else if m.message != "" {
				// Message processing is now handled asynchronously in input.go
				// Just continue with the same session
				inputModel = m
				continue
			}
		}

		// Something unexpected happened, break to avoid infinite loop
		break
	}

	return nil
}

// showHelpScreen displays the full-screen help interface
func showHelpScreen() error {
	helpModel := NewHelpModel()
	p := tea.NewProgram(helpModel, tea.WithAltScreen())

	_, err := p.Run()
	if err != nil {
		return fmt.Errorf("failed to run help screen: %w", err)
	}

	return nil
}

// showModelSelectScreen displays the full-screen model selection interface
func showModelSelectScreen() (string, error) {
	modelSelectModel := NewModelSelectModel()
	p := tea.NewProgram(modelSelectModel, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("failed to run model selection screen: %w", err)
	}

	// Extract the selected model
	if m, ok := finalModel.(ModelSelectModel); ok {
		return m.GetSelectedModel(), nil
	}

	return "", nil
}

// StartUIWithoutAltScreen runs the UI without alternative screen mode
// Useful for development or when you want to preserve terminal history
func StartUIWithoutAltScreen() error {
	model := NewMenuModel()
	p := tea.NewProgram(model)

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to start UI: %w", err)
	}

	return nil
}

// CheckTerminalCapabilities ensures the terminal supports required features
func CheckTerminalCapabilities() error {
	// Check if we're running in a terminal
	if !isatty(os.Stdout.Fd()) {
		return fmt.Errorf("magikarp requires a terminal to run")
	}

	return nil
}

// isatty checks if the file descriptor is a terminal
func isatty(_ uintptr) bool {
	// Simple check - in a real implementation you might want to use
	// a more robust terminal detection library
	return os.Getenv("TERM") != ""
}
