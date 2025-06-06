package cw

import "github.com/grewwc/go_tools/src/typesw"

type Map[K, V any] struct {
	data map[any]any
}

func NewMap[K, V any]() *Map[K, V] {
	return &Map[K, V]{
		data: make(map[any]any),
	}
}

func (m *Map[K, V]) Get(key K) V {
	if val, ok := m.data[key]; ok {
		return val.(V)
	}
	return *new(V)
}

func (m *Map[K, V]) GetOrDefault(key K, defaultVal V) V {
	if val, ok := m.data[key]; ok {
		return val.(V)
	}
	return defaultVal
}

func (m *Map[K, V]) Contains(key K) bool {
	_, ok := m.data[key]
	return ok
}

func (m *Map[K, V]) Keys() []K {
	res := make([]K, 0, len(m.data))
	for k := range m.data {
		res = append(res, k.(K))
	}
	return res
}

func (m *Map[K, V]) Values() []V {
	s := NewSet()
	for _, v := range m.data {
		s.Add(v)
	}
	res := make([]V, 0, s.Size())
	for val := range s.Iter().Iterate() {
		res = append(res, val.(V))
	}
	return res
}

func (m *Map[K, V]) Put(key K, value V) bool {
	_, ok := m.data[key]
	m.data[key] = value
	return ok
}

func (m *Map[K, V]) PutIfAbsent(key K, value V) bool {
	if m.Contains(key) {
		return false
	}
	m.data[key] = value
	return true
}

func (m *Map[K, V]) Size() int {
	return len(m.data)
}

func (m *Map[K, V]) Delete(key K) bool {
	if !m.Contains(key) {
		return false
	}
	delete(m.data, key)
	return true
}

func (m *Map[K, V]) DeleteAll(keys ...K) {
	for _, key := range keys {
		delete(m.data, key)
	}
}

func (m *Map[K, V]) Iter() typesw.IterableT[K] {
	return &interfaceKeyMapIterator[K, interface{}]{
		data: m.data,
	}
}

func (m *Map[K, V]) IterEntry() typesw.IterableT[typesw.IMapEntry[K, V]] {
	f := func() chan typesw.IMapEntry[K, V] {
		ch := make(chan typesw.IMapEntry[K, V])
		go func() {
			for k, v := range m.data {
				ch <- &MapEntry[K, V]{k.(K), v.(V)}
			}
			close(ch)
		}()
		return ch
	}
	return typesw.FuncToIterable(f)
}

func (m *Map[K, V]) ReverseKV() *Map[V, K] {
	res := NewMap[V, K]()
	for t := range m.IterEntry().Iterate() {
		res.Put(t.Val(), t.Key())
	}
	return res
}

func (m *Map[K, V]) Clear() {
	m.data = make(map[any]any)
}
