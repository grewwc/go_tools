package typew

type Iterable interface {
	Iterate() <-chan interface{}
}
