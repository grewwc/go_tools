package typesW

import (
	"sync"

	"github.com/petermattis/goid"
)

// ReentrantMutex 是可重入的互斥锁实现
type ReentrantMutex struct {
	mu      sync.Mutex
	owner   int64 // 当前持有锁的goroutine ID
	count   int32 // 当前goroutine的持有计数
	waiters *sync.Cond
}

// NewReentrantMutex 创建新的可重入互斥锁
func NewReentrantMutex() *ReentrantMutex {
	m := &ReentrantMutex{}
	m.waiters = sync.NewCond(&m.mu)
	return m
}

// Lock 获取锁
func (m *ReentrantMutex) Lock() {
	current := goid.Get()
	m.mu.Lock()
	// 如果当前goroutine已经是持有者，直接增加计数
	if m.owner == current {
		m.count++
		m.mu.Unlock()
		return
	}

	// 等待直到锁可用
	for m.owner != 0 {
		m.waiters.Wait()
	}

	// 成为新持有者
	m.owner = current
	m.count = 1
	m.mu.Unlock()
}

// Unlock 释放锁
func (m *ReentrantMutex) Unlock() {
	current := goid.Get()
	m.mu.Lock()
	// 检查是否是当前持有者
	if m.owner != current {
		m.mu.Unlock()
		return
	}

	// 减少计数
	m.count--
	if m.count == 0 {
		// 当计数为0时，释放锁并唤醒等待者
		m.owner = 0
		m.mu.Unlock()
		m.waiters.Broadcast()
	} else {
		m.mu.Unlock()
	}
}
