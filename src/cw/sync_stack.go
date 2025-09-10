package cw

import (
	"sync/atomic"
	"unsafe"
)

type SyncStack[T any] struct {
	head *ListNode[T]
	cnt  int64
}

func NewSyncStack[T any]() *SyncStack[T] {
	ret := &SyncStack[T]{
		head: &ListNode[T]{},
	}
	return ret
}

func (l *SyncStack[T]) Push(val T) {
	if l == nil {
		return
	}
	newHead := &ListNode[T]{
		value: val,
	}
	l.push(newHead)
}

func (l *SyncStack[T]) push(node *ListNode[T]) {
	for {
		addr := unsafe.Pointer(&l.head.next)
		head := atomic.LoadPointer((*unsafe.Pointer)(addr))
		oldHead := (*ListNode[T])(head)
		node.next = oldHead

		if atomic.CompareAndSwapPointer((*unsafe.Pointer)(addr), head, unsafe.Pointer(node)) {
			break
		}
	}
	atomic.AddInt64(&l.cnt, 1)
}

func (l *SyncStack[T]) Pop() *ListNode[T] {
	if l == nil || atomic.LoadInt64(&l.cnt) == 0 {
		return nil
	}

	for {
		addr := unsafe.Pointer(&l.head.next)
		head := (*ListNode[T])(atomic.LoadPointer((*unsafe.Pointer)(addr)))
		if head == nil {
			return nil
		}
		next := (*ListNode[T])(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&head.next))))
		if atomic.CompareAndSwapPointer((*unsafe.Pointer)(addr), unsafe.Pointer(head), unsafe.Pointer(next)) {
			atomic.AddInt64(&l.cnt, -1)
			return head
		}
	}
}

func (l *SyncStack[T]) Top() *ListNode[T] {
	if l == nil || atomic.LoadInt64(&l.cnt) == 0 {
		return nil
	}
	addr := unsafe.Pointer(&l.head.next)
	ptr := atomic.LoadPointer((*unsafe.Pointer)(addr))
	return (*ListNode[T])(ptr)
}

func (l *SyncStack[T]) Size() int {
	if l == nil {
		return 0
	}
	return int(atomic.LoadInt64(&l.cnt))
}

func (l *SyncStack[T]) Len() int {
	if l == nil {
		return 0
	}
	return l.Size()
}

func (l *SyncStack[T]) Snapshot() *LinkedList[T] {
	if l == nil || atomic.LoadInt64(&l.cnt) == 0 {
		return nil
	}

	addr := unsafe.Pointer(&l.head.next)
	head := (*ListNode[T])(atomic.LoadPointer((*unsafe.Pointer)(addr)))
	ret := NewLinkedList[T]()
	for head != nil {
		ret.PushBack(head.value)
		head = head.next
	}
	return ret
}
