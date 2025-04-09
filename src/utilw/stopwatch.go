package utilw

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/grewwc/go_tools/src/conw"
)

// StopWatch is "auto-reset" stopwatch
type StopWatch struct {
	mu      *sync.RWMutex
	start   int64
	records *conw.OrderedMap
}

func NewStopWatch() *StopWatch {
	return &StopWatch{
		records: conw.NewOrderedMap(),
		start:   time.Now().UnixNano(),
		mu:      &sync.RWMutex{},
	}
}

func (s *StopWatch) curr() int64 {
	return time.Now().UnixNano()
}

func (s *StopWatch) Record() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.records.Put("", s.curr())
}

func (s *StopWatch) RecordLabel(label string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.records.Put(label, s.curr())
}

func (s *StopWatch) NumRecords() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.records.Size()
}

func (s *StopWatch) Nanos() float64 {
	return float64(s.curr() - atomic.LoadInt64(&s.start))
}

func (s *StopWatch) Micros() float64 {
	return float64(s.curr()-atomic.LoadInt64(&s.start)) / 1e3
}

func (s *StopWatch) Mills() float64 {
	return s.Micros() / 1e3
}

func (s *StopWatch) Seconds() float64 {
	return s.Mills() / 1e3
}

func (s *StopWatch) Minutes() float64 {
	return s.Seconds() / 60.0
}

func (s *StopWatch) Hours() float64 {
	return s.Minutes() / 60
}

func (s *StopWatch) Days() float64 {
	return s.Hours() / 24.0
}

func (s *StopWatch) Tell() {
	fmt.Println(format(float64(s.curr() - atomic.LoadInt64(&s.start))))
}

func format(cost float64) string {
	unit := "ns"
	modify := false
	if cost > 1e3 {
		modify = true
		cost /= 1e3
		unit = "us"
	}
	if modify && cost > 1e3 {
		cost /= 1e3
		unit = "ms"
		modify = true
	} else {
		modify = false
	}

	if modify && cost > 1e3 {
		cost /= 1e3
		unit = "s"
	} else {
		modify = false
	}
	if modify && cost > 60 {
		cost /= 60
		unit = "min"
	} else {
		modify = false
	}
	if modify && cost > 24 {
		cost /= 24
		unit = "hour"
	} else {
		modify = false
	}
	return fmt.Sprintf("%.3f (%s)", cost, unit)
}

func (s *StopWatch) TellAll() {
	start := atomic.LoadInt64(&s.start)
	s.mu.RLock()
	defer s.mu.RUnlock()
	for entry := range s.records.Iterate() {
		fmt.Printf("%s: %s\n", entry.Key().(string), format(float64(entry.Val().(int64)-start)))
	}
	fmt.Println(format(float64(s.curr() - start)))
}

func (s *StopWatch) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.records.Clear()
	s.start = s.curr()
}
