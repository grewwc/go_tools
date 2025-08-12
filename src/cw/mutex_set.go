package cw

import (
	"sync"
)

type MutexSet[T comparable] struct {
	data *SetT[T]
	mu   *sync.RWMutex
}

func NewMutexSet[T comparable](items ...T) *MutexSet[T] {
	s := NewSetT[T]()
	for _, item := range items {
		s.Add(item)
	}
	return &MutexSet[T]{
		data: s,
		mu:   &sync.RWMutex{},
	}
}

func (s *MutexSet[T]) Add(item T) {
	s.mu.Lock()
	s.data.Add(item)
	s.mu.Unlock()
}

func (s *MutexSet[T]) AddAll(items ...T) {
	s.mu.Lock()
	for _, item := range items {
		s.data.Add(item)
	}
	s.mu.Unlock()
}

func (s *MutexSet[T]) Contains(item T) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data.Contains(item)
}

func (s *MutexSet[T]) Iterate() <-chan T {
	ret := make(chan T)
	go func() {
		defer close(ret)
		s.mu.RLock()
		defer s.mu.RUnlock()
		for val := range s.data.data {
			ret <- val
		}
	}()
	return ret
}

func (s *MutexSet[T]) ForEach(f func(val T)) {
	s.data.ForEach(f)
}


func (s *MutexSet[T]) IsMutualExclude(another *MutexSet[T]) bool {
	if another == nil {
		return true
	}
	s.mu.RLock()
	another.mu.RLock()
	defer s.mu.RUnlock()
	defer another.mu.RUnlock()
	return s.data.MutualExclude(another.data)
}

func (s *MutexSet[T]) Delete(item T) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.data.Delete(item)
}

func (s *MutexSet[T]) DeleteAll(items ...T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, item := range items {
		s.data.Delete(item)
	}
}

func (s *MutexSet[T]) Intersect(another *MutexSet[T]) *MutexSet[T] {
	if another == nil {
		return nil
	}
	result := NewMutexSet[T]()
	s.mu.RLock()
	another.mu.RLock()
	defer s.mu.RUnlock()
	defer another.mu.RUnlock()
	for k := range s.data.data {
		if another.data.Contains(k) {
			result.Add(k)
		}
	}
	return result
}

func (s *MutexSet[T]) Union(another *MutexSet[T]) *MutexSet[T] {
	result := NewMutexSet[T]()
	s.mu.RLock()
	for k := range s.data.data {
		result.Add(k)
	}
	s.mu.RUnlock()
	if another == nil {
		return result
	}
	another.mu.RLock()
	defer another.mu.RUnlock()
	for k := range another.data.data {
		result.Add(k)
	}
	return result
}

func (s *MutexSet[T]) IsSuperSet(another *MutexSet[T]) bool {
	if another == nil {
		return true
	}
	s.mu.RLock()
	another.mu.RLock()
	defer s.mu.RUnlock()
	defer another.mu.RUnlock()
	for k := range another.data.data {
		if !s.data.Contains(k) {
			return false
		}
	}
	return true
}

func (s *MutexSet[T]) IsSubSet(another *MutexSet[T]) bool {
	if another == nil {
		return false
	}
	return another.IsSuperSet(s)
}

func (s *MutexSet[T]) Empty() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.data.data) == 0
}

func (s *MutexSet[T]) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.data.data)
}

func (s *MutexSet[T]) Clear() {
	s.mu.Lock()
	s.data.data = make(map[T]bool, 8)
	s.mu.Unlock()
}

func (s *MutexSet[T]) ToSlice() []T {
	res := make([]T, 0, s.Size())
	s.mu.RLock()
	defer s.mu.RUnlock()
	for k := range s.data.data {
		res = append(res, k)
	}
	return res
}
