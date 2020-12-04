package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/grewwc/go_tools/src/utilsW"
	"github.com/grewwc/go_tools/src/windowsW"
)

func init() {
	windowsW.EnableVirtualTerminal()
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("pass files as arguments")
		os.Exit(1)
	}
	for _, name := range args {
		filenames, err := filepath.Glob(name)
		if err != nil {
			log.Println(err)
			continue
		}
		for _, filename := range filenames {
			if utilsW.IsDir(filename) {
				continue
			}
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
}
