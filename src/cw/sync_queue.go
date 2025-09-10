package cw

type SyncQueue[T any] struct {
	front, back *SyncStack[T]
}

func NewSyncQueue[T any]() *SyncQueue[T] {
	ret := &SyncQueue[T]{
		front: NewSyncStack[T](),
		back:  NewSyncStack[T](),
	}
	return ret
}

func (l *SyncQueue[T]) Enqueue(val T) {
	if l == nil {
		return
	}
	l.front.Push(val)
}

func (l *SyncQueue[T]) Dequeue() *ListNode[T] {
	if l == nil {
		return nil
	}
	for l.front.Len() > 0 {
		l.back.push(l.front.Pop())
	}
	return l.back.Pop()
}

func (l *SyncQueue[T]) Front() *ListNode[T] {
	if l == nil {
		return nil
	}
	return l.front.Top()
}

func (l *SyncQueue[T]) Size() int {
	if l == nil {
		return 0
	}
	return l.front.Len() + l.back.Len()
}

func (l *SyncQueue[T]) Len() int {
	if l == nil {
		return 0
	}
	return l.Size()
}

func (l *SyncQueue[T]) Snapshot() *LinkedList[T] {
	if l == nil {
		return nil
	}

	ret := l.front.Snapshot()
	ret.Reverse()
	l2 := l.back.Snapshot()
	ret.Merge(ret.tail, l2)
	return ret
}
