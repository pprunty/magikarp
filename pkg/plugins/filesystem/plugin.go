package filesystem

import (
	"github.com/pprunty/magikarp/pkg/agent"
	"github.com/pprunty/magikarp/pkg/loader"
	"github.com/pprunty/magikarp/pkg/plugins/filesystem/read_file"
	"github.com/pprunty/magikarp/pkg/plugins/filesystem/edit_file"
	"github.com/pprunty/magikarp/pkg/plugins/filesystem/list_files"
)

type fsPlugin struct {
	*agent.BasePlugin
}

func New() agent.Plugin {
	p := &fsPlugin{
		BasePlugin: agent.NewBasePlugin("filesystem", "Read and edit files on disk"),
	}
	p.AddTool(read_file.Definition())
	p.AddTool(edit_file.Definition())
	p.AddTool(list_files.Definition())
	return p
}

func init() {
	loader.Register(New())
} 