package test

import (
	"testing"

	"github.com/grewwc/go_tools/src/cw"
)

var (
	st = cw.NewStack(10)
)

func BenchmarkStackAdd(b *testing.B) {
	for i := 0; i < b.N; i++ {
		st.Push(1024)
	}
}
