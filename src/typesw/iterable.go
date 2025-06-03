package typesw

type IterableT[T any] interface {
	Iterate() <-chan T
}

type Iterable = IterableT[interface{}]

type it[T any] func() <-chan T

func (i it[T]) Iterate() <-chan T {
	return i()
}

func ToIterable[T any](f func() <-chan T) IterableT[T] {
	return it[T](f)
}
