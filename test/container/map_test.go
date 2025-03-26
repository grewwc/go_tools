package test

import (
	"sync"
	"testing"

	"github.com/grewwc/go_tools/src/containerW"
	"github.com/grewwc/go_tools/src/typesW"
)

func TestConcurrentMap(t *testing.T) {
	// 创建一个大小为 10 的 ConcurrentHashMap
	m := containerW.NewConcurrentHashMap[int, int](typesW.CreateDefaultHash[int](), nil)
	N := 10000
	// 启动多个 goroutine 模拟并发操作
	var wg sync.WaitGroup
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

}

func BenchmarkConcurrentMap(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		m := containerW.NewConcurrentHashMap[int, int](nil, func(i, j int) int {
			return i - j
		})
		// m := containerW.NewMutexMap[int, int]()
		N := 100000
		for i := 0; i < N; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				m.Put(i, i*i)
			}(i)
		}
		wg.Wait()
		for i := 0; i < N; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				m.Delete(i)
			}(i)
		}
		wg.Wait()
	}
}
