package cw

import (
	"container/list"

	"github.com/grewwc/go_tools/src/typesw"
)

type LinkedList struct {
	*list.List
}

func NewLinkedList(items ...interface{}) *LinkedList {
	res := &LinkedList{list.New()}
	for _, item := range items {
		res.Add(item)
	}
	return res
}

func (q *LinkedList) Add(item interface{}) {
	q.PushBack(item)
}

func (q *LinkedList) AddAll(items ...interface{}) {
	for _, item := range items {
		q.Add(item)
	}
}

func (q *LinkedList) Empty() bool {
	return q.Size() == 0
}

func (q *LinkedList) Size() int {
	return q.Len()
}

func (q *LinkedList) Len() int {
	return q.Size()
}

func (q *LinkedList) Iter() typesw.Iterable {
	return &listIterator[any]{data: q.List}
}

func (q *LinkedList) ToStringSlice() []string {
	res := make([]string, 0, q.Len())
	for s := range q.Iter().Iterate() {
		res = append(res, s.(string))
	}
	return res
}

func (q *LinkedList) ShallowCopy() *LinkedList {
	res := NewLinkedList()
	for item := range q.Iter().Iterate() {
		res.Add(item)
	}
	return res
}

func (q *LinkedList) Delete(item interface{}) bool {
	var removed bool
	for curr := q.Front(); curr != nil; curr = curr.Next() {
		if curr.Value == item {
			q.List.Remove(curr)
			removed = true
		}
	}
	return removed
}

func (q *LinkedList) Contains(target interface{}) bool {
	for item := range q.Iter().Iterate() {
		if item == target {
			return true
		}
	}
	return false
}

func (q *LinkedList) Equals(other *LinkedList) bool {
	if other == nil {
		return false
	}
	if q.Len() != other.Len() {
		return false
	}
	ca, cb := q.Front(), other.Front()
	for i := 0; i < q.Size(); i++ {
		if ca.Value != cb.Value {
			return false
		}
		ca = ca.Next()
		cb = cb.Next()
	}
	return true
}

func (q *LinkedList) get(idx int) interface{} {
	curr := q.Front()
	for i := 0; i < idx; i++ {
		curr = curr.Next()
		if curr == nil {
			return nil
		}
	}
	return curr
}

func (q *LinkedList) Get(idx int) interface{} {
	elem := q.get(idx)
	if elem == nil {
		return nil
	}
	return elem.(*list.Element).Value
}

func (q *LinkedList) Set(idx int, value interface{}) interface{} {
	elem := q.get(idx)
	if elem == nil {
		return nil
	}
	elemT := elem.(*list.Element)
	prev := elemT.Value
	elemT.Value = value
	return prev
}

func (q *LinkedList) Remove(idx int) interface{} {
	elem := q.get(idx)
	if elem == nil {
		return nil
	}
	return q.List.Remove(elem.(*list.Element))
}
