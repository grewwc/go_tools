package containerW

import (
	"container/list"
	"fmt"
)

type OrderedSet struct {
	m map[interface{}]*list.Element
	l *list.List
}

func NewOrderedSet() *OrderedSet {
	l := list.New()
	return &OrderedSet{make(map[interface{}]*list.Element), l}
}

func (s *OrderedSet) Add(v interface{}) {
	if _, exist := s.m[v]; !exist {
		e := s.l.PushBack(v)
		s.m[v] = e
	}
}

func (s *OrderedSet) AddAll(vs ...interface{}) {
	for _, v := range vs {
		s.Add(v)
	}
}

func (s OrderedSet) Iterate() <-chan interface{} {
	c := make(chan interface{})
	go func() {
		defer close(c)
		l := s.l.Len()
		elem := s.l.Front()
		if elem == nil {
			return
		}
		for i := 0; i < l; i++ {
			c <- elem.Value
			elem = elem.Next()
		}
	}()
	return c
}

func (s *OrderedSet) Contains(v interface{}) bool {
	if _, exist := s.m[v]; exist {
		return true
	}
	return false
}

func (s *OrderedSet) Delete(v interface{}) bool {
	if val, ok := s.m[v]; ok {
		delete(s.m, v)
		s.l.Remove(val)
		return true
	}
	return false
}

func (s *OrderedSet) DeleteAll(vs ...interface{}) {
	for _, v := range vs {
		s.Delete(v)
	}
}

func (s OrderedSet) Intersect(another OrderedSet) *OrderedSet {
	result := NewOrderedSet()
	for k := range s.m {
		if another.Contains(k) {
			result.Add(k)
		}
	}
	return result
}

func (s OrderedSet) Union(another OrderedSet) *OrderedSet {
	result := NewOrderedSet()
	for k := range s.m {
		result.Add(k)
	}
	for k := range another.m {
		result.Add(k)
	}
	return result
}

func (s OrderedSet) IsSuperSet(another OrderedSet) bool {
	for k := range another.m {
		if !s.Contains(k) {
			return false
		}
	}
	return true
}

func (s OrderedSet) IsSubSet(another OrderedSet) bool {
	return another.IsSuperSet(s)
}

func (s *OrderedSet) Empty() bool {
	return len(s.m) == 0
}

func (s *OrderedSet) Size() int {
	return len(s.m)
}

func (s *OrderedSet) Clear() {
	s.m = make(map[interface{}]*list.Element)
	s.l.Init()
}

func (s *OrderedSet) String() string {
	res := make([]interface{}, 0, len(s.m))
	front := s.l.Front()
	if front == nil {
		return "\n"
	}
	for i := 0; i < s.Size(); i++ {
		res = append(res, front.Value)
		front = front.Next()
	}

	return fmt.Sprintf("%v\n", res)
}

func (s OrderedSet) ShallowCopy() *OrderedSet {
	result := NewOrderedSet()
	front := s.l.Front()
	if front == nil {
		return result
	}
	for i := 0; i < s.Size(); i++ {
		result.Add(front.Value)
		front = front.Next()
	}
	return result
}

func (s *OrderedSet) Subtract(another OrderedSet) {
	for k := range another.m {
		s.Delete(k)
	}
}

func (s OrderedSet) ToSlice() []interface{} {
	res := make([]interface{}, 0, s.Size())
	for v := range s.Iterate() {
		res = append(res, v)
	}
	return res
}

// ToStringSlice is not type safe
func (s OrderedSet) ToStringSlice() []string {
	res := make([]string, 0, s.Size())
	for v := range s.Iterate() {
		res = append(res, v.(string))
	}
	return res
}

func (s OrderedSet) Equals(another OrderedSet) bool {
	return s.IsSubSet(another) && s.IsSubSet(another)
}
