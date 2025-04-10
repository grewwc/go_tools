package algow

import (
	"github.com/grewwc/go_tools/src/typew"
)

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

func BisectLeft[T any](arr []T, target T, cmp typew.CompareFunc[T]) int {
	if len(arr) == 0 {
		return -1
	}
	if cmp == nil {
		cmp = typew.CreateDefaultCmp[T]()
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

func BisectRight[T any](arr []T, target T, cmp typew.CompareFunc[T]) int {
	if len(arr) == 0 {
		return -1
	}
	if cmp == nil {
		cmp = typew.CreateDefaultCmp[T]()
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

func LongestIncreasingSubsequence[T any](arr []T, cmp typew.CompareFunc[T]) []T {
	if len(arr) == 0 {
		return []T{}
	}
	if cmp == nil {
		cmp = typew.CreateDefaultCmp[T]()
	}
	sub := make([]T, 0)
	for _, num := range arr {
		idx := BisectLeft(sub, num, cmp)
		// fmt.Print(sub, "-->")
		if idx < 0 || idx >= len(sub) {
			sub = append(sub, num)
		} else {
			sub[idx] = num
		}
		// fmt.Println(sub)
	}
	return sub
}

func EditDistance[T any](a1, a2 []T, cmp typew.CompareFunc[T]) int {
	if cmp == nil {
		cmp = typew.CreateDefaultCmp[T]()
	}
	m, n := len(a1), len(a2)
	if m < n {
		return EditDistance(a2, a1, cmp)
	}

	prev := make([]int, n+1)
	for j := 0; j <= n; j++ {
		prev[j] = j
	}
	for i := 1; i <= m; i++ {
		curr := make([]int, n+1)
		curr[0] = i
		for j := 1; j <= n; j++ {
			cost := 0
			if cmp(a1[i-1], a2[j-1]) != 0 {
				cost++
			}
			replace := prev[j-1] + cost
			insert := curr[j-1] + 1
			remove := prev[j] + 1
			curr[j] = Min(insert, remove, replace)
		}
		prev = curr
	}
	return prev[n]
}
