package typesW

type Iterable interface {
	Iterate() <-chan interface{}
}
