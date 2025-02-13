package utilsW

import (
	"math"
	"sync"
	"time"
)

type RateLimiter struct {
	// constant
	rate             float64
	unit             time.Duration
	capacity         float64
	_rate_to_nanosec float64

	tokens       float64
	lastModified time.Time
	mu           *sync.RWMutex
}

func (r *RateLimiter) update() {
	r.mu.Lock()
	defer r.mu.Unlock()
	curr := time.Now()
	sec := float64(time.Since(r.lastModified).Nanoseconds()) / r._rate_to_nanosec
	accumulated := sec * r.rate
	r.tokens = math.Min(accumulated+r.tokens, r.capacity)
	r.lastModified = curr
}

func (r *RateLimiter) Acquire() bool {
	r.update()
	r.mu.RLock()
	if r.tokens >= 1 {
		r.mu.RUnlock()
		r.mu.Lock()
		r.tokens--
		r.lastModified = time.Now()
		r.mu.Unlock()
		return true
	}
	needWaitNanoSec := int64((1 - r.tokens) / r.rate * r._rate_to_nanosec)
	r.mu.RUnlock()
	time.Sleep(time.Nanosecond * time.Duration(needWaitNanoSec))
	return r.Acquire()
}

func (r *RateLimiter) AcquireTimeout(timeout time.Duration) bool {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() bool {
		defer wg.Done()
		return r.Acquire()
	}()
	if err := TimeoutWait(&wg, timeout); err != nil {
		return false
	}
	return true
}

func NewRateLimiter(capacity, rate float64, unit time.Duration) *RateLimiter {
	return &RateLimiter{
		mu:               &sync.RWMutex{},
		capacity:         capacity,
		rate:             rate,
		unit:             unit,
		lastModified:     time.Now(),
		_rate_to_nanosec: float64(unit.Nanoseconds()) / float64(time.Nanosecond.Nanoseconds()),
	}
}
