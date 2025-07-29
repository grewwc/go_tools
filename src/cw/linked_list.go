package cw

import (
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

func (l *LinkedList[T]) Remove(node *ListNode[T]) T {
	// empty
	if l.head == nil {
		return node.value
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
	next := curr.next
	curr.next = curr.next.next
	next.next = nil
	if node == l.tail {
		l.tail = curr
	}
	l.size--
	return next.value
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

func (l *LinkedList[T]) Clear() {
	l.size = 0
	l.head = nil
	l.tail = nil
}
