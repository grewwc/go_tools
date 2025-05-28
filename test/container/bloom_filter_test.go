package test

import (
	"sync"
	"testing"

	"github.com/grewwc/go_tools/src/cw"
)

func TestBloomFilter(t *testing.T) {
	N := 100
	f := cw.NewBloomFilter[int](N * 100)
	var wg sync.WaitGroup
	for i := 0; i < N; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			f.Add(i)
		}(i)
	}
	wg.Wait()

	for i := 0; i < N/2; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			f.Delete(i)
		}(i)
	}
	wg.Wait()

	for i := N / 2; i < N; i++ {
		if !f.MayExist(i) {
			t.Error("wrong")
		}
	}
	cnt := 0
	for i := 0; i < 10*N; i++ {
		if f.MayExist(i) {
			cnt++
		}
	}
	t.Logf("stats: %d/%d", cnt, N)
}
