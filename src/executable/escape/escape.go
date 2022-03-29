package main

import (
	"flag"
	"fmt"
	"net/url"

	"github.com/fatih/color"
	"github.com/grewwc/go_tools/src/terminalW"
)

func main() {
	fs := flag.NewFlagSet("fs", flag.ExitOnError)
	fs.Bool("un", false, "unescape the url string")
	fs.Bool("p", false, "PathEscape (default is QueryEscape)")
	parsed := terminalW.ParseArgsCmd("un", "p")
	escape := true
	if parsed == nil || parsed.ContainsFlagStrict("h") {
		fs.PrintDefaults()
		return
	}
	if parsed != nil && parsed.ContainsFlag("un") {
		escape = false
	}
	var res string
	var err error
	pos := parsed.Positional.ToStringSlice()
	if len(pos) != 1 {
		panic("must have 1 positional argument")
	}
	if escape {
		if parsed.ContainsFlagStrict("p") {
			res = url.PathEscape(pos[0])
		} else {
			res = url.QueryEscape(pos[0])
		}
		fmt.Println(color.HiBlueString(res))
		return
	}
	// unescape
	if parsed.ContainsFlagStrict("p") {
		res, err = url.PathUnescape(pos[0])
	} else {
		res, err = url.QueryUnescape(pos[0])
	}
	if err != nil {
		panic(err)
	}
	fmt.Println(color.HiBlueString(res))
}
