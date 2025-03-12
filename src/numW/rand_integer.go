package numW

import (
	"math/rand"
	"time"
)

// RandInt
// lo: include, hi: exclude
func RandInt(lo, hi, N int) []int {
	rand.Seed(time.Now().UnixNano())
	result := make([]int, N)
	for i := 0; i < N; i++ {
		result[i] = rand.Intn(hi-lo) + lo
	}
	return result
}
