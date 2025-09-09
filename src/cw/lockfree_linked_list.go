package cw

import (
	"sync/atomic"
	"unsafe"

	optional "github.com/grewwc/go_tools/src/optionalw"
)

type doubleLinkNode[T any] struct {
	val        T
	prev, next *doubleLinkNode[T]
}

type SyncLinkedList[T any] struct {
	head, tail *doubleLinkNode[T]
	cnt        int64
}

func NewSyncLinkedlist[T any]() *SyncLinkedList[T] {
	ret := &SyncLinkedList[T]{
		head: &doubleLinkNode[T]{},
		tail: &doubleLinkNode[T]{},
	}
	ret.head.next = ret.tail
	ret.tail.prev = ret.head
	return ret
}

func (l *SyncLinkedList[T]) PushFront(val T) {
	if l == nil {
		return
	}
	newHead := &doubleLinkNode[T]{
		val: val,
	}
	for {
		addr := unsafe.Pointer(&l.head.next)
		head := atomic.LoadPointer((*unsafe.Pointer)(addr))
		oldHead := (*doubleLinkNode[T])(head)
		// newHead.prev = l.head
		newHead.next = oldHead

		if atomic.CompareAndSwapPointer((*unsafe.Pointer)(addr), head, unsafe.Pointer(newHead)) {
			// oldHead.prev = newHead
			atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&oldHead.prev)), unsafe.Pointer(newHead))
			break
		}
	}
	atomic.AddInt64(&l.cnt, 1)
}

func (l *SyncLinkedList[T]) PushBack(val T) {
	if l == nil {
		return
	}
	newTail := &doubleLinkNode[T]{
		val: val,
	}
	for {
		addr := unsafe.Pointer(&l.tail.prev)
		tail := atomic.LoadPointer((*unsafe.Pointer)(addr))
		oldTail := (*doubleLinkNode[T])(tail)
		newTail.next = l.tail
		newTail.prev = oldTail

		if atomic.CompareAndSwapPointer((*unsafe.Pointer)(addr), tail, unsafe.Pointer(newTail)) {
			// oldTail.next = newTail
			atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&oldTail.next)), unsafe.Pointer(newTail))
			break
		}
	}
	atomic.AddInt64(&l.cnt, 1)
}

func (l *SyncLinkedList[T]) PopFront() *optional.Optional[T] {
	if l == nil || atomic.LoadInt64(&l.cnt) == 0 {
		return nil
	}

	for {
		addr := unsafe.Pointer(&l.head.next)
		head := (*doubleLinkNode[T])(atomic.LoadPointer((*unsafe.Pointer)(addr)))
		next := (*doubleLinkNode[T])(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&head.next))))
		if atomic.CompareAndSwapPointer((*unsafe.Pointer)(addr), unsafe.Pointer(head), unsafe.Pointer(next)) {
			atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&next.prev)), unsafe.Pointer(head))
			atomic.AddInt64(&l.cnt, -1)
			return optional.Of(head.val)
		}
	}
}

func (l *SyncLinkedList[T]) PopBack() *optional.Optional[T] {
	if l == nil || atomic.LoadInt64(&l.cnt) == 0 {
		return nil
	}
	for {
		addr := unsafe.Pointer(&l.tail.prev)
		tail := (*doubleLinkNode[T])(atomic.LoadPointer((*unsafe.Pointer)(addr)))
		prev := (*doubleLinkNode[T])(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&tail.prev))))
		if atomic.CompareAndSwapPointer((*unsafe.Pointer)(addr), unsafe.Pointer(tail), unsafe.Pointer(prev)) {
			atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&tail.next)), unsafe.Pointer(tail))
			atomic.AddInt64(&l.cnt, -1)
			return optional.Of(tail.val)
		}
	}
}

func (l *SyncLinkedList[T]) PeekFront() *optional.Optional[T] {
	if l == nil || atomic.LoadInt64(&l.cnt) == 0 {
		return nil
	}
	addr := unsafe.Pointer(&l.head.next)
	ptr := atomic.LoadPointer((*unsafe.Pointer)(addr))
	if ptr == nil {
		return nil
	}
	return optional.Of((*doubleLinkNode[T])(ptr).val)
}

func (l *SyncLinkedList[T]) PeekBack() *optional.Optional[T] {
	if l == nil || atomic.LoadInt64(&l.cnt) == 0 {
		return nil
	}
	addr := unsafe.Pointer(&l.tail.prev)
	ptr := atomic.LoadPointer((*unsafe.Pointer)(addr))
	if ptr == nil {
		return nil
	}
	return optional.Of((*doubleLinkNode[T])(ptr).val)
}

func (l *SyncLinkedList[T]) Size() int {
	if l == nil {
		return 0
	}
	return int(atomic.LoadInt64(&l.cnt))
}

func (l *SyncLinkedList[T]) Len() int {
	if l == nil {
		return 0
	}
	return l.Size()
}

func (l *SyncLinkedList[T]) Snapshot() *LinkedList[T] {
	if l == nil || atomic.LoadInt64(&l.cnt) == 0 {
		return nil
	}

	addr := unsafe.Pointer(&l.head.next)
	head := (*doubleLinkNode[T])(atomic.LoadPointer((*unsafe.Pointer)(addr)))
	ret := NewLinkedList[T]()
	for head != nil && head != l.tail {
		ret.PushBack(head.val)
		head = head.next
	}
	return ret
}
