package search

import (
    "github.com/pprunty/magikarp/pkg/agent"
    "github.com/pprunty/magikarp/pkg/loader"
)

type searchPlugin struct {
    *agent.BasePlugin
}

func New() agent.Plugin {
    p := &searchPlugin{
        BasePlugin: agent.NewBasePlugin("search", "Plugin description"),
    }
    // Add your tools here
    // p.AddTool(tool1.Definition())
    // p.AddTool(tool2.Definition())
    return p
}

func init() {
    loader.Register(New())
}