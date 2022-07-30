package algorithmW

import (
	"github.com/grewwc/go_tools/src/containerW"
	"github.com/grewwc/go_tools/src/containerW/typesW"
)

// InsertionSort ints
func InsertionSort(arr []int) {
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
func QuickSort(arr []int) {
	if len(arr) < 8 {
		InsertionSort(arr)
		return
	}
	lt, gt := ThreeWayPartitionInts(arr)
	QuickSort(arr[:lt])
	QuickSort(arr[gt+1:])
}

// ShellSort ints
func ShellSort(arr []int) {
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

func HeapSort(arr []int, reverse bool) {
	var h containerW.Heap
	if reverse {
		h = containerW.NewMaxHeapCap(len(arr))
	} else {
		h = containerW.NewMinHeapCap(len(arr))
	}
	for _, val := range arr {
		h.Insert(typesW.IntComparable(val))
	}

	i := 0
	for !h.IsEmpty() {
		arr[i] = int(h.Pop().(typesW.IntComparable))
		i++
	}
}

func TopK(arr []interface{}, k int, minK bool) []interface{} {
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
		cmp := next.(typesW.Comparable).Compare(val.(typesW.Comparable))
		if !minK {
			cmp *= -1
		}
		if cmp > 0 {
			h.Pop()
			h.Insert(val)
		}
	}
	return h.ToList()
}
