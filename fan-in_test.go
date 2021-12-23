package harmony_test

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/butuzov/harmony"
)

func testFillChan(t *testing.T, ch chan int, n func() int, ticker *time.Ticker, timerStart, timerStop *time.Timer) {
	<-timerStart.C
	t.Helper()

	defer close(ch)
	defer timerStart.Stop()
	defer timerStop.Stop()
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			tmp := n()
			if tmp%10 == 0 {
				return
			}
			ch <- tmp

		case <-timerStop.C:
			return
		}
	}
}

func TestFanIn(t *testing.T) {
	// given: given two channels
	ch1 := make(chan int) // fills with n in decrising order (suppose to fill first half of slice)
	ch2 := make(chan int) // fills with n in incrising order (suppose to fill second half of slice)
	chOut := harmony.FanIn(ch1, ch2)

	// then fill them with data
	go testFillChan(t,
		ch1,
		func() func() int { n := 100; return func() int { n--; return n } }(),
		time.NewTicker(10*time.Microsecond), // fill rate (100µs)
		time.NewTimer(100*time.Millisecond), // start at  (100ms)
		time.NewTimer(200*time.Millisecond), // stop at   (200ms)
	)
	go testFillChan(t,
		ch2,
		func() func() int { n := 100; return func() int { n++; return n } }(),
		time.NewTicker(10*time.Microsecond), // fill rate (100µs)
		time.NewTimer(200*time.Millisecond), // start at  (200ms)
		time.NewTimer(300*time.Millisecond), // stop at   (300ms)
	)

	// when: with sync
	var got []int
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for val := range chOut {
			got = append(got, val)
		}
	}()
	wg.Wait()
	// then: check if out channel same

	want := []int{99, 98, 97, 96, 95, 94, 93, 92, 91, 101, 102, 103, 104, 105, 106, 107, 108, 109}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("failed: want (%v) vs got (%v) ", want, got)
	}
}

func TestFanInWithContext(t *testing.T) {
	// given: given two channels
	ch1 := make(chan int) // fills with n in decrising order (suppose to fill first half of slice)
	ch2 := make(chan int) // fills with n in incrising order (suppose to fill second half of slice)

	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond) // time of start of ch2 fill
	t.Cleanup(cancel)

	chOut := harmony.FanInWithContext(ctx, ch1, ch2)

	// then fill them with data
	go testFillChan(t,
		ch1,
		func() func() int { n := 100; return func() int { n--; return n } }(),
		time.NewTicker(10*time.Microsecond), // fill rate (100µs)
		time.NewTimer(100*time.Millisecond), // start at  (100ms)
		time.NewTimer(200*time.Millisecond), // stop at   (200ms)
	)
	go testFillChan(t,
		ch2,
		func() func() int { n := 100; return func() int { n++; return n } }(),
		time.NewTicker(10*time.Microsecond), // fill rate (100µs)
		time.NewTimer(300*time.Millisecond), // start at  (300ms)
		time.NewTimer(400*time.Millisecond), // stop at   (400ms)
	)

	var got []int
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for val := range chOut {
			got = append(got, val)
		}
	}()
	wg.Wait()

	// then: check if out channel same
	want := []int{99, 98, 97, 96, 95, 94, 93, 92, 91}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("failed: want (%v) vs got (%v) ", want, got)
	}
}

func ExampleFanInWithContext() {
	ch1 := make(chan int)
	ch2 := make(chan int)

	// Context going to timeout in 70 milliseconds.
	ctx, cancel := context.WithTimeout(context.Background(), 70*time.Millisecond)
	defer cancel()

	ch := harmony.FanInWithContext(ctx, ch1, ch2)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(ch1)

		for i := 0; i < 5; i++ {
			ch1 <- i
			time.Sleep(10 * time.Millisecond)
		}
	}()

	go func() {
		wg.Wait()
		defer close(ch2)

		for i := 5; i <= 10; i++ {
			ch2 <- i
			time.Sleep(10 * time.Millisecond)
		}
	}()

	var res []int
	done := make(chan struct{})
	go func() {
		defer close(done)

		for v := range ch {
			res = append(res, v)
		}
	}()

	<-done
	fmt.Println(res)
	// Output: [0 1 2 3 4 5 6]
}
