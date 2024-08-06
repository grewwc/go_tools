package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/grewwc/go_tools/src/terminalW"
	"github.com/grewwc/go_tools/src/utilsW"
	"golang.design/x/clipboard"
)

var (
	binary      bool = false
	toClipboard bool = false
)

func init() {
	if err := clipboard.Init(); err != nil {
		panic(err)
	}
}

func isCopyAction(parsed *terminalW.ParsedResults) bool {
	return parsed.ContainsFlag("c")
}

func checkInput(parsed *terminalW.ParsedResults) {
	if parsed.ContainsFlag("c") && parsed.ContainsFlag("p") {
		fmt.Println(color.HiRedString("-c/-p, only 1 argument can be set"))
		os.Exit(1)
	}
	if parsed.ContainsFlag("c") {
		toClipboard = true
		if parsed.ContainsFlag("b") {
			binary = true
		}
	}
	if parsed.Positional.Size() > 1 {
		fmt.Println(color.HiRedString("have at most 1 positional arg"))
	}
}

func copyToClipboard(parsed *terminalW.ParsedResults) {
	// get the type
	t := clipboard.FmtText
	if binary {
		t = clipboard.FmtImage
	}
	// get the data
	var data []byte
	var err error
	var filename string
	arg := parsed.Positional.ToStringSlice()
	if len(arg) == 1 {
		filename = arg[0]
	} else { // len(arg) == 0
		fmt.Print(">>> input the filename: ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		filename = scanner.Text()
	}
	data, err = os.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	clipboard.Write(t, data)
	fmt.Println("<<< DONE copying to clipboard")
}

func readFromClipboard(parsed *terminalW.ParsedResults) {
	var filename string
	if parsed.Positional.Size() < 1 {
		fmt.Print(">>> input the filename: ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		filename = scanner.Text()
	} else {
		filename = parsed.Positional.ToStringSlice()[0]
	}
	filename = strings.TrimSpace(filename)
	// get the type
	t := clipboard.FmtText
	if parsed.ContainsFlag("b") {
		t = clipboard.FmtImage
	}
	b := clipboard.Read(t)
	write := true
	if utilsW.IsExist(filename) {
		if !utilsW.PromptYesOrNo(fmt.Sprintf("file: %s already exists, do you want to overwrite it? (y/n)",
			color.HiRedString(filename))) {
			write = false
		}
	}
	// overwrite
	if write {
		if len(b) == 0 {
			_, err := utilsW.RunCmdWithTimeout(fmt.Sprintf("pngpaste %s", filename), 10*time.Second)
			fmt.Println("here", err)
			if err != nil {
				panic(err)
			}
		}

		if err := utilsW.WriteToFile(filename, b); err != nil {
			log.Fatalln(err)
		}
	}
	fmt.Printf("<<< DONE pasting from clipboard, write to file: %s\n", filename)
}

func main() {
	fmt.Println("begin")
	os.Exit(0)
	fs := flag.NewFlagSet("fs", flag.ExitOnError)

	fs.Bool("t", true, "text data")
	fs.Bool("b", false, "binary data")
	fs.Bool("c", true, "copy to clipboard")
	fs.Bool("p", false, "paste from clipboard")

	parsed := terminalW.ParseArgsCmd("t", "b", "h", "c", "p")
	if parsed == nil {
		fs.PrintDefaults()
		return
	}

	checkInput(parsed)

	if isCopyAction(parsed) {
		copyToClipboard(parsed)
	} else {
		readFromClipboard(parsed)
	}
}
