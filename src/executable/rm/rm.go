package main

import (
	"fmt"
	"log"
	"os"

	"github.com/grewwc/go_tools/src/terminalW"
	"github.com/grewwc/go_tools/src/utilsW"
)

func removeSingle(filename string, parsedResults terminalW.ParsedResults) {
	if parsedResults.ContainsFlagStrict("-rf") {
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
	parsedResults := terminalW.ParseArgsCmd("rf")
	if parsedResults == nil {
		fmt.Println("usage: rm -rf ")
		return
	}
	args := parsedResults.Positional.ToStringSlice()
	for _, filename := range args {
		for d, filenames := range utilsW.LsDirGlob(filename) {
			if d == "./" {
				for _, fname := range filenames {
					removeSingle(fname, *parsedResults)
				}
			} else {
				removeSingle(d, *parsedResults)
			}
		}
	}
}
