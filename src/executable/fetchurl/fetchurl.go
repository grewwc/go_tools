package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func main() {
	if len(os.Args) != 2 && len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: fetchurl url [save file name]")
		return
	}
	url := os.Args[1]
	var fname string
	if len(os.Args) == 3 {
		fname = os.Args[2]
	}
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	writesTo := []io.Writer{}
	if fname != "" {
		f, err := os.OpenFile(fname, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			log.Fatalln(err)
		}
		writesTo = append(writesTo, f)
	} else {
		writesTo = append(writesTo, os.Stdout)
	}
	multiWriter := io.MultiWriter(writesTo...)
	io.Copy(multiWriter, resp.Body)
	fmt.Println()
}
