package terminal

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// StartUI initializes and runs the Bubble Tea program
func StartUI() error {
	// Show welcome box with version and start directly with default model (Claude)
	fmt.Print(renderWelcomeBoxWithVersion() + "\n\n")
	
	// Start chat input with default model (first available model)
	defaultModel := "claude-3-5-sonnet-20240620" // Default fallback
	availableModels := GetAvailableModels()
	if len(availableModels) > 0 {
		defaultModel = availableModels[0]
	}
	return startChatInput(defaultModel)
}

// startChatInput launches the text input screen for the selected provider
func startChatInput(provider string) error {
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
				}
				continue
			} else if m.quitting {
				// User wants to quit the session
				break
			} else if m.message != "" {
				// User entered a message, process it and continue
				fmt.Printf("\nProvider: %s\nMessage: %s\n", provider, m.message)
				// TODO: Here you would integrate with the actual AI provider
				// Continue with the same session
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