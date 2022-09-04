package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/fatih/color"
	"github.com/grewwc/go_tools/src/terminalW"
	"github.com/grewwc/go_tools/src/utilsW"
	"golang.design/x/clipboard"
)

func init() {
	if err := clipboard.Init(); err != nil {
		panic(err)
	}
}

func isCopyAction(parsed *terminalW.ParsedResults) bool {
	return parsed.ContainsFlagStrict("c")
}

func checkInput(parsed *terminalW.ParsedResults) {
	if parsed.ContainsAllFlagStrict("c", "p") ||
		!parsed.ContainsAnyFlagStrict("c", "p") {
		fmt.Println(color.HiRedString("-c/-p, only 1 argument can be set"))
		os.Exit(1)
	}
	if parsed.Positional.Size() > 1 {
		fmt.Println(color.HiRedString("have at most 1 positional arg"))
	}
}

func writeToClipboard(parsed *terminalW.ParsedResults) {
	// get the type
	t := clipboard.FmtText
	if parsed.ContainsFlagStrict("b") {
		t = clipboard.FmtImage
	}
	// get the data
	var data []byte
	var err error
	pos := parsed.Positional.ToStringSlice()
	if len(pos) == 1 {
		if t == clipboard.FmtText {
			data = []byte(pos[0])
		} else {
			data, err = ioutil.ReadFile(pos[0])
			if err != nil {
				panic(err)
			}
		}
	} else {
		if t == clipboard.FmtText {
			fmt.Println(">>> input the text: ")
		} else {
			fmt.Print(">>> input the filename: ")
		}
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		line := scanner.Text()
		if t == clipboard.FmtText {
			data = []byte(line)
		} else {
			data, err = ioutil.ReadFile(line)
			if err != nil {
				panic(err)
			}
		}
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
	// get the type
	t := clipboard.FmtText
	if parsed.ContainsFlagStrict("b") {
		t = clipboard.FmtImage
	}
	b := clipboard.Read(t)
	toWrite := true
	if utilsW.IsExist(filename) {
		if !utilsW.PromptYesOrNo(fmt.Sprintf("file: %s already exists, do you want to overwrite it?(y/n)",
			color.HiRedString(filename))) {
			toWrite = false
		}
	}
	// overwrite
	if toWrite {
		if err := utilsW.WriteToFile(filename, b); err != nil {
			panic(err)
		}
	}
	fmt.Println("<<< DONE pasting from clipboard")
}

func main() {
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
		writeToClipboard(parsed)
	} else {
		readFromClipboard(parsed)
	}
}
