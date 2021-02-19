package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/grewwc/go_tools/src/utilsW"
)

func create(fname string) error {
	if !utilsW.IsExist(filepath.Dir(fname)) {
		os.MkdirAll(filepath.Dir(fname), os.ModeDir)
	}
	_, err := os.Create(fname)
	return err
}

func main() {
	filenames := os.Args[1:]
	for _, filename := range filenames {
		if utilsW.IsExist(filename) {
			log.Printf("file: %s existed\n", filename)
			continue
		}
		if err := create(filename); err != nil {
			log.Println(err)
		}
	}
}
