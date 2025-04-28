package agent

// BasePlugin provides a basic implementation of the Plugin interface
type BasePlugin struct {
	name        string
	description string
	tools       []ToolDefinition
}

// NewBasePlugin creates a new BasePlugin
func NewBasePlugin(name, description string) *BasePlugin {
	return &BasePlugin{
		name:        name,
		description: description,
		tools:       make([]ToolDefinition, 0),
	}
}

// Name returns the plugin name
func (p *BasePlugin) Name() string {
	return p.name
}

// Tools returns the list of tools provided by this plugin
func (p *BasePlugin) Tools() []ToolDefinition {
	return p.tools
}

// AddTool adds a tool to the plugin
func (p *BasePlugin) AddTool(tool ToolDefinition) {
	p.tools = append(p.tools, tool)
} 