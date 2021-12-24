package harmony

import (
	"context"
)

// Queue[T any] returns an unbuffered channel tpopulated by
// func genFn. It's similar to Future pattern, but doesn't have
// limit to just one result. Queue[T any] leeking goroutine, because of no way
// to stop generator function, please use QueueWithContext if you want prevent
// leaeking go routine.
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
