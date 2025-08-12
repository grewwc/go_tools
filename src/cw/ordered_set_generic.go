package cw

import (
	"fmt"
	"strings"

	"github.com/grewwc/go_tools/src/typesw"
)

type OrderedSetT[T comparable] struct {
	data *OrderedMapT[T, bool]
}

func NewOrderedSetT[T comparable](items ...T) *OrderedSetT[T] {
	res := &OrderedSetT[T]{data: NewOrderedMapT[T, bool]()}
	for _, item := range items {
		res.data.Put(item, true)
	}
	return res
}

// Add 如果已经存在，则忽略
func (s *OrderedSetT[T]) Add(v T) {
	s.data.PutIfAbsent(v, true)
}

func (s *OrderedSetT[T]) AddAll(vs ...T) {
	for _, v := range vs {
		s.Add(v)
	}
}

func (s *OrderedSetT[T]) Iter() typesw.IterableT[T] {
	return typesw.FuncToIterable(func() chan T {
		ch := make(chan T)
		go func() {
			defer close(ch)
			for entry := range s.data.Iter().Iterate() {
				ch <- entry.Key()
			}
		}()
		return ch
	})
}

func (s *OrderedSetT[T]) ForEach(f func(val T)) {
	s.data.ForEach(f)
}

func (s *OrderedSetT[T]) Data() *OrderedMapT[T, bool] {
	return s.data
}

func (s *OrderedSetT[T]) Contains(v T) bool {
	return s.data.Contains(v)
}

func (s *OrderedSetT[T]) Delete(v T) bool {
	return s.data.Delete(v)
}

func (s *OrderedSetT[T]) DeleteAll(vs ...T) {
	for _, v := range vs {
		s.Delete(v)
	}
}

func (s *OrderedSetT[T]) Intersect(another *OrderedSetT[T]) *OrderedSetT[T] {
	result := NewOrderedSetT[T]()
	for k := range s.Iter().Iterate() {
		if another.Contains(k) {
			result.Add(k)
		}
	}
	return result
}

func (s *OrderedSetT[T]) MutualExclude(another *OrderedSetT[T]) bool {
	for k := range s.Iter().Iterate() {
		if another.Contains(k) {
			return false
		}
	}
	return true
}

func (s *OrderedSetT[T]) Union(another *OrderedSetT[T]) *OrderedSetT[T] {
	result := NewOrderedSetT[T]()
	for k := range s.Iter().Iterate() {
		result.Add(k)
	}
	for k := range another.Iter().Iterate() {
		result.Add(k)
	}
	return result
}

func (s *OrderedSetT[T]) IsSuperSet(another *OrderedSetT[T]) bool {
	for k := range another.Iter().Iterate() {
		if !s.Contains(k) {
			return false
		}
	}
	return true
}

func (s *OrderedSetT[T]) IsSubSet(another *OrderedSetT[T]) bool {
	return another.IsSuperSet(s)
}

func (s *OrderedSetT[T]) Empty() bool {
	return s == nil || s.data.Empty()
}

func (s *OrderedSetT[T]) Size() int {
	if s == nil {
		return 0
	}
	return s.data.Size()
}

func (s *OrderedSetT[T]) Len() int {
	return s.Size()
}

func (s *OrderedSetT[T]) Clear() {
	if s == nil {
		return
	}
	s.data.Clear()
}

func (s *OrderedSetT[T]) String() string {
	res := s.ToStringSlice()
	return strings.Join(res, " ")
}

func (s *OrderedSetT[T]) ShallowCopy() *OrderedSetT[T] {
	if s == nil {
		return nil
	}
	if s.Empty() {
		return NewOrderedSetT[T]()
	}
	result := NewOrderedSetT[T]()
	for val := range s.Iter().Iterate() {
		result.Add(val)
	}
	return result
}

func (s *OrderedSetT[T]) Subtract(another *OrderedSetT[T]) {
	for k := range another.Iter().Iterate() {
		s.Delete(k)
	}
}

func (s *OrderedSetT[T]) ToSlice() []T {
	res := make([]T, 0, s.Size())
	if s.Empty() {
		return res
	}
	for v := range s.Iter().Iterate() {
		res = append(res, v)
	}
	return res
}

// ToStringSlice is not type safe
func (s *OrderedSetT[T]) ToStringSlice() []string {
	res := make([]string, 0, s.Len())
	if s.Empty() {
		return res
	}
	for val := range s.Iter().Iterate() {
		res = append(res, fmt.Sprintf("%v", val))
	}

	return res
}

func (s *OrderedSetT[T]) Equals(another *OrderedSetT[T]) bool {
	return s.IsSubSet(another) && another.IsSubSet(s)
}
