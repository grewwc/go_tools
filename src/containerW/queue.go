package containerW

import "container/list"

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
		panic("empty queue")
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
