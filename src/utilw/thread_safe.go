package utilw

import (
	"sync"
)

type ThreadSafeVal struct {
	val interface{}
	mu  sync.RWMutex
}

func NewThreadSafeVal(val interface{}) *ThreadSafeVal {
	return &ThreadSafeVal{val, sync.RWMutex{}}
}

func (obj *ThreadSafeVal) Get() interface{} {
	obj.mu.RLock()
	defer obj.mu.RUnlock()
	return obj.val
}

func (obj *ThreadSafeVal) Set(val interface{}) {
	obj.mu.Lock()
	defer obj.mu.Unlock()
	obj.val = val
}
