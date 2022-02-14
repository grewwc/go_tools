package main

import (
	"flag"
	"fmt"
	"net/url"

	"github.com/grewwc/go_tools/src/terminalW"
)

func main() {
	fs := flag.NewFlagSet("fs", flag.ExitOnError)
	fs.Bool("un", false, "unescape the url string")
	parsed := terminalW.ParseArgsCmd()
	escape := true
	if parsed == nil || parsed.ContainsFlagStrict("h"){
		fs.PrintDefaults()
		return
	}
	if parsed != nil || parsed.ContainsFlag("un"){
		escape = false
	}
	var res string
	var err error
	pos := parsed.Positional.ToStringSlice()
	if len(pos) != 1{
		panic("must have 1 positional argument")
	}
	if escape{
		res = url.QueryEscape(pos[0])
		fmt.Println(res)
		return 
	}
	// unescape 
	res, err = url.QueryUnescape(pos[0])
	if err != nil{
		panic(err)
	}
	fmt.Println(res)
}