package algoriwhtmW

import (
	"testing"

	"github.com/grewwc/go_tools/src/algoW"
	"github.com/grewwc/go_tools/src/numW"
)

// Mocking the random number generator for deterministic testing
type MockRand struct {
	next int
}

func (m *MockRand) RandInt(lo, hi, n int) []int {
	result := make([]int, n)
	for i := 0; i < n; i++ {
		result[i] = m.next
		m.next++
		if m.next >= hi {
			m.next = lo
		}
	}
	return result
}

func allSmall(nums []int, val int) bool {
	for _, num := range nums {
		if num >= val {
			return false
		}
	}
	return true
}

func allEqualLarge(nums []int, val int) bool {
	for _, num := range nums {
		if num < val {
			return false
		}
	}
	return true
}

func TestPartition(t *testing.T) {
	for i := 0; i < 100; i++ {
		nums := numW.RandInt(0, 10, 500)
		p := algoW.Partition(nums, 0, len(nums))
		val := nums[p]
		// all values before p should smaller than val
		if !allSmall(nums[:p], val) {
			t.Errorf("partition failed")
		}
		if !allEqualLarge(nums[p+1:], val) {
			t.Errorf("partition failed")
		}
	}
	numW.RandInt(0, 100, 100)
}
