package algoriwhtmW

import (
	"math/rand"
	"sort"
	"testing"

	"github.com/grewwc/go_tools/src/algoW"
	"github.com/grewwc/go_tools/src/containerW"
	"github.com/grewwc/go_tools/src/sortW"
)

const (
	N = 500000
)

func TestShellSort(t *testing.T) {
	for i := 0; i < 100; i++ {
		arr := algoW.RandInt(0, 100, 500)
		sortW.ShellSort(arr)
		if !sort.IntsAreSorted(arr) {
			t.Errorf("arr is not sorted, %v", arr)
		}
	}
}

func TestTreeSort(t *testing.T) {
	for i := 0; i < 100; i++ {
		st := containerW.NewRbTree[int](nil)
		arr := algoW.RandInt(0, 100, 500)
		for _, val := range arr {
			st.Insert(val)
		}
		arr = st.SearchRange(0, 100)
		if !sort.IntsAreSorted(arr) {
			t.Errorf("arr is not sorted, %v", arr)
		}
	}
}

func TestHeapSort(t *testing.T) {
	for i := 0; i < 100; i++ {
		arr := algoW.RandInt(-100, 100, 500)
		sortW.HeapSort(arr, false)
		if !sort.IntsAreSorted(arr) {
			t.Errorf("arr is not sorted, %v", arr)
		}
	}
}

func TestQuickSort(t *testing.T) {
	for i := 0; i < 100; i++ {
		arr := algoW.RandInt(0, 500, 1000)
		sortW.QuickSort(arr)
		if !sort.IntsAreSorted(arr) {
			t.Errorf("arr is not sorted, %v", arr)
		}
	}
}

func TestRadixSort(t *testing.T) {
	for i := 0; i < 100; i++ {
		arr := algoW.RandInt(0, 500, 1000)
		sortW.RadixSort(arr)
		if !sort.IntsAreSorted(arr) {
			t.Errorf("arr is not sorted, %v", arr)
		}
	}
}

func TestCountSort(t *testing.T) {
	for i := 0; i < 100; i++ {
		arr := algoW.RandInt(-500, 500, 10000)
		sortW.CountSort(arr)
		if !sort.IntsAreSorted(arr) {
			t.Errorf("arr is not sorted, %v", arr)
		}
	}
}

func BenchmarkQuickSort(b *testing.B) {
	arr := algoW.RandInt(0, 10000, N)
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		rand.Shuffle(len(arr), func(i, j int) {
			arr[i], arr[j] = arr[j], arr[i]
		})
		b.StartTimer()
		sortW.QuickSort(arr)
	}
}

func BenchmarkQuickSortCmp(b *testing.B) {
	arr := algoW.RandInt(0, 10000, N)
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		rand.Shuffle(len(arr), func(i, j int) {
			arr[i], arr[j] = arr[j], arr[i]
		})
		b.StartTimer()
		sortW.Sort(arr, func(i, j int) int { return i - j })
	}
}

func BenchmarkShellSort(b *testing.B) {
	arr := algoW.RandFloat64(N)
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		rand.Shuffle(len(arr), func(i, j int) {
			arr[i], arr[j] = arr[j], arr[i]
		})
		b.StartTimer()
		sortW.ShellSort(arr)
	}
}

func BenchmarkSort(b *testing.B) {
	arr := algoW.RandInt(0, 10000, N)
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		rand.Shuffle(len(arr), func(i, j int) {
			arr[i], arr[j] = arr[j], arr[i]
		})
		b.StartTimer()
		sort.Ints(arr)
	}
}

func TestKth(t *testing.T) {
	arr := algoW.RandInt(0, 1000, 500)
	for i := 0; i < 100; i++ {
		k := algoW.RandInt(0, len(arr), 1)[0]
		val := algoW.Kth(arr, k, nil)
		sortW.Sort(arr, nil)
		if val != arr[k] {
			t.Fatal("wrong")
		}
	}
}

func BenchmarkKth(b *testing.B) {
	arr := algoW.RandInt(0, 10000, N)
	for i := 0; i < b.N; i++ {
		k := algoW.RandInt(0, len(arr), 1)[0]
		algoW.Kth(arr, k, func(i, j int) int { return i - j })
	}
}
