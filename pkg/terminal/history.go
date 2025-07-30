package terminal

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	maxHistorySize = 100
	historyFile    = "input_history"
)

// HistoryManager handles persistent storage of input history
type HistoryManager struct {
	history []string
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

	hm := &HistoryManager{
		history: make([]string, 0),
		histDir: histDir,
	}

	// Load existing history
	if err := hm.LoadFromFile(); err != nil {
		// Don't fail if we can't load history, just start with empty
		// But still return the manager
	}

	return hm, nil
}

// AddMessage adds a message to history (avoiding duplicates and empty messages)
func (hm *HistoryManager) AddMessage(message string) error {
	message = strings.TrimSpace(message)
	if message == "" {
		return nil // Don't add empty messages
	}

	// Remove duplicate if it exists
	for i, hist := range hm.history {
		if hist == message {
			// Remove the duplicate entry
			hm.history = append(hm.history[:i], hm.history[i+1:]...)
			break
		}
	}

	// Add to the end (most recent)
	hm.history = append(hm.history, message)

	// Trim to max size if needed
	if len(hm.history) > maxHistorySize {
		hm.history = hm.history[len(hm.history)-maxHistorySize:]
	}

	// Save to file
	return hm.SaveToFile()
}

// GetHistory returns the full history slice
func (hm *HistoryManager) GetHistory() []string {
	return hm.history
}

// GetHistoryCount returns the number of items in history
func (hm *HistoryManager) GetHistoryCount() int {
	return len(hm.history)
}

// GetMessageAt returns the message at the given index (0 = oldest, len-1 = newest)
func (hm *HistoryManager) GetMessageAt(index int) string {
	if index < 0 || index >= len(hm.history) {
		return ""
	}
	return hm.history[index]
}

// GetHistoryFile returns the path to the history file
func (hm *HistoryManager) GetHistoryFile() string {
	return filepath.Join(hm.histDir, historyFile)
}

// SaveToFile saves the current history to disk
func (hm *HistoryManager) SaveToFile() error {
	file, err := os.Create(hm.GetHistoryFile())
	if err != nil {
		return fmt.Errorf("failed to create history file: %w", err)
	}
	defer file.Close()

	for _, message := range hm.history {
		if _, err := fmt.Fprintln(file, message); err != nil {
			return fmt.Errorf("failed to write to history file: %w", err)
		}
	}

	return nil
}

// LoadFromFile loads history from disk
func (hm *HistoryManager) LoadFromFile() error {
	file, err := os.Open(hm.GetHistoryFile())
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist yet, that's fine
			return nil
		}
		return fmt.Errorf("failed to open history file: %w", err)
	}
	defer file.Close()

	hm.history = make([]string, 0)
	scanner := bufio.NewScanner(file)
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			hm.history = append(hm.history, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read history file: %w", err)
	}

	// Ensure we don't exceed max size
	if len(hm.history) > maxHistorySize {
		hm.history = hm.history[len(hm.history)-maxHistorySize:]
	}

	return nil
}

// ClearHistory clears all history
func (hm *HistoryManager) ClearHistory() error {
	hm.history = make([]string, 0)
	return hm.SaveToFile()
}