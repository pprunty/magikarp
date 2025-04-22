package history

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/chzyer/readline"
)

const (
	maxHistorySize = 100  // Keep only last 100 commands
	historyFile    = ".magikarp_history"
)

// HistoryManager handles command history
type HistoryManager struct {
	rl      *readline.Instance
	histDir string
}

// NewHistoryManager creates a new history manager
func NewHistoryManager() (*HistoryManager, error) {
	// Get user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	// Create .magikarp directory if it doesn't exist
	histDir := filepath.Join(homeDir, ".magikarp")
	if err := os.MkdirAll(histDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create history directory: %w", err)
	}

	// Configure readline
	rl, err := readline.NewEx(&readline.Config{
		Prompt:          "\u001b[94mYou\u001b[0m: ",
		HistoryFile:     filepath.Join(histDir, historyFile),
		HistoryLimit:    maxHistorySize,
		InterruptPrompt: "^C",
		EOFPrompt:      "exit",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create readline instance: %w", err)
	}

	return &HistoryManager{
		rl:      rl,
		histDir: histDir,
	}, nil
}

// ReadLine reads a line from input with history support
func (h *HistoryManager) ReadLine() (string, error) {
	line, err := h.rl.Readline()
	if err != nil {
		return "", err
	}
	return line, nil
}

// Close closes the readline instance
func (h *HistoryManager) Close() error {
	return h.rl.Close()
}

// GetHistoryFile returns the path to the history file
func (h *HistoryManager) GetHistoryFile() string {
	return filepath.Join(h.histDir, historyFile)
}

// ClearHistory clears the command history
func (h *HistoryManager) ClearHistory() error {
	return os.WriteFile(h.GetHistoryFile(), []byte{}, 0644)
} 