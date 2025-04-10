package algow

import "github.com/grewwc/go_tools/src/typew"

// NthElement return nth element (increasing order)
// Original slice will be changed
// kth: [0, len(nums) )
func Kth[T any](arr []T, kth int, cmp typew.CompareFunc[T]) T {
	if cmp == nil {
		cmp = typew.CreateDefaultCmp[T]()
	}
	lt, gt := ThreeWayPartitionCmp(arr, cmp)
	if lt <= kth && kth <= gt {
		return arr[lt]
	}
	if kth < lt {
		return Kth(arr[:lt], kth, cmp)
	}
	return Kth(arr[gt+1:], kth-gt-1, cmp)
}
