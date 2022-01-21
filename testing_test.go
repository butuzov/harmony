package harmony_test

import (
	"context"
	"sync"
	"time"
)

type CancelFunc = func()

var testTableContext = map[string]struct {
	fncCtx   func() (context.Context, context.CancelFunc)
	expected bool
}{
	"context canceled": {
		fncCtx: func() (context.Context, context.CancelFunc) {
			return context.WithTimeout(context.Background(), time.Millisecond/10)
		},
		expected: false, // feed channel isn't finished
	},
	"normal execution": {
		fncCtx: func() (context.Context, context.CancelFunc) {
			return context.WithCancel(context.Background())
		},
		expected: true, // feed channel did its work.
	},
}

var testTableDone = map[string]struct {
	fncDone  func() (chan struct{}, CancelFunc)
	expected bool
}{
	"done is closed": {
		fncDone: func() (chan struct{}, CancelFunc) {
			var (
				ch     = make(chan struct{})
				once   sync.Once
				cancel = func() {
					once.Do(func() { close(ch) })
				}
			)

			time.AfterFunc(time.Millisecond/100, cancel)

			return ch, cancel
		},
		expected: false, // feed channel isn't finished
	},
	"normal execution": {
		fncDone: func() (chan struct{}, CancelFunc) {
			var (
				ch     = make(chan struct{})
				once   sync.Once
				cancel = func() {
					once.Do(func() { close(ch) })
				}
			)

			return ch, cancel
		},
		expected: true, // feed channel did its work.
	},
}

var delayBusyWork = func() { time.Sleep(5 * time.Microsecond) }

func generateNumberSequence(start, finish int) <-chan int {
	ch := make(chan int)

	go func() {
		defer close(ch)
		for i := start; i <= finish; i++ {
			ch <- i
			delayBusyWork()
		}
	}()

	return ch
}
