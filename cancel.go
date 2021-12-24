package harmony

import "context"

// CancelWithContext will return a new channel unbuffered channel of type `T`
// that serves as a pipeline for the incoming channel. Channel is closed once
// the context is canceled or the incoming channel is closed. This pattern
// usually called `Done`, `OrDone`, `Cancel`.
func CancelWithContext[T any](ctx context.Context, incoming <-chan T) <-chan T {
	ch := make(chan T)

	go func() {
		defer close(ch)

		for {
			select {
			case <-ctx.Done():
				return
			case val, ok := <-incoming:
				if !ok {
					return
				}

				select {
				case <-ctx.Done():
				case ch <- val:
				}

			}
		}
	}()

	return ch
}
