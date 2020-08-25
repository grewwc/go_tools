package pipeline

// RepeatFn takes a done channel to signal if finished
// "generator" is a generator function (returns interface{})
func RepeatFn(
	done chan interface{},
	generator func() interface{},
) <-chan interface{} {
	outStream := make(chan interface{})
	go func() {
		defer close(outStream)
		for {
			select {
			case <-done:
				return
			case outStream <- generator():
			}
		}
	}()
	return outStream
}

// Take takes a "n" variables from "inStream"
// can be controled by "done" channel
func Take(
	done chan interface{},
	inStream <-chan interface{},
	n int,
) <-chan interface{} {
	outStream := make(chan interface{})
	inStreamOrDone := orDone(done, inStream)
	go func() {
		defer close(outStream)
		for i := 0; i < n; i++ {
			select {
			case <-done:
				return
			case outStream <- <-inStreamOrDone:
			}
		}
	}()

	return outStream
}

// Tee duplicate "in" channel to 2 output channels
func Tee(in <-chan interface{}) (<-chan interface{}, <-chan interface{}) {
	out1 := make(chan interface{})
	out2 := make(chan interface{})
	go func() {
		defer close(out1)
		defer close(out2)
		for item := range in {
			out1, out2 := out1, out2
			for i := 0; i < 2; i++ {
				select {
				case out1 <- item:
					out1 = nil
				case out2 <- item:
					out2 = nil
				}
			}
		}
	}()
	return out1, out2
}
