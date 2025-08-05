package cw

import (
	"fmt"

	"github.com/grewwc/go_tools/src/typesw"
)

type SetT[T comparable] struct {
	data map[T]bool
}

func (s *SetT[T]) Add(item T) {
	s.data[item] = true
}

func (s *SetT[T]) AddAll(items ...T) {
	for _, item := range items {
		s.Add(item)
	}
}

func (s *SetT[T]) Contains(item T) bool {
	if _, exist := s.data[item]; exist {
		return true
	}
	return false
}

func (s *SetT[T]) Iter() typesw.IterableT[T] {
	return typesw.FuncToIterable(func() chan T {
		ch := make(chan T)
		go func() {
			defer close(ch)
			for k := range s.data {
				ch <- k
			}
		}()
		return ch
	})
}

func (s *SetT[T]) MutualExclude(another *SetT[T]) bool {
	for k := range s.data {
		if another.Contains(k) {
			return false
		}
	}
	return true
}

func (s *SetT[T]) Delete(item T) bool {
	if s.Contains(item) {
		delete(s.data, item)
		return true
	}
	return false
}

func (s *SetT[T]) DeleteAll(items ...T) {
	for _, item := range items {
		s.Delete(item)
	}
}

func (s *SetT[T]) Intersect(another *SetT[T]) *SetT[T] {
	if another == nil {
		return nil
	}
	result := NewSetT[T]()
	for k := range s.data {
		if another.Contains(k) {
			result.Add(k)
		}
	}
	return result
}

func (s *SetT[T]) Union(another *SetT[T]) *SetT[T] {
	result := NewSetT[T]()
	result.data = make(map[T]bool, s.Len())
	for k := range s.data {
		result.Add(k)
	}
	if another == nil {
		return result
	}
	for k := range another.data {
		result.Add(k)
	}
	return result
}

func (s *SetT[T]) IsSuperSet(another *SetT[T]) bool {
	for k := range another.data {
		if !s.Contains(k) {
			return false
		}
	}
	return true
}

func (s *SetT[T]) IsSubSet(another *SetT[T]) bool {
	return another.IsSuperSet(s)
}

func (s *SetT[T]) Empty() bool {
	if s == nil {
		return true
	}
	return len(s.data) == 0
}

func (s *SetT[T]) Size() int {
	if s == nil {
		return 0
	}
	return len(s.data)
}

func (s *SetT[T]) Len() int {
	return s.Size()
}

func (s *SetT[T]) Clear() {
	s.data = make(map[T]bool, 8)
}

func (s *SetT[T]) String() string {
	res := make([]T, 0, len(s.data))
	for k := range s.data {
		res = append(res, k)
	}
	return fmt.Sprintf("%v\n", res)
}

func (s *SetT[T]) ShallowCopy() *SetT[T] {
	result := NewSetT[T]()
	for k := range s.data {
		result.Add(k)
	}
	return result
}

func (s *SetT[T]) Subtract(another *SetT[T]) {
	if another == nil {
		return
	}
	for k := range another.data {
		s.Delete(k)
	}
}

func (s *SetT[T]) ToSlice() []T {
	res := make([]T, 0, s.Size())
	for k := range s.data {
		res = append(res, k)
	}
	return res
}

// ToStringSlice is not type safe
func (s *SetT[T]) ToStringSlice() []string {
	res := make([]string, 0, s.Size())
	for k := range s.data {
		res = append(res, fmt.Sprintf("%v", k))
	}
	return res
}

func (s *SetT[T]) Equals(another *SetT[T]) bool {
	return s.IsSubSet(another) && another.IsSubSet(s)
}

func (s *SetT[T]) Data() map[T]bool {
	return s.data
}

func NewSetT[T comparable](items ...T) *SetT[T] {
	s := SetT[T]{data: make(map[T]bool, len(items))}
	s.AddAll(items...)
	return &s
}

func (s *SetT[T]) Reserve(size int) {
	original := s.data
	s.data = make(map[T]bool, size)
	for key := range original {
		s.data[key] = true
	}
}
