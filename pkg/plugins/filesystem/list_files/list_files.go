package list_files

import (
    _ "embed"
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "time"

    "github.com/pprunty/magikarp/pkg/agent"
)

/* ──────────── embed & schema ──────────── */

//go:embed tool.json
var rawSchema []byte

// input represents the parameters for the list_files tool
type input struct {
    Path          string `json:"path,omitempty"`           // Base path to list files from
    Depth         int    `json:"depth,omitempty"`          // How deep to recurse (0 = unlimited)
    IncludeHidden bool   `json:"include_hidden,omitempty"` // Whether to include hidden files/dirs
    MaxEntries    int    `json:"max_entries,omitempty"`    // Maximum number of entries to return
    IncludeStats  bool   `json:"include_stats,omitempty"`  // Include file size and modification time
    Pattern       string `json:"pattern,omitempty"`        // Optional glob pattern to filter results
}

// fileEntry represents a single file or directory in the result
type fileEntry struct {
    Path         string     `json:"path"`                    // Path relative to the base path
    Type         string     `json:"type"`                    // "file" or "dir"
    Size         int64      `json:"size,omitempty"`          // File size in bytes (only if include_stats=true)
    ModTime      *time.Time `json:"mod_time,omitempty"`      // Last modification time (only if include_stats=true)
    ErrorMessage string     `json:"error,omitempty"`         // Error message if any
}

// resultSummary provides information about the overall listing operation
type resultSummary struct {
    TotalFiles      int    `json:"total_files"`            // Total number of files found
    TotalDirs       int    `json:"total_dirs"`             // Total number of directories found
    BasePath        string `json:"base_path"`              // The absolute base path that was searched
    Truncated       bool   `json:"truncated"`              // Whether results were truncated
    MaxEntriesLimit int    `json:"max_entries_limit"`      // The max entries limit that was applied
}

// listResult combines the list of entries with summary information
type listResult struct {
    Entries []fileEntry  `json:"entries"`                 // List of file/directory entries
    Summary resultSummary `json:"summary"`                // Summary information
}

func Definition() agent.ToolDefinition {
    var schema map[string]any
    if err := json.Unmarshal(rawSchema, &schema); err != nil {
        fmt.Printf("Error unmarshaling list_files schema: %v\n", err)
    }

    return agent.ToolDefinition{
        Name:        "list_files",
        Description: schema["description"].(string),
        InputSchema: schema["input_schema"].(map[string]any),
        Function:    run,
    }
}

/* ──────────── implementation ──────────── */

func run(ctx context.Context, m map[string]any) (*agent.ToolResult, error) {
    // Decode and validate user input
    var in input
    inputBytes, err := json.Marshal(m)
    if err != nil {
        return agent.NewToolResult("list_files", fmt.Sprintf("Error processing input parameters: %v", err), true), nil
    }

    if err := json.Unmarshal(inputBytes, &in); err != nil {
        return agent.NewToolResult("list_files", fmt.Sprintf("Error parsing input parameters: %v", err), true), nil
    }

    // Apply defaults and validate input parameters
    if in.Path == "" {
        in.Path = "."
    }

    if in.Depth < 0 {
        in.Depth = 0
    }

    if in.MaxEntries <= 0 {
        in.MaxEntries = 200 // Default
    }

    if in.MaxEntries > 1000 {
        in.MaxEntries = 1000 // Hard cap
    }

    // Security: ensure path is local
    if !filepath.IsLocal(in.Path) {
        return agent.NewToolResult("list_files", "Path must be local for security reasons", true), nil
    }

    // Get absolute path for better error messages
    absPath, err := filepath.Abs(in.Path)
    if err != nil {
        return agent.NewToolResult("list_files", fmt.Sprintf("Unable to resolve absolute path: %v", err), true), nil
    }

    // Check if path exists
    if _, err := os.Stat(in.Path); os.IsNotExist(err) {
        return agent.NewToolResult("list_files", fmt.Sprintf("Path does not exist: %s", in.Path), true), nil
    }

    // Compile glob pattern if provided
    var matcher func(string) bool = nil
    if in.Pattern != "" {
        matcher = func(name string) bool {
            matched, _ := filepath.Match(in.Pattern, filepath.Base(name))
            return matched
        }
    }

    // Initialize result
    out := listResult{
        Entries: []fileEntry{},
        Summary: resultSummary{
            BasePath:        absPath,
            MaxEntriesLimit: in.MaxEntries,
        },
    }

    // Walk the directory tree
    baseDepth := depth(in.Path)
    truncated := false

    err = filepath.WalkDir(in.Path, func(p string, d os.DirEntry, err error) error {
        // Handle walk errors (permission denied, etc.)
        if err != nil {
            // Continue walking but record the error in a special entry
            out.Entries = append(out.Entries, fileEntry{
                Path:         p,
                ErrorMessage: err.Error(),
            })
            return nil
        }

        // Check for context cancellation
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
        }

        // Enforce depth limit
        currDepth := depth(p) - baseDepth
        if in.Depth > 0 && currDepth > in.Depth {
            if d.IsDir() {
                return filepath.SkipDir
            }
            return nil
        }

        // Skip hidden files/directories if not included
        if !in.IncludeHidden && strings.HasPrefix(filepath.Base(p), ".") && p != in.Path {
            if d.IsDir() {
                return filepath.SkipDir
            }
            return nil
        }

        // Skip the root directory itself
        if p == in.Path {
            return nil
        }

        // Apply pattern filter if specified
        if matcher != nil && !matcher(p) && !d.IsDir() {
            return nil
        }

        // Create entry with basic info
        entry := fileEntry{
            Path: p,
            Type: map[bool]string{true: "dir", false: "file"}[d.IsDir()],
        }

        // Update summary counters
        if d.IsDir() {
            out.Summary.TotalDirs++
        } else {
            out.Summary.TotalFiles++
        }

        // Add additional stats if requested
        if in.IncludeStats {
            info, err := d.Info()
            if err == nil {
                entry.Size = info.Size()
                modTime := info.ModTime()
                entry.ModTime = &modTime
            }
        }

        // Add to results
        out.Entries = append(out.Entries, entry)

        // Check if we've reached the maximum entries
        if len(out.Entries) >= in.MaxEntries {
            truncated = true
            return errors.New("max-entries-reached")
        }

        return nil
    })

    // Handle walk errors
    if err != nil && !errors.Is(err, context.Canceled) && err.Error() != "max-entries-reached" {
        return agent.NewToolResult("list_files", fmt.Sprintf("Error listing files: %v", err), true), nil
    }

    // Update truncation status in summary
    out.Summary.Truncated = truncated

    // Convert result to JSON
    resultJSON, err := json.MarshalIndent(out, "", "  ")
    if err != nil {
        return agent.NewToolResult("list_files", "Error encoding results to JSON", true), nil
    }

    // Ensure output size doesn't exceed reasonable limits (5 MB)
    if len(resultJSON) > 5*1024*1024 {
        resultJSON = resultJSON[:5*1024*1024]
        resultJSON = append(resultJSON, []byte("\n... output truncated due to size limitations")...)
    }

    return agent.NewToolResult("list_files", string(resultJSON), false), nil
}

/* helpers */

// depth returns the directory depth of a path
func depth(p string) int {
    return len(strings.Split(filepath.Clean(p), string(os.PathSeparator)))
}

// must is a helper function for handling errors
func must[T any](v T, err error) T {
    if err != nil {
        panic(err)
    }
    return v
}