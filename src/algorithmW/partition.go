package algorithmW

import (
	"math/rand"

	"github.com/grewwc/go_tools/src/randW"
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

// Partition partition array into two parts [lo, hi)
func Partition[T constraints.Ordered](nums []T, lo, hi int) int {
	if hi-lo <= 1 {
		return lo
	}
	start := lo - 1
	r := randW.RandInt(lo, hi, 1)[0]
	nums[r], nums[hi-1] = nums[hi-1], nums[r]
	pivot := nums[hi-1]
	for i := lo; i < hi-1; i++ {
		if nums[i] < pivot {
			start++
			nums[start], nums[i] = nums[i], nums[start]
		}
	}
	start++
	nums[start], nums[hi-1] = nums[hi-1], nums[start]
	return start
}
