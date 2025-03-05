package algoriwhtmW

import (
	"math/rand/v2"
	"sort"
	"testing"

	"github.com/grewwc/go_tools/src/randW"
	"github.com/grewwc/go_tools/src/sortW"
	"github.com/grewwc/go_tools/src/typesW"
)

const (
	N = 500000
)

func TestShellSort(t *testing.T) {
	for i := 0; i < 100; i++ {
		arr := randW.RandInt(0, 100, 500)
		sortW.ShellSort(arr)
		if !sort.IntsAreSorted(arr) {
			t.Errorf("arr is not sorted, %v", arr)
		}
	}
}

func TestHeapSort(t *testing.T) {
	for i := 0; i < 100; i++ {
		arr := randW.RandInt(-100, 100, 500)
		sortW.HeapSort(arr, false)
		if !sort.IntsAreSorted(arr) {
			t.Errorf("arr is not sorted, %v", arr)
		}
	}
}

func TestQuickSort(t *testing.T) {
	for i := 0; i < 100; i++ {
		arr := randW.RandInt(0, 500, 1000)
		sortW.QuickSort(arr)
		if !sort.IntsAreSorted(arr) {
			t.Errorf("arr is not sorted, %v", arr)
		}
	}
}

func TestRadixSort(t *testing.T) {
	for i := 0; i < 100; i++ {
		arr := randW.RandInt(0, 500, 1000)
		sortW.RadixSort(arr)
		if !sort.IntsAreSorted(arr) {
			t.Errorf("arr is not sorted, %v", arr)
		}
	}
}

func TestCountSort(t *testing.T) {
	for i := 0; i < 100; i++ {
		arr := randW.RandInt(-500, 500, 10000)
		sortW.CountSort(arr)
		if !sort.IntsAreSorted(arr) {
			t.Errorf("arr is not sorted, %v", arr)
		}
	}
}

func TestShellSortComparable(t *testing.T) {
	for i := 0; i < 100; i++ {
		intsArr := randW.RandInt(0, 100, 500)
		arr := make([]typesW.Comparable, 0, len(intsArr))
		for _, v := range intsArr {
			arr = append(arr, typesW.IntComparable(v))
		}
		sortW.ShellSortComparable(arr)
		if !sortW.AreSorted(arr) {
			t.Errorf("arr is not sorted, %v", arr)
		}
	}
}

func TestQuickSortComparable(t *testing.T) {
	for i := 0; i < 100; i++ {
		intsArr := randW.RandInt(0, 100, 500)
		arr := make([]typesW.Comparable, 0, len(intsArr))
		for _, v := range intsArr {
			arr = append(arr, typesW.IntComparable(v))
		}
		sortW.ShellSortComparable(arr)
		if !sortW.AreSorted(arr) {
			t.Errorf("arr is not sorted, %v", arr)
		}
	}
}

func BenchmarkQuickSort(b *testing.B) {
	arr := randW.RandInt(0, 10000, N)
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		rand.Shuffle(len(arr), func(i, j int) {
			arr[i], arr[j] = arr[j], arr[i]
		})
		b.StartTimer()
		sortW.QuickSort(arr)
	}
}

func BenchmarkShellSort(b *testing.B) {
	arr := randW.RandFloat64(N)
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
	arr := randW.RandInt(0, 10000, N)
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		rand.Shuffle(len(arr), func(i, j int) {
			arr[i], arr[j] = arr[j], arr[i]
		})
		b.StartTimer()
		sort.Ints(arr)
	}
}
