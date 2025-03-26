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
	Data *TreeMap[K, V]
}

type ConcurrentHashMap[K, V any] struct {
	Buckets []*buck[K, V] // 桶数组，每个桶是一个独立的哈希表
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
		Buckets: buckets,
		hasher:  hasher,
		mutex:   &sync.RWMutex{},
		cmp:     cmp,
	}
}

func (cm *ConcurrentHashMap[K, V]) hash(k K) int {
	res := cm.hasher(k) % len(cm.Buckets)
	if res < 0 {
		return -res
	}
	return res
}

func (cm *ConcurrentHashMap[K, V]) rehash(size int) {
	prev := cm.Buckets
	cm.Buckets = make([]*buck[K, V], size)
	for _, bucket := range prev {
		if bucket.Data == nil {
			continue
		}
		for entry := range bucket.Data.IterateEntry() {
			index := cm.hash(entry.k)
			if cm.Buckets[index] == nil {
				cm.Buckets[index] = &buck[K, V]{mu: &sync.RWMutex{}, Data: NewTreeMap[K, V](cm.cmp)}
			}
			cm.Buckets[index].Data.Put(entry.k, entry.v)
		}
	}
	for i := range cm.Buckets {
		if cm.Buckets[i] == nil {
			cm.Buckets[i] = &buck[K, V]{mu: &sync.RWMutex{}}
		}
	}
}

func (cm *ConcurrentHashMap[K, V]) Put(key K, value V) bool {
	grow := false
	cm.mutex.RLock()
	index := cm.hash(key)
	cm.Buckets[index].mu.Lock()
	if cm.Buckets[index].Data == nil {
		cm.Buckets[index].Data = NewTreeMap[K, V](cm.cmp)
	}
	res := cm.Buckets[index].Data.Put(key, value)
	cm.Buckets[index].mu.Unlock()
	if res {
		atomic.AddInt64(&cm.cnt, 1)
		if atomic.LoadInt64(&cm.cnt) >= int64(len(cm.Buckets))*growThreash {
			grow = true
			cm.mutex.RUnlock()
			cm.mutex.Lock()
			if atomic.LoadInt64(&cm.cnt) >= int64(len(cm.Buckets))*growThreash {
				cm.rehash(growThreash * len(cm.Buckets))
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
	cm.Buckets[index].mu.Lock()
	if cm.Buckets[index].Data == nil {
		cm.Buckets[index].Data = NewTreeMap[K, V](cm.cmp)
	}
	res := cm.Buckets[index].Data.PutIfAbsent(key, value)
	cm.Buckets[index].mu.Unlock()
	if res {
		atomic.AddInt64(&cm.cnt, 1)
		if atomic.LoadInt64(&cm.cnt) >= int64(len(cm.Buckets))*growThreash {
			grow = true
			cm.mutex.RUnlock()
			cm.mutex.Lock()
			if atomic.LoadInt64(&cm.cnt) >= int64(len(cm.Buckets))*growThreash {
				cm.rehash(growThreash * len(cm.Buckets))
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
	cm.Buckets[index].mu.RLock()
	if cm.Buckets[index].Data == nil {
		return *new(V)
	}
	defer cm.Buckets[index].mu.RUnlock()
	return cm.Buckets[index].Data.Get(key)
}

func (cm *ConcurrentHashMap[K, V]) GetOrDefault(key K, defaultVal V) V {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	index := cm.hash(key)
	cm.Buckets[index].mu.RLock()
	if cm.Buckets[index].Data == nil {
		return defaultVal
	}
	defer cm.Buckets[index].mu.RUnlock()
	return cm.Buckets[index].Data.GetOrDefault(key, defaultVal)
}

func (cm *ConcurrentHashMap[K, V]) Delete(key K) bool {
	shrink := false
	cm.mutex.RLock()
	index := cm.hash(key)
	cm.Buckets[index].mu.Lock()
	if cm.Buckets[index].Data == nil {
		cm.mutex.RUnlock()
		cm.Buckets[index].mu.Unlock()
		return false
	}
	res := cm.Buckets[index].Data.Delete(key)
	cm.Buckets[index].mu.Unlock()
	if res {
		atomic.AddInt64(&cm.cnt, -1)
		if cm.Buckets[index].Data.Size() == 0 {
			cm.Buckets[index].Data = nil
		}
		if atomic.LoadInt64(&cm.cnt) < int64(len(cm.Buckets)/shrinkThreash) {
			shrink = true
			cm.mutex.RUnlock()
			cm.mutex.Lock()
			r := int64(len(cm.Buckets) / shrinkThreash)
			if atomic.LoadInt64(&cm.cnt) < r {
				cm.rehash(algoW.Max(len(cm.Buckets)/shrinkThreash, initCap))
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
	cm.Buckets[index].mu.RLock()
	defer cm.Buckets[index].mu.RUnlock()
	return cm.Buckets[index].Data.Contains(key)
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
	cm.Buckets = make([]*buck[K, V], initCap)
}
