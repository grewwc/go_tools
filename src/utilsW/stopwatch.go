package utilsW

import (
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/grewwc/go_tools/src/containerW"
)

// StopWatch is "auto-reset" stopwatch
type StopWatch struct {
	start   int64
	records *containerW.ConcurrentHashMap[string, int64]
}

func NewStopWatch() *StopWatch {
	return &StopWatch{
		records: containerW.NewConcurrentHashMap[string, int64](nil,
			func(a, b string) int {
				return strings.Compare(a, b)
			}),
		start: time.Now().UnixNano(),
	}
}

func (s *StopWatch) curr() int64 {
	return time.Now().UnixNano()
}

func (s *StopWatch) Record() {
	s.records.Put("", s.curr())
}

func (s *StopWatch) RecordLabel(label string) {
	s.records.Put(label, s.curr())
}

func (s *StopWatch) NumRecords() int {
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
	for tup := range s.records.IterateEntry() {
		ts := tup.Get(1).(int64)
		fmt.Printf("%s: %s\n", tup.Get(0).(string), format(float64(ts-start)))
	}
	fmt.Println(format(float64(s.curr() - start)))
}

func (s *StopWatch) Clear() {
	s.records.Clear()
	atomic.StoreInt64(&s.start, s.curr())
}
