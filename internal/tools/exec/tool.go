package exec

import (
	"github.com/pprunty/magikarp/internal/tools"
	"github.com/pprunty/magikarp/internal/tools/exec/bash"
)

type execToolbox struct {
	*tools.BaseToolbox
}

func New() tools.Toolbox {
	tb := &execToolbox{
		BaseToolbox: tools.NewBaseToolbox("execution", "Execute shell commands"),
	}
	tb.AddTool(bash.Definition())
	return tb
}

func init() {
	tools.Register(New())
}
