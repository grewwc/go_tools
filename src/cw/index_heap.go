package cw

import (
	"github.com/grewwc/go_tools/src/typesw"
)

type IndexHeap[T any] struct {
	h typesw.IHeap[Tuple]
}

func NewIndexHeap[T any](cmp typesw.CompareFunc[T]) *IndexHeap[T] {
	if cmp == nil {
		cmp = typesw.CreateDefaultCmp[T]()
	}
	newCmp := func(a, b Tuple) int {
		return cmp(a.Get(1).(T), b.Get(1).(T))
	}
	h := NewHeap(newCmp)
	return &IndexHeap[T]{
		h: h,
	}
}

func (h *IndexHeap[T]) Insert(index int, val T) {
	if index < 0 {
		return
	}
	h.h.Insert(*NewTuple(index, val))
}

func (h *IndexHeap[T]) Size() int {
	return h.h.Size()
}

func (h *IndexHeap[T]) IsEmpty() bool {
	return h.h.IsEmpty()
}

func (h *IndexHeap[T]) Pop() int {
	if h.IsEmpty() {
		return -1
	}
	t := h.h.Pop()
	return t.Get(0).(int)
}

func (h *IndexHeap[T]) Top() int {
	if h.IsEmpty() {
		return -1
	}
	t := h.h.Top()
	return t.Get(0).(int)
}
