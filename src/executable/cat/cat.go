package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/fatih/color"
	"github.com/grewwc/go_tools/src/terminalW"
)

func init() {
	terminalW.EnableVirtualTerminal()
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("pass files as arguments")
		os.Exit(1)
	}

	for _, filename := range args {
		fmt.Println(color.HiGreenString("======>\t%s\n", filename))
		f, err := os.Open(filename)
		if err != nil {
			log.Println(err)
			continue
		}
		io.Copy(os.Stdout, f)
		f.Close()
		fmt.Printf("\n\n")
	}
}
