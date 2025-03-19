package containerW

import (
	"github.com/grewwc/go_tools/src/typesW"
	"golang.org/x/exp/constraints"
)

type Heap[T any] struct {
	data []T
	cmp  typesW.CompareFunc
}

type IHeap[T any] interface {
	Insert(T)
	Pop() T
	Size() int
	IsEmpty() bool
	ToList() []T
	Next() T
}

func newHeapCap[T any](cap int, cmp typesW.CompareFunc) *Heap[T] {
	return &Heap[T]{
		data: make([]T, 1, cap+1),
		cmp:  cmp,
	}
}

func (h *Heap[T]) Size() int {
	return len(h.data) - 1
}

func (h *Heap[T]) Insert(val T) {
	h.data = append(h.data, val)
	swim(h.data, len(h.data)-1, h.cmp)
}

func (h *Heap[T]) Next() T {
	if len(h.data) == 1 {
		return *new(T)
	}
	return h.data[1]
}

func (h *Heap[T]) ToList() []T {
	res := make([]T, h.Size()-1)
	copy(res, h.data[1:])
	return res
}

func (h *Heap[T]) Pop() T {
	if h.Size() <= 0 {
		return *new(T)
	}
	res := h.Next()
	// swap
	h.data[1], h.data[len(h.data)-1] = h.data[len(h.data)-1], h.data[1]
	h.data = h.data[:len(h.data)-1]
	sink(h.data, 1, h.cmp)
	return res
}

func (h *Heap[T]) IsEmpty() bool {
	return len(h.data) == 1
}

func swim[T any](arr []T, idx int, cmp typesW.CompareFunc) {
	if idx <= 1 {
		return
	}

	if cmp(arr[idx/2], arr[idx]) > 0 {
		arr[idx/2], arr[idx] = arr[idx], arr[idx/2]
		swim(arr, idx/2, cmp)
		sink(arr, idx, cmp)
	}
}

func sink[T any](arr []T, idx int, cmp typesW.CompareFunc) {
	childIdx := idx * 2
	if childIdx >= len(arr) {
		return
	}
	if childIdx+1 < len(arr) && cmp(arr[childIdx], arr[childIdx+1]) > 0 {
		childIdx++
	}
	if cmp(arr[idx], arr[childIdx]) > 0 {
		arr[childIdx], arr[idx] = arr[idx], arr[childIdx]
		sink(arr, childIdx, cmp)
	}
}

func NewHeap[T constraints.Ordered](cmp typesW.CompareFunc) IHeap[T] {
	if cmp == nil {
		cmp = typesW.CreateDefaultCmp[T]()
	}
	return newHeapCap[T](8, cmp)
}
