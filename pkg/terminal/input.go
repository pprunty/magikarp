package terminal

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// InputModel represents the text input state
type InputModel struct {
	textInput           textinput.Model
	provider            string
	quitting            bool
	message             string
	width               int
	height              int
	messages            []string // Store conversation history
	historyManager      *HistoryManager
	historyIndex        int       // Current position in history (newest = len-1)
	inHistoryMode       bool      // Whether we're navigating history
	originalInput       string    // Store original input when entering history mode
	ctrlCPressed        bool      // Track if Ctrl+C was recently pressed
	ctrlCTime           time.Time // When Ctrl+C was pressed
	showExitPrompt      bool      // Show the exit prompt message
	showingSlashCommands bool          // Whether slash command menu is visible
	slashCommandCursor  int            // Current position in slash command menu
	availableCommands   []SlashCommand // Available slash commands
	filteredCommands    []SlashCommand // Filtered slash commands based on input
	triggerHelpScreen    bool      // Whether to trigger help screen
	triggerModelSelect   bool      // Whether to trigger model selection screen
	speechMode          bool      // Whether speech mode is enabled
}

// NewInputModel creates a new input model for the selected provider
func NewInputModel(provider string) InputModel {
	ti := textinput.New()
	ti.Placeholder = ""
	ti.Focus()
	ti.CharLimit = 500
	ti.Width = 76

	// Initialize history manager
	histManager, err := NewHistoryManager()
	if err != nil {
		// If history fails to initialize, continue without it
		histManager = nil
	}

	return InputModel{
		textInput:           ti,
		provider:            provider,
		width:               80,         // Default width
		height:              24,         // Default height
		messages:            []string{}, // Initialize empty message history
		historyManager:      histManager,
		historyIndex:        -1, // Not in history mode
		inHistoryMode:       false,
		showingSlashCommands: false,
		slashCommandCursor:  0,
		availableCommands:   GetAvailableCommands(),
		filteredCommands:    GetAvailableCommands(),
		triggerHelpScreen:    false,
		triggerModelSelect:   false,
		speechMode:          false, // Speech mode starts disabled
	}
}

// timeoutMsg is sent when the Ctrl+C timeout expires
type timeoutMsg struct{}

func (m InputModel) Init() tea.Cmd {
	return textinput.Blink
}

// timeoutCmd returns a command that sends a timeout message after 2 seconds
func timeoutCmd() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return timeoutMsg{}
	})
}

func (m InputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case timeoutMsg:
		// Timeout expired, reset Ctrl+C state
		m.ctrlCPressed = false
		m.showExitPrompt = false
		return m, nil
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Update text input width to fit the new terminal width
		// Account for border (2 chars) + padding (2 chars) + margin (2 chars)
		m.textInput.Width = max(18, m.width-6)
	// Remove mouse scroll handling - let terminal handle it naturally
	case tea.KeyMsg:
		// Handle specific slash command navigation keys
		if m.showingSlashCommands {
			switch msg.String() {
			case "up", "k":
				if len(m.filteredCommands) > 0 {
					m.slashCommandCursor--
					if m.slashCommandCursor < 0 {
						m.slashCommandCursor = len(m.filteredCommands) - 1
					}
				}
				return m, nil
			case "down", "j":
				if len(m.filteredCommands) > 0 {
					m.slashCommandCursor++
					if m.slashCommandCursor >= len(m.filteredCommands) {
						m.slashCommandCursor = 0
					}
				}
				return m, nil
			case "enter":
				if len(m.filteredCommands) > 0 && m.slashCommandCursor < len(m.filteredCommands) {
					selectedCommand := m.filteredCommands[m.slashCommandCursor]
					m.showingSlashCommands = false
					m.textInput.SetValue("")
					
					switch selectedCommand.Name {
					case "/exit":
						m.quitting = true
						return m, tea.Quit
					case "/help":
						m.triggerHelpScreen = true
						return m, tea.Quit
					case "/model":
						m.triggerModelSelect = true
						return m, tea.Quit
					case "/speech":
						m.speechMode = !m.speechMode
						return m, nil
					}
				}
				return m, nil
			case "esc":
				m.showingSlashCommands = false
				m.textInput.SetValue("")
				return m, nil
			}
			// For all other keys, continue to normal input processing
		}
		
		// Handle regular input
		switch msg.String() {
		case "ctrl+c":
			if m.ctrlCPressed && time.Since(m.ctrlCTime) <= 2*time.Second {
				// Second Ctrl+C within timeout window - exit
				m.quitting = true
				return m, tea.Quit
			} else {
				// First Ctrl+C or timeout expired - clear input and show prompt
				m.textInput.SetValue("")
				m.ctrlCPressed = true
				m.ctrlCTime = time.Now()
				m.showExitPrompt = true
				// Exit history mode if active
				if m.inHistoryMode {
					m.exitHistoryMode()
				}
				return m, timeoutCmd()
			}
		case "enter":
			// Reset Ctrl+C state on any other action
			m.ctrlCPressed = false
			m.showExitPrompt = false

			if m.textInput.Value() != "" {
				// Check if user typed "exit" to quit
				if m.textInput.Value() == "exit" {
					m.quitting = true
					return m, tea.Quit
				}
				
				// Check if user typed "help" to show help screen
				if m.textInput.Value() == "help" {
					m.triggerHelpScreen = true
					return m, tea.Quit
				}
				
				// Add message to conversation history
				m.messages = append(m.messages, m.textInput.Value())
				m.message = m.textInput.Value()

				// Save to input history
				if m.historyManager != nil {
					m.historyManager.AddMessage(m.textInput.Value())
				}

				// Exit history mode if we were in it
				if m.inHistoryMode {
					m.exitHistoryMode()
				}

				// Clear the input for next message
				m.textInput.SetValue("")
			}
		case "up":
			// Reset Ctrl+C state on any other action
			m.ctrlCPressed = false
			m.showExitPrompt = false

			// Navigate to previous message in history
			if m.historyManager != nil {
				m.navigateHistory(-1)
			}
		case "down":
			// Reset Ctrl+C state on any other action
			m.ctrlCPressed = false
			m.showExitPrompt = false

			// Navigate to next message in history (only if in history mode)
			if m.historyManager != nil && m.inHistoryMode {
				m.navigateHistory(1)
			}
		default:
			// Reset Ctrl+C state on any other key press
			m.ctrlCPressed = false
			m.showExitPrompt = false
		}
	}

	// Check if user is typing (exit history mode if so)
	if m.inHistoryMode {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			// Exit history mode on any regular typing
			switch msg.String() {
			case "ctrl+c", "esc", "enter", "up", "down":
				// Don't exit history mode for these keys
			default:
				// User is typing, exit history mode
				m.exitHistoryMode()
			}
		}
	}

	// Update text input first to allow continued typing
	m.textInput, cmd = m.textInput.Update(msg)
	
	inputValue := m.textInput.Value()
	
	// Check if user typed "/" to trigger slash commands or is typing a slash command
	if strings.HasPrefix(inputValue, "/") {
		if !m.showingSlashCommands {
			m.showingSlashCommands = true
		}
		
		// Filter commands based on current input
		m.filteredCommands = FilterCommands(inputValue)
		
		// Reset cursor if it's out of bounds due to filtering
		if m.slashCommandCursor >= len(m.filteredCommands) {
			m.slashCommandCursor = 0
		}
	} else if m.showingSlashCommands && !strings.HasPrefix(inputValue, "/") {
		// Hide slash commands if user deleted the "/"
		m.showingSlashCommands = false
	}
	
	return m, cmd
}

// ShouldTriggerHelp returns true if help screen should be triggered
func (m InputModel) ShouldTriggerHelp() bool {
	return m.triggerHelpScreen
}

// ShouldTriggerModelSelect returns true if model selection screen should be triggered
func (m InputModel) ShouldTriggerModelSelect() bool {
	return m.triggerModelSelect
}

// formatSlashCommand formats a slash command with aligned description
func formatSlashCommand(command, description string) string {
	// Define the width for command alignment (like Claude Code)
	const alignmentWidth = 20
	
	// Calculate padding needed to align descriptions
	commandLength := len(stripANSI(command))
	padding := alignmentWidth - commandLength
	if padding < 0 {
		padding = 1 // At least one space
	}
	
	paddingStr := strings.Repeat(" ", padding)
	return "  " + command + paddingStr + description
}

// stripANSI removes ANSI color codes to get actual string length
func stripANSI(s string) string {
	// Simple regex to remove common ANSI escape sequences
	// This is a basic implementation for length calculation
	result := ""
	inEscape := false
	for _, r := range s {
		if r == '\033' {
			inEscape = true
			continue
		}
		if inEscape {
			if r == 'm' {
				inEscape = false
			}
			continue
		}
		result += string(r)
	}
	return result
}

// max returns the larger of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Removed unused min function

// Removed maxScroll function - no longer needed for natural terminal flow

// navigateHistory moves through input history
func (m *InputModel) navigateHistory(direction int) {
	if m.historyManager == nil {
		return
	}

	historyCount := m.historyManager.GetHistoryCount()
	if historyCount == 0 {
		return
	}

	// If not in history mode, enter it and store current input
	if !m.inHistoryMode {
		m.inHistoryMode = true
		m.originalInput = m.textInput.Value()
		// Start from the most recent (newest) entry
		m.historyIndex = historyCount - 1
	} else {
		// Navigate within history
		m.historyIndex += direction
	}

	// Handle bounds
	if m.historyIndex < -1 {
		m.historyIndex = -1
		// Show original input when going before history
		m.textInput.SetValue(m.originalInput)
		m.inHistoryMode = false
		return
	} else if m.historyIndex >= historyCount {
		// Going past newest message - show empty input
		m.historyIndex = -1
		m.textInput.SetValue("")
		m.inHistoryMode = false
		return
	}

	// Set the text input to the history message
	if m.historyIndex >= 0 {
		historyMessage := m.historyManager.GetMessageAt(m.historyIndex)
		m.textInput.SetValue(historyMessage)
	}
}

// exitHistoryMode exits history navigation mode
func (m *InputModel) exitHistoryMode() {
	m.inHistoryMode = false
	m.historyIndex = -1
	m.originalInput = ""
}

func (m InputModel) View() string {
	if m.triggerHelpScreen || m.triggerModelSelect {
		// Don't show anything when triggering help or model selection screen
		return ""
	}
	
	if m.quitting {
		// Show conversation history on exit
		s := "\n"
		// Display all message history
		if len(m.messages) > 0 {
			for i, msg := range m.messages {
				s += messageStyle.Render(fmt.Sprintf("> %s", msg)) + "\n"
				s += aiResponseStyle.Render(fmt.Sprintf("⏺ Processing your request... (message %d)", i+1)) + "\n"
				s += "\n" // Blank line between exchanges
			}
		}
		return s
	}

	s := ""

	// Display all message history (natural terminal flow)
	if len(m.messages) > 0 {
		s += "\n"
		// Display all messages without viewport restrictions
		for i, msg := range m.messages {
			s += messageStyle.Render(fmt.Sprintf("> %s", msg)) + "\n"
			s += aiResponseStyle.Render(fmt.Sprintf("⏺ Processing your request... (message %d)", i+1)) + "\n"
			s += "\n" // Blank line between exchanges
		}
	} else {
		s += "\n"
	}

	// Add border around text input with dynamic width
	// Calculate exact width to prevent double borders
	availableWidth := max(20, m.width-4) // Account for border chars and margins
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("8")).
		Padding(0, 1).
		Width(availableWidth)

	inputWithBorder := borderStyle.Render(m.textInput.View())
	s += inputWithBorder
	s += "\n"
	
	// Show slash command menu if active
	if m.showingSlashCommands && len(m.filteredCommands) > 0 {
		s += "\n"
		for i, command := range m.filteredCommands {
			if i == m.slashCommandCursor {
				// Highlight selected command with purple color for both name and description
				commandPart := slashCommandActiveStyle.Render(command.Name)
				descPart := slashCommandActiveStyle.Render(command.Description)
				s += formatSlashCommand(commandPart, descPart) + "\n"
			} else {
				// Normal command display with gray color for both name and description
				commandPart := slashCommandNormalStyle.Render(command.Name)
				descPart := slashCommandNormalStyle.Render(command.Description)
				s += formatSlashCommand(commandPart, descPart) + "\n"
			}
		}
		s += "\n"
	}
	
	
	s += "\n"

	// Show specific model name based on provider with speech mode indicator
	modelName := GetModelDisplayName(m.provider)
	
	speechIndicator := ""
	if m.speechMode {
		speechIndicator = " " + speechModeOnStyle.Render("•") + " " + modelRunningStyle.Render("speech mode on")
	} else {
		speechIndicator = " " + speechModeOffStyle.Render("•") + " " + modelRunningStyle.Render("speech mode off")
	}

	s += modelRunningStyle.Render("• " + modelName) + speechIndicator
	s += "\n"

	// Show help text or exit prompt
	if m.showExitPrompt {
		s += exitPromptStyle.Render("Press Ctrl+C again to exit")
	} else if m.showingSlashCommands {
		s += helpStyle.Render("↑/↓: navigate • enter: select • esc: cancel")
	} else if m.inHistoryMode && m.historyManager != nil {
		s += helpStyle.Render("↑/↓: navigate • any key: exit history • ctrl+c: clear")
	} else {
		s += helpStyle.Render("↑/↓: history • /: commands • ctrl+c: clear")
	}
	s += "\n"

	return s
}

// Additional styling for input screen
var (
	chatHeaderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Bold(true).
			Italic(true)

	inputPromptStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Bold(false)

	modelRunningStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#626262"))

	messageStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFF00")).
			Bold(true)

	aiResponseStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575"))

	historyIndicatorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF6B35")).
				Bold(true)

	exitPromptStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B35")).
			Bold(true)

	slashCommandHeaderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Bold(true)

	helpDisplayStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))

	// Slash command specific styles
	slashCommandNormalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")) // Gray for normal items

	slashCommandActiveStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9B59B6")) // Purple for active items
	
	// Speech mode indicator styles
	speechModeOffStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")) // Red circle for speech mode off
	
	speechModeOnStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")) // Green circle for speech mode on
)
