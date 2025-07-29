package cw

import (
	"github.com/grewwc/go_tools/src/typesw"
)

/*Stack is not thread safe*/
type Stack[T any] struct {
	data *LinkedList[T]
}

func NewStack[T any]() *Stack[T] {
	data := NewLinkedList[T]()
	return &Stack[T]{data}
}

func (s *Stack[T]) Push(item T) {
	s.data.PushFront(item)
}

func (s *Stack[T]) Pop() T {
	result := s.Top()
	return result
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
