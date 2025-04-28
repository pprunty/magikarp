
package main
import (
    "fmt"
    "os"
    icli "github.com/pprunty/magikarp/pkg/cli"
    "github.com/urfave/cli/v2"
)
func main(){
    app:=&cli.App{
        Name:"mgk",
        Usage:"Magikarp scaffolding CLI",
        Commands: []*cli.Command{
            icli.NewCreateToolCmd(),
            icli.NewCreatePluginCmd(),
            icli.NewCreateAgentCmd(),
        },
    }
    if err:=app.Run(os.Args); err!=nil{
        fmt.Fprintln(os.Stderr,err)
        os.Exit(1)
    }
}
