package typesW

type Comparable interface {
	Compare(interface{}) int
}

type IntComparable int

func (i IntComparable) Compare(other interface{}) int {
	return int(i - other.(IntComparable))
}

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
}
type IConcurrentMap[K, V any] interface {
	IMap[K, V]
}
