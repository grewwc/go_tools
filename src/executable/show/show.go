package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/grewwc/go_tools/src/stringsW"
	"github.com/grewwc/go_tools/src/terminalW"
	"github.com/grewwc/go_tools/src/utilsW"
)

func main() {
	parsed := terminalW.ParseArgsCmd("h")
	if parsed != nil && parsed.ContainsAnyFlagStrict("h") {
		fmt.Println("print all the binary files of go_tools")
		return
	}

	dir := utilsW.GetDirOfTheFile()
	dir = filepath.Join(dir, "..", "..", "..", "bin")
	var allExecutables []string
	for _, name := range utilsW.LsDir(dir) {
		allExecutables = append(allExecutables, name)
	}
	_, w, err := utilsW.GetTerminalSize()
	if err != nil {
		panic(err)
	}
	fmt.Println(color.HiRedString(stringsW.Wrap(strings.Join(allExecutables, " "), w, 3, "  ")))
}
