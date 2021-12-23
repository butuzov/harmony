package harmony

import (
	"context"
	"sync"
)

// FanIn[T any] returns unbuffered channel of type T which serves as delivery pipeline
// for the values received from at least 2 incoming channels, its closed once
// all of the incoming channels closed.
func FanIn[T any](ch1, ch2 <-chan T, channels ...<-chan T) <-chan T {
	return fanIn(context.Background(), ch1, ch2, channels...)
}

// FanInWithContext[T any] returns unbuffered channel of type T which serves as
// delivery pipeline for the values received from at least 2 incoming channels,
// its closed once all of the incoming channels closed or context cancelled.
func FanInWithContext[T any](ctx context.Context, ch1, ch2 <-chan T, channels ...<-chan T) <-chan T {
	return fanIn(ctx, ch1, ch2, channels...)
}

func fanIn[T any](ctx context.Context, ch1, ch2 <-chan T, channels ...<-chan T) <-chan T {
	var wg sync.WaitGroup
	ch := make(chan T)

	// note(butuzov): for a sake of simplicity in gengrics worlds
	// todo(butuzov): remove once tooling allows
	wg.Add(2)
	go pass(ctx, &wg, ch, ch1)
	go pass(ctx, &wg, ch, ch2)

	for _, incomingChannel := range channels {
		wg.Add(1)
		go pass(ctx, &wg, ch, incomingChannel)
	}

	go closeOnDone(&wg, ch)

	return ch
}

func pass[T any](ctx context.Context, wg *sync.WaitGroup, out chan<- T, in <-chan T) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case val, ok := <-in:
			if !ok {
				return
			}

			select {
			case <-ctx.Done():
				return
			case out <- val:
			}

		}
	}
}

func closeOnDone[T any](wg *sync.WaitGroup, ch chan T) {
	wg.Wait()
	close(ch)
}
