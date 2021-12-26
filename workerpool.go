package harmony

import (
	"context"
)

// WorkerPool returns channel of generic type `T` which excepts jobs of the same type
// for some number of workers that do workerFn. If you want to stop WorkerPool, close
// the jobQueue channel.
func WorkerPool[T any](totalWorkers int, workerFn func(T)) chan<- T {
	return WorkerPoolWithContext(context.Background(), totalWorkers, workerFn)
}

// WorkerPool returns channel of generic type `T` which excepts jobs of the same type
// for some number of workers that do workerFn. If you want to stop WorkerPool, close
// the jobQueue channel or cancel the context.
func WorkerPoolWithContext[T any](ctx context.Context, totalWorkers int, workerFn func(T)) chan<- T {
	ch := make(chan T) // channel for the jobs.
	busyWorkers := make(chan token, totalWorkers)

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
