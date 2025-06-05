package cw

import (
	"fmt"

	"github.com/grewwc/go_tools/src/typesw"
)

type Edge[T any] struct {
	v1, v2   T
	weight   float64
	directed bool
}

func newEdge[T any](v1, v2 T, weight float64, directed bool) *Edge[T] {
	return &Edge[T]{
		v1:     v1,
		v2:     v2,
		weight: weight,

		directed: directed,
	}
}

func (e *Edge[T]) V1() T {
	return e.v1
}

func (e *Edge[T]) V2() T {
	return e.v2
}

func (e *Edge[T]) Other(u T, cmp typesw.CompareFunc[T]) T {
	if cmp(e.v1, u) == 0 {
		return e.v2
	}
	if cmp(e.v2, u) == 0 {
		return e.v1
	}
	return e.V1()
}

func (e *Edge[T]) Weight() float64 {
	return e.weight
}

func (e *Edge[T]) String() string {
	symbol := "-"
	if e.directed {
		symbol = "->"
	}
	return fmt.Sprintf("(%v%s%v) %.3f", e.v1, symbol, e.v2, e.weight)
}

func (e *Edge[T]) cmp(other *Edge[T]) int {
	diff := e.weight - other.weight
	if diff < 0 {
		return -1
	}
	if diff > 0 {
		return 1
	}
	return 0
}
