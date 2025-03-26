package algoW

import (
	"math/rand"
	"time"

	"golang.org/x/exp/constraints"
)

type number interface {
	constraints.Integer | constraints.Float
}

func Max[T constraints.Ordered](args ...T) T {
	if len(args) == 0 {
		return *new(T)
	}
	res := args[0]
	for _, val := range args[1:] {
		if val > res {
			res = val
		}
	}
	return res
}

func Min[T constraints.Ordered](args ...T) T {
	if len(args) == 0 {
		return *new(T)
	}
	res := args[0]
	for _, val := range args[1:] {
		if val < res {
			res = val
		}
	}
	return res
}

func RandFloat64(num int) []float64 {
	result := make([]float64, num)
	for i := 0; i < num; i++ {
		result[i] = rand.Float64()
	}
	return result
}

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

func Range[T number](start, end, step T) []T {
	if step == 0 {
		return nil
	}
	var res []T
	if step > 0 {
		if start > end {
			return nil
		}
		res = make([]T, 0, int((end-start)/step+1))
		for i := start; i < end; i += step {
			res = append(res, i)
		}
	} else {
		if start < end {
			return nil
		}
		res = make([]T, 0, int((start-end)/step+1))
		for i := start; i > end; i += step {
			res = append(res, i)
		}
	}
	return res
}
