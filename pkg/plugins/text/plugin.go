package text

import (
	"github.com/pprunty/magikarp/pkg/agent"
	"github.com/pprunty/magikarp/pkg/loader"
	"github.com/pprunty/magikarp/pkg/plugins/text/search_text"
	"github.com/pprunty/magikarp/pkg/plugins/text/analyze_text"
)

type textPlugin struct {
	*agent.BasePlugin
}

func New() agent.Plugin {
	p := &textPlugin{
		BasePlugin: agent.NewBasePlugin("text", "Text manipulation and analysis"),
	}
	p.AddTool(search_text.Definition())
	p.AddTool(analyze_text.Definition())
	return p
}

func init() {
	loader.Register(New())
} 