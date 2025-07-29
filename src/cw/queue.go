package cw

import (
	"fmt"

	"github.com/grewwc/go_tools/src/typesw"
)

type Queue[T any] struct {
	data *LinkedList[T]
}

func NewQueue[T any](items ...T) *Queue[T] {
	res := &Queue[T]{NewLinkedList[T](items...)}
	return res
}

func (q *Queue[T]) Enqueue(item T) {
	q.data.PushBack(item)
}

func (q *Queue[T]) Front() T {
	if q.Empty() {
		return *new(T)
	}
	return q.data.Front().Value()
}

func (q *Queue[T]) Dequeue() T {
	if q.Empty() {
		return *new(T)
	}
	return q.data.PopFront().Value()
}

func (q *Queue[T]) Empty() bool {
	return q.Size() == 0
}

func (q *Queue[T]) Size() int {
	if q == nil {
		return 0
	}
	return q.data.Len()
}

func (q *Queue[T]) Iterate() typesw.IterableT[T] {
	if q.Empty() {
		return typesw.EmptyIterable[T]()
	}
	return typesw.FuncToIterable(func() chan T {
		ch := make(chan T)
		go func() {
			defer close(ch)
			for curr := q.data.Front(); curr != nil; curr = curr.Next() {
				ch <- curr.Value()
			}
		}()
		return ch
	})
}

func (q *Queue[T]) ToStringSlice() []string {
	res := make([]string, 0, q.data.Len())
	for s := range q.Iterate().Iterate() {
		res = append(res, fmt.Sprintf("%v", s))
	}
	return res
}

func (q *Queue[T]) ShallowCopy() *Queue[T] {
	res := NewQueue[T]()
	for item := range q.Iterate().Iterate() {
		res.Enqueue(item)
	}
	return res
}
