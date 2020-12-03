package main

import (
	"log"
	"os"

	"github.com/grewwc/go_tools/src/utilsW"
)

func main() {
	filenames := os.Args[1:]
	for _, filename := range filenames {
		if utilsW.IsExist(filename) {
			log.Printf("file: %s existed\n", filename)
			continue
		}
		if _, err := os.Create(filename); err != nil {
			log.Println(err)
		}

	}
}
