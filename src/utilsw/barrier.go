package utilsw

import (
	"sync"
	"sync/atomic"
	"time"
)

type Barrier struct {
	parties    int
	broken     atomic.Bool
	numWaiting atomic.Int32
	sema       *Semaphore

	action func()
	once   sync.Once
}

func NewBarrier(parties int, action func()) *Barrier {
	return &Barrier{
		parties: parties,
		sema:    NewSemaphore(0),
		action:  action,
		once:    sync.Once{},
	}

}

func (b *Barrier) Wait() {
	b.numWaiting.Add(1)
	if b.numWaiting.Load() == int32(b.parties) {
		b.sema.cond.Broadcast()
		if b.action != nil {
			b.once.Do(b.action)
		}
	}
	b.sema.Acquire()
	b.numWaiting.Add(-1)
}

func (b *Barrier) WaitTimeout(timeout time.Duration) bool {
	b.numWaiting.Add(1)
	waitSuccess := b.sema.AcquireTimeout(timeout)
	if !waitSuccess {
		b.broken.Store(true)
	}
	b.numWaiting.Add(-1)
	return waitSuccess
}

func (b *Barrier) Reset() {
	b.sema.cond.Broadcast()
	b.sema = NewSemaphore(0)
	b.broken.Store(false)
	b.numWaiting.Store(0)
	b.once = sync.Once{}
}

func (b *Barrier) GetParties() int {
	return b.parties
}

func (b *Barrier) GetNumWaiting() int {
	return int(b.numWaiting.Load())
}

func (b *Barrier) IsBroken() bool {
	return b.broken.Load()
}
