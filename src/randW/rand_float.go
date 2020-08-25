package randW

import "math/rand"

func RandFloat64(num int) []float64 {
	result := make([]float64, num)
	for i := 0; i < num; i++ {
		result[i] = rand.Float64()
	}
	return result
}
