package test

import (
	"testing"

	"github.com/grewwc/go_tools/src/cw"
)

func TestBloomFilter(t *testing.T) {
	N := 1000
	f := cw.NewBloomFilter[int](N * 100)
	for i := 0; i < N; i++ {
		f.Add(i)
	}

	for i := 0; i < N/2; i++ {
		f.Delete(i)
	}

	for i := N / 2; i < N; i++ {
		if !f.Contains(i) {
			t.Error("wrong")
		}
	}
	cnt := 0
	for i := 0; i < 10*N; i++ {
		if f.Contains(i) {
			cnt++
		}
	}
	t.Logf("stats: %d/%d", cnt, N)
}
