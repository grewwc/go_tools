package algorithmW

import (
	"math/rand"

	"golang.org/x/exp/constraints"
)

// ThreeWayPartition return range of pivot value (both inclusive)
func ThreeWayPartitionInts[T constraints.Ordered](nums []T) (int, int) {
	rand.Shuffle(len(nums), func(i, j int) {
		nums[i], nums[j] = nums[j], nums[i]
	})
	pivot := nums[0]
	lt, gt := 0, len(nums)-1
	i := 1
	for i <= gt {
		cur := nums[i]
		if cur < pivot {
			// swap(nums, i, lt)
			nums[i], nums[lt] = nums[lt], nums[i] // for faster speed
			i++
			lt++
		} else if cur == pivot {
			i++
		} else {
			// swap(nums, i, gt)
			nums[i], nums[gt] = nums[gt], nums[i] // for faster speed
			gt--
		}
	}
	return lt, gt
}
