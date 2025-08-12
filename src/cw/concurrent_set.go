package cw

import "github.com/grewwc/go_tools/src/typesw"

var empty = struct{}{}

type ConcurrentHashSet[T any] struct {
	data *ConcurrentHashMap[T, struct{}]
}

func NewConcurrentHashSet[T any](hasher typesw.HashFunc[T], cmp typesw.CompareFunc[T]) *ConcurrentHashSet[T] {
	return &ConcurrentHashSet[T]{
		data: NewConcurrentHashMap[T, struct{}](hasher, cmp),
	}
}

func (s *ConcurrentHashSet[T]) Add(item T) {
	s.data.Put(item, empty)
}

func (s *ConcurrentHashSet[T]) AddAll(items ...T) {
	for _, item := range items {
		s.Add(item)
	}
}

func (s *ConcurrentHashSet[T]) Contains(item T) bool {
	return s.data.Contains(item)
}

func (s *ConcurrentHashSet[T]) Iter() typesw.IterableT[T] {
	f := func() chan T {
		ret := make(chan T)
		go func() {
			defer close(ret)
			for val := range s.data.Iter().Iterate() {
				ret <- val
			}
		}()

		return ret
	}
	return typesw.FuncToIterable(f)
}

func (s *ConcurrentHashSet[T]) ForEach(f func(val T)) {
	s.data.ForEach(f)
}

func (s *ConcurrentHashSet[T]) IsMutualExclude(another *ConcurrentHashSet[T]) bool {
	if another == nil {
		return true
	}
	for item := range s.Iter().Iterate() {
		if another.Contains(item) {
			return false
		}
	}
	return true
}

func (s *ConcurrentHashSet[T]) Delete(item T) bool {
	return s.data.Delete(item)
}

func (s *ConcurrentHashSet[T]) DeleteAll(items ...T) {
	for _, item := range items {
		s.data.Delete(item)
	}
}

func (s *ConcurrentHashSet[T]) Intersect(another *ConcurrentHashSet[T]) *ConcurrentHashSet[T] {
	if another == nil {
		return nil
	}
	result := NewConcurrentHashSet(s.data.hash, s.data.cmp)
	for k := range s.Iter().Iterate() {
		if another.data.Contains(k) {
			result.Add(k)
		}
	}
	return result
}

func (s *ConcurrentHashSet[T]) Union(another *ConcurrentHashSet[T]) *ConcurrentHashSet[T] {
	result := NewConcurrentHashSet(s.data.hasher, s.data.cmp)
	for k := range s.data.Iter().Iterate() {
		result.Add(k)
	}
	if another == nil {
		return result
	}
	for k := range another.data.Iter().Iterate() {
		result.Add(k)
	}
	return result
}

func (s *ConcurrentHashSet[T]) IsSuperSet(another *ConcurrentHashSet[T]) bool {
	if another == nil {
		return true
	}
	for k := range another.data.Iter().Iterate() {
		if !s.data.Contains(k) {
			return false
		}
	}
	return true
}

func (s *ConcurrentHashSet[T]) IsSubSet(another *ConcurrentHashSet[T]) bool {
	if another == nil {
		return false
	}
	return another.IsSuperSet(s)
}

func (s *ConcurrentHashSet[T]) Empty() bool {
	return s.Size() == 0
}

func (s *ConcurrentHashSet[T]) Size() int {
	return s.data.Size()
}

func (s *ConcurrentHashSet[T]) Clear() {
	s.data.Clear()
}

func (s *ConcurrentHashSet[T]) ToSlice() []T {
	res := make([]T, 0, s.Size())
	for k := range s.data.Iter().Iterate() {
		res = append(res, k)
	}
	return res
}
