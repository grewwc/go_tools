package cw

import (
	"fmt"

	"github.com/grewwc/go_tools/src/typesw"
)

type ListNode[T any] struct {
	value T
	next  *ListNode[T]
}

func (n *ListNode[T]) Next() *ListNode[T] {
	return n.next
}

func (n *ListNode[T]) Value() T {
	return n.value
}

// LinkedList is single LinkedList, so some operations are not supported.
type LinkedList[T any] struct {
	head, tail *ListNode[T]
	size       int
}

func NewLinkedList[T any](vals ...T) *LinkedList[T] {
	res := &LinkedList[T]{}
	for _, val := range vals {
		res.PushBack(val)
	}
	return res
}

func (l *LinkedList[T]) PushFront(val T) *ListNode[T] {
	if l == nil {
		return nil
	}
	node := &ListNode[T]{value: val}
	l.size++
	if l.head == nil {
		l.head = node
		l.tail = node
		return node
	}
	node.next = l.head
	l.head = node
	return node
}

func (l *LinkedList[T]) PushBack(val T) *ListNode[T] {
	if l == nil {
		return nil
	}
	node := &ListNode[T]{value: val}
	l.size++
	if l.tail == nil {
		l.head = node
		l.tail = node
		return node
	}
	l.tail.next = node
	l.tail = node
	return node
}

func (l *LinkedList[T]) PopFront() *ListNode[T] {
	if l.Empty() {
		return nil
	}
	l.size--
	front := l.head
	l.head = front.next
	if l.tail == front {
		l.tail = nil
	}
	return front
}

func (l *LinkedList[T]) Merge(mergePoint *ListNode[T], other *LinkedList[T]) *LinkedList[T] {
	if mergePoint == nil || other.Empty() {
		return l
	}
	if l.Empty() {
		return other
	}
	next := mergePoint.next
	mergePoint.next = other.Front()
	other.tail.next = next
	l.tail = other.tail
	other.Clear()
	l.size += other.size
	return l
}

func (l *LinkedList[T]) Reverse() *LinkedList[T] {
	if l.Empty() {
		return l
	}
	head, tail := l.head, l.tail
	var prev *ListNode[T] = nil
	for curr := l.head; curr != nil; {
		next := curr.next
		curr.next = prev
		prev = curr
		curr = next
	}
	l.head = tail
	l.tail = head
	return l
}

func (l *LinkedList[T]) Empty() bool {
	return l == nil || l.size == 0
}

func (l *LinkedList[T]) Len() int {
	if l == nil {
		return 0
	}
	return l.size
}

func (l *LinkedList[T]) Iter() typesw.IterableT[*ListNode[T]] {
	return typesw.FuncToIterable(func() chan *ListNode[T] {
		ch := make(chan *ListNode[T])
		go func() {
			defer close(ch)
			if l.Empty() {
				return
			}
			for curr := l.Front(); curr != nil; curr = curr.Next() {
				ch <- curr
			}
		}()
		return ch
	})
}

func (l *LinkedList[T]) Delete(val T, cmp typesw.CompareFunc[T]) bool {
	if l.Empty() {
		return false
	}
	if cmp == nil {
		cmp = typesw.CreateDefaultCmp[T]()
	}
	dummy := &ListNode[T]{
		next: l.head,
	}
	curr := dummy
	for ; curr != nil && curr.Next() != nil && cmp(curr.Next().Value(), val) != 0; curr = curr.Next() {
	}
	if curr == nil || curr.Next() == nil {
		return false
	}
	d := curr.next
	curr.next = d.next
	if d == l.tail {
		l.tail = curr
	}
	if d == l.head {
		l.head = d.next
	}
	return true
}

func (l *LinkedList[T]) Remove(node *ListNode[T]) T {
	// empty
	if l.head == nil {
		return node.value
	}
	if node == nil {
		return *new(T)
	}
	dummy := &ListNode[T]{next: l.head}
	curr := dummy
	for ; curr != nil && curr.next != node; curr = curr.next {
	}
	// assert curr.next == node || curr == nil

	// node not found
	if curr == nil {
		return node.value
	}
	d := curr.next
	curr.next = curr.next.next
	if d == l.tail {
		l.tail = curr
	}
	if d == l.head {
		l.head = curr.next
	}
	l.size--
	return d.value
}

func (l *LinkedList[T]) Front() *ListNode[T] {
	if l == nil {
		return nil
	}
	return l.head
}

func (l *LinkedList[T]) Back() *ListNode[T] {
	if l == nil {
		return nil
	}
	return l.tail
}

func (l *LinkedList[T]) ShallowCopy() *LinkedList[T] {
	if l == nil {
		return nil
	}
	res := NewLinkedList[T]()
	for node := range l.Iter().Iterate() {
		res.PushBack(node.Value())
	}
	return res
}

func (l *LinkedList[T]) Clear() {
	l.size = 0
	l.head = nil
	l.tail = nil
}

func (l *LinkedList[T]) Equals(other *LinkedList[T], cmp typesw.CompareFunc[T]) bool {
	if l.Len() != other.Len() {
		return false
	}
	if cmp == nil {
		cmp = typesw.CreateDefaultCmp[T]()
	}
	if cmp == nil {
		return false
	}
	for t := range Zip(l.Iter(), other.Iter()).Iterate() {
		if cmp(t.Get(0).(*ListNode[T]).Value(), t.Get(1).(*ListNode[T]).Value()) != 0 {
			return false
		}
	}
	return true
}

func (l *LinkedList[T]) Contains(val T, cmp typesw.CompareFunc[T]) bool {
	if l.Empty() {
		return false
	}
	for node := range l.Iter().Iterate() {
		if cmp(node.Value(), val) == 0 {
			return true
		}
	}
	return false
}

func (l *LinkedList[T]) ToStringSlice() []string {
	if l == nil {
		return nil
	}
	res := make([]string, 0, l.Len())
	if l.Empty() {
		return res
	}
	for node := range l.Iter().Iterate() {
		res = append(res, fmt.Sprintf("%v", node.Value()))
	}
	return res
}
