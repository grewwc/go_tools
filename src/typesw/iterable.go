package typesw

type IterableT[T any] interface {
	Iterate() chan T
}

type Iterable = IterableT[interface{}]

type it[T any] func() chan T

func (i it[T]) Iterate() chan T {
	return i()
}

func FuncToIterable[T any](f func() chan T) IterableT[T] {
	return it[T](f)
}

func ToIterable[T any](it Iterable) IterableT[T] {
	f := func() chan T {
		ch := make(chan T)
		go func() {
			defer close(ch)
			for val := range it.Iterate() {
				ch <- val.(T)
			}
		}()
		return ch
	}
	return FuncToIterable(f)
}

func EmptyIterable[T any]() IterableT[T] {
	return FuncToIterable(func() chan T {
		ch := make(chan T)
		close(ch)
		return ch
	})
}
