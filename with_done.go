// Code generated by cmd/internal/gendone. DO NOT EDIT.n
//go:generate make doc
package harmony

import (
	"sync"
)

// --- Bridge Pattern  ---------------------------------------------------------

// BridgeWithDone will return chan of generic type `T` used a pipe for the
// values received from the sequence of channels. Close channel (received from
// `incoming`) in order to  switch for a new one. Goroutines exists on close of
// `incoming` or done chan closed.
func BridgeWithDone[T any](done <-chan struct{}, incoming <-chan (<-chan T)) <-chan T {
	outgoing := make(chan T)

	go func() {
		defer close(outgoing)

		for {
			var stream <-chan T

			// Reading new stream on each iteration over the incoming, once incoming
			// drained we close goroutine.
			select {
			case <-done:
			case tmp, ok := <-incoming:
				if !ok {
					return
				}
				stream = tmp
			}

			// now we ae trying to actually read from channel, and pass value next to
			// its receive channel, if fail to read back to new iteration of this
			// loop.
			for val := range OrWithDone(done, stream) {
				outgoing <- val
			}
		}
	}()

	return outgoing
}

// --- Fan-in Pattern  ---------------------------------------------------------

// FanInWithDone returns unbuffered channel of generic type `T` which serves as
// delivery pipeline for the values received from at least 2 incoming channels,
// its closed once all of the incoming channels closed or done is closed
func FanInWithDone[T any](done <-chan struct{}, ch1, ch2 <-chan T, channels ...<-chan T) <-chan T {
	var wg sync.WaitGroup
	ch := make(chan T)

	// note(butuzov): for a sake of simplicity in gengrics worlds
	// todo(butuzov): remove once tooling allows
	wg.Add(2)
	go passWithDone(done, &wg, ch, ch1)
	go passWithDone(done, &wg, ch, ch2)

	for _, incomingChannel := range channels {
		wg.Add(1)
		go passWithDone(done, &wg, ch, incomingChannel)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	return ch
}

func passWithDone[T any](done <-chan struct{}, wg *sync.WaitGroup, out chan<- T, in <-chan T) {
	defer wg.Done()

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
}

// --- Future Pattern  ---------------------------------------------------------

// FututeWithDone[T any] will return buffered channel of size 1 and generic type `T`,
// which will eventually contain the results of the execution `futureFn``, or be closed
// in case if context cancelled.
func FututeWithDone[T any](done <-chan struct{}, futureFn func() T) <-chan T {
	ch := make(chan T, 1)

	go func() {
		defer close(ch)
		ch <- futureFn()
	}()

	return OrWithDone(done, ch)
}

// --- OrDone Pattern  ---------------------------------------------------------

// OrWithDone will return a new unbuffered channel of type `T`
// that serves as a pipeline for the incoming channel. Channel is closed once
// the context is canceled or the incoming channel is closed. This is variation
// or the pattern that usually called `OrWithDone` or`Cancel`.
func OrWithDone[T any](done <-chan struct{}, incoming <-chan T) <-chan T {
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

// --- Pipeline Pattern  -------------------------------------------------------

// PipelineWithDone returns the channel of generic type `T2` that can serve
// as a pipeline for the next stage. It's implemented in almost same manner as
// a `WorkerPool` and allows to specify number of workers that going to proseed
// values received from the incoming channel. Outgoing channel is going to be
// closed once the incoming chan is closed or context canceld.
func PipelineWithDone[T1, T2 any](
	done <-chan struct{},
	incoming <-chan T1,
	totalWorkers int,
	workerFn func(T1) T2,
) <-chan T2 {

	outgoing := make(chan T2)
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
			var job T1

			select {
			case <-done:
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

// QueueWithDone returns an unbuffered channel that is populated by
// func `genFn`. Chan is closed once context is Done. It's similar to `Future`
// pattern, but doesn't have a limit to just one result.
func QueueWithDone[T any](done <-chan struct{}, genFn func() T) <-chan T {
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

// --- Tee Pattern -------------------------------------------------------------

// TeeWithDone will return two channels of generic type `T` used to fan-out
// data from the incoming channel. Channels needs to be read in order next
// iteration over incoming chanel happen.
func TeeWithDone[T any](done <-chan struct{}, incoming <-chan T) (<-chan T, <-chan T) {
	ch1, ch2 := make(chan T), make(chan T)

	go func() {
		defer close(ch1)
		defer close(ch2)

		for val := range OrWithDone(done, incoming) {
			ch1, ch2 := ch1, ch2
			for i := 0; i < 2; i++ {
				select {
				// this case statement can add issue with disproportional ch1, ch2 sends
				// e.g. first chan got val, second didn't got chance due context.Done.
				case <-done:
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

// WorkerPoolWithDone accepts channel of generic type `T` which is used to
// serve jobs to max workersTotal workers. Goroutines stop: Once channel closed and drained,
// or done is closed
func WorkerPoolWithDone[T any](
	done <-chan struct{},
	jobQueue chan T,
	maxWorkers int,
	workFunc func(T),
) {
	busyWorkers := make(chan token, maxWorkers) // semaphore for the workers

	go func() {
		// wait for all workers to finish.
		defer func() {
			for i := 0; i < maxWorkers; i++ {
				busyWorkers <- token{}
			}
		}()

		for {
			var job T

			select {
			case <-done:
				return
			case tmp, ok := <-jobQueue:
				if !ok {
					return
				}
				job = tmp
			}

			// Execution is blocked until we able to get one more work permit for this
			// job. Pattern described by Bryan C. Miles
			busyWorkers <- token{}
			go func() {
				workFunc(job)
				<-busyWorkers
			}()
		}
	}()
}
