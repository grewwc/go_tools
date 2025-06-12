package cw

import (
	"container/list"

	"github.com/grewwc/go_tools/src/typesw"
)

// sliceIterator
type sliceIterator[T any] struct {
	data    []T
	reverse bool

	chT chan T
}

func (it *sliceIterator[T]) Iterate() <-chan T {
	it.chT = make(chan T)
	go func() {
		if !it.reverse {
			for _, val := range it.data {
				it.chT <- val
			}
		} else {
			for i := len(it.data) - 1; i >= 0; i-- {
				it.chT <- it.data[i]
			}
		}
		quiteClose(it.chT)
	}()
	return it.chT
}

func (it *sliceIterator[T]) Stop() {
	quiteClose(it.chT)
}

// listIterator
type listIterator[T any] struct {
	data    *list.List
	reverse bool

	ch chan T
}

func (it *listIterator[T]) Iterate() <-chan T {
	if it == nil || it.data == nil {
		return typesw.EmptyIterable[T]().Iterate()
	}
	it.ch = make(chan T)
	go func() {
		if !it.reverse {
			for curr := it.data.Front(); curr != nil; curr = curr.Next() {
				it.ch <- curr.Value.(T)
			}
		} else {
			for curr := it.data.Back(); curr != nil; curr = curr.Prev() {
				it.ch <- curr.Value.(T)
			}
		}
		quiteClose(it.ch)
	}()
	return it.ch
}

func (it *listIterator[T]) Stop() {
	quiteClose(it.ch)
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
