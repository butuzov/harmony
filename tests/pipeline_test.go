package harmony_test

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/butuzov/harmony"
)

func TestPipelineWithContext_Closed(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Nanosecond)
	t.Cleanup(cancel)

	incoming := func() <-chan int {
		ch := make(chan int)

		go func() {
			for i := 0; i < 10; i++ {
				ch <- i
			}
		}()

		return ch
	}()

	var pipe <-chan int

	pipe = harmony.PipelineWithContext(ctx, incoming, 1, func(job int) int {
		time.Sleep(time.Microsecond)
		return job
	})

	result := []int{}
	for val := range pipe {
		result = append(result, val)
	}

	if len(result) == 10 {
		t.Errorf("full result not expected, got (%v)", result)
	}
}

func TestPipelineWithDone_Done(t *testing.T) {
	done := make(chan struct{})
	time.AfterFunc(5*time.Nanosecond, func() { close(done) })

	incoming := func() <-chan int {
		ch := make(chan int)

		go func() {
			for i := 0; i < 10; i++ {
				ch <- i
			}
		}()

		return ch
	}()

	var pipe <-chan int

	pipe = harmony.PipelineWithDone(done, incoming, 1, func(job int) int {
		time.Sleep(time.Microsecond)
		return job
	})

	result := []int{}
	for val := range pipe {
		result = append(result, val)
	}

	if len(result) == 10 {
		t.Errorf("full result not expected, got (%v)", result)
	}
}

func ExamplePipelineWithContext() {
	lim := 10

	seedCh := func() <-chan int {
		ch := make(chan int)

		go func() {
			defer close(ch)

			for i := 0; i <= lim; i++ {
				ch <- i
			}
		}()

		return ch
	}()

	ctx := context.Background()

	var pipe <-chan int

	// we starting many subWorkers because we don't care about order of results
	// read from the pipe.
	pipe = seedCh
	// does nothing
	pipe = harmony.PipelineWithContext(ctx, pipe, 2, func(n int) int { return n })
	// does multiplication
	pipe = harmony.PipelineWithContext(ctx, pipe, 3, func(n int) int { return n * 2 })
	// does addition
	pipe = harmony.PipelineWithContext(ctx, pipe, 3, func(n int) int { return n + 2 })

	result := []int{}
	for val := range pipe {
		result = append(result, val)
	}

	sort.Ints(result)
	fmt.Printf("%v", result)
	// Output: [2 4 6 8 10 12 14 16 18 20 22]
}

func ExamplePipelineWithDone() {
	lim := 10

	seedCh := func() <-chan int {
		ch := make(chan int)

		go func() {
			defer close(ch)

			for i := 0; i <= lim; i++ {
				ch <- i
			}
		}()

		return ch
	}()

	done := make(chan struct{})
	defer close(done)

	var pipe <-chan int

	// we starting many subWorkers because we don't care about order of results
	// read from the pipe.
	pipe = seedCh
	// does nothing
	pipe = harmony.PipelineWithDone(done, pipe, 2, func(n int) int { return n })
	// does multiplication
	pipe = harmony.PipelineWithDone(done, pipe, 3, func(n int) int { return n * 2 })
	// does addition
	pipe = harmony.PipelineWithDone(done, pipe, 3, func(n int) int { return n + 2 })

	result := []int{}
	for val := range pipe {
		result = append(result, val)
	}

	sort.Ints(result)
	fmt.Printf("%v", result)
	// Output: [2 4 6 8 10 12 14 16 18 20 22]
}
