package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/fatih/color"
	"github.com/grewwc/go_tools/src/terminalw"
	"github.com/grewwc/go_tools/src/utilw"
)

const (
	helpMsg = "usage: totime 1603372219690"
)

func main() {
	parser := terminalw.NewParser()
	parser.Bool("ts", false, "get current timestamp")
	parser.ParseArgsCmd()
	if parser.Empty() || parser.ContainsFlagStrict("-h") {
		parser.PrintDefaults()
		fmt.Println(color.GreenString(helpMsg))
		return
	}
	if parser.ContainsFlagStrict("ts") {
		fmt.Printf("%v (ms)\n", (time.Now().Local().UnixNano())/int64(1e6))
		return
	}
	posArr := parser.Positional.ToStringSlice()
	if len(posArr) != 1 {
		panic(helpMsg)
	}
	unixTime, err := strconv.Atoi(posArr[0])
	if err != nil {
		fmt.Println(utilw.ToUnix(posArr[0]), "s")
		return
	}
	res := time.Unix(int64(unixTime), 0)
	if res.After(time.Date(2500, time.January, 1, 0, 0, 0, 0, time.Local)) {
		res = time.Unix(int64(unixTime/1000), 0)
	}
	fmt.Println(res)
}
