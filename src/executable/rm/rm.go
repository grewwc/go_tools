//go:build windows
// +build windows

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/grewwc/go_tools/src/terminalw"
	"github.com/grewwc/go_tools/src/utilw"
)

func removeSingle(filename string, parser terminalw.Parser) {
	if parser.ContainsFlagStrict("-rf") {
		err := os.RemoveAll(filename)
		if err != nil {
			log.Println(err)
		}
	} else {
		err := os.Remove(filename)
		if err != nil {
			log.Println(err)
		}
	}
}
func main() {
	parser := terminalw.NewParser()
	parser.ParseArgsCmd("rf")
	if parser == nil {
		fmt.Println("usage: rm -rf ")
		return
	}
	args := parser.Positional.ToStringSlice()
	for _, filename := range args {
		for d, filenames := range utilw.LsDirGlob(filename) {
			if d == "./" {
				for _, fname := range filenames {
					removeSingle(fname, *parser)
				}
			} else {
				removeSingle(d, *parser)
			}
		}
	}
}
