
package ip

import (
    "github.com/pprunty/magikarp/pkg/agent"
    "github.com/pprunty/magikarp/pkg/loader"
    "github.com/pprunty/magikarp/pkg/plugins/ip/resize"
    "github.com/pprunty/magikarp/pkg/plugins/ip/convert"
)

type ipPlugin struct {
    *agent.BasePlugin
}

func New() agent.Plugin {
    p := &ipPlugin{
        BasePlugin: agent.NewBasePlugin("ip", "Image processing utilities (resize / convert)"),
    }
    p.AddTool(resize.Definition())
    p.AddTool(convert.Definition())
    return p
}

func init() {
    loader.Register(New())
}
