package sortW

import (
	"go_tools/src/algorithmW"
)

func QuicksortInts(nums []int) {
	l := len(nums)
	if l <= 1 {
		return
	}

	if l <= 8 {
		InsertionSortInts(nums)
		return
	}
	lt, gt := algorithmW.ThreeWayPartitionInts(nums)
	QuicksortInts(nums[:lt])
	QuicksortInts(nums[gt+1:])
}
