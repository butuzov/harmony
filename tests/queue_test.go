package harmony_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/butuzov/harmony"
)

func TestQueueWithContext(t *testing.T) {
	// returns next power of n on each call.
	pow := func(n uint64) func() uint64 {
		total := uint64(1)

		return func() uint64 {
			total *= n
			return total
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	t.Cleanup(cancel)

	powerRes := make([]uint64, 10)
	powerCh := harmony.QueueWithContext(ctx, pow(2))
	for i := 0; i < cap(powerRes); i++ {
		powerRes[i] = <-powerCh
	}

	want := []uint64{2, 4, 8, 16, 32, 64, 128, 256, 512, 1024}
	if !reflect.DeepEqual(want, powerRes) {
		t.Errorf("got %v vs want %v", powerRes, want)
	}
}

func TestQueueWithDone(t *testing.T) {
	// returns next power of n on each call.
	pow := func(n uint64) func() uint64 {
		total := uint64(1)

		return func() uint64 {
			total *= n
			return total
		}
	}

	done := make(chan struct{})

	powerRes := make([]uint64, 10)
	powerCh := harmony.QueueWithDone(done, pow(2))
	for i := 0; i < cap(powerRes); i++ {
		powerRes[i] = <-powerCh
	}
	close(done)

	want := []uint64{2, 4, 8, 16, 32, 64, 128, 256, 512, 1024}
	if !reflect.DeepEqual(want, powerRes) {
		t.Errorf("got %v vs want %v", powerRes, want)
	}
}

func ExampleQueueWithDone() {
	// fin returns function  that returns Fibonacci sequence up to n element,
	// it returns 0 after limit reached.
	fib := func(limit int) func() int {
		a, b, nTh := 0, 1, 1
		return func() int {
			if nTh > limit {
				return 0
			}

			nTh++
			a, b = b, a+b
			return a
		}
	}

	first10FibNumbers := make([]int, 10)
	incoming := harmony.QueueWithContext(context.Background(), fib(10))
	for i := 0; i < cap(first10FibNumbers); i++ {
		first10FibNumbers[i] = <-incoming
	}

	fmt.Println(first10FibNumbers)
	// Output: [1 1 2 3 5 8 13 21 34 55]
}
