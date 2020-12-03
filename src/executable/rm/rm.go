package main

import (
	"fmt"
	"log"
	"os"

	"github.com/grewwc/go_tools/src/terminalW"
)

func main() {
	parsedResults := terminalW.ParseArgsCmd("rf")
	args := parsedResults.Positional.ToStringSlice()
	var filename string
	switch len(args) {
	case 1:
		filename = args[0]
	default:
		fmt.Println("need at least 1 argument")
		return
	}
	if parsedResults.ContainsFlag("-rf") {
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
