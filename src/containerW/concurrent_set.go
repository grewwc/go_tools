package containerW

import (
	"sync"
)

type ConcurrentSet[T any] struct {
	data *Set
	mu   *sync.RWMutex
}

func NewConcurrentSet[T any](items ...T) *ConcurrentSet[T] {
	s := NewSet()
	for _, item := range items {
		s.Add(item)
	}
	return &ConcurrentSet[T]{
		data: s,
		mu:   &sync.RWMutex{},
	}
}

func (s *ConcurrentSet[T]) Add(item T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data.Add(item)
}

func (s *ConcurrentSet[T]) AddAll(items ...T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, item := range items {
		s.data.Add(item)
	}
}

func (s *ConcurrentSet[T]) Contains(item T) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data.Contains(item)
}

func (s *ConcurrentSet[T]) Iterate() <-chan T {
	ret := make(chan T)
	go func() {
		defer close(ret)
		s.mu.RLock()
		defer s.mu.RUnlock()
		for val := range s.data.data {
			ret <- val.(T)
		}
	}()
	return ret
}

func (s *ConcurrentSet[T]) IsMutualExclude(another *ConcurrentSet[T]) bool {
	if another == nil {
		return true
	}
	s.mu.RLock()
	another.mu.RLock()
	defer s.mu.RUnlock()
	defer another.mu.RUnlock()
	return s.data.MutualExclude(another.data)
}

func (s *ConcurrentSet[T]) Delete(item T) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.data.Delete(item)
}

func (s *ConcurrentSet[T]) DeleteAll(items ...T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, item := range items {
		s.data.Delete(item)
	}
}

func (s *ConcurrentSet[T]) Intersect(another *ConcurrentSet[T]) *ConcurrentSet[T] {
	if another == nil {
		return nil
	}
	result := NewConcurrentSet[T]()
	s.mu.RLock()
	another.mu.RLock()
	defer s.mu.RUnlock()
	defer another.mu.RUnlock()
	for k := range s.data.data {
		if another.data.Contains(k) {
			result.Add(k.(T))
		}
	}
	return result
}

func (s *ConcurrentSet[T]) Union(another *ConcurrentSet[T]) *ConcurrentSet[T] {
	result := NewConcurrentSet[T]()
	s.mu.RLock()
	for k := range s.data.data {
		result.Add(k.(T))
	}
	s.mu.RUnlock()
	if another == nil {
		return result
	}
	another.mu.RLock()
	defer another.mu.RUnlock()
	for k := range another.data.data {
		result.Add(k.(T))
	}
	return result
}

func (s *ConcurrentSet[T]) IsSuperSet(another *ConcurrentSet[T]) bool {
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

func (s *ConcurrentSet[T]) IsSubSet(another *ConcurrentSet[T]) bool {
	if another == nil {
		return false
	}
	return another.IsSuperSet(s)
}

func (s *ConcurrentSet[T]) Empty() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.data.data) == 0
}

func (s *ConcurrentSet[T]) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.data.data)
}

func (s *ConcurrentSet[T]) Clear() {
	s.mu.Lock()
	s.data.data = make(map[interface{}]bool, 8)
	s.mu.Unlock()
}
