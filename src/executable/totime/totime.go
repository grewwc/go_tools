package main

import (
	"flag"
	"fmt"
	"strconv"
	"time"

	"github.com/fatih/color"
	"github.com/grewwc/go_tools/src/terminalW"
)

const (
	helpMsg = "usage: totime 1603372219690"
)

func main() {
	fs := flag.NewFlagSet("fs", flag.ExitOnError)
	fs.Bool("ts", false, "get current timestamp")
	parsed := terminalW.ParseArgsCmd()
	if parsed == nil || parsed.ContainsFlagStrict("-h") {
		fs.PrintDefaults()
		fmt.Println(color.GreenString(helpMsg))
		return
	}
	if parsed.ContainsFlagStrict("ts") {
		fmt.Println(time.Now().Unix())
		return
	}
	posArr := parsed.Positional.ToStringSlice()
	if len(posArr) != 1 {
		panic(helpMsg)
	}
	unixTime, err := strconv.Atoi(posArr[0])
	if err != nil {
		panic(err)
	}
	res := time.Unix(int64(unixTime), 0)
	if res.After(time.Date(2500, time.January, 1, 0, 0, 0, 0, time.Local)) {
		res = time.Unix(int64(unixTime/1000), 0).UTC()
	}
	fmt.Println(res)
}
