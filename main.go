package main

import (
	"github.com/pprunty/magikarp/cmd"
	_ "github.com/pprunty/magikarp/internal/tools/core"
	_ "github.com/pprunty/magikarp/internal/tools/exec"
	_ "github.com/pprunty/magikarp/internal/tools/filesystem"
)

func main() {
	cmd.Execute()
}
