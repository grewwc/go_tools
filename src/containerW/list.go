package containerW

import (
	"container/list"
)

type LinkedList struct {
	data *list.List
}

func NewLinkedList(items ...interface{}) *LinkedList {
	res := &LinkedList{list.New()}
	for _, item := range items {
		res.Add(item)
	}
	return res
}

func (q *LinkedList) Add(item interface{}) {
	q.data.PushBack(item)
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
	return q.data.Len()
}

func (q *LinkedList) Len() int {
	return q.Size()
}

func (q *LinkedList) Iterate() <-chan interface{} {
	ch := make(chan interface{})
	go func() {
		defer close(ch)
		for curr := q.data.Front(); curr != nil; curr = curr.Next() {
			ch <- curr.Value
		}
	}()
	return ch
}

func (q *LinkedList) ToStringSlice() []string {
	res := make([]string, 0, q.data.Len())
	for s := range q.Iterate() {
		res = append(res, s.(string))
	}
	return res
}

func (q *LinkedList) ShallowCopy() *LinkedList {
	res := NewLinkedList()
	for item := range q.Iterate() {
		res.Add(item)
	}
	return res
}

func (q *LinkedList) Delete(item interface{}) bool {
	var removed bool
	for curr := q.data.Front(); curr != nil; curr = curr.Next() {
		if curr.Value == item {
			q.data.Remove(curr)
			removed = true
		}
	}
	return removed
}

func (q *LinkedList) Contains(target interface{}) bool {
	for item := range q.Iterate() {
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
	if q.data.Len() != other.data.Len() {
		return false
	}
	ca, cb := q.data.Front(), other.data.Front()
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
	curr := q.data.Front()
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
	return q.data.Remove(elem.(*list.Element))
}
