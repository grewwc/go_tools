package utilsw

import (
	"github.com/grewwc/go_tools/src/cw"
	"github.com/grewwc/go_tools/src/typesw"
)

type ThreadLocal[T any] struct {
	m typesw.IConcurrentMap[int, T]
}

func NewThreadLocal[T any](val T) *ThreadLocal[T] {
	result := &ThreadLocal[T]{
		m: cw.NewConcurrentHashMap[int, T](nil, func(i, j int) int {
			return i - j
		}),
	}
	result.m.Put(Goid(), val)
	return result
}

func (t *ThreadLocal[T]) Set(val T) {
	t.m.Put(Goid(), val)
}

func (t *ThreadLocal[T]) SetIfAbsent(val T) {
	t.m.PutIfAbsent(Goid(), val)
}

func (t *ThreadLocal[T]) Get() T {
	return t.m.Get(Goid())
}

func (t *ThreadLocal[T]) GetOrDefault(defaultVal T) T {
	return t.m.GetOrDefault(Goid(), defaultVal)
}

func (t *ThreadLocal[T]) Remove() {
	t.m.Delete(Goid())
}

func (t *ThreadLocal[T]) Contains() bool {
	return t.m.Contains(Goid())
}
