package typesW

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
}
type IConcurrentMap[K, V any] interface {
	IMap[K, V]
}
