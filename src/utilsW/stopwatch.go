package utilsW

import (
	"fmt"
	"os"
	"time"
)

// StopWatch is "auto-reset" stopwatch
type StopWatch struct {
	start, stop int64
	running     bool
}

func NewStopWatch() StopWatch {
	return StopWatch{}
}

func (s *StopWatch) Start() {
	s.start = time.Now().UnixNano()
	s.running = true
}

func (s *StopWatch) Stop() {
	s.stop = time.Now().UnixNano()
	s.running = false
}

func (s StopWatch) Nanos() float64 {
	if s.running {
		fmt.Fprintln(os.Stderr, "stopwatch is running")
		return -1
	}
	return float64(s.stop - s.start)
}

func (s StopWatch) Micros() float64 {
	if s.running {
		fmt.Fprintln(os.Stderr, "stopwatch is running")
		return -1
	}
	return float64(s.stop-s.start) / 1e3
}

func (s StopWatch) Mills() float64 {
	if s.running {
		fmt.Fprintln(os.Stderr, "stopwatch is running")
		return -1
	}
	return float64(s.stop-s.start) / 1e6
}

func (s StopWatch) Seconds() float64 {
	if s.running {
		fmt.Fprintln(os.Stderr, "stopwatch is running")
		return -1
	}
	return float64(s.stop-s.start) / 1e9
}

func (s StopWatch) Minutes() float64 {
	if s.running {
		fmt.Fprintln(os.Stderr, "stopwatch is running")
		return -1
	}
	return s.Seconds() / 60.0
}

func (s StopWatch) Hours() float64 {
	if s.running {
		fmt.Fprintln(os.Stderr, "stopwatch is running")
		return -1
	}
	return s.Minutes() / 60
}

func (s StopWatch) Days() float64 {
	if s.running {
		fmt.Fprintln(os.Stderr, "stopwatch is running")
		return -1
	}
	return s.Hours() / 24.0
}
