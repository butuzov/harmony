//go:generate go run ./cmd/internal/gendone/
package harmony_test

import (
	"context"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/butuzov/harmony"
)

// --- Bridge Pattern  ---------------------------------------------------------
func TestBridgeWithContext(t *testing.T) {
	for name, test := range testTableContext {
		name, test := name, test

		t.Run(name, func(t *testing.T) {
			ctx, cancel := test.fncCtx()
			t.Cleanup(cancel)

			// given: chan chan generator
			incoming := func(n int) chan (<-chan int) {
				ch := make(chan (<-chan int))

				go func() {
					defer close(ch)

					for i := 0; i < n; i++ {
						ch <- generateNumberSequence(i*10, ((i+1)*10)-1)
					}
				}()

				return ch
			}

			// when: channel is drained
			var results []int
			for val := range harmony.BridgeWithContext(ctx, incoming(10)) {
				results = append(results, val)
			}

			// then: we checking the results
			if (len(results) == 100) != test.expected {
				t.Errorf("unexpected: results len(results) is %d", len(results))
			}
		})
	}
}

// --- Fan-in Pattern  ---------------------------------------------------------
func TestFanInWithContext(t *testing.T) {
	for name, test := range testTableContext {
		name, test := name, test

		t.Run(name, func(t *testing.T) {
			ctx, cancel := test.fncCtx()
			t.Cleanup(cancel)

			// given:
			chOut := harmony.FanInWithContext(ctx,
				generateNumberSequence(1, 10),
				generateNumberSequence(1, 10),
				generateNumberSequence(1, 10),
				generateNumberSequence(1, 10),
			)

			// when:
			var results []int
			for val := range chOut {
				results = append(results, val)
			}

			// then:
			if (len(results) == 40) != test.expected {
				t.Errorf("unexpected: len(results) is %d", len(results))
			}
		})
	}
}

// --- Future Pattern  ---------------------------------------------------------
func TestFututeWithContext(t *testing.T) {
	for name, test := range testTableContext {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := test.fncCtx()
			t.Cleanup(cancel)

			ch := harmony.FututeWithContext(ctx, func() int {
				time.Sleep(time.Millisecond)
				return 42
			})

			var val int
			var mu sync.Mutex
			go func() {
				mu.Lock()
				defer mu.Unlock()
				for {
					select {
					case tmpVal, ok := <-ch:
						if !ok {
							return
						}
						val = tmpVal
						return
					default:
					}
				}
			}()
			time.Sleep(time.Millisecond)

			var zeroVal int
			mu.Lock()
			defer mu.Unlock()
			if !(val == zeroVal) != test.expected {
				t.Errorf("unexpected: value of val is %+v", val)
			}
		})
	}
}

// --- OrDone Pattern  ---------------------------------------------------------
func TestOrWithContext(t *testing.T) {
	for name, test := range testTableContext {
		name, test := name, test

		t.Run(name, func(t *testing.T) {
			ctx, cancel := test.fncCtx()
			t.Cleanup(cancel)

			// given: sequence that ends with some random number.
			limit := 1_000
			outgoing := harmony.OrWithContext(ctx, generateNumberSequence(1, limit))

			// when: we run out processor we can finish in time or context can get canceld.
			var lastReadNumber int
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()

				for {
					select {
					case <-ctx.Done():
						return
					case val, ok := <-outgoing:
						if !ok {
							return
						}
						delayBusyWork()
						lastReadNumber = val
					}
				}
			}()

			// then: wating till consumer done, and check the last read number.
			wg.Wait()

			if (lastReadNumber == limit) != test.expected {
				t.Errorf("unexpected: want(%t) vs got(%t)", test.expected, lastReadNumber == limit)
			}
		})
	}
}

// --- Pipeline Pattern  -------------------------------------------------------
func TestPipelineWithContext(t *testing.T) {
	for name, test := range testTableContext {
		name, test := name, test

		t.Run(name, func(t *testing.T) {
			ctx, cancel := test.fncCtx()
			t.Cleanup(cancel)

			// given: simple pipeline
			jobFunc := func(job int) int {
				delayBusyWork()
				return job
			}

			pipe := harmony.PipelineWithContext(ctx, generateNumberSequence(1, 10), 1, jobFunc)

			// when: pipe is drained
			results := []int{}
			for val := range pipe {
				results = append(results, val)
			}

			// results are expected
			if (len(results) == 10) != test.expected {
				t.Errorf("unexpected: results len(results) is %d", len(results))
			}
		})
	}
}

// --- Queue Pattern -----------------------------------------------------------
func TestQueueWithContext(t *testing.T) {
	// returns next power of n on each call.
	pow := func(n uint64) func() uint64 {
		total := uint64(1)

		return func() uint64 {
			delayBusyWork()
			total *= n
			delayBusyWork()
			return total
		}
	}

	for name, test := range testTableContext {
		name, test := name, test

		t.Run(name, func(t *testing.T) {
			ctx, cancel := test.fncCtx()
			t.Cleanup(cancel)

			powerRes := make([]uint64, 10)
			powerCh := harmony.QueueWithContext(ctx, pow(2))
			for i := 0; i < cap(powerRes); i++ {
				powerRes[i] = <-powerCh
			}

			want := []uint64{2, 4, 8, 16, 32, 64, 128, 256, 512, 1024}
			if reflect.DeepEqual(want, powerRes) != test.expected {
				t.Errorf("got %v vs want (%t)%v", powerRes, test.expected, want)
			}
		})
	}
}

// --- Tee Pattern -------------------------------------------------------------
func TestTeeWithContext(t *testing.T) {
	for name, test := range testTableContext {
		name, test := name, test

		t.Run(name, func(t *testing.T) {
			ctx, cancel := test.fncCtx()
			t.Cleanup(cancel)

			// given: tee generating two channels, and reader that
			var wg sync.WaitGroup

			ch1, ch2 := harmony.TeeWithContext(ctx, generateNumberSequence(0, 9)) // seq (0...9)

			reader := func(ctx context.Context, in <-chan int) (*[]int, func()) {
				var n []int
				return &n, func() {
					defer wg.Done()

					for {
						select {
						case <-ctx.Done():
							return
						case val, ok := <-in:
							if !ok {
								return
							}
							n = append(n, val)
						}
					}
				}
			}

			// when:

			wg.Add(2)
			resCh1, readerCh1 := reader(ctx, ch1)
			go readerCh1()

			resCh2, readerCh2 := reader(ctx, ch2)
			go readerCh2()

			// then results are equal & has expected elements
			wg.Wait()

			if (len(*resCh1) == 10) != test.expected {
				t.Errorf("unexpected: len(results) is %d", len(*resCh1))
			}

			if (len(*resCh2) == 10) != test.expected {
				t.Errorf("unexpected: len(results) is %d", len(*resCh2))
			}
		})
	}
}

// --- WorkerPool Pattern  -----------------------------------------------------
func TestWorkerPoolWithContext(t *testing.T) {
	for name, test := range testTableContext {
		name, test := name, test

		t.Run(name, func(t *testing.T) {
			ctx, cancel := test.fncCtx()
			t.Cleanup(cancel)

			var flag int32

			jobQueue := make(chan int)

			// given: we running 100 workers that do nothing but wait for small delay each.
			harmony.WorkerPoolWithContext(ctx, jobQueue, 100, func(n int) {
				delayBusyWork()
			})

			go func() {
				defer func() {
					atomic.AddInt32(&flag, 1)
					close(jobQueue)
					cancel()
				}()

				for i := 0; i < 100; i++ {
					jobQueue <- i
				}
			}()

			<-ctx.Done()

			// then: we either have flag changed (to 1) (job queue channel is closed/drained)
			//                 or not (context cancelled)
			got := atomic.LoadInt32(&flag) == 1
			if got != test.expected {
				t.Errorf("unexpeced: want(%t) vs got(%t)", test.expected, got)
			}
		})
	}
}
