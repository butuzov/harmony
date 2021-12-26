package harmony

import (
	"context"
)

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

// PipelineWithCDone returns the channel of generic type `T` that can serve
// as a pipeline for a next stage. It's implemented in same manner as a
// `WorkerPool` and allows to specify number of workers that going to proseed
// values received from the incoming channel. Outgoing channel is going to be
// closed once the incoming chan is closed or done is closed.
func PipelineWithDone[T any](
	done chan struct{},
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
