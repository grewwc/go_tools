package sortw

import (
	"golang.org/x/exp/constraints"

	"github.com/grewwc/go_tools/src/typesw"
)

func calcMinRun(n int) int {
	r := 0
	for n >= 64 {
		r |= n & 1
		n >>= 1
	}
	return n + r
}

func TimeSortCmp[T any](arr []T, cmp typesw.CompareFunc[T]) {
	if cmp == nil {
		cmp = typesw.CreateDefaultCmp[T]()
	}
	minRun := calcMinRun(len(arr))
	if len(arr) <= minRun {
		InsertionSortComparator(arr, cmp)
		return
	}
	for start := 0; start < len(arr); start += minRun {
		end := start + minRun
		if end > len(arr) {
			end = len(arr)
		}
		InsertionSortComparator(arr[start:end], cmp)
	}
	size := minRun
	for size < len(arr) {
		for start := 0; start < len(arr); start += size * 2 {
			mid := start + size
			end := start + size*2
			if mid >= len(arr) {
				break
			}
			if end > len(arr) {
				end = len(arr)
			}
			merge(arr, start, mid, end, cmp)
		}
		size *= 2
	}
}

func TimeSort[T constraints.Ordered](arr []T) {
	minRun := calcMinRun(len(arr))
	if len(arr) <= minRun {
		InsertionSort(arr)
		return
	}
	for start := 0; start < len(arr); start += minRun {
		end := start + minRun
		if end > len(arr) {
			end = len(arr)
		}
		InsertionSort(arr[start:end])
	}
	cmp := typesw.CreateDefaultCmp[T]()
	size := minRun
	merge := func(arr []T, start, mid, end int, cmp typesw.CompareFunc[T]) {
		left := arr[start:mid]
		right := arr[mid:end]
		i, j := 0, 0
		for k := start; k < end; k++ {
			if i < len(left) && (j >= len(right) || cmp(left[i], right[j]) <= 0) {
				arr[k] = left[i]
				i++
			} else {
				arr[k] = right[j]
				j++
			}
		}
	}
	for size < len(arr) {
		for start := 0; start < len(arr); start += size * 2 {
			mid := start + size
			end := start + size*2
			if mid >= len(arr) {
				break
			}
			if end > len(arr) {
				end = len(arr)
			}
			merge(arr, start, mid, end, cmp)
		}
		size *= 2
	}
}
