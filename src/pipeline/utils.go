package pipeline

func OrDone[T any](
	done <-chan T,
	in <-chan T,
) <-chan T {
	out := make(chan T)
	go func() {
		defer close(out)
		for {
			select {
			case <-done:
				return
			case val, ok := <-in:
				if !ok {
					return
				}
				select {
				case <-done:
					return
				case out <- val:
				}
			}
		}
	}()

	return out
}
