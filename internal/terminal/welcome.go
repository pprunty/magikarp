package terminal

import (
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/pprunty/magikarp/internal/config"
	"github.com/pprunty/magikarp/internal/orchestration"
)

// renderWelcomeBox creates the Claude Code-style welcome box
func renderWelcomeBox() string {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "unknown"
	}

	// Check if we should show provider status
	showProviderStatus := shouldShowProviderStatus()
	var providerStatus string
	if showProviderStatus {
		providerStatus = getProviderStatus()
	}

	// Style the content with opacity for welcome message
	content := welcomeTextStyle.Render("✱ Welcome to Magikarp Coding Agent") + "\n\n"
	content += descriptionStyle.Render("  Magikarp is an open-source CLI and autonomous coding assistant which supports\n  multiple LLM providers.") + "\n\n"
	content += descriptionStyle.Render("  Please reference the repository for documentation and contribution guidelines at") + "\n"
	content += "  " + linkStyle.Render("https://github.com/pprunty/magikarp") + "\n\n"
	content += grayTextStyle.Render("  cwd: "+cwd) + "\n\n"
	if showProviderStatus {
		content += providerStatus
	}

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
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#626262")).
		Padding(0, 1).
		Width(width)

	return style.Render(content)
}

// shouldShowProviderStatus checks config to determine if provider status should be shown
func shouldShowProviderStatus() bool {
	configPath := findConfigFile()
	if configPath == "" {
		return true // Default to showing provider status
	}
	
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return true // Default to showing provider status on error
	}
	
	return cfg.Terminal.DetailProviders
}

// getProviderStatus returns formatted provider status with grid layout
func getProviderStatus() string {
	providers := []struct {
		name string
		env  string
	}{
		{"Anthropic", "ANTHROPIC_API_KEY"},
		{"OpenAI", "OPENAI_API_KEY"},
		{"Gemini", "GEMINI_API_KEY"},
		{"Mistral", "MISTRAL_API_KEY"},
		{"Alibaba", "ALIBABA_API_KEY"},
	}

	// Get actual provider initialization status
	providerInitStatus := getActualProviderStatus()

	var status []string
	const colWidth = 20 // Width for each column

	// Create grid layout (2 columns)
	for i := 0; i < len(providers); i += 2 {
		line := "  "

		// First column
		provider1 := providers[i]
		name1 := grayTextStyle.Render(provider1.name + ":")
		padding1 := colWidth - len(provider1.name) - 1
		if padding1 < 1 {
			padding1 = 1
		}

		var indicator1 string
		if isInitialized, exists := providerInitStatus[strings.ToLower(provider1.name)]; exists && isInitialized {
			indicator1 = setKeyStyle.Render("✓")
		} else {
			indicator1 = unsetKeyStyle.Render("✗")
		}

		line += name1 + strings.Repeat(" ", padding1) + indicator1

		// Second column (if exists)
		if i+1 < len(providers) {
			provider2 := providers[i+1]
			name2 := grayTextStyle.Render(provider2.name + ":")
			padding2 := colWidth - len(provider2.name) - 1
			if padding2 < 1 {
				padding2 = 1
			}

			var indicator2 string
			if isInitialized, exists := providerInitStatus[strings.ToLower(provider2.name)]; exists && isInitialized {
				indicator2 = setKeyStyle.Render("✓")
			} else {
				indicator2 = unsetKeyStyle.Render("✗")
			}

			line += "    " + name2 + strings.Repeat(" ", padding2) + indicator2
		}

		status = append(status, line)
	}

	return strings.Join(status, "\n")
}

// getActualProviderStatus gets the real initialization status from the registry
func getActualProviderStatus() map[string]bool {
	// Try to load config and get provider status
	configPath := findConfigFile()
	if configPath == "" {
		return make(map[string]bool) // Return empty if no config
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return make(map[string]bool) // Return empty if config load fails
	}

	// Initialize the registry (this is safe to call multiple times)
	_ = orchestration.Init(cfg)

	// Get actual provider status from registry
	return orchestration.GetInitializedProviders(cfg)
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
	welcomeTextStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Bold(false).
				Faint(false)

	descriptionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#626262"))

	linkStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#82A2BE")).
			Underline(true)

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
