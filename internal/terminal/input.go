package terminal

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pprunty/magikarp/internal/orchestration"
	"github.com/pprunty/magikarp/internal/providers"
	"github.com/pprunty/magikarp/internal/tools"
)

// wrapText wraps text to the specified width on word boundaries
func wrapText(text string, width int) string {
	if disableBeautify || width <= 0 {
		// Skip wrapping when beautification is disabled
		return text
	}

	// Preserve explicit newlines by splitting into paragraphs first
	paragraphs := strings.Split(text, "\n")
	var wrappedParagraphs []string

	for _, p := range paragraphs {
		p = strings.TrimSpace(p)
		if p == "" {
			wrappedParagraphs = append(wrappedParagraphs, "")
			continue
		}

		words := strings.Fields(p)
		var lines []string
		var currentLine strings.Builder
		lineLength := 0

		for _, word := range words {
			wordLen := len(word)
			if lineLength > 0 && lineLength+1+wordLen > width {
				lines = append(lines, currentLine.String())
				currentLine.Reset()
				lineLength = 0
			}

			if lineLength > 0 {
				currentLine.WriteString(" ")
				lineLength++
			}
			currentLine.WriteString(word)
			lineLength += wordLen
		}

		if currentLine.Len() > 0 {
			lines = append(lines, currentLine.String())
		}

		wrappedParagraphs = append(wrappedParagraphs, strings.Join(lines, "\n"))
	}

	return strings.Join(wrappedParagraphs, "\n")
}

// Debug logging for input handling
var inputDebug = os.Getenv("MAGIKARP_DEBUG") == "1"
var inputDebugFile *os.File

func init() {
	if inputDebug {
		var err error
		inputDebugFile, err = os.OpenFile("magikarp_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err == nil {
			timestamp := time.Now().Format("2006/01/02 15:04:05")
			fmt.Fprintf(inputDebugFile, "%s [Input] Init: debug enabled\n", timestamp)
			inputDebugFile.Sync()
		}
	}
}

func inputDebugLog(format string, args ...interface{}) {
	if inputDebug && inputDebugFile != nil {
		timestamp := time.Now().Format("2006/01/02 15:04:05")
		fmt.Fprintf(inputDebugFile, "%s [Input] "+format+"\n", append([]interface{}{timestamp}, args...)...)
		inputDebugFile.Sync()
	}
}

// ConversationPair represents a user message and AI response pair
type ConversationPair struct {
	UserMessage  string
	AIResponse   string
	IsProcessing bool // Whether this conversation is currently being processed
}

// Spinner state
var spinnerChars = []string{"◰", "◳", "◲", "◱"}
var currentSpinnerIndex = 0

// spinnerTickMsg is sent every 200ms to update the spinner
type spinnerTickMsg struct{}

// InputModel represents the text input state
type InputModel struct {
	textInput            textinput.Model
	provider             string
	quitting             bool
	message              string
	width                int
	height               int
	messages             []string           // Store user message history for input history
	conversation         []ConversationPair // Store full conversation
	historyManager       *HistoryManager
	historyIndex         int            // Current position in history (newest = len-1)
	inHistoryMode        bool           // Whether we're navigating history
	originalInput        string         // Store original input when entering history mode
	ctrlCPressed         bool           // Track if Ctrl+C was recently pressed
	ctrlCTime            time.Time      // When Ctrl+C was pressed
	showExitPrompt       bool           // Show the exit prompt message
	showingSlashCommands bool           // Whether slash command menu is visible
	slashCommandCursor   int            // Current position in slash command menu
	availableCommands    []SlashCommand // Available slash commands
	filteredCommands     []SlashCommand // Filtered slash commands based on input
	triggerHelpScreen    bool           // Whether to trigger help screen
	triggerModelSelect   bool           // Whether to trigger model selection screen
	speechMode           bool           // Whether speech mode is enabled
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
		textInput:            ti,
		provider:             provider,
		width:                80,         // Default width
		height:               24,         // Default height
		messages:             []string{}, // Initialize empty message history
		conversation:         []ConversationPair{},
		historyManager:       histManager,
		historyIndex:         -1, // Not in history mode
		inHistoryMode:        false,
		showingSlashCommands: false,
		slashCommandCursor:   0,
		availableCommands:    GetAvailableCommands(),
		filteredCommands:     GetAvailableCommands(),
		triggerHelpScreen:    false,
		triggerModelSelect:   false,
		speechMode:           false, // Speech mode starts disabled
	}
}

// aiResponseMsg is sent when we receive an AI response
type aiResponseMsg struct {
	response string
	isError  bool
}

// processingMsg is sent when we start processing a message
type processingMsg struct{}

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

// spinnerTickCmd returns a command that sends a spinner tick message every 200ms
func spinnerTickCmd() tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg {
		return spinnerTickMsg{}
	})
}

func (m InputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	inputDebugLog("Update called with msg type: %T", msg)

	switch msg := msg.(type) {
	case aiResponseMsg:
		// Received AI response, update the conversation
		if msg.isError {
			m.SetAIResponse(fmt.Sprintf("Error: %s", msg.response))
		} else {
			m.SetAIResponse(msg.response)
		}
		return m, nil
	case processingMsg:
		// Start processing - this is just for UI feedback
		return m, nil
	case timeoutMsg:
		// Timeout expired, reset Ctrl+C state
		m.ctrlCPressed = false
		m.showExitPrompt = false
		return m, nil
	case spinnerTickMsg:
		// Update spinner state
		currentSpinnerIndex++
		if currentSpinnerIndex >= len(spinnerChars) {
			currentSpinnerIndex = 0
		}

		// Continue ticking if we have any processing conversations
		hasProcessing := false
		for _, pair := range m.conversation {
			if pair.IsProcessing {
				hasProcessing = true
				break
			}
		}

		if hasProcessing {
			return m, spinnerTickCmd()
		}
		return m, nil
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Update text input width to fit the new terminal width
		// Account for border (2 chars) + padding (2 chars) + margin (2 chars)
		m.textInput.Width = max(18, m.width-6)
	// Remove mouse scroll handling - let terminal handle it naturally
	case tea.KeyMsg:
		inputDebugLog("KeyMsg received: %s", msg.String())
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
				inputDebugLog("Enter pressed in slash command mode")
				if len(m.filteredCommands) > 0 && m.slashCommandCursor < len(m.filteredCommands) {
					selectedCommand := m.filteredCommands[m.slashCommandCursor]
					
					// Save the slash command to history before executing it
					if m.historyManager != nil {
						m.historyManager.AddMessage(selectedCommand.Name)
					}
					
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
						SetSpeechModeEnabled(m.speechMode)
						// Update placeholder based on speech mode
						if m.speechMode {
							m.textInput.Placeholder = "Listening..."
						} else {
							m.textInput.Placeholder = ""
						}
						return m, nil
					case "/tools":
						// Toggle tools globally - call via exported function
						ToggleTools()
						// Add a user message to show the toggle status in the conversation
						if GetToolsEnabled() {
							m.AddConversationPair("/tools", "System: Tools enabled")
						} else {
							m.AddConversationPair("/tools", "System: Tools disabled")
						}
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
			inputDebugLog("Enter pressed, text input value: '%s'", m.textInput.Value())
			// Reset Ctrl+C state on any other action
			m.ctrlCPressed = false
			m.showExitPrompt = false

			if m.textInput.Value() != "" {
				inputDebugLog("Processing non-empty message")
				// Check if user typed "exit" to quit
				if m.textInput.Value() == "exit" {
					inputDebugLog("Exit command detected")
					m.quitting = true
					return m, tea.Quit
				}

				// Check if user typed "help" to show help screen
				if m.textInput.Value() == "help" {
					inputDebugLog("Help command detected")
					m.triggerHelpScreen = true
					return m, tea.Quit
				}

				// Add message to conversation history
				m.messages = append(m.messages, m.textInput.Value())
				userMessage := m.textInput.Value()

				// Add conversation pair with empty AI response initially
				m.AddConversationPair(userMessage, "")

				inputDebugLog("Message set to: '%s'", userMessage)

				// Save to input history
				if m.historyManager != nil {
					m.historyManager.AddMessage(userMessage)
				}

				// Exit history mode if we were in it
				if m.inHistoryMode {
					m.exitHistoryMode()
				}

				// Clear the input for next message
				m.textInput.SetValue("")
				inputDebugLog("Input cleared, starting AI processing")

				// Start async AI processing and spinner
				return m, tea.Batch(
					func() tea.Msg { return processingMsg{} },
					processMessageAsync(userMessage, m.provider),
					spinnerTickCmd(),
				)
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

// AddConversationPair adds a user message and AI response pair to the conversation
func (m *InputModel) AddConversationPair(userMsg, aiResponse string) {
	m.conversation = append(m.conversation, ConversationPair{
		UserMessage:  userMsg,
		AIResponse:   aiResponse,
		IsProcessing: aiResponse == "", // If no AI response yet, it's processing
	})
}

// SetAIResponse sets the AI response for the most recent conversation pair
func (m *InputModel) SetAIResponse(aiResponse string) {
	if len(m.conversation) > 0 {
		m.conversation[len(m.conversation)-1].AIResponse = aiResponse
		m.conversation[len(m.conversation)-1].IsProcessing = false
	}
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
		// Display all conversation pairs
		if len(m.conversation) > 0 {
			for _, pair := range m.conversation {
				// Wrap user message
				userMsg := wrapText(pair.UserMessage, m.width-6) // Account for "> " prefix and margins
				s += messageStyle.Render(fmt.Sprintf("> %s", userMsg)) + "\n"

				if pair.AIResponse != "" {
					// Wrap AI response
					aiMsg := wrapText(pair.AIResponse, m.width-6) // Account for "⏺ " prefix and margins
					s += aiResponseStyle.Render(fmt.Sprintf("⏺ %s", aiMsg)) + "\n"
				} else if pair.IsProcessing {
					s += aiResponseStyle.Render("Processing interrupted...") + "\n"
				}
				s += "\n" // Blank line between exchanges
			}
		}
		return s
	}

	s := ""

	// Display conversation history (natural terminal flow)
	if len(m.conversation) > 0 {
		s += "\n"
		// Display all conversation pairs
		for _, pair := range m.conversation {
			// Wrap user message
			userMsg := wrapText(pair.UserMessage, m.width-6) // Account for "> " prefix and margins
			s += messageStyle.Render(fmt.Sprintf("> %s", userMsg)) + "\n"

			if pair.AIResponse != "" {
				// Wrap AI response
				aiMsg := wrapText(pair.AIResponse, m.width-6) // Account for "⏺ " prefix and margins
				s += aiResponseStyle.Render(fmt.Sprintf("⏺ %s", aiMsg)) + "\n"
			} else if pair.IsProcessing {
				s += aiResponseStyle.Render(fmt.Sprintf("%s Processing...", spinnerChars[currentSpinnerIndex])) + "\n"
			}
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
	if SpeechModeEnabled() {
		speechIndicator = " " + speechModeOnStyle.Render("•") + " " + modelRunningStyle.Render("speech-to-text on")
	} else {
		speechIndicator = " " + speechModeOffStyle.Render("•") + " " + modelRunningStyle.Render("speech-to-text off")
	}

	// Tools indicator (reuse the same color styles)
	toolsIndicator := ""
	if GetToolsEnabled() {
		toolsIndicator = " " + speechModeOnStyle.Render("•") + " " + modelRunningStyle.Render("tools on")
	} else {
		toolsIndicator = " " + speechModeOffStyle.Render("•") + " " + modelRunningStyle.Render("tools off")
	}

	s += modelRunningStyle.Render("• "+modelName) + speechIndicator + toolsIndicator
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

// Add an init function after style vars to de-activate styling when beautification is disabled
func init() {
	if disableBeautify {
		plain := lipgloss.NewStyle()
		messageStyle = plain
		aiResponseStyle = plain
		modelRunningStyle = plain
		slashCommandNormalStyle = plain
		slashCommandActiveStyle = plain
		speechModeOnStyle = plain
		speechModeOffStyle = plain
	}
}

// processMessageAsync processes a user message with the AI provider asynchronously
func processMessageAsync(userMessage, provider string) tea.Cmd {
	return func() tea.Msg {
		// Get provider instance
		p, err := orchestration.ProviderFor(provider)
		if err != nil {
			return aiResponseMsg{
				response: fmt.Sprintf("Error getting provider: %v", err),
				isError:  true,
			}
		}

		// Load system prompt – prefer value from loaded config.yaml
		sysPrompt := "You are a helpful coding assistant."
		if globalConfig != nil && globalConfig.System != "" {
			sysPrompt = globalConfig.System
		}

		inputDebugLog("System prompt used: %s", sysPrompt)

		// Build messages
		messages := []providers.ChatMessage{
			{Role: providers.RoleSystem, Content: sysPrompt},
			{Role: providers.RoleUser, Content: userMessage},
		}

		// Get tools if enabled
		var providerTools []providers.Tool
		if GetToolsEnabled() {
			allTools := tools.GetAllTools()
			providerTools = make([]providers.Tool, len(allTools))
			for i, tool := range allTools {
				providerTools[i] = providers.Tool{
					Name:        tool.Name,
					Description: tool.Description,
					InputSchema: tool.InputSchema,
				}
			}
		} else {
			// Always expose core tools even when general tools are disabled
			core := tools.GetCoreTools()
			providerTools = make([]providers.Tool, len(core))
			for i, tool := range core {
				providerTools[i] = providers.Tool{
					Name:        tool.Name,
					Description: tool.Description,
					InputSchema: tool.InputSchema,
				}
			}
		}

		// update global current model for query tools
		SetCurrentModel(provider)

		// Call the provider
		assistantMsgs, toolCalls, err := p.Chat(context.Background(), messages, providerTools)
		if err != nil {
			return aiResponseMsg{
				response: fmt.Sprintf("Chat error: %v", err),
				isError:  true,
			}
		}

		// If tools requested, execute them
		if len(toolCalls) > 0 {
			var results []providers.ToolResult
			var used []string
			for _, call := range toolCalls {
				def, ok := tools.GetToolByName(call.Name)
				if !ok {
					results = append(results, providers.ToolResult{ID: call.ID, Content: "tool not found", IsError: true})
					continue
				}
				// parse input json
				var inputMap map[string]interface{}
				_ = json.Unmarshal(call.Input, &inputMap)
				res, _ := def.Function(context.Background(), inputMap)
				res.ID = call.ID
				results = append(results, *res)

				// Build display name with parameters, truncate if too long
				paramPreview := ""
				if len(inputMap) > 0 {
					if b, err := json.Marshal(inputMap); err == nil {
						s := string(b)
						if len(s) > 60 {
							s = s[:57] + "..."
						}
						paramPreview = "(" + s + ")"
					}
				}
				used = append(used, call.Name+paramPreview)
			}

			assistantMsgs, _, err = p.SendToolResult(context.Background(), append(messages, assistantMsgs...), results)
			if err != nil {
				return aiResponseMsg{response: fmt.Sprintf("Tool result error: %v", err), isError: true}
			}
			// Build summary line always
			summary := fmt.Sprintf("[Used tools: %s]", strings.Join(used, ", "))

			content := summary

			if GetToolsOutputEnabled() {
				// Build tool outputs string
				var toolOutputs []string
				for _, r := range results {
					prefix := ""
					if r.IsError {
						prefix = "(tool error) "
					} else {
						prefix = "(tool result) "
					}
					// Ensure multi-line content is indented nicely
					lines := strings.Split(strings.TrimSpace(r.Content), "\n")
					for i, l := range lines {
						if i == 0 {
							toolOutputs = append(toolOutputs, prefix+l)
						} else {
							toolOutputs = append(toolOutputs, "              "+l)
						}
					}
				}

				// Trim overly long outputs for better UI experience
				if len(toolOutputs) > maxToolOutputLines {
					trimmed := toolOutputs[:maxToolOutputLines]
					trimmed = append(trimmed, fmt.Sprintf("... (%d more lines truncated)", len(toolOutputs)-maxToolOutputLines))
					toolOutputs = trimmed
				}
				combined := strings.Join(toolOutputs, "\n")
				if len(combined) > maxToolOutputChars {
					combined = combined[:maxToolOutputChars] + "\n... (output truncated)"
				}

				content = summary + "\n" + combined
			}

			assistantMsgs = append([]providers.ChatMessage{{Role: providers.RoleAssistant, Content: content}}, assistantMsgs...)
		}

		// Combine assistant messages into a single response
		var responseText strings.Builder
		for _, msg := range assistantMsgs {
			if msg.Content != "" {
				if responseText.Len() > 0 {
					responseText.WriteString("\n")
				}
				responseText.WriteString(msg.Content)
			}
		}

		return aiResponseMsg{response: responseText.String(), isError: false}
	}
}

// Feature toggle: disable text beautification (colors/wrapping) when MAGIKARP_PLAIN=1
var disableBeautify = os.Getenv("MAGIKARP_PLAIN") == "1"

const (
	maxToolOutputLines = 40   // show at most 40 lines from any combined tool output
	maxToolOutputChars = 4000 // and at most 4000 characters overall
)
