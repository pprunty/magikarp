package terminal

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// MenuModel represents the main menu state
type MenuModel struct {
	choices  []string
	cursor   int
	choice   string
	quitting bool
}

// Initialize the menu model
func NewMenuModel() MenuModel {
	return MenuModel{
		choices: []string{
			"Chat with Claude",
			"Chat with GPT",
			"Chat with Gemini",
			"Settings",
			"Exit",
		},
	}
}

// Init is the first function that will be called when the program starts
func (m MenuModel) Init() tea.Cmd {
	return nil
}

// Update handles incoming messages and updates the model state
func (m MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.quitting = true
			return m, tea.Quit

		case "up", "k":
			m.cursor--
			if m.cursor < 0 {
				m.cursor = len(m.choices) - 1
			}

		case "down", "j":
			m.cursor++
			if m.cursor >= len(m.choices) {
				m.cursor = 0
			}

		case "enter", " ":
			m.choice = m.choices[m.cursor]
			if m.cursor == 4 { // Exit option
				m.quitting = true
			}
			return m, tea.Quit
		}
	}

	return m, nil
}

// View renders the UI
func (m MenuModel) View() string {
	if m.quitting {
		return quitTextStyle.Render("Thanks for using Magikarp! 🐟\n")
	}

	s := ""

	// Claude Code-style welcome box
	s += renderWelcomeBox() + "\n\n"

	// Provider selection
	s += sectionTitleStyle.Render(" Select a provider:")
	s += "\n\n"

	// Menu options
	for i, choice := range m.choices {
		if m.cursor == i {
			s += focusedStyle.Render("> ") + selectedItemStyle.Render(choice) + "\n"
		} else {
			s += "  " + itemStyle.Render(choice) + "\n"
		}
	}

	// Footer
	s += "\n"
	s += helpStyle.Render("↑/↓: navigate • enter: select • q/esc/ctrl+c: quit")
	s += "\n"

	return s
}


// Styling
var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B35")).
			Bold(true).
			Padding(0, 1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Italic(true)



	sectionTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true)
)