package cw

// sliceIterator
type sliceIterator[T any] struct {
	data []T

	chT chan T
}

func (it *sliceIterator[T]) Iterate() <-chan T {
	it.chT = make(chan T)
	go func() {
		for _, val := range it.data {
			it.chT <- val
		}
		quiteClose(it.chT)
	}()
	return it.chT
}

func (it *sliceIterator[T]) Stop() {
	quiteClose(it.chT)
}

// interfaceKeyMapIterator
type interfaceKeyMapIterator[K, V any] struct {
	data map[any]V

	chK chan K
}

func (it *interfaceKeyMapIterator[K, V]) Iterate() <-chan K {
	it.chK = make(chan K)
	go func() {
		for key := range it.data {
			it.chK <- key.(K)
		}
		quiteClose(it.chK)
	}()
	return it.chK
}

func (it *interfaceKeyMapIterator[K, V]) Stop() {
	quiteClose(it.chK)
}

// helper functions
// quiteClose close channel without panic
func quiteClose[T any](ch1 chan T) {
	func() {
		defer func() {
			recover()
		}()
		close(ch1)
	}()
}
