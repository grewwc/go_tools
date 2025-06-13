package cw

import (
	"container/list"
	"fmt"

	"github.com/grewwc/go_tools/src/typesw"
)

type OrderedSet struct {
	m map[interface{}]*list.Element
	l *list.List
}

func NewOrderedSet(items ...interface{}) *OrderedSet {
	l := list.New()
	res := &OrderedSet{make(map[interface{}]*list.Element, 8), l}
	res.AddAll(items...)
	return res
}

// Add 如果已经存在，则忽略
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

func (s OrderedSet) Iter() typesw.Iterable {
	return &listIterator[any]{
		data:    s.l,
		reverse: false,
	}
}

func (s OrderedSet) ReverseIter() typesw.Iterable {
	return &listIterator[any]{
		data:    s.l,
		reverse: true,
	}
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

func (s OrderedSet) MutualExclude(another OrderedSet) bool {
	for k := range s.m {
		if another.Contains(k) {
			return false
		}
	}
	return true
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

	return fmt.Sprintf("%v", res)
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
	for v := range s.Iter().Iterate() {
		res = append(res, v)
	}
	return res
}

// ToStringSlice is not type safe
func (s OrderedSet) ToStringSlice() []string {
	l := s.Size()
	res := make([]string, 0, l)
	cur := s.l.Front()
	for i := 0; i < l; i++ {
		res = append(res, cur.Value.(string))
		cur = cur.Next()
	}

	return res
}

func (s OrderedSet) Equals(another OrderedSet) bool {
	return s.IsSubSet(another) && another.IsSubSet(s)
}
