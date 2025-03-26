package containerW

import (
	"sync"
	"sync/atomic"

	"github.com/grewwc/go_tools/src/algoW"
	"github.com/grewwc/go_tools/src/typesW"
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
	hasher  typesW.HashFunc[K]
	cnt     int64

	cmp typesW.CompareFunc[K]
}

func NewConcurrentHashMap[K, V any](hasher typesW.HashFunc[K], cmp typesW.CompareFunc[K]) *ConcurrentHashMap[K, V] {
	buckets := make([]*buck[K, V], initCap)
	if hasher == nil {
		hasher = typesW.CreateDefaultHash[K]()
	}
	if cmp == nil {
		cmp = typesW.CreateDefaultCmp[K]()
	}
	for i := range buckets {
		buckets[i] = &buck[K, V]{mu: &sync.RWMutex{}}
	}
	return &ConcurrentHashMap[K, V]{
		buckets: buckets,
		hasher:  hasher,
		mutex:   &sync.RWMutex{},
		cmp:     cmp,
	}
}

func (cm *ConcurrentHashMap[K, V]) hash(k K) int {
	res := cm.hasher(k) % len(cm.buckets)
	if res < 0 {
		return -res
	}
	return res
}

func (cm *ConcurrentHashMap[K, V]) rehash(size int) {
	prev := cm.buckets
	cm.buckets = make([]*buck[K, V], size)
	for _, bucket := range prev {
		if bucket.data == nil {
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
	for i := range cm.buckets {
		if cm.buckets[i] == nil {
			cm.buckets[i] = &buck[K, V]{mu: &sync.RWMutex{}}
		}
	}
}

func (cm *ConcurrentHashMap[K, V]) Put(key K, value V) bool {
	grow := false
	cm.mutex.RLock()
	index := cm.hash(key)
	cm.buckets[index].mu.Lock()
	if cm.buckets[index].data == nil {
		cm.buckets[index].data = NewTreeMap[K, V](cm.cmp)
	}
	res := cm.buckets[index].data.Put(key, value)
	cm.buckets[index].mu.Unlock()
	if res {
		atomic.AddInt64(&cm.cnt, 1)
		if atomic.LoadInt64(&cm.cnt) >= int64(len(cm.buckets))*growThreash {
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
	cm.buckets[index].mu.Lock()
	if cm.buckets[index].data == nil {
		cm.buckets[index].data = NewTreeMap[K, V](cm.cmp)
	}
	res := cm.buckets[index].data.PutIfAbsent(key, value)
	cm.buckets[index].mu.Unlock()
	if res {
		atomic.AddInt64(&cm.cnt, 1)
		if atomic.LoadInt64(&cm.cnt) >= int64(len(cm.buckets))*growThreash {
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
	cm.buckets[index].mu.RLock()
	if cm.buckets[index].data == nil {
		return *new(V)
	}
	defer cm.buckets[index].mu.RUnlock()
	return cm.buckets[index].data.Get(key)
}

func (cm *ConcurrentHashMap[K, V]) GetOrDefault(key K, defaultVal V) V {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	index := cm.hash(key)
	cm.buckets[index].mu.RLock()
	if cm.buckets[index].data == nil {
		return defaultVal
	}
	defer cm.buckets[index].mu.RUnlock()
	return cm.buckets[index].data.GetOrDefault(key, defaultVal)
}

func (cm *ConcurrentHashMap[K, V]) Delete(key K) bool {
	shrink := false
	cm.mutex.RLock()
	index := cm.hash(key)
	cm.buckets[index].mu.Lock()
	if cm.buckets[index].data == nil {
		cm.mutex.RUnlock()
		cm.buckets[index].mu.Unlock()
		return false
	}
	res := cm.buckets[index].data.Delete(key)
	cm.buckets[index].mu.Unlock()
	if res {
		atomic.AddInt64(&cm.cnt, -1)
		if cm.buckets[index].data.Size() == 0 {
			cm.buckets[index].data = nil
		}
		if atomic.LoadInt64(&cm.cnt) < int64(len(cm.buckets)/shrinkThreash) {
			shrink = true
			cm.mutex.RUnlock()
			cm.mutex.Lock()
			r := int64(len(cm.buckets) / shrinkThreash)
			if atomic.LoadInt64(&cm.cnt) < r {
				cm.rehash(algoW.Max(len(cm.buckets)/shrinkThreash, initCap))
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
	cm.buckets[index].mu.RLock()
	defer cm.buckets[index].mu.RUnlock()
	return cm.buckets[index].data.Contains(key)
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
			if bucket.data == nil {
				bucket.mu.RUnlock()
				continue
			}
			for k := range bucket.data.Iterate() {
				ch <- k
			}
			bucket.mu.RUnlock()
		}
	}()
	return ch
}
