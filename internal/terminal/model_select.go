package terminal

import (
	"sort"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TreeItem represents an item in the tree structure
type TreeItem struct {
	Text      string // Display text
	Value     string // Actual value (model name or empty for providers)
	IsProvider bool   // True if this is a provider header
	IsLast     bool   // True if this is the last model in a provider group
}

// ModelSelectModel represents the full-screen model selection interface
type ModelSelectModel struct {
	width          int
	height         int
	cursor         int
	treeItems      []TreeItem
	selectedModel  string
	quitting       bool
}

// NewModelSelectModel creates a new model selection model
func NewModelSelectModel() ModelSelectModel {
	treeItems := buildTreeItems()
	
	// Find the first selectable model (not a provider header) 
	initialCursor := 0
	for i, item := range treeItems {
		if !item.IsProvider {
			initialCursor = i
			break
		}
	}
	
	return ModelSelectModel{
		width:       80,
		height:      24,
		cursor:      initialCursor,
		treeItems:   treeItems,
		selectedModel: "",
		quitting:    false,
	}
}

// buildTreeItems creates the tree structure from available models
func buildTreeItems() []TreeItem {
	providerModels := GetAvailableModelsByProvider()
	var items []TreeItem
	
	// Sort provider names for consistent display
	providerNames := make([]string, 0, len(providerModels))
	for providerName := range providerModels {
		providerNames = append(providerNames, providerName)
	}
	sort.Strings(providerNames)
	
	for _, providerName := range providerNames {
		models := providerModels[providerName]
		
		// Add provider header
		items = append(items, TreeItem{
			Text:       providerName,
			Value:      "",
			IsProvider: true,
			IsLast:     false,
		})
		
		// Sort models within provider
		sort.Strings(models)
		
		// Add models under provider
		for i, model := range models {
			isLast := i == len(models)-1
			var prefix string
			if isLast {
				prefix = "└── "
			} else {
				prefix = "├── "
			}
			
			items = append(items, TreeItem{
				Text:       prefix + model,
				Value:      model,
				IsProvider: false,
				IsLast:     isLast,
			})
		}
	}
	
	return items
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
			// Move up, skipping provider headers
			originalCursor := m.cursor
			for {
				m.cursor--
				if m.cursor < 0 {
					m.cursor = len(m.treeItems) - 1
				}
				// Stop if we're on a selectable item (model) or if we've looped back to start
				if m.cursor < len(m.treeItems) && (!m.treeItems[m.cursor].IsProvider || m.cursor == originalCursor) {
					break
				}
			}
		case "down", "j":
			// Move down, skipping provider headers
			originalCursor := m.cursor
			for {
				m.cursor++
				if m.cursor >= len(m.treeItems) {
					m.cursor = 0
				}
				// Stop if we're on a selectable item (model) or if we've looped back to start
				if m.cursor < len(m.treeItems) && (!m.treeItems[m.cursor].IsProvider || m.cursor == originalCursor) {
					break
				}
			}
		case "enter":
			if len(m.treeItems) > 0 && m.cursor < len(m.treeItems) {
				item := m.treeItems[m.cursor]
				// Only select if it's a model (not a provider header)
				if !item.IsProvider && item.Value != "" {
					m.selectedModel = item.Value
					m.quitting = true
					return m, tea.Quit
				}
			}
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

	// Tree structure
	for i, item := range m.treeItems {
		if item.IsProvider {
			// Provider header (always white, never highlighted)
			s += modelSelectProviderStyle.Render("  "+item.Text) + "\n"
		} else {
			// Model item
			if i == m.cursor {
				// Highlighted/selected model
				s += modelSelectActiveStyle.Render("  "+item.Text) + "\n"
			} else {
				// Normal model
				s += modelSelectNormalStyle.Render("  "+item.Text) + "\n"
			}
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

	modelSelectProviderStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true)

	modelSelectActiveStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9B59B6")).
		Bold(true)

	modelSelectNormalStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")) // Gray to match slash commands

	modelSelectHelpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262"))
)