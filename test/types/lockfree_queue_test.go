package test

import (
	"sync"
	"testing"

	"github.com/grewwc/go_tools/src/cw"
	"github.com/grewwc/go_tools/src/sortw"
)

var parallel = 32

func BenchmarkLockFreeQueue(b *testing.B) {
	n := 10000
	// latch := typesw.NewCountDownLatch(n)
	wg := sync.WaitGroup{}
	wg.Add(n)
	ch := make(chan struct{}, parallel)
	l := cw.NewSyncLinkedlist[int]()
	for i := 0; i < n; i++ {
		ch <- struct{}{}
		go func(i int) {
			defer wg.Done()
			l.PushFront(i)
			<-ch
		}(i)
	}
	wg.Wait()

}

func TestLockFreeQueue(t *testing.T) {
	n := 10000
	// latch := typesw.NewCountDownLatch(n)
	wg := sync.WaitGroup{}
	wg.Add(n)
	ch := make(chan struct{}, parallel)
	l := cw.NewSyncLinkedlist[int]()
	for i := 0; i < n; i++ {
		ch <- struct{}{}
		go func(i int) {
			defer wg.Done()
			l.PushFront(i)
			<-ch
		}(i)
	}
	wg.Wait()

	if l.Len() != n {
		t.Errorf("len check failed. expected:%v, found:%v", n, l.Len())
	}

	s := l.Snapshot().ToSlice()
	sortw.Sort(s, nil)
	if s[0] != 0 || s[n-1] != n-1 {
		t.Error("data failed")
	}

	wg.Add(n)

	for i := 0; i < n; i++ {
		ch <- struct{}{}
		go func() {
			defer wg.Done()
			// if i%3 == 0 {
			// l.PopFront()
			// } else {
			l.PopBack()
			// }
			<-ch
		}()
	}

	wg.Wait()

	if l.Len() != 0 {
		t.Errorf("len should be 0. size:%v", l.Len())
	}

}
