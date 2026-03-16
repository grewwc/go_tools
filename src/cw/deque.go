package cw

import (
	"container/list"
	"fmt"
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
	front, ok := dq.TryFront()
	if !ok {
		panic("empty deque")
	}
	return front
}

func (dq *Deque) TryFront() (interface{}, bool) {
	front := dq.data.Front()
	if front == nil {
		return nil, false
	}
	return front.Value, true
}

func (dq *Deque) PopFront() interface{} {
	front := dq.Front()
	dq.data.Remove(dq.data.Front())
	return front
}

func (dq *Deque) TryPopFront() (interface{}, bool) {
	front, ok := dq.TryFront()
	if !ok {
		return nil, false
	}
	dq.data.Remove(dq.data.Front())
	return front, true
}

func (dq *Deque) Back() interface{} {
	back, ok := dq.TryBack()
	if !ok {
		panic("empty deque")
	}
	return back
}

func (dq *Deque) TryBack() (interface{}, bool) {
	back := dq.data.Back()
	if back == nil {
		return nil, false
	}
	return back.Value, true
}

func (dq *Deque) PopBack() interface{} {
	front := dq.Back()
	dq.data.Remove(dq.data.Back())
	return front
}

func (dq *Deque) TryPopBack() (interface{}, bool) {
	back, ok := dq.TryBack()
	if !ok {
		return nil, false
	}
	dq.data.Remove(dq.data.Back())
	return back, true
}

func (dq *Deque) Empty() bool {
	return dq.Size() == 0
}

func (dq *Deque) Size() int {
	return dq.data.Len()
}

func (dq *Deque) Len() int {
	return dq.Size()
}

func (dq *Deque) ToSlice() []interface{} {
	res := make([]interface{}, 0, dq.Size())
	for l := dq.data.Front(); l != nil; l = l.Next() {
		res = append(res, l.Value)
	}
	return res
}

func (dq *Deque) ToStringSlice() []string {
	res := make([]string, 0, dq.Size())
	for l := dq.data.Front(); l != nil; l = l.Next() {
		res = append(res, fmt.Sprintf("%v", l.Value))
	}
	return res
}
