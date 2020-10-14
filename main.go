package main

import (
	"fmt"

	"github.com/grewwc/go_tools/src/terminalW"
)

func main() {
	res := terminalW.ParseArgs()
	fmt.Println(res.Positional)
	fmt.Println(res.Optional)
}
