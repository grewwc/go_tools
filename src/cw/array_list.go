package cw

import "github.com/grewwc/go_tools/src/typesw"

type ArrayList struct {
	data []interface{}
}

func NewArrayList(items ...interface{}) *ArrayList {
	l := len(items)
	data := make([]interface{}, 0, l)
	data = append(data, items...)
	return &ArrayList{
		data: data,
	}
}

func (q *ArrayList) Add(item interface{}) {
	q.data = append(q.data, item)
}

func (q *ArrayList) AddAll(items ...interface{}) {
	for _, item := range items {
		q.Add(item)
	}
}

func (q *ArrayList) Empty() bool {
	return q.Size() == 0
}

func (q *ArrayList) Size() int {
	return len(q.data)
}

func (q *ArrayList) Len() int {
	return q.Size()
}

func (q *ArrayList) Iter() typesw.Iterable {
	return &sliceIterator[interface{}]{
		data: q.data,
	}
}

func (q *ArrayList) ToStringSlice() []string {
	res := make([]string, 0, q.Len())
	for s := range q.Iter().Iterate() {
		res = append(res, s.(string))
	}
	return res
}

func (q *ArrayList) ShallowCopy() typesw.IList {
	res := NewArrayList()
	for item := range q.Iter().Iterate() {
		res.Add(item)
	}
	return res
}

func (q *ArrayList) Delete(item interface{}) bool {
	var removed bool
	for i, e := range q.data {
		if e == item {
			removed = true
			q.data = append(q.data[:i], q.data[i+1:]...)
		}
	}
	return removed
}

func (q *ArrayList) Contains(target interface{}) bool {
	for item := range q.Iter().Iterate() {
		if item == target {
			return true
		}
	}
	return false
}

func (q *ArrayList) Equals(other typesw.IList) bool {
	if other == nil {
		return false
	}
	if q.Len() != other.Len() {
		return false
	}
	for i := 0; i < q.Size(); i++ {
		if q.Get(i) != other.Get(i) {
			return false
		}
	}
	return true
}

func (q *ArrayList) Get(idx int) interface{} {
	if idx >= len(q.data) {
		return nil
	}
	return q.data[idx]
}

func (q *ArrayList) Set(idx int, value interface{}) interface{} {
	if idx >= len(q.data) {
		return nil
	}
	prev := q.Get(idx)
	q.data[idx] = value
	return prev
}

func (q *ArrayList) Remove(idx int) interface{} {
	if idx >= len(q.data) || idx < 0 {
		return nil
	}
	return append(q.data[:idx], q.data[idx+1:]...)
}
