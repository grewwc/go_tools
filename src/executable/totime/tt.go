package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/grewwc/go_tools/src/terminalw"
	"github.com/grewwc/go_tools/src/utilsw"
)

const (
	helpMsg = "usage: tt 1603372219690"
)

func main() {
	parser := terminalw.NewParser()
	parser.ParseArgsCmd()
	if parser.ContainsFlagStrict("-h") {
		parser.PrintDefaults()
		fmt.Println(helpMsg)
		return
	}
	if parser.Empty() {
		fmt.Printf("%v ms\n", (time.Now().Local().UnixNano())/int64(1e6))
		return
	}
	posArr := parser.Positional.ToStringSlice()
	if len(posArr) != 1 {
		panic(helpMsg)
	}
	unixTime, err := strconv.Atoi(posArr[0])
	if err != nil {
		res, err := time.ParseInLocation(utilsw.DateTimeFormat, posArr[0], time.Local)
		if err != nil {
			panic(err)
		}
		fmt.Println(res.UnixMilli()/1000, "s")
		return
	}
	res := time.Unix(int64(unixTime), 0)
	if res.After(time.Date(2500, time.January, 1, 0, 0, 0, 0, time.Local)) {
		res = time.Unix(int64(unixTime/1000), 0)
	}
	fmt.Println(res)
}
