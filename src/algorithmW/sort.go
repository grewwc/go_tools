package algorithmW

import (
	"github.com/grewwc/go_tools/src/containerW"
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

// QuickSort ints
func QuickSort[T constraints.Ordered](arr []T) {
	if len(arr) < 8 {
		InsertionSort(arr)
		return
	}
	lt, gt := ThreeWayPartitionInts(arr)
	QuickSort(arr[:lt])
	QuickSort(arr[gt+1:])
}

// ShellSort ints
func ShellSort[T constraints.Ordered](arr []T) {
	h := 1
	l := len(arr)
	if l <= 1 {
		return
	}
	for 3*h < l {
		h = 3*h + 1
	}
	for h >= 1 {
		for i := h; i < l-1; i++ {
			for j := i + 1; j >= h; j -= h {
				if arr[j] < arr[j-h] {
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
