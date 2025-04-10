package test

import (
	"sync"
	"testing"

	"github.com/grewwc/go_tools/src/cw"
	"github.com/grewwc/go_tools/src/typew"
)

func TestConcurrentMap(t *testing.T) {
	// 创建一个大小为 10 的 ConcurrentHashMap
	m := cw.NewConcurrentHashMap[int, int](typew.CreateDefaultHash[int](), nil)
	N := 100000
	// 启动多个 goroutine 模拟并发操作
	var wg sync.WaitGroup
	for l := 0; l < 10; l++ {
		for i := 0; i < N; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				m.Put(i, i*i) // 插入键值对
			}(i)
		}
		wg.Wait()
		if m.Size() != N {
			t.Errorf("size wrong. Expected: %d, Actual: %d", N, m.Size())
		}
		for i := 0; i < N; i++ {
			if m.GetOrDefault(i, -1) == -1 {
				t.Errorf("%d is not in map. size: %d", i, m.Size())
			}
		}
		// delete
		for i := 0; i < N/2; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				if !m.Contains(i) {
					t.Error("contains wrong")
				}
				m.Delete(i)
				if m.Contains(i) {
					t.Error("contains wrong")
				}
			}(i)
		}
		wg.Wait()
		if m.Size() != N/2 {
			t.Errorf("size wrong. Expected: %d, Actual: %d", N/2, m.Size())
		}
	}

}

func BenchmarkConcurrentMap(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		m := cw.NewConcurrentHashMap[int, int](nil, func(i, j int) int {
			return i - j
		})
		// m := cw.NewMutexMap[int, int]()
		N := 100000
		for i := 0; i < N; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				m.Put(i, i*i)
				m.Get(i * 2)
			}(i)
		}
		wg.Wait()
		for i := 0; i < N; i++ {
			wg.Add(1)
			go func(i int) {
				if !m.Contains(i) {
					b.Error("wrong")
				}
				defer wg.Done()
				m.Delete(i)
			}(i)
		}
		wg.Wait()
	}
}
