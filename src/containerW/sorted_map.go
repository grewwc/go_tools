package containerW

type sortedMapEntry[K any, V any] struct {
	k K
	v V
}

type SortedMap[K any, V any] struct {
	rbTree *RbTree[*sortedMapEntry[K, V]]
}

func NewSortedMap[K, V any](cmp compareFunc) *SortedMap[K, V] {
	if cmp == nil {
		cmp = createDefaultCmp[K]()
	}
	cmpWrapper := func(a, b interface{}) int {
		return cmp(a.(*sortedMapEntry[K, V]).k, b.(*sortedMapEntry[K, V]).k)
	}
	return &SortedMap[K, V]{
		rbTree: NewRbTree[*sortedMapEntry[K, V]](cmpWrapper),
	}
}

func (m *SortedMap[K, V]) Get(key K) V {
	ret := m.rbTree.Search(&sortedMapEntry[K, V]{k: key})
	if ret == nil {
		return *new(V)
	}
	return ret.val.v
}

func (m *SortedMap[K, V]) GetOrDefault(key K, defaultVal V) V {
	ret := m.rbTree.Search(&sortedMapEntry[K, V]{k: key})
	if ret == nil {
		return defaultVal
	}
	return ret.val.v
}

func (m *SortedMap[K, V]) Contains(key K) bool {
	return m.rbTree.Contains(&sortedMapEntry[K, V]{k: key})
}

func (m *SortedMap[K, V]) Put(key K, value V) bool {
	node := sortedMapEntry[K, V]{k: key, v: value}
	n := m.rbTree.Search(&node)
	if n == nil {
		m.rbTree.Insert(&node)
		return false
	}
	n.val.v = value
	return true
}

func (m *SortedMap[K, V]) PutIfAbsent(key K, value V) bool {
	node := sortedMapEntry[K, V]{k: key, v: value}
	if m.rbTree.Contains(&node) {
		return false
	}
	m.rbTree.Insert(&node)
	return true
}

func (m *SortedMap[K, V]) Size() int {
	return m.rbTree.size
}

func (m *SortedMap[K, V]) Delete(key K) bool {
	node := sortedMapEntry[K, V]{k: key}
	n := m.rbTree.Search(&node)
	if n == nil {
		return false
	}
	m.rbTree.Delete(&node)
	return true
}

func (m *SortedMap[K, V]) DeleteAll(keys ...K) {
	for _, key := range keys {
		m.Delete(key)
	}
}

func (m *SortedMap[K, V]) Iterate() <-chan K {
	ch := make(chan K)
	go func() {
		defer close(ch)
		for val := range m.rbTree.Iterate() {
			ch <- val.k
		}
	}()

	return ch
}

func (m *SortedMap[K, V]) Clear() {
	m.rbTree.Clear()
}

func (m *SortedMap[K, V]) SearchRange(lower, upper K) []K {
	lowerEntry := sortedMapEntry[K, V]{k: lower}
	upperEntry := sortedMapEntry[K, V]{k: upper}
	entry := m.rbTree.SearchRange(&lowerEntry, &upperEntry)
	ret := make([]K, 0, len(entry))
	for _, e := range entry {
		ret = append(ret, e.k)
	}
	return ret
}
