package typesw

type Iterable interface {
	Iterate() <-chan interface{}
}
