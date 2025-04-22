package registry

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

// ProcessRunner runs tools using an external process
type ProcessRunner struct {
	toolDesc     *ToolDescriptor
	toolDir      string
	builtFlag    bool
	templateArgs *template.Template
}

// NewProcessRunner creates a new process runner for a tool
func NewProcessRunner(desc *ToolDescriptor) *ProcessRunner {
	// Create the template for Run command
	tmpl, err := template.New("run").Parse(desc.Run)
	if err != nil {
		// Use a fallback template that just passes the full JSON
		tmpl = template.Must(template.New("run").Parse("{{.}}"))
	}

	return &ProcessRunner{
		toolDesc:     desc,
		toolDir:      filepath.Join("pkg", "tools", desc.Name),
		builtFlag:    false,
		templateArgs: tmpl,
	}
}

// BuildTool builds the tool if necessary
func (r *ProcessRunner) BuildTool() error {
	// Skip if already built or no build command
	if r.builtFlag || r.toolDesc.Build == "" {
		return nil
	}

	// Run the build command
	cmd := exec.Command("sh", "-c", r.toolDesc.Build)
	cmd.Dir = r.toolDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to build tool: %w", err)
	}

	r.builtFlag = true
	return nil
}

// Run executes the tool with the given arguments
func (r *ProcessRunner) Run(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
	// Build the tool if necessary
	if err := r.BuildTool(); err != nil {
		return nil, err
	}

	// Parse the arguments to a map for template substitution
	var argsMap map[string]interface{}
	if err := json.Unmarshal(args, &argsMap); err != nil {
		return nil, fmt.Errorf("failed to parse tool arguments: %w", err)
	}

	// Expand the command template with args
	var cmdBuilder strings.Builder
	if err := r.templateArgs.Execute(&cmdBuilder, argsMap); err != nil {
		return nil, fmt.Errorf("failed to expand command template: %w", err)
	}
	
	// Create the command
	cmdStr := cmdBuilder.String()
	cmd := exec.CommandContext(ctx, "sh", "-c", cmdStr)
	cmd.Dir = r.toolDir
	
	// Set stdin to be the raw JSON args
	cmd.Stdin = bytes.NewReader(args)
	
	// Capture stdout for the result
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	// Run the command
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("tool execution failed: %v\n%s", err, stderr.String())
	}
	
	// Parse the result as JSON
	result := stdout.Bytes()
	
	// Validate that the result is valid JSON
	var jsonCheck interface{}
	if err := json.Unmarshal(result, &jsonCheck); err != nil {
		return nil, fmt.Errorf("tool returned invalid JSON: %v\n%s", err, result)
	}
	
	return result, nil
}

// GoFunctionRunner runs Go functions directly
type GoFunctionRunner struct {
	function func(json.RawMessage) (json.RawMessage, error)
}

// NewGoFunctionRunner creates a new runner for a Go function
func NewGoFunctionRunner(fn func(json.RawMessage) (json.RawMessage, error)) *GoFunctionRunner {
	return &GoFunctionRunner{
		function: fn,
	}
}

// Run executes the Go function with the given arguments
func (r *GoFunctionRunner) Run(_ context.Context, args json.RawMessage) (json.RawMessage, error) {
	return r.function(args)
}

// RegisterGoTool registers a Go function as a tool
func RegisterGoTool(name, description string, inputs, outputs map[string]interface{}, fn func(json.RawMessage) (json.RawMessage, error)) error {
	runner := NewGoFunctionRunner(fn)
	
	return RegisterTool(name, ToolDefinition{
		Name:        name,
		Description: description,
		Language:    "go",
		InputSchema: inputs,
		OutputSchema: outputs,
		Runner:      runner,
	})
} 