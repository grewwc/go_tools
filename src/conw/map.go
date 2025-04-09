package conw

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

func (m *Map[K, V]) Iterate() <-chan K {
	sz := m.Size()
	if sz > 32 {
		sz = 32
	}
	ch := make(chan K)
	go func() {
		defer close(ch)
		for k := range m.data {
			ch <- k.(K)
		}
	}()
	return ch
}

func (m *Map[K, V]) Clear() {
	m.data = make(map[any]any)
}
