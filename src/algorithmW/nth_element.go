package algorithmW

// NthElement return nth element (increasing order)
// Original slice will be changed
func NthElementInts(nums []int, nth int) int {
	lt, gt := ThreeWayPartitionInts(nums)
	// transfer nth to index
	nth--
	if lt <= nth && nth <= gt {
		return nums[lt]
	}
	if nth < lt {
		return NthElementInts(nums[:lt], nth+1)
	}
	return NthElementInts(nums[gt+1:], nth-gt)
}
