package cw

import (
	"github.com/grewwc/go_tools/src/typesw"
)

/*Stack is not thread safe*/
type Stack[T any] struct {
	data *LinkedList[T]
}

func NewStack[T any](vals ...T) *Stack[T] {
	data := NewLinkedList[T]()
	res := &Stack[T]{data}
	for _, val := range vals {
		res.Push(val)
	}
	return res
}

func (s *Stack[T]) Push(item T) {
	s.data.PushFront(item)
}

func (s *Stack[T]) Pop() T {
	result := s.data.PopFront()
	return result.Value()
}
func (s *Stack[T]) Top() T {
	if s.Empty() {
		return *new(T)
	}
	return s.data.Front().Value()
}

func (s *Stack[T]) Empty() bool {
	return s.Size() == 0
}

func (s *Stack[T]) Size() int {
	return s.data.Len()
}

func (s *Stack[T]) Len() int {
	return s.Size()
}

func (s *Stack[T]) Iter() typesw.IterableT[T] {
	return typesw.FuncToIterable(func() chan T {
		ch := make(chan T)
		go func() {
			defer close(ch)
			if s.Empty() {
				return
			}
			for curr := s.data.Front(); curr != nil; curr = curr.Next() {
				ch <- curr.Value()
			}
		}()
		return ch
	})
}
