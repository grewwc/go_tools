package conw

import (
	"container/list"
	"log"
)

type Deque struct {
	data *list.List
}

func NewDeque() *Deque {
	return &Deque{list.New()}
}

func (dq *Deque) PushFront(item interface{}) {
	dq.data.PushFront(item)
}

func (dq *Deque) PushBack(item interface{}) {
	dq.data.PushBack(item)
}

func (dq *Deque) Front() interface{} {
	front := dq.data.Front()
	if front == nil {
		log.Fatalln("empty deque")
	}
	return front.Value
}

func (dq *Deque) PopFront() interface{} {
	front := dq.Front()
	dq.data.Remove(dq.data.Front())
	return front
}

func (dq *Deque) Back() interface{} {
	front := dq.data.Back()
	if front == nil {
		panic("empty deque")
	}
	return front.Value
}

func (dq *Deque) PopBack() interface{} {
	front := dq.Back()
	dq.data.Remove(dq.data.Back())
	return front
}

func (dq *Deque) Empty() bool {
	return dq.Size() == 0
}

func (dq *Deque) Size() int {
	return dq.data.Len()
}

func (dq *Deque) ToSlice() []interface{} {
	l := dq.data.Front()
	res := make([]interface{}, 0, dq.Size())
	for l.Next() != nil {
		res = append(res, l.Value)
	}
	return res
}

func (dq *Deque) ToStringSlice() []string {
	l := dq.data.Front()
	res := make([]string, 0, dq.Size())
	for l.Next() != nil {
		res = append(res, l.Value.(string))
	}
	return res
}
