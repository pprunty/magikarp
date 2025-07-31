package filesystem

import (
	"github.com/pprunty/magikarp/internal/tools"
	"github.com/pprunty/magikarp/internal/tools/filesystem/read_file"
)

type fsToolbox struct {
	*tools.BaseToolbox
}

func New() tools.Toolbox {
	tb := &fsToolbox{
		BaseToolbox: tools.NewBaseToolbox("filesystem", "File system operations"),
	}
	tb.AddTool(read_file.Definition())
	return tb
}

func init() {
	tools.Register(New())
}
