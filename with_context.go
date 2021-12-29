//go:generate go run ./cmd/internal/gendone/
package harmony

import (
	"context"
	"sync"
)

// --- Bridge Pattern  ---------------------------------------------------------

// BridgeWithContext will return chan of generic type `T` used a pipe for the
// values received from the sequence of channels. Close channel (received from
// `incoming`) in order to  switch for a new one. Goroutines exists on close of
// `incoming` or context canceled.
func BridgeWithContext[T any](ctx context.Context, incoming <-chan (<-chan T)) <-chan T {
	outgoing := make(chan T)

	go func() {
		defer close(outgoing)

		for {
			var stream <-chan T

			// Reading new stream on each iteration over the incoming, once incoming
			// drained we close goroutine.
			select {
			case <-ctx.Done():
			case tmp, ok := <-incoming:
				if !ok {
					return
				}
				stream = tmp
			}

			// now we ae trying to actually read from channel, and pass value next to
			// its receive channel, if fail to read back to new iteration of this
			// loop.
			for val := range OrWithContext(ctx, stream) {
				outgoing <- val
			}
		}
	}()

	return outgoing
}

// --- Fan-in Pattern  ---------------------------------------------------------

// FanInWithContext returns unbuffered channel of generic type `T` which serves as
// delivery pipeline for the values received from at least 2 incoming channels,
// its closed once all of the incoming channels closed or context cancelled.
func FanInWithContext[T any](ctx context.Context, ch1, ch2 <-chan T, channels ...<-chan T) <-chan T {
	var wg sync.WaitGroup
	ch := make(chan T)

	// note(butuzov): for a sake of simplicity in gengrics worlds
	// todo(butuzov): remove once tooling allows
	wg.Add(2)
	go passWithContext(ctx, &wg, ch, ch1)
	go passWithContext(ctx, &wg, ch, ch2)

	for _, incomingChannel := range channels {
		wg.Add(1)
		go passWithContext(ctx, &wg, ch, incomingChannel)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	return ch
}

func passWithContext[T any](ctx context.Context, wg *sync.WaitGroup, out chan<- T, in <-chan T) {
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

// --- Future Pattern  ---------------------------------------------------------

// FututeWithContext[T any] will return buffered channel of size 1 and generic type `T`,
// which will eventually contain the results of the execution `futureFn``, or be closed
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

// --- OrDone Pattern  ---------------------------------------------------------

// OrWithDone will return a new unbuffered channel of type `T`
// that serves as a pipeline for the incoming channel. Channel is closed once
// the context is canceled or the incoming channel is closed. This is variation
// or the pattern that usually called `OrWithDone` or`Cancel`.
func OrWithContext[T any](ctx context.Context, incoming <-chan T) <-chan T {
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

// --- Pipeline Pattern  -------------------------------------------------------

// PipelineWithContext returns the channel of generic type `T` that can serve
// as a pipeline for the next stage. It's implemented in same manner as a
// `WorkerPool` and allows to specify number of workers that going to proseed
// values received from the incoming channel. Outgoing channel is going to be
// closed once the incoming chan is closed or context canceld.
func PipelineWithContext[T any](
	ctx context.Context,
	incoming <-chan T,
	totalWorkers int,
	workerFn func(T) T,
) <-chan T {

	outgoing := make(chan T)
	workers := make(chan token, totalWorkers)

	go func() {
		// close outgoing channel
		defer close(outgoing)

		// wait for each worker to finish work.
		defer func() {
			for i := 0; i < totalWorkers; i++ {
				workers <- token{}
			}
		}()

		for {
			var job T

			select {
			case <-ctx.Done():
				return
			case tmp, ok := <-incoming:
				if !ok {
					return
				}
				job = tmp
			}

			workers <- token{}
			go func() {
				outgoing <- workerFn(job)
				<-workers
			}()
		}
	}()

	return outgoing
}

// --- Queue Pattern -----------------------------------------------------------

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

// --- Tee Pattern -------------------------------------------------------------

// TeeWithContext will return two channels of generic type `T` used to fan-out
// data from the incoming channel. Channels needs to be read in order next
// iteration over incoming chanel happen.
func TeeWithContext[T any](ctx context.Context, incoming <-chan T) (<-chan T, <-chan T) {
	ch1, ch2 := make(chan T), make(chan T)

	go func() {
		defer close(ch1)
		defer close(ch2)

		for val := range OrWithContext(ctx, incoming) {
			ch1, ch2 := ch1, ch2
			for i := 0; i < 2; i++ {
				select {
				// this case statement can add issue with disproportional ch1, ch2 sends
				// e.g. first chan got val, second didn't got chance due context.Done.
				case <-ctx.Done():
					return
				case ch1 <- val:
					ch1 = nil
				case ch2 <- val:
					ch2 = nil
				}
			}
		}
	}()

	return ch1, ch2
}

// --- WorkerPool Pattern  -----------------------------------------------------

// WorkerPoolWithContext returns channel of generic type `T` which excepts jobs
// of the same type for some number of workers that do workerFn. If you want to
// stop WorkerPool, close the jobQueue channel or cancel the context.
func WorkerPoolWithContext[T any](
	ctx context.Context,
	totalWorkers int,
	workerFn func(T),
) chan<- T {
	ch := make(chan T)                            // channel for the jobs.
	busyWorkers := make(chan token, totalWorkers) // semaphore for the workers

	go func() {
		// wait for all workers to finish.
		defer func() {
			for i := 0; i < totalWorkers; i++ {
				busyWorkers <- token{}
			}
		}()

		for {
			var job T

			select {
			case <-ctx.Done():
				return
			case tmp, ok := <-ch:
				if !ok {
					return
				}
				job = tmp
			}

			// Execution is blocked until we able to get one more work permit for this
			// job. Pattern described by Bryan C. Miles
			busyWorkers <- token{}
			go func() {
				workerFn(job)
				<-busyWorkers
			}()

		}
	}()

	return ch
}
