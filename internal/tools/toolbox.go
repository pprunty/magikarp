package tools

import "github.com/pprunty/magikarp/internal/providers"

// Toolbox represents a collection of related tool definitions.
// A toolbox groups multiple tools under a name/description.
type Toolbox interface {
	Name() string
	Description() string
	Tools() []providers.ToolDefinition
}

// BaseToolbox is a helper implementing Toolbox.
type BaseToolbox struct {
	name        string
	description string
	tools       []providers.ToolDefinition
}

// NewBaseToolbox returns a new BaseToolbox.
func NewBaseToolbox(name, description string) *BaseToolbox {
	return &BaseToolbox{
		name:        name,
		description: description,
		tools:       []providers.ToolDefinition{},
	}
}

func (b *BaseToolbox) Name() string                       { return b.name }
func (b *BaseToolbox) Description() string                { return b.description }
func (b *BaseToolbox) Tools() []providers.ToolDefinition  { return b.tools }
func (b *BaseToolbox) AddTool(t providers.ToolDefinition) { b.tools = append(b.tools, t) }

var registry []Toolbox

// Register adds a toolbox to the global registry.
func Register(tb Toolbox) { registry = append(registry, tb) }

// GetAllTools returns every tool definition registered across all toolboxes.
func GetAllTools() []providers.ToolDefinition {
	var out []providers.ToolDefinition
	for _, tb := range registry {
		out = append(out, tb.Tools()...)
	}
	return out
}

// GetCoreTools returns tool definitions from the toolbox named "core".
func GetCoreTools() []providers.ToolDefinition {
	var out []providers.ToolDefinition
	for _, tb := range registry {
		if tb.Name() == "core" {
			out = append(out, tb.Tools()...)
		}
	}
	return out
}

// GetToolByName finds a tool by name.
func GetToolByName(name string) (providers.ToolDefinition, bool) {
	for _, tb := range registry {
		for _, t := range tb.Tools() {
			if t.Name == name {
				return t, true
			}
		}
	}
	return providers.ToolDefinition{}, false
}

// Toolboxes lists registered toolboxes.
func Toolboxes() []Toolbox { return registry }
