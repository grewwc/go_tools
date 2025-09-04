package test

import (
	"testing"

	"github.com/grewwc/go_tools/src/algow"
	"github.com/grewwc/go_tools/src/cw"
	"github.com/grewwc/go_tools/src/typesw"
)

func TestBlockingQueue(t *testing.T) {
	n := 2000
	data := make([]int, n)
	for i := 0; i < n; i++ {
		data[i] = i * i
	}

	l := typesw.NewCountDownLatch(1)
	arr := make([]int, n)
	q := cw.NewBlockingQueue[int](3)

	go func() {
		for i := 0; i < n; i++ {
			arr[i] = q.PopFirst()
			// fmt.Println(i, arr[i])
		}
		l.CountDown()
	}()

	for _, val := range data {
		q.AddLast(val)
	}

	l.Wait()

	if !algow.Equals(data, arr, nil) {
		t.Fatal("not equal", data[len(data)-10:], arr[len(arr)-10:])
	}

	q.AddFirst(1)
	q.AddFirst(2)

	if q.PeekFirst().ValueOr(-1) != 2 {
		t.Fail()
	}

	if q.PeekLast().ValueOr(-1) != 1 {
		t.Fail()
	}
}
