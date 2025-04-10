package sortw

import (
	"math"
	"reflect"
	"sync"
	"unsafe"

	"github.com/grewwc/go_tools/src/algow"
	"github.com/grewwc/go_tools/src/cw"
	"github.com/grewwc/go_tools/src/typew"
	"golang.org/x/exp/constraints"
)

// InsertionSort ints
func InsertionSort[T constraints.Ordered](arr []T) {
	l := len(arr)
	if l <= 1 {
		return
	}
	if l == 2 {
		if arr[0] > arr[1] {
			arr[0], arr[1] = arr[1], arr[0]
		}
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

func InsertionSortComparator[T any](arr []T, comparator typew.CompareFunc[T]) {
	l := len(arr)
	if l <= 1 {
		return
	}
	for i := 0; i < l-1; i++ {
		for j := i + 1; j > 0; j-- {
			if comparator(arr[j], arr[j-1]) < 0 {
				arr[j], arr[j-1] = arr[j-1], arr[j]
			} else {
				break
			}
		}
	}

}

func mustToIntegerSlice[From constraints.Ordered, To constraints.Integer](from []From) []To {
	return *((*[]To)(unsafe.Pointer(&from)))
}

func minMax[T constraints.Ordered](arr []T) (T, T) {
	maxVal := arr[0]
	minVal := arr[0]
	for _, val := range arr[1:] {
		if val > maxVal {
			maxVal = val
		}
		if val < minVal {
			minVal = val
		}
	}
	return minVal, maxVal
}

// QuickSort uses multi cores
func QuickSort[T constraints.Ordered](arr []T) {
	if len(arr) <= 1 {
		return
	}
	name := reflect.TypeOf(*new(T)).Name()
	minVal, maxVal := minMax(arr)
	minValInt := *(*int)(unsafe.Pointer(&minVal))
	maxValInt := *(*int)(unsafe.Pointer(&maxVal))
	isint := name == "int" || name == "int8" || name == "int16" || name == "int32" || name == "int64" ||
		name == "uint" || name == "uint8" || name == "uint16" || name == "uint32" || name == "uint64"
	thresh := int(1e5) + 1
	if isint && maxValInt < thresh {
		countSortWithThreash(mustToIntegerSlice[T, int](arr), minValInt, maxValInt)
		return
	}
	quickSort(arr, true, nil)
}

func Sort[T any](arr []T, cmp typew.CompareFunc[T]) {
	if cmp == nil {
		cmp = typew.CreateDefaultCmp[T]()
	}
	if cmp == nil {
		cmp = typew.CreateDefaultCmp[T]()
	}
	quickSortComparator(arr, cmp, true, nil)
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

func HeapSort[T constraints.Ordered](arr []T, reverse bool) {
	var h cw.IHeap[T]
	cmp := typew.CreateDefaultCmp[T]()
	if reverse {
		cmp = func(i, j T) int {
			return -cmp(i, j)
		}
	}
	h = cw.NewHeap(cmp)
	for _, val := range arr {
		h.Insert(val)
	}

	i := 0
	for !h.IsEmpty() {
		arr[i] = h.Pop()
		i++
	}
}

func countSortWithThreash[T constraints.Integer](arr []T, min, max T) {
	if len(arr) <= 1 {
		return
	}
	count := make([]int64, max-min+1)
	for _, val := range arr {
		count[val-min]++
	}
	var index int64
	for i, val := range count {
		for j := 0; int64(j) < val; j++ {
			arr[index] = T(i) + T(min)
			index++
		}
	}
}

func CountSort[T constraints.Integer](arr []T) {
	if len(arr) <= 1 {
		return
	}
	minVal, maxVal := minMax(arr)
	countSortWithThreash(arr, minVal, maxVal)
}

func TopK[T constraints.Ordered](arr []T, k int, minK bool) []T {
	if k < 1 {
		return nil
	}
	var h cw.IHeap[T]
	cmp := typew.CreateDefaultCmp[T]()
	if minK {
		cmp = func(i, j T) int {
			return -cmp(i, j)
		}
	}
	h = cw.NewHeap[T](cmp)
	for i, val := range arr {
		if i < k {
			h.Insert(val)
			continue
		}
		next := h.Next()
		cmp := 0
		if next < val {
			cmp = -1
		} else if next == val {
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

func AreSorted[T constraints.Ordered](arr []T) bool {
	for i := 0; i < len(arr)-1; i++ {
		if arr[i] < arr[i+1] {
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

func calcSortedRatioComparator[T any](arr []T, comparator typew.CompareFunc[T]) float32 {
	if len(arr) <= 1 {
		return 1
	}
	cnt := 1
	for i := 0; i < len(arr)-1; i++ {
		if comparator(arr[i], arr[i+1]) < 0 {
			cnt++
		}
	}
	return float32(cnt) / float32(len(arr))
}

func quickSort[T constraints.Ordered](arr []T, calclRatio bool, wg *sync.WaitGroup) {
	if wg != nil {
		defer wg.Done()
	}
	if len(arr) < 48 {
		InsertionSort(arr)
		return
	}
	if calclRatio && calcSortedRatio(arr) >= 0.95 {
		InsertionSort(arr)
		return
	}

	lt, gt := algow.ThreeWayPartitionCmp(arr, func(i, j T) int {
		if i < j {
			return -1
		}
		if i == j {
			return 0
		}
		return 1
	})
	left, right := arr[:lt], arr[gt+1:]
	n := 4096
	if len(left) < n || len(right) < n {
		quickSort(left, false, nil)
		quickSort(right, false, nil)
		return
	}
	var wg1 sync.WaitGroup
	wg1.Add(2)
	go quickSort(left, false, &wg1)
	go quickSort(right, false, &wg1)
	wg1.Wait()
}

func quickSortComparator[T any](arr []T, comparator typew.CompareFunc[T], calcRatio bool, wg *sync.WaitGroup) {
	if wg != nil {
		defer wg.Done()
	}
	if len(arr) < 32 {
		InsertionSortComparator(arr, comparator)
		return
	}
	if calcRatio && calcSortedRatioComparator(arr, comparator) >= 0.95 {
		InsertionSortComparator(arr, comparator)
		return
	}

	lt, gt := algow.ThreeWayPartitionCmp(arr, comparator)
	left, right := arr[:lt], arr[gt+1:]
	n := 4096
	if len(left) < n || len(right) < n {
		quickSortComparator(left, comparator, false, nil)
		quickSortComparator(right, comparator, false, nil)
		return
	}
	wg1 := sync.WaitGroup{}
	wg1.Add(2)
	go quickSortComparator(left, comparator, false, &wg1)
	go quickSortComparator(right, comparator, false, &wg1)
	wg1.Wait()
}
