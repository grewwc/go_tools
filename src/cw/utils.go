package cw

import "github.com/grewwc/go_tools/src/typesw"

func Zip[T any](it1, it2 typesw.IterableT[T]) typesw.IterableT[*Tuple] {
	return typesw.FuncToIterable(func() chan *Tuple {
		ch := make(chan *Tuple)
		go func() {
			defer close(ch)
			ch2 := it2.Iterate()
			for v1 := range it1.Iterate() {
				if v2, ok := <-ch2; ok {
					ch <- NewTuple(v1, v2)
				} else {
					break
				}
			}
		}()
		return ch
	})
}
