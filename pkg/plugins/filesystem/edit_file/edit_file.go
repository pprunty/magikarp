package edit_file

import (
    _ "embed"
    "context"
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "strings"

    "github.com/pprunty/magikarp/pkg/agent"
)

//go:embed tool.json
var schema []byte

// Input represents the parameters for the edit_file tool
type input struct {
    Path        string `json:"path"`        // The file path to edit
    Content     string `json:"content,omitempty"`     // The entire content to write to the file (overrides all other operations)
    Search      string `json:"search,omitempty"`      // Text to search for when doing a replacement
    ReplaceWith string `json:"replace_with,omitempty"` // Text to replace the search text with
    Append      string `json:"append,omitempty"`      // Text to append to the file
    Prepend     string `json:"prepend,omitempty"`     // Text to prepend to the file
}

// Definition returns the tool definition for the edit_file tool
func Definition() agent.ToolDefinition {
    var params map[string]agent.Parameter
    if err := json.Unmarshal(schema, &params); err != nil {
        // Handle error properly - log it
        fmt.Printf("Error unmarshaling schema: %v\n", err)
    }

    return agent.ToolDefinition{
        Name:        "edit_file",
        Description: "Edit a text file by search/replace, append, prepend, or overwrite operations",
        Parameters:  params,
        Function:    run,
    }
}

// run executes the file editing operation
func run(ctx context.Context, inputData map[string]interface{}) (*agent.ToolResult, error) {
    // Convert input data to our structured input type
    var in input
    inputBytes, err := json.Marshal(inputData)
    if err != nil {
        return agent.NewToolResult("edit_file", fmt.Sprintf("Error processing input parameters: %v", err), true), nil
    }

    if err := json.Unmarshal(inputBytes, &in); err != nil {
        return agent.NewToolResult("edit_file", fmt.Sprintf("Error parsing input parameters: %v", err), true), nil
    }

    // Validate path is provided and local
    if in.Path == "" {
        return agent.NewToolResult("edit_file", "Path parameter is required", true), nil
    }

    if !filepath.IsLocal(in.Path) {
        return agent.NewToolResult("edit_file", "Path must be a local file path", true), nil
    }

    // Check file exists before attempting operations
    fileInfo, err := os.Stat(in.Path)
    var orig string

    if os.IsNotExist(err) {
        // File doesn't exist - will be created if we have content to add
        orig = ""
    } else if err != nil {
        return agent.NewToolResult("edit_file", fmt.Sprintf("Error accessing file: %v", err), true), nil
    } else if fileInfo.IsDir() {
        return agent.NewToolResult("edit_file", "Path points to a directory, not a file", true), nil
    } else {
        // Read existing file
        data, err := os.ReadFile(in.Path)
        if err != nil {
            return agent.NewToolResult("edit_file", fmt.Sprintf("Error reading file: %v", err), true), nil
        }
        orig = string(data)
    }

    // Determine the operation to perform
    var changed string
    switch {
    case in.Content != "":
        // Complete content replacement overrides other operations
        changed = in.Content
    case in.Search != "":
        if in.ReplaceWith == "" {
            return agent.NewToolResult("edit_file", "replace_with parameter is required when search is provided", true), nil
        }
        changed = strings.ReplaceAll(orig, in.Search, in.ReplaceWith)
    case in.Prepend != "" && in.Append != "":
        // Both prepend and append
        changed = in.Prepend + orig + in.Append
    case in.Prepend != "":
        // Just prepend
        changed = in.Prepend + orig
    case in.Append != "":
        // Just append
        changed = orig + in.Append
    default:
        return agent.NewToolResult("edit_file", "No operation provided. Specify content, search/replace_with, append, or prepend", true), nil
    }

    // Check if there are any changes to make
    if changed == orig && orig != "" {
        return agent.NewToolResult("edit_file", "No changes made to the file", false), nil
    }

    // Create directory if it doesn't exist
    dir := filepath.Dir(in.Path)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return agent.NewToolResult("edit_file", fmt.Sprintf("Error creating directory structure: %v", err), true), nil
    }

    // Write the changes to the file
    if err := os.WriteFile(in.Path, []byte(changed), 0644); err != nil {
        return agent.NewToolResult("edit_file", fmt.Sprintf("Error writing to file: %v", err), true), nil
    }

    // Return success with details of what was done
    operation := determineOperation(in)
    return agent.NewToolResult("edit_file", fmt.Sprintf("File successfully %s: %s", operation, in.Path), false), nil
}

// determineOperation returns a string describing what operation was performed
func determineOperation(in input) string {
    switch {
    case in.Content != "":
        return "overwritten"
    case in.Search != "":
        return "updated with search/replace"
    case in.Prepend != "" && in.Append != "":
        return "modified with prepend and append"
    case in.Prepend != "":
        return "prepended"
    case in.Append != "":
        return "appended"
    default:
        return "updated"
    }
}