package utilsw

import (
	"context"
	"sync"
	"time"
)

type Semaphore struct {
	cond  *sync.Cond
	mu    *sync.Mutex
	Count int
}

func NewSemaphore(cnt int) *Semaphore {
	mu := &sync.Mutex{}
	sem := &Semaphore{
		mu:    mu,
		cond:  sync.NewCond(mu),
		Count: cnt,
	}
	return sem
}

func (s *Semaphore) Acquire() {
	s.mu.Lock()
	for s.Count <= 0 {
		s.cond.Wait()
	}
	s.Count--
	s.mu.Unlock()
}

func (s *Semaphore) AcquireTimeout(timeout time.Duration) bool {
	s.mu.Lock()
	if s.Count > 0 {
		s.Count--
		s.mu.Unlock()
		return true
	}
	s.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	done := make(chan struct{})
	timeoutCh := make(chan struct{}, 1)
	timeoutCh <- struct{}{}
	go func() {
		s.mu.Lock()
		for s.Count <= 0 {
			s.cond.Wait()
		}
		s.Count--
		if _, ok := <-timeoutCh; !ok {
			s.cond.Signal()
		}
		s.mu.Unlock()
		close(done)
	}()
	var aquired bool
	select {
	case <-ctx.Done():
		aquired = false
	case <-done:
		aquired = true
	}
	close(timeoutCh)
	return aquired

}

func (s *Semaphore) Release() {
	s.mu.Lock()
	// fmt.Println("before,", s.Count)
	s.Count++
	// fmt.Println("after,", s.Count)
	s.cond.Signal()
	s.mu.Unlock()
}
