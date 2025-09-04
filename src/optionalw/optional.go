package optional

type Optional[T any] struct {
	val    T
	hasVal bool
}

func Of[T any](val T) *Optional[T] {
	return &Optional[T]{
		val:    val,
		hasVal: true,
	}
}

func (op *Optional[T]) HasValue() bool {
	if op == nil {
		return false
	}
	return op.hasVal
}

func (op *Optional[T]) ValueOr(defaultVal T) T {
	if op == nil {
		return defaultVal
	}
	if op.hasVal {
		return op.val
	}
	return defaultVal
}
