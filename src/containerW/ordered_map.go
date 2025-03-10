package containerW

import (
	"container/list"
	"fmt"
)

// OrderedMap is a map that maintains the order of insertion.
type OrderedMap struct {
	m map[interface{}]*list.Element
	l *list.List
}

type MapEntry struct {
	k, v interface{}
}

func (e MapEntry) Key() interface{} {
	return e.k
}

func (e MapEntry) Val() interface{} {
	return e.v
}

func NewOrderedMap() *OrderedMap {
	l := list.New()
	res := &OrderedMap{make(map[interface{}]*list.Element), l}
	return res
}

func (s *OrderedMap) Put(k, v interface{}) {
	if node, exist := s.m[k]; !exist {
		e := s.l.PushBack(&MapEntry{k, v})
		s.m[k] = e
	} else {
		node.Value = &MapEntry{k, v}
	}
}

func (s *OrderedMap) PutIfAbsent(k, v interface{}) {
	if _, ok := s.m[k]; ok {
		return
	}
	s.Put(k, v)
}

func (s *OrderedMap) Get(k interface{}) interface{} {
	return s.m[k].Value.(*MapEntry).v
}

func (s *OrderedMap) GetOrDefault(k, defaultVal interface{}) interface{} {
	if val, ok := s.m[k]; ok {
		return val.Value.(*MapEntry).v
	}
	return defaultVal
}

func (s OrderedMap) Iterate() <-chan *MapEntry {
	c := make(chan *MapEntry)
	go func() {
		defer close(c)
		l := s.l.Len()
		elem := s.l.Front()
		if elem == nil {
			return
		}
		for i := 0; i < l; i++ {
			c <- elem.Value.(*MapEntry)
			elem = elem.Next()
		}
	}()
	return c
}

func (s *OrderedMap) Contains(k interface{}) bool {
	if _, exist := s.m[k]; exist {
		return true
	}
	return false
}

func (s *OrderedMap) Delete(k interface{}) bool {
	if val, ok := s.m[k]; ok {
		delete(s.m, k)
		s.l.Remove(val)
		return true
	}
	return false
}

func (s *OrderedMap) DeleteAll(ks ...interface{}) {
	for _, k := range ks {
		s.Delete(k)
	}
}

func (s *OrderedMap) Empty() bool {
	return len(s.m) == 0
}

func (s *OrderedMap) Size() int {
	return len(s.m)
}

func (s *OrderedMap) Clear() {
	s.m = make(map[interface{}]*list.Element)
	s.l.Init()
}

func (s *OrderedMap) String() string {
	res := make([]interface{}, 0, len(s.m))
	front := s.l.Front()
	if front == nil {
		return "\n"
	}
	for i := 0; i < s.Size(); i++ {
		k := front.Value.(*MapEntry).k
		res = append(res, fmt.Sprintf("%v: %v", k, s.Get(k)))
		front = front.Next()
	}

	return fmt.Sprintf("%v", res)
}

func (s OrderedMap) ShallowCopy() *OrderedMap {
	result := NewOrderedMap()
	front := s.l.Front()
	if front == nil {
		return result
	}
	for i := 0; i < s.Size(); i++ {
		k := front.Value.(*MapEntry).k
		result.Put(k, s.Get(k))
		front = front.Next()
	}
	return result
}
