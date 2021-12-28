package harmony_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/butuzov/harmony"
)

func TestTeeWithContext(t *testing.T) {
	// numbers generator
	gen := func() chan int {
		ch := make(chan int)

		go func() {
			defer close(ch)
			for i := 1; i <= 9; i++ {
				ch <- i
			}
		}()

		return ch
	}()

	var wg sync.WaitGroup

	ch1, ch2 := harmony.TeeWithContext(context.Background(), gen)

	expectedSum := 46
	sum, sumFunc := func(in <-chan int) (*int, func()) {
		n := 1
		return &n, func() {
			defer wg.Done()

			for val := range in {
				n += val
			}
		}
	}(ch1)

	expectedFact := 362_880
	fact, factFunc := func(in <-chan int) (*int, func()) {
		n := 1
		return &n, func() {
			defer wg.Done()

			for val := range in {
				n *= val
			}
		}
	}(ch2)

	wg.Add(2)

	go sumFunc()
	go factFunc()

	wg.Wait()

	if *sum != expectedSum {
		t.Errorf("Sum: expected(%d) vs got(%d)", expectedSum, *sum)
	}

	if *fact != expectedFact {
		t.Errorf("Product: expected(%d) vs got(%d)", expectedFact, *fact)
	}
}

func TestTeeWithContext_ContextDone(t *testing.T) {
	// numbers generator
	gen := func() chan int {
		ch := make(chan int)

		go func() {
			defer close(ch)
			for i := 1; i <= 9; i++ {
				ch <- i
				time.Sleep(100 * time.Microsecond)
			}
		}()

		return ch
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Nanosecond)
	t.Cleanup(cancel)

	ch1, ch2 := harmony.TeeWithContext(ctx, gen)
	wg := sync.WaitGroup{}

	reader := func(ch <-chan int) (*[]int, func()) {
		wg.Add(1)
		res := make([]int, 0, 100)
		return &res, func() {
			defer wg.Done()

			for val := range ch {
				res = append(res, val)
			}
		}
	}

	res1, goReader1 := reader(ch1)
	go goReader1()
	res2, goReader2 := reader(ch2)
	go goReader2()

	wg.Wait()

	if len(*res1) == 10 || len(*res2) == 10 {
		t.Errorf("unexpected results: res1(%v) && res2(%v)", res1, res2)
	}
}

func TestTeeWithDone_Done(t *testing.T) {
	// numbers generator
	gen := func() chan int {
		ch := make(chan int)

		go func() {
			defer close(ch)
			for i := 1; i <= 9; i++ {
				ch <- i
				time.Sleep(100 * time.Microsecond)
			}
		}()

		return ch
	}()

	done := make(chan struct{})
	time.AfterFunc(5*time.Nanosecond, func() { close(done) })

	ch1, ch2 := harmony.TeeWithDone(done, gen)
	wg := sync.WaitGroup{}

	reader := func(ch <-chan int) (*[]int, func()) {
		wg.Add(1)
		res := make([]int, 0, 100)
		return &res, func() {
			defer wg.Done()

			for val := range ch {
				res = append(res, val)
			}
		}
	}

	res1, goReader1 := reader(ch1)
	go goReader1()
	res2, goReader2 := reader(ch2)
	go goReader2()

	wg.Wait()

	if len(*res1) == 10 || len(*res2) == 10 {
		t.Errorf("unexpected results: res1(%v) && res2(%v)", res1, res2)
	}
}
