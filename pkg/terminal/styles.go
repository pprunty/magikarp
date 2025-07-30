package terminal

import "github.com/charmbracelet/lipgloss"

// Shared styles used across multiple components
var (
	// Menu and selection styles
	focusedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF6B35")).
		Bold(true)

	selectedItemStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#808080")).
		Bold(false)

	itemStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF"))

	helpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Italic(true)

	quitTextStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#04B575")).
		Bold(true)
)