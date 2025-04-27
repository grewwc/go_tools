package typesw

import (
	"log"
	"time"
)

// Semaphore 是一个控制并发访问的信号量
type Semaphore struct {
	ch chan struct{}
}

// NewSemaphore 创建一个最大并发数为 max 的信号量
func NewSemaphore(max int) *Semaphore {
	if max <= 0 {
		log.Fatalln("max must be greater than 0")
	}
	sem := &Semaphore{
		ch: make(chan struct{}, max),
	}
	// 初始化时填满 channel，表示初始有 max 个许可
	for i := 0; i < max; i++ {
		sem.ch <- struct{}{}
	}
	return sem
}

// Acquire 获取一个许可，当无许可时阻塞
func (s *Semaphore) Acquire() {
	<-s.ch // 从 channel 接收，当 channel 有元素时获取成功
}

func (s *Semaphore) AcquireTimeout(timeout time.Duration) bool {
	select {
	case <-s.ch:
		return true
	case <-time.After(timeout):
		return false
	}
}

// Release 释放一个许可
func (s *Semaphore) Release() {
	s.ch <- struct{}{} // 向 channel 发送元素，释放一个许可
}

func (s *Semaphore) ReleaseTimeout(timeout time.Duration) bool {
	select {
	case s.ch <- struct{}{}:
		return true
	case <-time.After(timeout):
		return false
	}
}
