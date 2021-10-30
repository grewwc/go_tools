package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/grewwc/go_tools/src/stringsW"
	"github.com/grewwc/go_tools/src/utilsW"
)

func main() {
	var cmdStr, dir string
	switch len(os.Args) {
	case 1:
		dir = "."
	case 2:
		dir = os.Args[1]
	default:
		fmt.Println("too much arguments")
		return
	}
	switch utilsW.GetPlatform() {
	case utilsW.WINDOWS:
		cmdStr = fmt.Sprintf(`cmd /C start "" %q`, dir)
	case utilsW.MAC:
		cmdStr = fmt.Sprintf("/usr/bin/open %q", dir)
	case utilsW.LINUX:
		cmdStr = fmt.Sprintf("xdg-open %q", dir)
	}
	cmdSlice := stringsW.SplitNoEmpty(cmdStr, " ")
	cmd := exec.Command(cmdSlice[0])
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	// fmt.Println()
	cmd.SysProcAttr.CmdLine = strings.Join(cmdSlice[1:], " ")
	if err := cmd.Run(); err != nil {
		log.Fatalln(err)
	}
}
