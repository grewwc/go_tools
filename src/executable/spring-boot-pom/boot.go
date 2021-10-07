package main

import (
	"flag"

	"github.com/grewwc/go_tools/src/terminalW"
)

const (
	parent = "parent"
	web    = "web"
)

func main() {
	fs := flag.NewFlagSet("fs", flag.ExitOnError)

	fs.Bool("ls", false, "list all possible choices")

	parsed := terminalW.ParseArgsCmd("ls")

	if parsed == nil {
		fs.PrintDefaults()
		return
	}

}
