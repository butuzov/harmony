package harmony

import "context"

// CancelWithContext will return new channel unbuffered channel of type T
// that serve as pipeline for the incoming channel. Channel is closed once
// context canceled or incoming channel closed. This pattern usually called
// `Done`, `OrDone`, `Cancel`.
func CancelWithContext[T any](ctx context.Context, incoming <-chan T) chan T {
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
