package numW

import "golang.org/x/exp/constraints"

type number interface {
	constraints.Integer | constraints.Float
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
