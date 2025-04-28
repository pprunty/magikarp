package execution

import (
	"github.com/pprunty/magikarp/pkg/agent"
	"github.com/pprunty/magikarp/pkg/loader"
	"github.com/pprunty/magikarp/pkg/plugins/execution/execute_command"
	"github.com/pprunty/magikarp/pkg/plugins/execution/list_tools"
)

type execPlugin struct {
	*agent.BasePlugin
}

func New() agent.Plugin {
	p := &execPlugin{
		BasePlugin: agent.NewBasePlugin("execution", "Execute commands and list tools"),
	}
	p.AddTool(execute_command.Definition())
	p.AddTool(list_tools.Definition())
	return p
}

func init() {
	loader.Register(New())
} 