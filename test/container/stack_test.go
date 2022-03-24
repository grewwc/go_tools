package test

import (
	"testing"

	"github.com/grewwc/go_tools/src/containerW"
)

var (
	st = containerW.NewStack(10)
)

func BenchmarkStackAdd(b *testing.B) {
	for i := 0; i < b.N; i++ {
		st.Push(1024)
	}
}
