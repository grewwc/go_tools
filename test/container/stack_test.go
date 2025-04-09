package test

import (
	"testing"

	"github.com/grewwc/go_tools/src/conw"
)

var (
	st = conw.NewStack(10)
)

func BenchmarkStackAdd(b *testing.B) {
	for i := 0; i < b.N; i++ {
		st.Push(1024)
	}
}
