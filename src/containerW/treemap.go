package containerW

import "github.com/grewwc/go_tools/src/typesW"

type sortedMapEntry[K any, V any] struct {
	k K
	v V
}

type TreeMap[K any, V any] struct {
	rbTree *RbTree[*sortedMapEntry[K, V]]
}

func NewTreeMap[K, V any](cmp typesW.CompareFunc[K]) *TreeMap[K, V] {
	if cmp == nil {
		cmp = typesW.CreateDefaultCmp[K]()
	}
	cmpWrapper := func(a, b *sortedMapEntry[K, V]) int {
		return cmp(a.k, b.k)
	}
	return &TreeMap[K, V]{
		rbTree: NewRbTree(cmpWrapper),
	}
}

func (m *TreeMap[K, V]) Get(key K) V {
	ret := m.rbTree.Search(&sortedMapEntry[K, V]{k: key})
	if ret == nil {
		return *new(V)
	}
	return ret.val.v
}

func (m *TreeMap[K, V]) GetOrDefault(key K, defaultVal V) V {
	ret := m.rbTree.Search(&sortedMapEntry[K, V]{k: key})
	if ret == nil {
		return defaultVal
	}
	return ret.val.v
}

func (m *TreeMap[K, V]) Contains(key K) bool {
	return m.rbTree.Contains(&sortedMapEntry[K, V]{k: key})
}

func (m *TreeMap[K, V]) Put(key K, value V) bool {
	node := sortedMapEntry[K, V]{k: key, v: value}
	n := m.rbTree.Search(&node)
	if n == nil {
		m.rbTree.Insert(&node)
		return false
	}
	n.val.v = value
	return true
}

func (m *TreeMap[K, V]) PutIfAbsent(key K, value V) bool {
	node := sortedMapEntry[K, V]{k: key, v: value}
	if m.rbTree.Contains(&node) {
		return false
	}
	m.rbTree.Insert(&node)
	return true
}

func (m *TreeMap[K, V]) Size() int {
	return m.rbTree.size
}

func (m *TreeMap[K, V]) Len() int {
	return m.Size()
}

func (m *TreeMap[K, V]) Delete(key K) bool {
	node := sortedMapEntry[K, V]{k: key}
	n := m.rbTree.Search(&node)
	if n == nil {
		return false
	}
	m.rbTree.Delete(&node)
	return true
}

func (m *TreeMap[K, V]) DeleteAll(keys ...K) {
	for _, key := range keys {
		m.Delete(key)
	}
}

func (m *TreeMap[K, V]) Iterate() <-chan K {
	ch := make(chan K)
	go func() {
		defer close(ch)
		for val := range m.rbTree.Iterate() {
			ch <- val.k
		}
	}()

	return ch
}

func (m *TreeMap[K, V]) Clear() {
	m.rbTree.Clear()
}

func (m *TreeMap[K, V]) SearchRange(lower, upper K) []K {
	lowerEntry := sortedMapEntry[K, V]{k: lower}
	upperEntry := sortedMapEntry[K, V]{k: upper}
	entry := m.rbTree.SearchRange(&lowerEntry, &upperEntry)
	ret := make([]K, 0, len(entry))
	for _, e := range entry {
		ret = append(ret, e.k)
	}
	return ret
}
