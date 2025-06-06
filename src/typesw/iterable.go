package typesw

type IterableT[T any] interface {
	Iterate() <-chan T
	Stop()
}

type Iterable = IterableT[interface{}]

func FuncToIterable[T any](f func() chan T) IterableT[T] {
	return &funcIterable[T]{
		f: f,
	}
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

type emptyIterable[T any] struct{}

func (it *emptyIterable[T]) Iterate() <-chan T {
	ch := make(chan T)
	close(ch)
	return ch
}

func (it *emptyIterable[T]) Stop() {}

type funcIterable[T any] struct {
	f  func() chan T
	ch chan T
}

func (it *funcIterable[T]) Iterate() <-chan T {
	if it.ch == nil {
		it.ch = it.f()
	}
	return it.ch
}

func (it *funcIterable[T]) Stop() {
	quiteClose(it.ch)
}

func EmptyIterable[T any]() IterableT[T] {
	return &emptyIterable[T]{}
}

func quiteClose[T any](ch chan T) {
	defer func() {
		recover()
	}()
	close(ch)
}
