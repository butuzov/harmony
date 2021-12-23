package harmony_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/butuzov/harmony"
)

func ExampleCancelWithContext() {
	icnoming := make(chan int)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Creating cancelable pipeline for incoming chan.
	chOut := harmony.CancelWithContext(ctx, icnoming)

	go func() {
		defer close(icnoming)
		for i := 1; i < 100; i++ {
			icnoming <- i
		}
	}()

	var res []int
	done := make(chan struct{})

	go func() {
		defer close(done)

		for val := range chOut {
			res = append(res, val)

			// We going to cancel execution once we reach any number devisable by 7
			if val%7 == 0 {
				cancel()
			}

			time.Sleep(time.Millisecond)

		}
	}()

	<-done

	fmt.Println(res)
	// Output: [1 2 3 4 5 6 7]
}

func TestCancelWithContext_ContextNotUsed(t *testing.T) {
	incoming := make(chan int)
	go func() {
		defer close(incoming)
		for i := 0; i < 10; i++ {
			incoming <- i
		}
	}()

	// Creating cancelable pipeline for incoming chan.
	chOut := harmony.CancelWithContext(context.Background(), incoming)
	var res []int
	want := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	for val := range chOut {
		res = append(res, val)
	}

	if !reflect.DeepEqual(res, want) {
		t.Errorf("want %v vs got %v", want, res)
	}
}
