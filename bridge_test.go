package harmony_test

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/butuzov/harmony"
)

func TestBridgeWithContext(t *testing.T) {
	genSequence := func(start, end int) <-chan int {
		ch := make(chan int)

		go func() {
			defer close(ch)

			for n := start; n < end; n++ {
				ch <- n
			}
		}()

		return ch
	}

	incoming := func(n int) chan (<-chan int) {
		ch := make(chan (<-chan int))

		go func() {
			defer close(ch)

			for i := 0; i < n; i++ {
				ch <- genSequence(i*10, (i+1)*10)
			}
		}()

		return ch
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	var results []int
	for val := range harmony.BridgeWithContext(ctx, incoming(10)) {
		results = append(results, val)
	}

	if len(results) != 100 {
		t.Errorf("unexpected: %v", results)
	}
}

func TestBridgeWithContext_Done(t *testing.T) {
	genSequence := func(start, end int) <-chan int {
		ch := make(chan int)

		go func() {
			defer close(ch)

			for n := start; n < end; n++ {
				ch <- n
				time.Sleep(time.Duration(rand.Int63n(10)) * time.Microsecond)
			}
		}()

		return ch
	}

	incoming := func(n int) chan (<-chan int) {
		ch := make(chan (<-chan int))

		go func() {
			defer close(ch)

			for i := 0; i < n; i++ {
				ch <- genSequence(i*10, (i+1)*10)
			}
		}()

		return ch
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	t.Cleanup(cancel)

	var results []int
	for val := range harmony.BridgeWithContext(ctx, incoming(10)) {
		results = append(results, val)
	}

	if len(results) == 100 {
		t.Errorf("unexpected: %v", results)
	}
}
