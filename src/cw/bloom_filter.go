package cw

import (
	"fmt"
	"sync/atomic"

	"github.com/grewwc/go_tools/src/algow"
	"github.com/grewwc/go_tools/src/typesw"
)

type BloomFilter[T any] struct {
	data      []uint64
	hasher    typesw.HashFunc[string]
	capacity  int
	hashTimes int
}

func NewBloomFilter[T any](capacity int) *BloomFilter[T] {
	hasher := typesw.CreateDefaultHash[string]()
	data := make([]uint64, capacity)
	return &BloomFilter[T]{
		hasher:    hasher,
		data:      data,
		capacity:  capacity,
		hashTimes: 3,
	}
}

func (f *BloomFilter[T]) MayExist(item T) bool {
	hashVal := f.hash(item)
	cnt := algow.Max(f.numDigit(hashVal), f.hashTimes)
	for i := 0; i < cnt; i++ {
		old := atomic.LoadUint64(&f.data[hashVal])
		if old == 0 {
			return false
		}
		hashVal = f.hasher(fmt.Sprintf("%v%d", item, hashVal)) % f.capacity
	}
	return true
}

func (f *BloomFilter[T]) Add(item T) bool {
	hashVal := f.hash(item)
	swapped := true
	cnt := algow.Max(f.numDigit(hashVal), f.hashTimes)
	for i := 0; i < cnt; i++ {
		old := atomic.LoadUint64(&f.data[hashVal])
		swapped = swapped && atomic.CompareAndSwapUint64(&f.data[hashVal], old, old+1)
		hashVal = f.hasher(fmt.Sprintf("%v%d", item, hashVal)) % f.capacity
	}
	return swapped
}

func (f *BloomFilter[T]) Delete(item T) bool {
	hashVal := f.hash(item)
	swapped := true
	cnt := algow.Max(f.numDigit(hashVal), f.hashTimes)
	for i := 0; i < cnt; i++ {
		old := atomic.LoadUint64(&f.data[hashVal])
		if old == 0 {
			return false
		}
		swapped = swapped && atomic.CompareAndSwapUint64(&f.data[hashVal], old, old-1)
		hashVal = f.hasher(fmt.Sprintf("%v%d", item, hashVal)) % f.capacity
	}
	return swapped
}

func (f *BloomFilter[T]) spread(h int) int {
	uh := uint(h)
	return int((uh ^ (uh >> 16)) & 0x7fffffff)
}

func (f *BloomFilter[T]) hash(item T) int {
	s := fmt.Sprintf("%v", item)
	return f.spread(f.hasher(s)) % f.capacity
}

func (f *BloomFilter[T]) numDigit(val int) int {
	ret := 0
	for val > 0 {
		ret++
		val = val >> 1
	}
	return ret
}
