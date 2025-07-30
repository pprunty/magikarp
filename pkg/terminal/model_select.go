package terminal

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ModelSelectModel represents the full-screen model selection interface
type ModelSelectModel struct {
	width          int
	height         int
	cursor         int
	availableModels []string
	selectedModel  string
	quitting       bool
}

// NewModelSelectModel creates a new model selection model
func NewModelSelectModel() ModelSelectModel {
	models := GetAvailableModels()
	return ModelSelectModel{
		width:          80,
		height:         24,
		cursor:         0,
		availableModels: models,
		selectedModel:  "",
		quitting:       false,
	}
}

// Init initializes the model selection model
func (m ModelSelectModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the model selection model
func (m ModelSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			m.cursor--
			if m.cursor < 0 {
				m.cursor = len(m.availableModels) - 1
			}
		case "down", "j":
			m.cursor++
			if m.cursor >= len(m.availableModels) {
				m.cursor = 0
			}
		case "enter":
			if len(m.availableModels) > 0 && m.cursor < len(m.availableModels) {
				m.selectedModel = m.availableModels[m.cursor]
			}
			m.quitting = true
			return m, tea.Quit
		case "esc", "q":
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

// GetSelectedModel returns the selected model name
func (m ModelSelectModel) GetSelectedModel() string {
	return m.selectedModel
}

// View renders the model selection screen
func (m ModelSelectModel) View() string {
	if m.quitting {
		return ""
	}

	s := ""

	// Welcome box at top
	s += renderWelcomeBox() + "\n\n"

	// Version display
	s += " " + versionStyle.Render(GetVersionDisplay()) + "\n\n"

	// Model list
	for i, model := range m.availableModels {
		if i == m.cursor {
			// Highlighted/selected model
			s += modelSelectActiveStyle.Render("  "+model) + "\n"
		} else {
			// Normal model
			s += modelSelectNormalStyle.Render("  "+model) + "\n"
		}
	}

	s += "\n"

	// Help text
	s += "\n"
	s += modelSelectHelpStyle.Render(" ↑/↓: navigate • enter: select • esc: cancel") + "\n\n"

	// Press Enter to continue
	s += continueStyle.Render(" Press Enter to select, Esc to cancel…")

	return s
}

// Model selection specific styles
var (
	modelSelectHeaderStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#04B575")).
		Bold(true)

	modelSelectActiveStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9B59B6")).
		Bold(true)

	modelSelectNormalStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")) // Gray to match slash commands

	modelSelectHelpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262"))
)