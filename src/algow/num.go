package algow

import (
	"math/rand"
	"time"

	"golang.org/x/exp/constraints"
)

type Number interface {
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
	var res T
	if len(args) == 0 {
		return res
	}
	res = args[0]
	for _, val := range args[1:] {
		if val < res {
			res = val
		}
	}
	return res
}

func Sum[T Number](nums ...T) T {
	var res T
	for _, num := range nums {
		res += num
	}
	return res
}

func Accumulate[T any](arr []T, f func(t1, t2 T) T) T {
	var res T
	if len(arr) == 0 {
		return res
	}
	for _, e := range arr {
		res = f(res, e)
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

func Range[T Number](start, end, step T) []T {
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

func Abs[T Number](val T) T {
	if val < 0 {
		return -val
	}
	return val
}

func Combinations[T any](arr []T, k int) [][]T {
	var result [][]T
	if k == 0 {
		return result
	}
	curr := make([]T, 0)
	var backtrack func(int)
	backtrack = func(start int) {
		if len(curr) == k {
			cp := make([]T, len(curr))
			copy(cp, curr)
			result = append(result, cp)
			return
		}
		for i := start; i < len(arr); i++ {
			curr = append(curr, arr[i])
			backtrack(i + 1)
			curr = curr[:len(curr)-1]
		}
	}
	backtrack(0)
	return result
}

func Permutatation[T any](arr []T, k int) [][]T {
	result := make([][]T, 0)
	if k < 1 {
		return result
	}
	curr := make([]T, 0)
	used := make([]bool, len(arr))

	var permutation func()
	permutation = func() {
		if k == len(curr) {
			cp := make([]T, len(arr))
			copy(cp, arr)
			result = append(result, cp)
			return
		}

		for i := 0; i < len(arr); i++ {
			if used[i] {
				continue
			}
			used[i] = true
			curr = append(curr, arr[i])
			permutation()
			used[i] = false
			curr = curr[:len(curr)-1]
		}
	}
	permutation()
	return result
}
