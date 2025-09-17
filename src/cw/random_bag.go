package cw

import "math/rand"

type RandomBag[T any] struct {
	m    *Map[T, int]
	data []T
}

func NewRandomBag[T any]() *RandomBag[T] {
	m := NewMap[T, int]()
	return &RandomBag[T]{
		m:    m,
		data: make([]T, 0),
	}
}

func (b *RandomBag[T]) Insert(item T) bool {
	if b.m.Contains(item) {
		return false
	}
	idx := len(b.data)
	b.data = append(b.data, item)
	b.m.Put(item, idx)
	return true
}

func (b *RandomBag[T]) Remove(item T) bool {
	if !b.m.Contains(item) {
		return false
	}
	idx := b.m.Get(item)
	b.m.Delete(item)
	b.data[idx], b.data[len(b.data)-1] = b.data[len(b.data)-1], b.data[idx]
	b.data = b.data[:len(b.data)-1]
	return true
}

func (b *RandomBag[T]) Peek() T {
	idx := rand.Intn(len(b.data))
	return b.data[idx]
}

func (b *RandomBag[T]) Take() T {
	idx := rand.Intn(len(b.data))
	res := b.data[idx]
	b.Remove(res)
	return res
}
