package main

import (
	"fmt"
	"github.com/grewwc/go_tools/src/containerW"
)

func main() {
	t := containerW.NewTrie()
	t.Insert("good")
	fmt.Println(t.Search("goods"))
}
