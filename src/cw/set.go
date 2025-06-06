package cw

import (
	"fmt"

	"github.com/grewwc/go_tools/src/typesw"
)

type Set struct {
	data map[interface{}]bool
}

func (s *Set) Add(item interface{}) {
	s.data[item] = true
}

func (s *Set) AddAll(items ...interface{}) {
	for _, item := range items {
		s.Add(item)
	}
}

func (s *Set) Contains(item interface{}) bool {
	if _, exist := s.data[item]; exist {
		return true
	}
	return false
}

func (s *Set) Iter() typesw.Iterable {
	return &interfaceKeyMapIterator[interface{}, bool]{
		data: s.data,
	}
}

func (s *Set) MutualExclude(another *Set) bool {
	for k := range s.data {
		if another.Contains(k) {
			return false
		}
	}
	return true
}

func (s *Set) Delete(item interface{}) bool {
	if s.Contains(item) {
		delete(s.data, item)
		return true
	}
	return false
}

func (s *Set) DeleteAll(items ...interface{}) {
	for _, item := range items {
		s.Delete(item)
	}
}

func (s *Set) Intersect(another *Set) *Set {
	result := NewSet()
	for k := range s.data {
		if another.Contains(k) {
			result.Add(k)
		}
	}
	return result
}

func (s *Set) Union(another *Set) *Set {
	result := NewSet()
	for k := range s.data {
		result.Add(k)
	}
	for k := range another.data {
		result.Add(k)
	}
	return result
}

func (s *Set) IsSuperSet(another *Set) bool {
	for k := range another.data {
		if !s.Contains(k) {
			return false
		}
	}
	return true
}

func (s *Set) IsSubSet(another *Set) bool {
	return another.IsSuperSet(s)
}

func (s *Set) Empty() bool {
	return len(s.data) == 0
}

func (s *Set) Size() int {
	return len(s.data)
}

func (s *Set) Clear() {
	s.data = make(map[interface{}]bool, 8)
}

func (s *Set) String() string {
	res := make([]interface{}, 0, len(s.data))
	for k := range s.data {
		res = append(res, k)
	}
	return fmt.Sprintf("%v\n", res)
}

func (s *Set) ShallowCopy() *Set {
	result := NewSet()
	for k := range s.data {
		result.Add(k)
	}
	return result
}

func (s *Set) Subtract(another Set) {
	for k := range another.data {
		s.Delete(k)
	}
}

func (s Set) ToSlice() []interface{} {
	res := make([]interface{}, 0, s.Size())
	for k := range s.data {
		res = append(res, k)
	}
	return res
}

// ToStringSlice is not type safe
func (s *Set) ToStringSlice() []string {
	res := make([]string, 0, s.Size())
	for k := range s.data {
		res = append(res, k.(string))
	}
	return res
}

func (s *Set) Equals(another *Set) bool {
	return s.IsSubSet(another) && another.IsSubSet(s)
}

func NewSet(items ...interface{}) *Set {
	s := Set{data: make(map[interface{}]bool, 8)}
	s.AddAll(items...)
	return &s
}
