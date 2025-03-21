package containerW

type Tuple struct {
	data []interface{}
}

func NewTuple(data ...interface{}) *Tuple {
	return &Tuple{
		data: data,
	}
}

func (t *Tuple) Size() int {
	return len(t.data)
}

func (t *Tuple) Len() int {
	return t.Size()
}

func (t *Tuple) Get(idx int) interface{} {
	if idx >= t.Size() {
		return nil
	}
	return t.data[idx]
}

func (t *Tuple) Set(idx int, val interface{}) {
	if idx >= t.Size() {
		return
	}
	t.data[idx] = val
}
