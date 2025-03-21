package main

import (
	"fmt"
	"net/url"

	"github.com/fatih/color"
	"github.com/grewwc/go_tools/src/terminalW"
)

func main() {
	parser := terminalW.NewParser()
	parser.Bool("un", false, "unescape the url string")
	parser.Bool("p", false, "PathEscape (default is QueryEscape)")
	parser.ParseArgsCmd("un", "p")
	escape := true
	if parser.Empty() || parser.ContainsFlagStrict("h") {
		parser.PrintDefaults()
		return
	}
	if parser.ContainsFlag("un") {
		escape = false
	}
	var res string
	var err error
	pos := parser.Positional.ToStringSlice()
	if len(pos) != 1 {
		panic("must have 1 positional argument")
	}
	if escape {
		if parser.ContainsFlagStrict("p") {
			res = url.PathEscape(pos[0])
		} else {
			res = url.QueryEscape(pos[0])
		}
		fmt.Println(color.HiBlueString(res))
		return
	}
	// unescape
	if parser.ContainsFlagStrict("p") {
		res, err = url.PathUnescape(pos[0])
	} else {
		res, err = url.QueryUnescape(pos[0])
	}
	if err != nil {
		panic(err)
	}
	fmt.Println(color.HiBlueString(res))
}
