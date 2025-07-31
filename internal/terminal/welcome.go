package terminal

import (
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderWelcomeBox creates the Claude Code-style welcome box
func renderWelcomeBox() string {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "unknown"
	}

	// Check API key status
	apiKeyStatus := getAPIKeyStatus()

	// Style the content with gray for subtitle and cwd
	content := "△ Welcome to Magikarp!\n\n"
	content += grayTextStyle.Render("  AI coding assistant with multiple LLM providers") + "\n\n"
	content += grayTextStyle.Render("  cwd: "+cwd) + "\n\n"
	content += apiKeyStatus

	// Calculate dynamic width based on content - need to account for new API key lines
	lines := strings.Split(content, "\n")
	width := 0
	for _, line := range lines {
		// Strip ANSI codes for length calculation
		cleanLine := stripANSIForWidth(line)
		if len(cleanLine) > width {
			width = len(cleanLine)
		}
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

// getAPIKeyStatus returns formatted API key status with aligned indicators
func getAPIKeyStatus() string {
	apiKeys := []struct {
		name string
		env  string
	}{
		{"Anthropic", "ANTHROPIC_API_KEY"},
		{"OpenAI", "OPENAI_API_KEY"},
		{"Gemini", "GEMINI_API_KEY"},
	}

	var status []string
	const alignmentWidth = 15 // Width for provider name alignment

	for _, key := range apiKeys {
		// Create the provider name with colon in gray
		providerPart := grayTextStyle.Render(key.name + ":")

		// Calculate padding for alignment
		padding := alignmentWidth - len(key.name) - 1 // -1 for the colon
		if padding < 0 {
			padding = 1 // At least one space
		}
		paddingStr := strings.Repeat(" ", padding)

		// Add status indicator
		if os.Getenv(key.env) != "" {
			status = append(status, "  "+providerPart+paddingStr+setKeyStyle.Render("✓"))
		} else {
			status = append(status, "  "+providerPart+paddingStr+unsetKeyStyle.Render("✗"))
		}
	}

	return strings.Join(status, "\n")
}

// stripANSIForWidth removes ANSI escape sequences for length calculations
func stripANSIForWidth(str string) string {
	// Simple implementation - in production you might want a more robust solution
	result := ""
	inEscape := false
	for _, r := range str {
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

	setKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")).
			Bold(true)

	unsetKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B35")).
			Bold(true)
)
