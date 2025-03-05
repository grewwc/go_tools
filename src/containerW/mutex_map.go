package containerW

import (
	"sync"
)

type MutexMap[K, V any] struct {
	data map[any]any
	mu   *sync.RWMutex
}

func NewMutexMap[K, V any]() *MutexMap[K, V] {
	return &MutexMap[K, V]{
		data: make(map[any]any),
		mu:   &sync.RWMutex{},
	}
}

func (m *MutexMap[K, V]) Get(key K) V {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if val, ok := m.data[key]; ok {
		return val.(V)
	}
	return *new(V)
}

func (m *MutexMap[K, V]) GetOrDefault(key K, defaultVal V) V {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if val, ok := m.data[key]; ok {
		return val.(V)
	}
	return defaultVal
}

func (m *MutexMap[K, V]) Contains(key K) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.data[key]
	return ok
}

func (m *MutexMap[K, V]) Put(key K, value V) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.data[key]
	m.data[key] = value
	return ok
}

func (m *MutexMap[K, V]) PutIfAbsent(key K, value V) bool {
	if m.Contains(key) {
		return false
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
	return true
}

func (m *MutexMap[K, V]) Size() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.data)
}

func (m *MutexMap[K, V]) Delete(key K) bool {
	if !m.Contains(key) {
		return false
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
	return true
}

func (m *MutexMap[K, V]) DeleteAll(keys ...K) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, key := range keys {
		delete(m.data, key)
	}
}

func (m *MutexMap[K, V]) Iterate() <-chan K {
	sz := m.Size()
	if sz > 32 {
		sz = 32
	}
	ch := make(chan K)
	go func() {
		defer close(ch)
		m.mu.RLock()
		defer m.mu.RUnlock()
		for k := range m.data {
			ch <- k.(K)
		}
	}()
	return ch
}
