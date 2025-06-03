package cw

import (
	"github.com/grewwc/go_tools/src/typesw"
)

type IndexHeap[Index, T any] struct {
	h typesw.IHeap[*Tuple]
	m *Map[Index, *Tuple]
}

func NewIndexHeap[Index, T any](cmp typesw.CompareFunc[T]) *IndexHeap[Index, T] {
	if cmp == nil {
		cmp = typesw.CreateDefaultCmp[T]()
	}
	newCmp := func(a, b *Tuple) int {
		return cmp(a.Get(1).(T), b.Get(1).(T))
	}
	h := NewHeap(newCmp)
	return &IndexHeap[Index, T]{
		h: h,
		m: NewMap[Index, *Tuple](),
	}
}

func (h *IndexHeap[Index, T]) Insert(index Index, val T) {
	if h.m.Contains(index) {
		return
	}
	tup := NewTuple(index, val)
	h.m.PutIfAbsent(index, tup)
	h.h.Insert(tup)
}

func (h *IndexHeap[Key, T]) Size() int {
	return h.h.Size()
}

func (h *IndexHeap[Key, T]) IsEmpty() bool {
	return h.h.IsEmpty()
}

func (h *IndexHeap[Index, T]) Pop() T {
	if h.IsEmpty() {
		return *new(T)
	}
	t := h.h.Pop()
	h.m.Delete(t.Get(0).(Index))
	return t.Get(1).(T)
}

func (h *IndexHeap[Index, T]) PopIndex() Index {
	if h.IsEmpty() {
		return *new(Index)
	}
	t := h.h.Pop()
	h.m.Delete(t.Get(0).(Index))
	return t.Get(0).(Index)
}

func (h *IndexHeap[Index, T]) Top() T {
	if h.IsEmpty() {
		return *new(T)
	}
	t := h.h.Top()
	return t.Get(1).(T)
}

func (h *IndexHeap[Index, T]) TopIndex() Index {
	if h.IsEmpty() {
		return *new(Index)
	}
	t := h.h.Top()
	return t.Get(0).(Index)
}

func (h *IndexHeap[Index, T]) Contains(index Index) bool {
	if h.IsEmpty() {
		return false
	}
	return h.m.Contains(index)
}

func (h *IndexHeap[Index, T]) Update(index Index, val T) {
	tup := h.m.GetOrDefault(index, nil)
	if tup == nil {
		h.Insert(index, val)
	}
	tup.Set(1, val)
}
