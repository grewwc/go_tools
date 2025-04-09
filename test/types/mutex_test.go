package test

import (
	"sync"
	"testing"
	"time"

	"github.com/grewwc/go_tools/src/typew"
)

var mu = typew.NewReentrantMutex()

var count = 0

func TestReentrantMutex(t *testing.T) {
	var wg sync.WaitGroup
	add := func() {
		defer wg.Done()
		for i := 0; i < 100000; i++ {
			mu.Lock()
			count++
			mu.Unlock()
			time.Sleep(10 * time.Microsecond)
		}
	}

	minus := func() {
		defer wg.Done()
		for i := 0; i < 100000; i++ {
			mu.Lock()
			count--
			mu.Unlock()
			time.Sleep(5 * time.Microsecond)
		}
	}

	wg.Add(2)
	go add()
	go minus()
	wg.Wait()
	if count != 0 {
		t.Fatalf("count is not 0 (%d) \n", count)
	}
}

func BenchmarkReentrantMutex(b *testing.B) {
	var wg sync.WaitGroup

	add := func() {
		defer wg.Done()
		for i := 0; i < 100000; i++ {
			mu.Lock()
			count++
			mu.Unlock()
		}
	}

	minus := func() {
		defer wg.Done()
		for i := 0; i < 100000; i++ {
			mu.Lock()
			count--
			mu.Unlock()
		}
	}

	for i := 0; i < b.N; i++ {
		wg.Add(2)
		go add()
		go minus()
		wg.Wait()
	}
}

func BenchmarkCompareFunc(b *testing.B) {
	f := typew.CreateDefaultCmp[int]()
	// f := func(a, b any) int {
	// 	return a.(int) - b.(int)
	// }
	for i := 0; i < b.N; i++ {
		// _ = 1 - 3
		f(1, 3)
	}
}
