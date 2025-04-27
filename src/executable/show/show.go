package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/grewwc/go_tools/src/strw"
	"github.com/grewwc/go_tools/src/terminalw"
	"github.com/grewwc/go_tools/src/utilsw"
)

func main() {
	parser := terminalw.NewParser()
	parser.ParseArgsCmd("h")
	if parser.ContainsAnyFlagStrict("h") {
		fmt.Println("print all the binary files of go_tools")
		return
	}

	dir := utilsw.GetDirOfTheFile()
	dir = filepath.Join(dir, "..", "..", "..", "bin")
	var allExecutables []string
	allExecutables = append(allExecutables, utilsw.LsDir(dir, nil, nil)...)
	_, w, err := utilsw.GetTerminalSize()
	if err != nil {
		panic(err)
	}
	fmt.Println(strw.Wrap(strings.Join(allExecutables, " "), w, 3, "  "))
}
