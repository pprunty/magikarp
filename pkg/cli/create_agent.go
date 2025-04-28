
package cli
import(
    "encoding/json"
    "fmt"
    "os"
    "github.com/urfave/cli/v2"
)
func NewCreateAgentCmd()*cli.Command{
    return &cli.Command{
        Name:"create-agent",
        Usage:"Create new agent definition",
        Flags: []cli.Flag{
            &cli.StringFlag{Name:"name",Usage:"agent name"},
        },
        Action: func(c *cli.Context) error{
            name:=c.String("name")
            if name==""{return fmt.Errorf("name required")}
            def:=map[string]any{
                "name":name,
                "description":"custom agent",
                "plugins":[]string{},
            }
            data,_:=json.MarshalIndent(def,"","  ")
            dir:=fmt.Sprintf("pkg/agents/%s",name)
            if err:=os.MkdirAll(dir,0755); err!=nil {return err}
            return os.WriteFile(dir+"/agent.json",data,0644)
        },
    }
}
