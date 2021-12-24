package harmony

import (
	"context"
)

// Queue returns an unbuffered channel populated by func genFn. It's similar to
// `Future` pattern but doesn't have a limit to just one result. Queue is leaking
// goroutine, and provided purly for consistency reasons. Use `QueueWithContext`
// to prevent leaking resources.
func Queue[T any](genFn func() T) <-chan T {
	return QueueWithContext(context.Background(), genFn)
}

// QueueWithContext[T any] returns an unbuffered channel thats is populated by
// func genFn. Chan is closed once context is Done. It's similar to Future
// pattern, but doesn't have limit to just one result.
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
