package cw

import "github.com/grewwc/go_tools/src/typesw"

type TreeSet[T any] struct {
	data *TreeMap[T, struct{}]
	cmp  typesw.CompareFunc[T]
}

func NewTreeSet[T any](cmp typesw.CompareFunc[T]) *TreeSet[T] {
	return &TreeSet[T]{
		data: NewTreeMap[T, struct{}](cmp),
	}
}

func (set *TreeSet[T]) Add(e T) {
	set.data.Put(e, struct{}{})
}

func (set *TreeSet[T]) AddAll(e ...T) {
	for _, v := range e {
		set.Add(v)
	}
}

func (set *TreeSet[T]) Delete(e T) bool {
	return set.data.Delete(e)
}

func (set *TreeSet[T]) DeleteAll(e ...T) {
	set.data.DeleteAll(e...)
}

func (set *TreeSet[T]) Contains(e T) bool {
	return set.data.Contains(e)
}

func (set *TreeSet[T]) AddIfAbsent(e T) bool {
	return set.data.PutIfAbsent(e, struct{}{})
}

func (set *TreeSet[T]) Size() int {
	return set.data.Size()
}

func (set *TreeSet[T]) Len() int {
	return set.Size()
}

func (set *TreeSet[T]) Iterate() <-chan T {
	return set.data.Iterate()
}

func (set *TreeSet[T]) Clear() {
	set.data.Clear()
}

func (s *TreeSet[T]) MutualExclude(another *TreeSet[T]) bool {
	for k := range s.Iterate() {
		if another.Contains(k) {
			return false
		}
	}
	return true
}

func (s *TreeSet[T]) Intersect(another *TreeSet[T]) *TreeSet[T] {
	result := NewTreeSet(s.cmp)
	for k := range s.Iterate() {
		if another.Contains(k) {
			result.Add(k)
		}
	}
	return result
}

func (s *TreeSet[T]) Union(another *TreeSet[T]) *TreeSet[T] {
	result := NewTreeSet(s.cmp)
	for k := range s.Iterate() {
		result.Add(k)
	}
	for k := range another.Iterate() {
		result.Add(k)
	}
	return result
}

func (s *TreeSet[T]) IsSuperSet(another *TreeSet[T]) bool {
	for k := range another.data.Iterate() {
		if !s.Contains(k) {
			return false
		}
	}
	return true
}

func (s *TreeSet[T]) IsSubSet(another *TreeSet[T]) bool {
	return another.IsSuperSet(s)
}

func (s *TreeSet[T]) Empty() bool {
	return s.Len() == 0
}

func (s *TreeSet[T]) ShallowCopy() *TreeSet[T] {
	result := NewTreeSet(s.cmp)
	for k := range s.Iterate() {
		result.Add(k)
	}
	return result
}

func (s *TreeSet[T]) Subtract(another *TreeSet[T]) {
	for k := range another.Iterate() {
		s.Delete(k)
	}
}

func (s *TreeSet[T]) ToSlice() []T {
	res := make([]T, 0, s.Size())
	for k := range s.Iterate() {
		res = append(res, k)
	}
	return res
}

func (s *TreeSet[T]) Equals(another *TreeSet[T]) bool {
	return s.IsSubSet(another) && s.IsSuperSet(another)
}
