package history

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/chzyer/readline"
)

// HistoryManager handles command history for the CLI
type HistoryManager struct {
	rl          *readline.Instance
	historyFile string
}

// NewHistoryManager creates a new history manager
func NewHistoryManager() (*HistoryManager, error) {
	// Create history directory if it doesn't exist
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	historyDir := filepath.Join(homeDir, ".magikarp")
	if err := os.MkdirAll(historyDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create history directory: %w", err)
	}

	historyFile := filepath.Join(historyDir, "history")
	
	// Create readline instance
	rl, err := readline.NewEx(&readline.Config{
		Prompt:          "> ",
		HistoryFile:     historyFile,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create readline instance: %w", err)
	}

	return &HistoryManager{
		rl:          rl,
		historyFile: historyFile,
	}, nil
}

// ReadLine reads a line from the input with history support
func (h *HistoryManager) ReadLine() (string, error) {
	return h.rl.Readline()
}

// AddHistory adds a line to the history
func (h *HistoryManager) AddHistory(line string) error {
	return h.rl.SaveHistory(line)
}

// Close closes the history manager
func (h *HistoryManager) Close() error {
	return h.rl.Close()
} 