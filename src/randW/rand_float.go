package randW

import (
	"math/rand"
	"time"
)

func RandFloat64(num int) []float64 {
	rand.Seed(time.Now().UnixNano())
	result := make([]float64, num)
	for i := 0; i < num; i++ {
		result[i] = rand.Float64()
	}
	return result
}
