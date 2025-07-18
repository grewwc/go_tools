package cw

import (
	"github.com/grewwc/go_tools/src/typesw"
)

type Heap[T any] struct {
	data []T
	cmp  typesw.CompareFunc[T]
}

func newHeapCap[T any](cap int, cmp typesw.CompareFunc[T]) *Heap[T] {
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

func (h *Heap[T]) Top() T {
	return h.Next()
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

func swim[T any](arr []T, idx int, cmp typesw.CompareFunc[T]) {
	if idx <= 1 {
		return
	}

	for p := idx / 2; p >= 1 && cmp(arr[p], arr[idx]) > 0; {
		arr[p], arr[idx] = arr[idx], arr[p]
		idx = p
		p = idx / 2
	}
}

func getC[T any](arr []T, cmp typesw.CompareFunc[T], idx int) int {
	c := idx * 2
	if c+1 < len(arr) && cmp(arr[c], arr[c+1]) > 0 {
		c++
	}
	return c
}

func sink[T any](arr []T, idx int, cmp typesw.CompareFunc[T]) {
	for c := getC(arr, cmp, idx); c < len(arr) && cmp(arr[idx], arr[c]) > 0; {
		arr[c], arr[idx] = arr[idx], arr[c]
		idx = c
		c = getC(arr, cmp, idx)
	}
}

func NewHeap[T any](cmp typesw.CompareFunc[T]) typesw.IHeap[T] {
	if cmp == nil {
		cmp = typesw.CreateDefaultCmp[T]()
	}
	if cmp == nil {
		panic("compare function is nil")
	}
	return newHeapCap(8, cmp)
}
