package cw

/*Stack is not thread safe*/
type Stack struct {
	data []interface{}
}

func NewStack(capacity int) *Stack {
	data := make([]interface{}, 0, capacity)
	return &Stack{data}
}

func (s *Stack) Push(item interface{}) {
	s.data = append(s.data, item)
}

func (s *Stack) Pop() interface{} {
	result := s.Top()
	s.data = s.data[:len(s.data)-1]
	return result
}
func (s *Stack) Top() interface{} {
	size := len(s.data)
	if size == 0 {
		panic("stack is empty")
	}
	return s.data[len(s.data)-1]
}

func (s *Stack) Empty() bool {
	return s.Size() == 0
}

func (s *Stack) Size() int {
	return len(s.data)
}

func (s *Stack) Resize() {
	s.data = s.data[:0]
}

func (s *Stack) Iterate() chan interface{} {
	res := make(chan interface{})
	go func() {
		for i := len(s.data) - 1; i >= 0; i-- {
			res <- s.data[i]
		}
		close(res)
	}()
	return res
}
