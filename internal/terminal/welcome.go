package terminal

import (
	"os"

	"github.com/charmbracelet/lipgloss"
)

// renderWelcomeBox creates the Claude Code-style welcome box
func renderWelcomeBox() string {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "unknown"
	}

	// Style the content with gray for subtitle and cwd
	content := "△ Welcome to Magikarp!\n\n"
	content += grayTextStyle.Render("  AI coding assistant with multiple LLM providers") + "\n\n"
	content += grayTextStyle.Render("  cwd: " + cwd)

	// Calculate dynamic width based on content
	maxLineLength := len("  AI coding assistant with multiple LLM providers")
	cwdLineLength := len("  cwd: " + cwd)
	titleLineLength := len("✻ Welcome to Magikarp!")

	width := maxLineLength
	if cwdLineLength > width {
		width = cwdLineLength
	}
	if titleLineLength > width {
		width = titleLineLength
	}

	// Add some padding
	width += 4

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#626262")).
		Padding(0, 1).
		Width(width)

	return style.Render(content)
}

// renderWelcomeBoxWithVersion creates welcome box with version display below
func renderWelcomeBoxWithVersion() string {
	welcomeBox := renderWelcomeBox()
	version := " " + versionDisplayStyle.Render(GetVersionDisplay())
	return welcomeBox + "\n\n" + version
}

// Styles for welcome box content
var (
	grayTextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))

	versionDisplayStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Bold(true)
)
