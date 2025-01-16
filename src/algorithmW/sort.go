package algorithmW

import (
	"math"

	"github.com/grewwc/go_tools/src/containerW"
	"github.com/grewwc/go_tools/src/containerW/typesW"
	"golang.org/x/exp/constraints"
)

// InsertionSort ints
func InsertionSort[T constraints.Ordered](arr []T) {
	l := len(arr)
	if l <= 1 {
		return
	}
	for i := 0; i < l-1; i++ {
		for j := i + 1; j > 0; j-- {
			if arr[j] < arr[j-1] {
				arr[j], arr[j-1] = arr[j-1], arr[j]
			} else {
				break
			}
		}
	}
}

func InsertionSortComparable[T typesW.Comparable](arr []T) {
	l := len(arr)
	if l <= 1 {
		return
	}
	for i := 0; i < l-1; i++ {
		for j := i + 1; j > 0; j-- {
			if arr[j].Compare(arr[j-1]) < 0 {
				arr[j], arr[j-1] = arr[j-1], arr[j]
			} else {
				break
			}
		}
	}

}

// QuickSort ints
func QuickSort[T constraints.Ordered](arr []T) {
	quickSort(arr, true)
}

func QuickSortComparable[T typesW.Comparable](arr []T) {
	if len(arr) < 8 {
		InsertionSortComparable(arr)
		return
	}
	lt, gt := ThreeWayPartitionComparable(arr)
	QuickSortComparable(arr[:lt])
	QuickSortComparable(arr[gt+1:])
}

// ShellSort ints
func ShellSort[T constraints.Ordered](arr []T) {
	h := 1
	l := len(arr)
	if l <= 1 {
		return
	}
	r := int(math.Max(math.Log10(float64(l)), 3))
	for r*h < l {
		h = r*h + 1
	}
	for h >= 1 {
		for i := 0; i < l-h; i += h {
			for j := i + h; j >= h; j -= h {
				if arr[j] < arr[j-h] {
					arr[j], arr[j-h] = arr[j-h], arr[j]
				} else {
					break
				}
			}
		}
		h /= r
	}
}

func ShellSortComparable[T typesW.Comparable](arr []T) {
	h := 1
	l := len(arr)
	if l <= 1 {
		return
	}
	for 3*h < l {
		h = 3*h + 1
	}
	for h >= 1 {
		for i := 0; i < l-1; i += h {
			for j := i + 1; j >= h; j -= h {
				if arr[j].Compare(arr[j-h]) < 0 {
					arr[j], arr[j-h] = arr[j-h], arr[j]
				} else {
					break
				}
			}
		}
		h /= 3
	}
}

func HeapSort[T constraints.Ordered](arr []T, reverse bool) {
	var h containerW.Heap
	if reverse {
		h = containerW.NewMaxHeapCap(len(arr))
	} else {
		h = containerW.NewMinHeapCap(len(arr))
	}
	for _, val := range arr {
		h.Insert(val)
	}

	i := 0
	for !h.IsEmpty() {
		arr[i] = h.Pop().(T)
		i++
	}
}

func TopK[T constraints.Ordered](arr []T, k int, minK bool) []T {
	if k < 1 {
		return nil
	}
	var h containerW.Heap
	if minK {
		h = containerW.NewMaxHeapCap(k)
	} else {
		h = containerW.NewMinHeapCap(k)
	}
	for i, val := range arr {
		if i < k {
			h.Insert(val)
			continue
		}
		next := h.Next()
		cmp := 0
		if next.(T) < val {
			cmp = -1
		} else if next.(T) == val {
			cmp = 0
		} else {
			cmp = -1
		}
		if !minK {
			cmp *= -1
		}
		if cmp > 0 {
			h.Pop()
			h.Insert(val)
		}
	}
	interfaceList := h.ToList()
	result := make([]T, len(interfaceList))

	return result
}

func AreSorted[T typesW.Comparable](arr []T) bool {
	for i := 0; i < len(arr)-1; i++ {
		if arr[i].Compare(arr[i+1]) > 0 {
			return false
		}
	}
	return true
}

func calcSortedRatio[T constraints.Ordered](arr []T) float32 {
	if len(arr) <= 1 {
		return 1
	}
	cnt := 1
	for i := 0; i < len(arr)-1; i++ {
		if arr[i] <= arr[i+1] {
			cnt++
		}
	}
	return float32(cnt) / float32(len(arr))
}

func calcSortedRatioComparable[T typesW.Comparable](arr []T) float32 {
	if len(arr) <= 1 {
		return 1
	}
	cnt := 1
	for i := 0; i < len(arr)-1; i++ {
		if arr[i].Compare(arr[i+1]) <= 0 {
			cnt++
		}
	}
	return float32(cnt) / float32(len(arr))
}

func quickSort[T constraints.Ordered](arr []T, calclRatio bool) {
	if len(arr) < 32 {
		InsertionSort(arr)
		return
	}
	if calclRatio && calcSortedRatio(arr) >= 0.95 {
		InsertionSort(arr)
		return
	}

	lt, gt := ThreeWayPartitionInts(arr)
	quickSort(arr[:lt], false)
	quickSort(arr[gt+1:], false)
}

func quickSortComparable[T typesW.Comparable](arr []T, calcRatio bool) {
	if len(arr) < 32 {
		InsertionSortComparable(arr)
		return
	}
	if calcRatio && calcSortedRatioComparable(arr) >= 0.95 {
		InsertionSortComparable(arr)
		return
	}

	lt, gt := ThreeWayPartitionComparable(arr)
	quickSortComparable(arr[:lt], false)
	quickSortComparable(arr[gt+1:], false)
}
