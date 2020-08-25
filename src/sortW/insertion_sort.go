package sortW

func InsertionSortInts(nums []int) {
	l := len(nums)
	for i := 1; i < l; i++ {
		for j := i; j > 0; j-- {
			if nums[j] < nums[j-1] {
				nums[j], nums[j-1] = nums[j-1], nums[j]
			}
		}
	}
}
