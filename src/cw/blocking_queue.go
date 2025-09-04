package cw

import (
	"context"
	"sync"
	"time"
)

type BlockingQueue[T any] struct {
	cap        int
	start, end int
	len        int
	data       []T
	lock       *sync.Mutex
	notFull    *sync.Cond
	notEmpty   *sync.Cond
}

func NewBlockingQueue[T any](cap int) *BlockingQueue[T] {
	q := &BlockingQueue[T]{
		cap:  cap,
		data: make([]T, cap),
		lock: &sync.Mutex{},
	}

	q.notFull = sync.NewCond(q.lock)
	q.notEmpty = sync.NewCond(q.lock)
	return q

}

func (q *BlockingQueue[T]) AddFirst(val T) {
	q.lock.Lock()
	for q.len == q.cap {
		q.notFull.Wait()
	}
	q.data[q.start] = val
	q.start--
	q.len++
	if q.start < 0 {
		q.start += q.cap
	}
	q.notEmpty.Signal()
	q.lock.Unlock()
}

func (q BlockingQueue[T]) OfferFirst(val T, timeout time.Duration) bool {
	return withTimeout(func() { q.AddFirst(val) }, timeout)
}

func (q *BlockingQueue[T]) AddLast(val T) {
	q.lock.Lock()

	for q.len == q.cap {
		q.notFull.Wait()
	}
	q.end++
	q.len++
	if q.end >= q.cap {
		q.end %= q.cap
	}
	q.data[q.end] = val
	q.notEmpty.Signal()
	q.lock.Unlock()
}

func (q *BlockingQueue[T]) OfferLast(val T, timeout time.Duration) bool {
	return withTimeout(func() { q.AddLast(val) }, timeout)
}

func (q *BlockingQueue[T]) PopFirst() T {
	q.lock.Lock()
	for q.len == 0 {
		q.notEmpty.Wait()
	}
	q.start++
	q.len--
	if q.start >= q.cap {
		q.start %= q.cap
	}
	res := q.data[q.start]
	q.notFull.Signal()
	q.lock.Unlock()
	return res
}

func (q *BlockingQueue[T]) PopLast() T {
	q.lock.Lock()

	for q.len == 0 {
		q.notEmpty.Wait()
	}
	res := q.data[q.end]
	q.end--
	q.len--
	if q.end < 0 {
		q.end += q.cap
	}
	q.notFull.Signal()
	q.lock.Unlock()
	return res
}

func (q *BlockingQueue[T]) PollFirst(timeout time.Duration) (T, bool) {
	ch := make(chan T)
	defer close(ch)
	waitSuccess := withTimeout(func() {
		ch <- q.PopFirst()
	}, timeout)
	if waitSuccess {
		return <-ch, true
	}
	return *new(T), false
}

func (q *BlockingQueue[T]) PollLast(timeout time.Duration) (T, bool) {
	ch := make(chan T)
	defer close(ch)
	waitSuccess := withTimeout(func() {
		ch <- q.PopLast()
	}, timeout)
	if waitSuccess {
		return <-ch, true
	}
	return *new(T), false
}

func (q *BlockingQueue[T]) Snapshot() *Queue[T] {
	q.lock.Lock()
	res := NewQueue[T]()
	for i := 0; i < q.len; i++ {
		q.start++
		if q.start >= q.cap {
			q.start %= q.cap
		}
		res.Enqueue(q.data[q.start])
	}
	q.lock.Unlock()
	return res
}

func (q *BlockingQueue[T]) IsEmpty() bool {
	q.lock.Lock()
	defer q.lock.Unlock()
	return q.len == q.cap
}

func withTimeout(f func(), timeout time.Duration) bool {
	ch := make(chan struct{})
	go func() {
		f()
		close(ch)
	}()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	select {
	case <-ctx.Done():
		return false
	case <-ch:
		return true
	}
}
