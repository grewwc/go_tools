package conw

import (
	"container/list"
)

type Queue struct {
	data *list.List
}

func NewQueue(items ...interface{}) *Queue {
	res := &Queue{list.New()}
	for _, item := range items {
		res.Enqueue(item)
	}
	return res
}

func (q *Queue) Enqueue(item interface{}) {
	q.data.PushBack(item)
}

func (q *Queue) Front() interface{} {
	front := q.data.Front()
	if front == nil {
		return nil
	}

	return front.Value
}

func (q *Queue) Dequeue() interface{} {
	front := q.Front()
	q.data.Remove(q.data.Front())
	return front
}

func (q *Queue) Empty() bool {
	return q.Size() == 0
}

func (q *Queue) Size() int {
	return q.data.Len()
}

func (q *Queue) Iterate() <-chan interface{} {
	ch := make(chan interface{})
	go func() {
		defer close(ch)
		for curr := q.data.Front(); curr != nil; curr = curr.Next() {
			ch <- curr.Value
		}
	}()
	return ch
}

func (q *Queue) ToStringSlice() []string {
	res := make([]string, 0, q.data.Len())
	for s := range q.Iterate() {
		res = append(res, s.(string))
	}
	return res
}

func (q *Queue) ShallowCopy() *Queue {
	res := NewQueue()
	for item := range q.Iterate() {
		if item != nil {
			res.Enqueue(item)
		}
	}
	return res
}
