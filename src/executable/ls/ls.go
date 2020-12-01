package main

import (
	"fmt"
	"log"

	"github.com/grewwc/go_tools/src/stringsW"
	"github.com/grewwc/go_tools/src/utilsW"
	"github.com/nsf/termbox-go"
)

func main() {
	var files string
	for _, file := range utilsW.LsDir(".") {
		files += file
		if utilsW.IsDir(file) {
			files += "/"
		}
		files += " "
	}
	if err := termbox.Init(); err != nil {
		log.Fatalln(err)
	}
	w, _ := termbox.Size()
	toPrint, err := stringsW.Wrap(files, w, 2)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(toPrint)
}
