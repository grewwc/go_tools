package test

import (
	"math/rand"
	"sort"
	"testing"

	"github.com/grewwc/go_tools/src/algow"
	"github.com/grewwc/go_tools/src/cw"
	"github.com/grewwc/go_tools/src/sortw"
)

const (
	N = 500000
)

func TestShellSort(t *testing.T) {
	for i := 0; i < 100; i++ {
		arr := algow.RandInt(0, 100, 500)
		sortw.ShellSort(arr)
		if !sort.IntsAreSorted(arr) {
			t.Errorf("arr is not sorted, %v", arr)
		}
	}
}

func TestTreeSort(t *testing.T) {
	for i := 0; i < 100; i++ {
		st := cw.NewRbTree[int](nil)
		arr := algow.RandInt(0, 100, 500)
		for _, val := range arr {
			st.Insert(val)
		}
		arr = []int{}
		for val := range st.SearchRange(0, 100).Iterate() {
			arr = append(arr, val)
		}
		if !sort.IntsAreSorted(arr) {
			t.Errorf("arr is not sorted, %v", arr)
		}
	}
}

func TestHeapSort(t *testing.T) {
	for i := 0; i < 100; i++ {
		arr := algow.RandInt(-100, 100, 500)
		sortw.HeapSort(arr, false)
		if !sort.IntsAreSorted(arr) {
			t.Errorf("arr is not sorted, %v", arr)
		}
	}
}

func TestQuickSort(t *testing.T) {
	for i := 0; i < 100; i++ {
		arr := algow.RandInt(0, 500, 1000)
		sortw.QuickSort(arr)
		if !sort.IntsAreSorted(arr) {
			t.Errorf("arr is not sorted, %v", arr)
		}
	}
}

func TestStableSort(t *testing.T) {
	for i := 0; i < 100; i++ {
		arr := algow.RandInt(0, 500, 1000)
		sortw.StableSort(arr, func(a, b int) int { return a - b })
		if !sort.IntsAreSorted(arr) {
			t.Errorf("arr is not sorted, %v", arr)
		}
	}
}

func TestRadixSort(t *testing.T) {
	for i := 0; i < 100; i++ {
		arr := algow.RandInt(0, 500, 1000)
		sortw.RadixSort(arr)
		if !sort.IntsAreSorted(arr) {
			t.Errorf("arr is not sorted, %v", arr)
		}
	}
}

func TestCountSort(t *testing.T) {
	for i := 0; i < 100; i++ {
		arr := algow.RandInt(-500, 500, 10000)
		sortw.CountSort(arr)
		if !sort.IntsAreSorted(arr) {
			t.Errorf("arr is not sorted, %v", arr)
		}
	}
}

func BenchmarkQuickSort(b *testing.B) {
	arr := algow.RandInt(0, 10000, N)
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		rand.Shuffle(len(arr), func(i, j int) {
			arr[i], arr[j] = arr[j], arr[i]
		})
		b.StartTimer()
		sortw.QuickSort(arr)
	}
}

func BenchmarkQuickSortCmp(b *testing.B) {
	arr := algow.RandInt(0, 10000, N)
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		rand.Shuffle(len(arr), func(i, j int) {
			arr[i], arr[j] = arr[j], arr[i]
		})
		b.StartTimer()
		sortw.Sort(arr, func(i, j int) int { return i - j })
	}
}

func BenchmarkShellSort(b *testing.B) {
	arr := algow.RandFloat64(N)
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		rand.Shuffle(len(arr), func(i, j int) {
			arr[i], arr[j] = arr[j], arr[i]
		})
		b.StartTimer()
		sortw.ShellSort(arr)
	}
}

func BenchmarkSort(b *testing.B) {
	arr := algow.RandInt(0, 10000, N)
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		rand.Shuffle(len(arr), func(i, j int) {
			arr[i], arr[j] = arr[j], arr[i]
		})
		b.StartTimer()
		sort.Ints(arr)
	}
}

func BenchmarkStableSort(b *testing.B) {
	arr := algow.RandInt(0, 10000, N)
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		rand.Shuffle(len(arr), func(i, j int) {
			arr[i], arr[j] = arr[j], arr[i]
		})
		b.StartTimer()
		sortw.StableSort(arr, func(a, b int) int { return a - b })
	}
}

func TestKth(t *testing.T) {
	arr := algow.RandInt(0, 1000, 500)
	for i := 0; i < 100; i++ {
		k := algow.RandInt(0, len(arr), 1)[0]
		val := algow.Kth(arr, k, nil)
		sortw.Sort(arr, nil)
		if val != arr[k] {
			t.Fatal("wrong")
		}
	}
}

func BenchmarkKth(b *testing.B) {
	arr := algow.RandInt(0, 10000, N)
	for i := 0; i < b.N; i++ {
		k := algow.RandInt(0, len(arr), 1)[0]
		algow.Kth(arr, k, func(i, j int) int { return i - j })
	}
}
