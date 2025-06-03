package cw

import (
	"github.com/grewwc/go_tools/src/typesw"
)

type edge[T any] struct {
	u, v   T
	weight float32
}

func newEdge[T any](u, v T) *edge[T] {
	return &edge[T]{
		u: u,
		v: v,
	}
}

func (e *edge[T]) either() T {
	return e.u
}

func (e *edge[T]) other(u T, cmp typesw.CompareFunc[T]) T {
	if cmp(e.u, u) == 0 {
		return e.v
	}
	if cmp(e.v, u) == 0 {
		return u
	}
	return e.either()
}

func (e *edge[T]) cmp(other *edge[T]) int {
	diff := e.weight - other.weight
	if diff < 0 {
		return -1
	}
	if diff > 0 {
		return 1
	}
	return 0
}
