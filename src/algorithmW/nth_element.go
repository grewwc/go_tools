package algorithmW

import "github.com/grewwc/go_tools/src/typesW"

// NthElement return nth element (increasing order)
// Original slice will be changed
// kth: [0, len(nums) )
func Kth[T any](arr []T, kth int, cmp typesW.CompareFunc[T]) T {
	if cmp == nil {
		cmp = typesW.CreateDefaultCmp[T]()
	}
	lt, gt := ThreeWayPartitionComparator(arr, cmp)
	if lt <= kth && kth <= gt {
		return arr[lt]
	}
	if kth < lt {
		return Kth(arr[:lt], kth, cmp)
	}
	return Kth(arr[gt+1:], kth-gt-1, cmp)
}
