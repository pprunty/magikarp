package cli
import(
    "fmt"
    "os"
    "github.com/urfave/cli/v2"
)

func NewCreatePluginCmd()*cli.Command{
    return &cli.Command{
        Name:"create-plugin",
        Usage:"Create new plugin skeleton",
        Flags: []cli.Flag{
            &cli.StringFlag{Name:"name",Usage:"plugin name"},
            &cli.StringFlag{Name:"description",Usage:"plugin description",Value:"Plugin description"},
        },
        Action: func(c *cli.Context) error{
            name:=c.String("name")
            if name==""{return fmt.Errorf("name required")}
            
            // Create plugin directory
            path:=fmt.Sprintf("pkg/plugins/%s",name)
            if err:=os.MkdirAll(path,0755); err!=nil {return err}
            
            // Create plugin.go
            pluginFile:=fmt.Sprintf("%s/plugin.go",path)
            skel:=fmt.Sprintf(`package %s

import (
    "github.com/pprunty/magikarp/pkg/agent"
    "github.com/pprunty/magikarp/pkg/loader"
)

type %sPlugin struct {
    *agent.BasePlugin
}

func New() agent.Plugin {
    p := &%sPlugin{
        BasePlugin: agent.NewBasePlugin("%s", "%s"),
    }
    // Add your tools here
    // p.AddTool(tool1.Definition())
    // p.AddTool(tool2.Definition())
    return p
}

func init() {
    loader.Register(New())
}`, name, name, name, name, c.String("description"))
            
            return os.WriteFile(pluginFile,[]byte(skel),0644)
        },
    }
}
