package typesw

import (
	"sync/atomic"
	"time"
)

type CountDownLatch struct {
	cnt  int32
	sema *Semaphore
}

func NewCountDownLatch(count int) *CountDownLatch {
	return &CountDownLatch{
		cnt:  int32(count),
		sema: NewSemaphore(-count + 1),
	}
}

func (l *CountDownLatch) CountDown() {
	l.sema.Release()
	atomic.AddInt32(&l.cnt, -1)
}

func (l *CountDownLatch) Wait() {
	l.sema.Acquire()
}

func (l *CountDownLatch) WaitTimeout(timeout time.Duration) bool {
	return l.sema.AcquireTimeout(timeout)
}

func (l *CountDownLatch) GetCount() int {
	return int(atomic.LoadInt32(&l.cnt))
}
