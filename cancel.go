package harmony

import "context"

// OrContextDone will return a new unbuffered channel of type `T`
// that serves as a pipeline for the incoming channel. Channel is closed once
// the context is canceled or the incoming channel is closed. This is variation
// or the pattern that usually called `OrDone` or`Cancel`.
func OrContextDone[T any](ctx context.Context, incoming <-chan T) <-chan T {
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
					return
				case ch <- val:
				}

			}
		}
	}()

	return ch
}

// OrDone will return a new unbuffered channel of type `T`
// that serves as a pipeline for the values from the incoming channel. Channel
// is closed once the done chan is closed or the incoming channel is closed.
// This pattern usually called `OrDone` or `Cancel`.
func OrDone[T any](done chan struct{}, incoming <-chan T) <-chan T {
	ch := make(chan T)

	go func() {
		defer close(ch)

		for {
			select {
			case <-done:
				return
			case val, ok := <-incoming:
				if !ok {
					return
				}

				select {
				case <-done:
					return
				case ch <- val:
				}

			}
		}
	}()

	return ch
}
