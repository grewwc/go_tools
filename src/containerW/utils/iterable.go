package containerW

type Iterable interface {
	Iterate() <-chan interface{}
}
