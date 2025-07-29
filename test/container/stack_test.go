package test

import (
	"container/list"
	"testing"

	"github.com/grewwc/go_tools/src/cw"
)

const (
	N = 100000
)

func BenchmarkStackAdd(b *testing.B) {
	st := cw.NewStack[int]()
	for i := 0; i < b.N; i++ {
		for j := 0; j < N; j++ {
			st.Push(j)
			if j%3 == 0 {
				st.Pop()
			}
		}
	}
}
func BenchmarkListAdd(b *testing.B) {
	st := list.New()
	for i := 0; i < b.N; i++ {
		for j := 0; j < N; j++ {
			st.PushFront(j)
			if j%3 == 0 {
				st.Remove(st.Front())
			}
		}
	}
}
