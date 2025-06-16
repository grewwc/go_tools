package cw

import (
	"fmt"

	"github.com/grewwc/go_tools/src/typesw"
	"golang.org/x/exp/constraints"
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

func max[T constraints.Ordered](args ...T) T {
	if len(args) == 0 {
		return *new(T)
	}
	res := args[0]
	for _, val := range args[1:] {
		if val > res {
			res = val
		}
	}
	return res
}

func (f *BloomFilter[T]) Contains(item T) bool {
	hashVal := f.hash(item)
	cnt := max(f.numDigit(hashVal), f.hashTimes)
	for i := 0; i < cnt; i++ {
		if f.data[hashVal] == 0 {
			return false
		}
		hashVal = f.hasher(fmt.Sprintf("%v%d", item, hashVal)) % f.capacity
	}
	return true
}

func (f *BloomFilter[T]) Add(item T) {
	hashVal := f.hash(item)
	cnt := max(f.numDigit(hashVal), f.hashTimes)
	for i := 0; i < cnt; i++ {
		f.data[hashVal]++
		hashVal = f.hasher(fmt.Sprintf("%v%d", item, hashVal)) % f.capacity
	}
}

func (f *BloomFilter[T]) Delete(item T) {
	hashVal := f.hash(item)
	cnt := max(f.numDigit(hashVal), f.hashTimes)
	for i := 0; i < cnt; i++ {
		f.data[hashVal]--
		hashVal = f.hasher(fmt.Sprintf("%v%d", item, hashVal)) % f.capacity
	}
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
