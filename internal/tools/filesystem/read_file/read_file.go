package read_file

import (
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/pprunty/magikarp/internal/providers"
)

//go:embed tool.json
var wrapper []byte // tool.json contains name/description/input_schema

/* ------------------------------------------------------------------ */

type input struct {
	Path           string `json:"path"`
	MaxSize        int    `json:"max_size,omitempty"`
	DetectEncoding bool   `json:"detect_encoding,omitempty"`
	IncludeStats   bool   `json:"include_stats,omitempty"`
}

func Definition() providers.ToolDefinition {
	var w map[string]any
	if err := json.Unmarshal(wrapper, &w); err != nil {
		fmt.Printf("Error unmarshaling read_file schema: %v\n", err)
	}

	schema := w["input_schema"].(map[string]any)

	return providers.ToolDefinition{
		Name:        "read_file",
		Description: w["description"].(string),
		InputSchema: schema,
		Function:    run,
	}
}

/* ------------------------------------------------------------------ */

func run(ctx context.Context, inMap map[string]any) (*providers.ToolResult, error) {
	// Parse input parameters
	var in input
	inputBytes, err := json.Marshal(inMap)
	if err != nil {
		return providers.NewToolResult("read_file", fmt.Sprintf("Error processing input parameters: %v", err), true), nil
	}

	if err := json.Unmarshal(inputBytes, &in); err != nil {
		return providers.NewToolResult("read_file", fmt.Sprintf("Error parsing input parameters: %v", err), true), nil
	}

	// Set default max size if not specified
	if in.MaxSize <= 0 {
		in.MaxSize = 100_000 // Default max size: 100 KB
	} else if in.MaxSize > 1_000_000 {
		in.MaxSize = 1_000_000 // Hard cap: 1 MB
	}

	// Validate path
	if in.Path == "" {
		return providers.NewToolResult("read_file", "Path parameter is required", true), nil
	}

	if !filepath.IsLocal(in.Path) {
		return providers.NewToolResult("read_file", "Path must be local for security reasons", true), nil
	}

	// Clean path and handle case-insensitive search if needed
	path := filepath.Clean(in.Path)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Try case-insensitive match
		dir := filepath.Dir(path)
		base := filepath.Base(path)

		// Make sure the directory exists
		entries, err := os.ReadDir(dir)
		if err != nil {
			return providers.NewToolResult("read_file",
				fmt.Sprintf("Directory not found or not accessible: %s (%v)", dir, err), true), nil
		}

		// Look for case-insensitive match
		found := false
		for _, e := range entries {
			if strings.EqualFold(e.Name(), base) {
				path = filepath.Join(dir, e.Name())
				found = true
				break
			}
		}

		if !found {
			return providers.NewToolResult("read_file",
				fmt.Sprintf("File not found: %s (no case-insensitive match found)", in.Path), true), nil
		}
	}

	// Check if path exists and is a file
	fileInfo, err := os.Stat(path)
	if err != nil {
		return providers.NewToolResult("read_file", fmt.Sprintf("Error accessing file: %v", err), true), nil
	}

	if fileInfo.IsDir() {
		return providers.NewToolResult("read_file", fmt.Sprintf("Path points to a directory, not a file: %s", path), true), nil
	}

	// Check file size before reading
	if fileInfo.Size() > int64(in.MaxSize) {
		return providers.NewToolResult("read_file",
			fmt.Sprintf("File size (%d bytes) exceeds maximum allowed size (%d bytes)",
				fileInfo.Size(), in.MaxSize), true), nil
	}

	// Read file content
	data, err := os.ReadFile(path)
	if err != nil {
		return providers.NewToolResult("read_file", fmt.Sprintf("Error reading file: %v", err), true), nil
	}

	// Validate UTF-8 encoding
	if !utf8.Valid(data) && !in.DetectEncoding {
		return providers.NewToolResult("read_file",
			"File contains invalid UTF-8 sequences. Set detect_encoding=true to attempt conversion.", true), nil
	}

	// Create response
	content := string(data)

	// If include_stats requested, create a JSON response with both content and metadata
	if in.IncludeStats {
		// Calculate content hash
		hasher := sha256.New()
		hasher.Write(data)
		contentHash := base64.StdEncoding.EncodeToString(hasher.Sum(nil))

		// Count lines
		lineCount := strings.Count(content, "\n") + 1

		// Create response with stats
		stats := map[string]interface{}{
			"content":      content,
			"path":         path,
			"size_bytes":   fileInfo.Size(),
			"lines":        lineCount,
			"modified_at":  fileInfo.ModTime(),
			"content_hash": contentHash,
			"is_binary":    !utf8.Valid(data) && in.DetectEncoding,
		}

		statsJSON, err := json.MarshalIndent(stats, "", "  ")
		if err != nil {
			return providers.NewToolResult("read_file", "Error generating stats JSON", true), nil
		}

		return providers.NewToolResult("read_file", string(statsJSON), false), nil
	}

	// Return just the content for simple requests
	return providers.NewToolResult("read_file", content, false), nil
}

/* helpers */
func contains(raw any, key string) bool {
	if arr, ok := raw.([]any); ok {
		for _, v := range arr {
			if v == key {
				return true
			}
		}
	}
	return false
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
