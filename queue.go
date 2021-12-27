package harmony

import (
	"context"
)

// Queue returns an unbuffered channel that is populated by
// func genFn. Chan is closed once context is Done. It's similar to `Future`
// pattern, but doesn't have a limit to just one result. Also: it's leaking
// gorotine.
func Queue[T any](genFn func() T) <-chan T {
	return QueueWithContext(context.Background(), genFn)
}

// QueueWithContext returns an unbuffered channel that is populated by
// func `genFn`. Chan is closed once context is Done. It's similar to `Future`
// pattern, but doesn't have a limit to just one result.
func QueueWithContext[T any](ctx context.Context, genFn func() T) <-chan T {
	ch := make(chan T)

	go func() {
		defer close(ch)

		for {
			select {
			case <-ctx.Done():
				return
			case ch <- genFn():
			}
		}
	}()

	return ch
}

// QueueWithContext returns an unbuffered channel that is populated by
// the func `genFn`. Chan is closed once done chan is closed. It's similar to
// `Future` pattern, but doesn't have a limit to just one result.
func QueueWithDone[T any](done chan struct{}, genFn func() T) <-chan T {
	ch := make(chan T)

	go func() {
		defer close(ch)

		for {
			select {
			case <-done:
				return
			case ch <- genFn():
			}
		}
	}()

	return ch
}
