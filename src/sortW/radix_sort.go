package sortW

import "math"

func getNumDigits(num int) int {
	if num == 0 {
		return 1
	}
	resultVal := 0
	for num > 0 {
		num /= 10
		resultVal++
	}
	return resultVal
}

func countSort(nums []int, exp int) {
	old := nums
	count := make([]int, 10)
	n := len(old)
	output := make([]int, n)
	for _, num := range old {
		digit := (num / exp) % 10
		count[digit]++
	}

	for i := 1; i < 10; i++ {
		count[i] += count[i-1]
	}

	// 必须是逆序的
	// 不然就不是稳定的了
	for i := n - 1; i >= 0; i-- {
		digit := (old[i] / exp) % 10
		output[count[digit]-1] = old[i]
		count[digit]--
	}

	copy(old, output)
}

// RadixSort int slice version
func RadixSort(nums []int) {
	getMaxVal := func(nums []int) int {
		resultVal := math.MinInt64
		for _, num := range nums {
			if resultVal < num {
				resultVal = num
			}
		}
		return resultVal
	}

	maxDigit := getNumDigits(getMaxVal(nums))
	exp := 1
	for i := 0; i < maxDigit; i++ {
		countSort(nums, exp)
		exp *= 10
	}
}
