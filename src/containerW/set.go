package containerW

import "fmt"

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

func (s Set) Intersect(another Set) *Set {
	result := NewSet()
	for k := range s.data {
		if another.Contains(k) {
			result.Add(k)
		}
	}
	return result
}

func (s Set) Union(another Set) *Set {
	result := NewSet()
	for k := range s.data {
		result.Add(k)
	}
	return result
}

func (s Set) IsSuperSet(another Set) bool {
	for k := range s.data {
		if !another.Contains(k) {
			return false
		}
	}
	return true
}

func (s Set) IsSubSet(another Set) bool {
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

func (s Set) ShallowCopy() *Set {
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

func NewSet() *Set {
	s := Set{data: make(map[interface{}]bool, 8)}
	return &s
}
