package harmony_test

import (
	"context"
	"fmt"
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
