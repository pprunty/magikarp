package text

import (
	"encoding/json"
	"strings"

	"github.com/pprunty/magikarp/pkg/agent"
)

// TextPlugin implements the Plugin interface for text manipulation
type TextPlugin struct{}

// New creates a new TextPlugin instance
func New() *TextPlugin {
	return &TextPlugin{}
}

// Name returns the name of the plugin
func (p *TextPlugin) Name() string {
	return "text"
}

// Description returns a description of what the plugin does
func (p *TextPlugin) Description() string {
	return "Provides tools for text manipulation and transformation"
}

// Tools returns the tools provided by this plugin
func (p *TextPlugin) Tools() []agent.ToolDefinition {
	return []agent.ToolDefinition{
		rot13Tool(),
	}
}

func rot13Tool() agent.ToolDefinition {
	return agent.ToolDefinition{
		Name:        "rot13",
		Description: "Apply ROT13 transformation to text",
		InputSchema: agent.GenerateSchema[Rot13Input](),
		Function:    rot13,
	}
}

type Rot13Input struct {
	Text string `json:"text" jsonschema_description:"The text to transform"`
}

func rot13(input []byte) (string, error) {
	rot13Input := Rot13Input{}
	err := json.Unmarshal(input, &rot13Input)
	if err != nil {
		return "", err
	}

	return rot13Transform(rot13Input.Text), nil
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