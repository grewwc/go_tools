package algoriwhtmW

import (
	"math/rand/v2"
	"sort"
	"testing"

	"github.com/grewwc/go_tools/src/algorithmW"
	"github.com/grewwc/go_tools/src/containerW/typesW"
	"github.com/grewwc/go_tools/src/randW"
)

const (
	N = 50000
)

func TestShellSort(t *testing.T) {
	for i := 0; i < 100; i++ {
		arr := randW.RandInt(0, 100, 500)
		algorithmW.ShellSort(arr)
		if !sort.IntsAreSorted(arr) {
			t.Errorf("arr is not sorted, %v", arr)
		}
	}

}

func TestQuickSort(t *testing.T) {
	for i := 0; i < 100; i++ {
		arr := randW.RandInt(0, 100, 500)
		algorithmW.QuickSort(arr)
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
		algorithmW.ShellSortComparable(arr)
		if !algorithmW.AreSorted(arr) {
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
		algorithmW.ShellSortComparable(arr)
		if !algorithmW.AreSorted(arr) {
			t.Errorf("arr is not sorted, %v", arr)
		}
	}
}

func BenchmarkQuickSort(b *testing.B) {
	arr := randW.RandFloat64(N)
	sort.Float64s(arr)
	for i := 0; i < b.N; i++ {
		rand.Shuffle(len(arr), func(i, j int) {
			arr[i], arr[j] = arr[j], arr[i]
		})
		algorithmW.QuickSort(arr)
	}
}

func BenchmarkShellSort(b *testing.B) {
	arr := randW.RandFloat64(N)
	for i := 0; i < b.N; i++ {
		rand.Shuffle(len(arr), func(i, j int) {
			arr[i], arr[j] = arr[j], arr[i]
		})
		algorithmW.ShellSort(arr)
	}
}

func BenchmarkSort(b *testing.B) {
	arr := randW.RandFloat64(N)
	sort.Float64s(arr)
	for i := 0; i < b.N; i++ {
		rand.Shuffle(len(arr), func(i, j int) {
			arr[i], arr[j] = arr[j], arr[i]
		})
		sort.Float64s(arr)
	}
}
