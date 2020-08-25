package randW

import "math/rand"

func RandInt(lo, hi, N int) []int {
	result := make([]int, N)
	for i := 0; i < N; i++ {
		result[i] = rand.Intn(hi-lo) + lo
	}
	return result	
}
