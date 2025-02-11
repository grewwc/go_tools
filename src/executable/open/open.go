//go:build windows
// +build windows

package main

import (
	"os/exec"

	"github.com/grewwc/go_tools/src/terminalW"
)

func main() {
	commands := []string{"/C", "start", ""}
	parsed := terminalW.ParseArgsCmd()
	if !parsed.Empty() {
		pos := parsed.Positional.ToStringSlice()
		if len(pos) > 1 {
			panic("can only open 1 file at a time")
		}
		commands = append(commands, pos[0])
	} else {
		commands[len(commands)-1] = "."
	}
	cmd := exec.Command("cmd.exe", commands...)
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}
