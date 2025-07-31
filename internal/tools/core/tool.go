package core

import (
	"github.com/pprunty/magikarp/internal/tools"
	"github.com/pprunty/magikarp/internal/tools/core/control_state"
	"github.com/pprunty/magikarp/internal/tools/core/get_model_version"
	"github.com/pprunty/magikarp/internal/tools/core/list_tools"
)

type coreToolbox struct{ *tools.BaseToolbox }

func New() tools.Toolbox {
	tb := &coreToolbox{tools.NewBaseToolbox("core", "Core Magikarp tools")}
	tb.AddTool(list_tools.Definition())
	tb.AddTool(get_model_version.Definition())
	tb.AddTool(control_state.Definition())
	return tb
}

func init() { tools.Register(New()) }
