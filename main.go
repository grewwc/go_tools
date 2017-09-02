package main

import "systools"

// "systools"

const (
	first = iota
	_
	_
	second = 1 << iota
)

func test(i int) int {
	return i
}

func main() {

	systools.Move("./to", "./from")
	// fmt.Println(filepath.Abs("."))
}
