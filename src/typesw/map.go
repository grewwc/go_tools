package typesw

type IMap[K, V any] interface {
	Get(key K) V
	GetOrDefault(key K, defaultVal V) V
	Put(key K, val V) bool
	PutIfAbsent(key K, val V) bool
	Contains(key K) bool
	Delete(key K) bool
	Size() int
	DeleteAll(keys ...K)
	Iterate() <-chan (K)
	Clear()
	Keys() []K
	Values() []V
}

type ISortedMap[K, V any] interface {
	IMap[K, V]
	SearchRange(lower, upper K) []K
}

type IConcurrentMap[K, V any] interface {
	IMap[K, V]
}

type IList interface {
	Add(interface{})
	AddAll(...interface{})
	Delete(interface{}) bool
	Len() int
	Empty() bool
	Iterate() <-chan interface{}
	ShallowCopy() IList
	Contains(interface{}) bool
	Equals(IList) bool
	Get(int) interface{}
	Set(int, interface{}) interface{}
	Remove(int) interface{}
	ToStringSlice() []string
}

type IHeap[T any] interface {
	Insert(T)
	Pop() T
	Size() int
	IsEmpty() bool
	ToList() []T
	Next() T
	Top() T
}
