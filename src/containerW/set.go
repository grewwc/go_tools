package containerW

import "fmt"

type Set struct {
	data map[interface{}]bool
}

func (s *Set) Add(item interface{}) {
	s.data[item] = true
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

func NewSet() *Set {
	s := Set{data: make(map[interface{}]bool, 8)}
	return &s
}
