package algoW

import "github.com/grewwc/go_tools/src/typesW"

func Fill[T any](arr []T, value T) {
	for i := 0; i < len(arr); i++ {
		arr[i] = value
	}
}

func Reverse[T any](arr []T) {
	for i := 0; i < len(arr)/2; i++ {
		arr[i], arr[len(arr)-i-1] = arr[len(arr)-i-1], arr[i]
	}
}

func BisectLeft[T any](arr []T, target T, cmp typesW.CompareFunc[T]) int {
	if len(arr) == 0 {
		return -1
	}
	if cmp == nil {
		cmp = typesW.CreateDefaultCmp[T]()
	}
	lo, hi := 0, len(arr)
	for lo < hi {
		mid := lo + (hi-lo)/2
		if cmp(target, arr[mid]) > 0 {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	return lo

}

func BisectRight[T any](arr []T, target T, cmp typesW.CompareFunc[T]) int {
	if len(arr) == 0 {
		return -1
	}
	if cmp == nil {
		cmp = typesW.CreateDefaultCmp[T]()
	}
	lo, hi := 0, len(arr)
	for lo < hi {
		mid := lo + (hi-lo)/2
		if cmp(target, arr[mid]) < 0 {
			hi = mid
		} else {
			lo = mid + 1
		}
	}
	return lo
}

func LongestIncreasingSubsequence[T any](arr []T, cmp typesW.CompareFunc[T]) []T {
	if len(arr) == 0 {
		return []T{}
	}
	if cmp == nil {
		cmp = typesW.CreateDefaultCmp[T]()
	}
	sub := make([]T, 0)
	for _, num := range arr {
		idx := BisectLeft(sub, num, cmp)
		if idx < 0 || idx >= len(sub) {
			sub = append(sub, num)
		} else {
			sub[idx] = num
		}
	}
	return sub
}
