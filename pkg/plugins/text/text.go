package text

import (
	"strings"

	"github.com/pprunty/magikarp/pkg/agent"
)

// TextPlugin implements the Plugin interface for text manipulation
type TextPlugin struct {
	*agent.BasePlugin
}

// New creates a new TextPlugin instance
func New() *TextPlugin {
	plugin := &TextPlugin{
		BasePlugin: agent.NewBasePlugin("text", "Provides tools for text manipulation and transformation"),
	}
	
	// Add tools during initialization
	plugin.AddTool("rot13", "Apply ROT13 transformation to text", 
		agent.GenerateSchema[Rot13Input](), 
		plugin.rot13)
	
	return plugin
}

// Initialize is called when the plugin is loaded
func (p *TextPlugin) Initialize() error {
	// Any initialization code can go here
	return nil
}

// Cleanup is called when the plugin is unloaded
func (p *TextPlugin) Cleanup() error {
	// Any cleanup code can go here
	return nil
}

type Rot13Input struct {
	Text string `json:"text" jsonschema_description:"The text to transform"`
}

func (p *TextPlugin) rot13(input []byte) (string, error) {
	var toolInput agent.ToolInput
	toolInput.Data = input
	
	var rot13Input Rot13Input
	if err := toolInput.UnmarshalInput(&rot13Input); err != nil {
		return "", err
	}

	transformed := rot13Transform(rot13Input.Text)
	result := agent.NewToolResult(true, "Text transformed successfully", transformed)
	return result.ToJSON()
}

func rot13Transform(text string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case r >= 'A' && r <= 'Z':
			return 'A' + (r-'A'+13)%26
		case r >= 'a' && r <= 'z':
			return 'a' + (r-'a'+13)%26
		default:
			return r
		}
	}, text)
} 