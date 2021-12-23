package harmony

import (
	"context"
)

// Future[T any] will return buffered channel of size 1 and type T, which will
// eventually contain the results of the execution futureFn
func Futute[T any](futureFn func() T) <-chan T {
	return FututeWithContext(context.Background(), futureFn)
}

// FututeWithContext[T any] will return buffered channel of size 1 and type T,
// which will eventually contain the results of the execution futureFn, or be closed
// in case if context cancelled.
func FututeWithContext[T any](ctx context.Context, futureFn func() T) <-chan T {
	ch := make(chan T, 1)

	go func() {
		defer close(ch)

		select {
		case <-ctx.Done():
		case ch <- futureFn():
		}
	}()

	return ch
}
