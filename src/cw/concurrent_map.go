package cw

import (
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/grewwc/go_tools/src/algow"
	"github.com/grewwc/go_tools/src/typesw"
)

const (
	growThreash   = 64
	shrinkThreash = 16
	initCap       = 64
)

// buck 是一个独立的哈希表
type buck[K, V any] struct {
	mu   *sync.RWMutex
	data *TreeMap[K, V]
}

type ConcurrentHashMap[K, V any] struct {
	buckets []*buck[K, V] // 桶数组，每个桶是一个独立的哈希表
	mutex   *sync.RWMutex
	hasher  typesw.HashFunc[K]
	cnt     int64

	cmp typesw.CompareFunc[K]
}

func NewConcurrentHashMap[K, V any](hasher typesw.HashFunc[K], cmp typesw.CompareFunc[K]) *ConcurrentHashMap[K, V] {
	buckets := make([]*buck[K, V], initCap)
	if hasher == nil {
		hasher = typesw.CreateDefaultHash[K]()
	}
	if cmp == nil {
		cmp = typesw.CreateDefaultCmp[K]()
	}
	return &ConcurrentHashMap[K, V]{
		buckets: buckets,
		hasher:  hasher,
		mutex:   &sync.RWMutex{},
		cmp:     cmp,
	}
}

func (cm *ConcurrentHashMap[K, V]) hash(k K) int {
	return spread(cm.hasher(k)) & (len(cm.buckets) - 1)
}

func (cm *ConcurrentHashMap[K, V]) rehash(size int) {
	prev := cm.buckets
	cm.buckets = make([]*buck[K, V], size)
	for _, bucket := range prev {
		if bucket == nil {
			continue
		}
		for entry := range bucket.data.IterateEntry() {
			index := cm.hash(entry.k)
			if cm.buckets[index] == nil {
				cm.buckets[index] = &buck[K, V]{mu: &sync.RWMutex{}, data: NewTreeMap[K, V](cm.cmp)}
			}
			cm.buckets[index].data.Put(entry.k, entry.v)
		}
	}
}

func spread(h int) int {
	uh := uint(h)
	return int((uh ^ (uh >> 16)) & 0x7fffffff)
}

func (cm *ConcurrentHashMap[K, V]) Put(key K, value V) bool {
	grow := false
	cm.mutex.RLock()
	index := cm.hash(key)
	ptr := (*unsafe.Pointer)(unsafe.Pointer(&cm.buckets[index]))
	sPtr := atomic.LoadPointer(ptr)
	if sPtr == nil {
		newNode := &buck[K, V]{data: NewTreeMap[K, V](cm.cmp), mu: &sync.RWMutex{}}
		newNode.data.Put(key, value)
		if atomic.CompareAndSwapPointer(ptr, nil, unsafe.Pointer(newNode)) {
			cm.mutex.RUnlock()
			atomic.AddInt64(&cm.cnt, 1)
			return true
		}
	}
	if sPtr = atomic.LoadPointer(ptr); sPtr == nil {
		return false
	}
	bucketPtr := (*buck[K, V])(sPtr)
	bucketPtr.mu.Lock()
	res := bucketPtr.data.Put(key, value)
	bucketPtr.mu.Unlock()
	if res {
		cnt := atomic.AddInt64(&cm.cnt, 1)
		if cnt >= int64(len(cm.buckets))*growThreash {
			grow = true
			cm.mutex.RUnlock()
			cm.mutex.Lock()
			if atomic.LoadInt64(&cm.cnt) >= int64(len(cm.buckets))*growThreash {
				cm.rehash(growThreash * len(cm.buckets))
			}
			cm.mutex.Unlock()
		}
	}
	if !grow {
		cm.mutex.RUnlock()
	}
	return res
}

func (cm *ConcurrentHashMap[K, V]) PutIfAbsent(key K, value V) bool {
	grow := false
	cm.mutex.RLock()
	index := cm.hash(key)
	ptr := (*unsafe.Pointer)(unsafe.Pointer(&cm.buckets[index]))
	sPtr := atomic.LoadPointer(ptr)
	if sPtr == nil {
		newNode := &buck[K, V]{data: NewTreeMap[K, V](cm.cmp), mu: &sync.RWMutex{}}
		newNode.data.Put(key, value)
		if atomic.CompareAndSwapPointer(ptr, nil, unsafe.Pointer(newNode)) {
			cm.mutex.RUnlock()
			atomic.AddInt64(&cm.cnt, 1)
			return true
		}
	}
	if sPtr = atomic.LoadPointer(ptr); sPtr == nil {
		return false
	}
	bucketPtr := (*buck[K, V])(sPtr)
	bucketPtr.mu.Lock()
	res := bucketPtr.data.PutIfAbsent(key, value)
	bucketPtr.mu.Unlock()
	if res {
		cnt := atomic.AddInt64(&cm.cnt, 1)
		if cnt >= int64(len(cm.buckets))*growThreash {
			grow = true
			cm.mutex.RUnlock()
			cm.mutex.Lock()
			if atomic.LoadInt64(&cm.cnt) >= int64(len(cm.buckets))*growThreash {
				cm.rehash(growThreash * len(cm.buckets))
			}
			cm.mutex.Unlock()
		}
	}
	if !grow {
		cm.mutex.RUnlock()
	}
	return res
}

func (cm *ConcurrentHashMap[K, V]) Get(key K) V {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	index := cm.hash(key)
	ptr := (*unsafe.Pointer)(unsafe.Pointer(&cm.buckets[index]))
	sPtr := atomic.LoadPointer(ptr)
	if sPtr == nil {
		return *new(V)
	}
	bucketPtr := (*buck[K, V])(sPtr)
	bucketPtr.mu.RLock()
	defer bucketPtr.mu.RUnlock()
	return bucketPtr.data.Get(key)
}

func (cm *ConcurrentHashMap[K, V]) Keys() []K {
	var res []K
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	for _, buck := range cm.buckets {
		buck.mu.Lock()
		res = append(res, buck.data.Keys()...)
		buck.mu.Unlock()
	}
	return res
}

func (cm *ConcurrentHashMap[K, V]) Values() []V {
	s := NewSet()
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	for _, buck := range cm.buckets {
		buck.mu.Lock()
		for _, val := range buck.data.Values() {
			s.Add(val)
		}
		buck.mu.Unlock()
	}
	res := make([]V, 0, s.Size())
	for val := range s.Iterate() {
		res = append(res, val.(V))
	}
	return res
}

func (cm *ConcurrentHashMap[K, V]) GetOrDefault(key K, defaultVal V) V {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	index := cm.hash(key)
	ptr := (*unsafe.Pointer)(unsafe.Pointer(&cm.buckets[index]))
	sPtr := atomic.LoadPointer(ptr)
	if sPtr == nil {
		return defaultVal
	}
	if sPtr = atomic.LoadPointer(ptr); sPtr == nil {
		return *new(V)
	}
	bucketPtr := (*buck[K, V])(sPtr)
	bucketPtr.mu.RLock()
	defer bucketPtr.mu.RUnlock()
	return bucketPtr.data.GetOrDefault(key, defaultVal)
}

func (cm *ConcurrentHashMap[K, V]) Delete(key K) bool {
	shrink := false
	cm.mutex.RLock()
	index := cm.hash(key)
	ptr := (*unsafe.Pointer)(unsafe.Pointer(&cm.buckets[index]))
	sPtr := atomic.LoadPointer(ptr)
	if sPtr == nil {
		cm.mutex.RUnlock()
		return false
	}
	bucketPtr := (*buck[K, V])(sPtr)
	mu := bucketPtr.mu
	mu.Lock()
	res := bucketPtr.data.Delete(key)
	if res && bucketPtr.data.Size() == 0 {
		atomic.StorePointer(ptr, nil)
		// atomic.CompareAndSwapPointer(ptr, unsafe.Pointer(bucketPtr), nil)
	}
	mu.Unlock()
	if res {
		cnt := atomic.AddInt64(&cm.cnt, -1)
		if cnt < int64(len(cm.buckets)/shrinkThreash) {
			shrink = true
			cm.mutex.RUnlock()
			cm.mutex.Lock()
			r := int64(len(cm.buckets) / shrinkThreash)
			if atomic.LoadInt64(&cm.cnt) < r {
				cm.rehash(algow.Max(len(cm.buckets)/shrinkThreash, initCap))
			}
			cm.mutex.Unlock()
		}
	}
	if !shrink {
		cm.mutex.RUnlock()
	}
	return res
}

func (cm *ConcurrentHashMap[K, V]) Contains(key K) bool {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	index := cm.hash(key)
	ptr := (*unsafe.Pointer)(unsafe.Pointer(&cm.buckets[index]))
	sPtr := atomic.LoadPointer(ptr)
	if sPtr == nil {
		return false
	}
	bucketPtr := (*buck[K, V])(sPtr)
	bucketPtr.mu.RLock()
	defer bucketPtr.mu.RUnlock()
	return bucketPtr.data.Contains(key)
}

func (cm *ConcurrentHashMap[K, V]) Size() int {
	return int(atomic.LoadInt64(&cm.cnt))
}

func (cm *ConcurrentHashMap[K, V]) DeleteAll(keys ...K) {
	for _, k := range keys {
		cm.Delete(k)
	}
}

func (cm *ConcurrentHashMap[K, V]) Clear() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	cm.cnt = 0
	cm.buckets = make([]*buck[K, V], initCap)
}

// Iterate is not thread safe
func (cm *ConcurrentHashMap[K, V]) Iterate() <-chan K {
	ch := make(chan K)
	go func() {
		defer close(ch)
		cm.mutex.RLock()
		defer cm.mutex.RUnlock()
		for _, bucket := range cm.buckets {
			if bucket == nil {
				continue
			}
			bucket.mu.RLock()
			for k := range bucket.data.Iterate() {
				ch <- k
			}
			bucket.mu.RUnlock()
		}
	}()
	return ch
}

// Iterate is not thread safe
func (cm *ConcurrentHashMap[K, V]) IterateEntry() <-chan *Tuple {
	ch := make(chan *Tuple)
	go func() {
		defer close(ch)
		cm.mutex.RLock()
		defer cm.mutex.RUnlock()
		for _, bucket := range cm.buckets {
			if bucket == nil {
				continue
			}
			bucket.mu.RLock()
			for entry := range bucket.data.IterateEntry() {
				ch <- NewTuple(entry.k, entry.v)
			}
			bucket.mu.RUnlock()
		}
	}()
	return ch
}
