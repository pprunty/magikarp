package terminal

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// HelpModel represents the full-screen help interface
type HelpModel struct {
	width    int
	height   int
	quitting bool
}

// NewHelpModel creates a new help model
func NewHelpModel() HelpModel {
	return HelpModel{
		width:  80,
		height: 24,
	}
}

// Init initializes the help model
func (m HelpModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the help model
func (m HelpModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "enter", "esc", "q":
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

// View renders the help screen
func (m HelpModel) View() string {
	if m.quitting {
		return ""
	}

	s := ""

	// Welcome box at top
	s += renderWelcomeBox() + "\n\n"

	// Version display
	s += " " + versionStyle.Render(GetVersionDisplay()) + "\n\n"

	// Help content
	s += helpContentStyle.Render(" Always review AI responses, especially when running code. Magikarp provides") + "\n"
	s += helpContentStyle.Render(" interactive AI assistance with multiple language model providers.") + "\n\n"

	s += helpSectionStyle.Render(" Usage Modes:") + "\n"
	s += helpItemStyle.Render(" • Interactive: magikarp (start chat session)") + "\n"
	s += helpItemStyle.Render(" • Command line: magikarp --help") + "\n\n"

	s += helpSectionStyle.Render(" Common Tasks:") + "\n"
	s += helpItemStyle.Render(" • Ask coding questions > How does this function work?") + "\n"
	s += helpItemStyle.Render(" • Get code suggestions > Help me implement...") + "\n"
	s += helpItemStyle.Render(" • Debug issues > Why is this not working?") + "\n"
	s += helpItemStyle.Render(" • Switch models > /model") + "\n"
	s += helpItemStyle.Render(" • Show help > help or /help") + "\n"
	s += helpItemStyle.Render(" • Exit application > exit or /exit") + "\n\n"

	s += helpSectionStyle.Render(" Interactive Mode Commands:") + "\n"
	commands := GetAvailableCommands()
	for _, cmd := range commands {
		s += helpCommandStyle.Render("  "+cmd.Name) + " - " + helpDescStyle.Render(cmd.Description) + "\n"
	}
	s += "\n"

	s += helpContentStyle.Render(" Learn more at: https://github.com/pprunty/magikarp") + "\n\n"

	// Press Enter to continue
	s += "\n"
	s += continueStyle.Render(" Press Enter to continue…")

	return s
}

// Help screen specific styles
var (
	versionStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true)

	helpContentStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF"))

	helpSectionStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true)

	helpItemStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF"))

	helpCommandStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#04B575")).
		Bold(true)

	helpDescStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262"))

	continueStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF6B35")).
		Bold(true)
)