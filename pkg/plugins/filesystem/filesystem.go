package filesystem

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pprunty/magikarp/pkg/agent"
)

// FileSystemPlugin implements the Plugin interface for file operations
type FileSystemPlugin struct {
	*agent.BasePlugin
}

// New creates a new FileSystemPlugin instance
func New() *FileSystemPlugin {
	plugin := &FileSystemPlugin{
		BasePlugin: agent.NewBasePlugin("filesystem", "Provides tools for file system operations like reading, listing, and editing files"),
	}

	// Add tools during initialization
	plugin.AddTool("read_file", "Read the contents of a given relative file path. Use this when you want to see what's inside a file. Do not use this with directory names.",
		agent.GenerateSchema[ReadFileInput](),
		plugin.readFile)

	plugin.AddTool("list_files", "List files and directories at a given path. If no path is provided, lists files in the current directory.",
		agent.GenerateSchema[ListFilesInput](),
		plugin.listFiles)

	plugin.AddTool("edit_file", `Make edits to a text file.

Replaces 'old_str' with 'new_str' in the given file. 'old_str' and 'new_str' MUST be different from each other.

If the file specified with path doesn't exist, it will be created.`,
		agent.GenerateSchema[EditFileInput](),
		plugin.editFile)

	return plugin
}

// Initialize is called when the plugin is loaded
func (p *FileSystemPlugin) Initialize() error {
	// Any initialization code can go here
	return nil
}

// Cleanup is called when the plugin is unloaded
func (p *FileSystemPlugin) Cleanup() error {
	// Any cleanup code can go here
	return nil
}

type ReadFileInput struct {
	Path string `json:"path" jsonschema_description:"The relative path of a file in the working directory."`
}

func (p *FileSystemPlugin) readFile(input []byte) (string, error) {
	var toolInput agent.ToolInput
	toolInput.Data = input

	var readFileInput ReadFileInput
	if err := toolInput.UnmarshalInput(&readFileInput); err != nil {
		return "", err
	}

	content, err := os.ReadFile(readFileInput.Path)
	if err != nil {
		return agent.NewToolResult(false, fmt.Sprintf("Failed to read file: %v", err), nil).ToJSON()
	}

	return agent.NewToolResult(true, "File read successfully", string(content)).ToJSON()
}

type ListFilesInput struct {
	Path string `json:"path,omitempty" jsonschema_description:"Optional relative path to list files from. Defaults to current directory if not provided."`
}

func (p *FileSystemPlugin) listFiles(input []byte) (string, error) {
	var toolInput agent.ToolInput
	toolInput.Data = input

	var listFilesInput ListFilesInput
	if err := toolInput.UnmarshalInput(&listFilesInput); err != nil {
		return "", err
	}

	dir := "."
	if listFilesInput.Path != "" {
		dir = listFilesInput.Path
	}

	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		if relPath != "." {
			if info.IsDir() {
				files = append(files, relPath+"/")
			} else {
				files = append(files, relPath)
			}
		}
		return nil
	})

	if err != nil {
		return agent.NewToolResult(false, fmt.Sprintf("Failed to list files: %v", err), nil).ToJSON()
	}

	return agent.NewToolResult(true, "Files listed successfully", files).ToJSON()
}

type EditFileInput struct {
	Path   string `json:"path" jsonschema_description:"The path to the file"`
	OldStr string `json:"old_str" jsonschema_description:"Text to search for - must match exactly and must only have one match exactly"`
	NewStr string `json:"new_str" jsonschema_description:"Text to replace old_str with"`
}

func (p *FileSystemPlugin) editFile(input []byte) (string, error) {
	var toolInput agent.ToolInput
	toolInput.Data = input

	var editFileInput EditFileInput
	if err := toolInput.UnmarshalInput(&editFileInput); err != nil {
		return "", err
	}

	if editFileInput.Path == "" || editFileInput.OldStr == editFileInput.NewStr {
		return agent.NewToolResult(false, "Invalid input parameters", nil).ToJSON()
	}

	content, err := os.ReadFile(editFileInput.Path)
	if err != nil {
		if os.IsNotExist(err) && editFileInput.OldStr == "" {
			result, err := p.createNewFile(editFileInput.Path, editFileInput.NewStr)
			if err != nil {
				return agent.NewToolResult(false, fmt.Sprintf("Failed to create file: %v", err), nil).ToJSON()
			}
			return agent.NewToolResult(true, "File created successfully", result).ToJSON()
		}
		return agent.NewToolResult(false, fmt.Sprintf("Failed to read file: %v", err), nil).ToJSON()
	}

	oldContent := string(content)
	newContent := strings.Replace(oldContent, editFileInput.OldStr, editFileInput.NewStr, -1)

	if oldContent == newContent && editFileInput.OldStr != "" {
		return agent.NewToolResult(false, "old_str not found in file", nil).ToJSON()
	}

	err = os.WriteFile(editFileInput.Path, []byte(newContent), 0644)
	if err != nil {
		return agent.NewToolResult(false, fmt.Sprintf("Failed to write file: %v", err), nil).ToJSON()
	}

	return agent.NewToolResult(true, "File edited successfully", "OK").ToJSON()
}

func (p *FileSystemPlugin) createNewFile(filePath, content string) (string, error) {
	dir := path.Dir(filePath)
	if dir != "." {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return "", fmt.Errorf("failed to create directory: %w", err)
		}
	}

	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}

	return fmt.Sprintf("Successfully created file %s", filePath), nil
} 