package cw

import "github.com/grewwc/go_tools/src/typesw"

type TreeMap[K any, V any] struct {
	rbTree *RbTree[*MapEntry[K, V]]
}

func NewTreeMap[K, V any](cmp typesw.CompareFunc[K]) *TreeMap[K, V] {
	if cmp == nil {
		cmp = typesw.CreateDefaultCmp[K]()
	}
	cmpWrapper := func(a, b *MapEntry[K, V]) int {
		return cmp(a.k, b.k)
	}
	return &TreeMap[K, V]{
		rbTree: NewRbTree(cmpWrapper),
	}
}

func (m *TreeMap[K, V]) Get(key K) V {
	ret := m.rbTree.Search(&MapEntry[K, V]{k: key})
	if ret == nil {
		return *new(V)
	}
	return ret.val.v
}

func (m *TreeMap[K, V]) GetOrDefault(key K, defaultVal V) V {
	ret := m.rbTree.Search(&MapEntry[K, V]{k: key})
	if ret == nil {
		return defaultVal
	}
	return ret.val.v
}

func (m *TreeMap[K, V]) Contains(key K) bool {
	return m.rbTree.Contains(&MapEntry[K, V]{k: key})
}

func (m *TreeMap[K, V]) Keys() []K {
	res := make([]K, 0, m.Size())
	for k := range m.Iter().Iterate() {
		res = append(res, k)
	}
	return res
}

func (m *TreeMap[K, V]) Values() []V {
	s := NewSet()
	for entry := range m.IterEntry().Iterate() {
		s.Add(entry.v)
	}
	res := make([]V, 0, s.Size())
	for val := range s.Iter().Iterate() {
		res = append(res, val.(V))
	}
	return res
}

func (m *TreeMap[K, V]) Put(key K, value V) bool {
	node := MapEntry[K, V]{k: key, v: value}
	n := m.rbTree.Search(&node)
	if n == nil {
		m.rbTree.Insert(&node)
		return true
	}
	n.val.v = value
	return false
}

func (m *TreeMap[K, V]) PutIfAbsent(key K, value V) bool {
	node := MapEntry[K, V]{k: key, v: value}
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
	node := MapEntry[K, V]{k: key}
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

func (m *TreeMap[K, V]) Iter() typesw.IterableT[K] {
	f := func() chan K {
		ch := make(chan K)
		go func() {
			defer close(ch)
			for val := range m.rbTree.Iter().Iterate() {
				ch <- val.k
			}
		}()

		return ch
	}
	return typesw.FuncToIterable(f)
}

func (m *TreeMap[K, V]) IterEntry() typesw.IterableT[*MapEntry[K, V]] {
	f := func() chan *MapEntry[K, V] {
		ch := make(chan *MapEntry[K, V])
		go func() {
			defer close(ch)
			for val := range m.rbTree.Iter().Iterate() {
				ch <- val
			}
		}()
		return ch
	}
	return typesw.FuncToIterable(f)
}

func (m *TreeMap[K, V]) ForEachEntry(f func(*MapEntry[K, V])) {
	for entry := range m.rbTree.Iter().Iterate() {
		f(entry)
	}
}

func (m *TreeMap[K, V]) ForEach(f func(k K)) {
	for entry := range m.rbTree.Iter().Iterate() {
		f(entry.k)
	}
}

func (m *TreeMap[K, V]) Clear() {
	m.rbTree.Clear()
}

func (m *TreeMap[K, V]) SearchRange(lower, upper K) typesw.IterableT[K] {
	f := func() chan K {
		lowerEntry := MapEntry[K, V]{k: lower}
		upperEntry := MapEntry[K, V]{k: upper}
		entry := m.rbTree.SearchRange(&lowerEntry, &upperEntry)
		ch := make(chan K)
		go func() {
			defer close(ch)
			for e := range entry.Iterate() {
				ch <- e.k
			}
		}()
		return ch
	}
	return typesw.FuncToIterable(f)
}
