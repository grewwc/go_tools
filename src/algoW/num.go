package algoW

import "golang.org/x/exp/constraints"

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
