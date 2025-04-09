package utilw

import (
	"bytes"

	"github.com/grewwc/go_tools/src/conw"
)

func ReverseBytes(arr []byte) {
	last := len(arr) - 1
	mid := len(arr) / 2
	for i := 0; i < mid; i++ {
		arr[i], arr[last-i] = arr[last-i], arr[i]
	}
}

func ReverseInts(arr []int) {
	last := len(arr) - 1
	mid := len(arr) / 2
	for i := 0; i < mid; i++ {
		arr[i], arr[last-i] = arr[last-i], arr[i]
	}
}

func ReverseInt64(arr []int) {
	last := len(arr) - 1
	mid := len(arr) / 2
	for i := 0; i < mid; i++ {
		arr[i], arr[last-i] = arr[last-i], arr[i]
	}
}

func ReverseInt32(arr []int) {
	last := len(arr) - 1
	mid := len(arr) / 2
	for i := 0; i < mid; i++ {
		arr[i], arr[last-i] = arr[last-i], arr[i]
	}
}

func ReverseFloat64(arr []float64) {
	last := len(arr) - 1
	mid := len(arr) / 2
	for i := 0; i < mid; i++ {
		arr[i], arr[last-i] = arr[last-i], arr[i]
	}
}

func ReverseFloat32(arr []float64) {
	last := len(arr) - 1
	mid := len(arr) / 2
	for i := 0; i < mid; i++ {
		arr[i], arr[last-i] = arr[last-i], arr[i]
	}
}

func ReverseString(s string) string {
	stack := conw.NewStack(len(s))
	for _, r := range s {
		stack.Push(r)
	}
	buf := bytes.Buffer{}
	for !stack.Empty() {
		buf.WriteRune(stack.Pop().(rune))
	}
	return buf.String()
}
