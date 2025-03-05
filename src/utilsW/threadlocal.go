package utilsW

import (
	"github.com/grewwc/go_tools/src/typesW"
)

type ThreadLocal[T any] struct {
	m typesW.IConcurrentMap[int, T]
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
