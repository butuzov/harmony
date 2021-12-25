package harmony_test

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/butuzov/harmony"
)

func TestWorkerPool(t *testing.T) {
	resCh := make(chan string)
	jobFn := func(s string) {
		time.Sleep(time.Duration(rand.Int63n(10)) * time.Millisecond)
		resCh <- strings.ToUpper(s)
	}

	jobsQueue := harmony.WorkerPool[string](3, jobFn)

	// assigning jobs
	input := []string{"foo", "bar", "baz", "boom", "faz"}
	go func() {
		for _, word := range input {
			jobsQueue <- word
		}
	}()

	var res []string
	for i := 0; i < len(input); i++ {
		res = append(res, <-resCh)
	}

	sort.Strings(res)

	want := []string{"BAR", "BAZ", "BOOM", "FAZ", "FOO"}
	if !reflect.DeepEqual(res, want) {
		t.Errorf("got(%v) vs want(%v)\n", res, want)
	}
}

func TestWorkerPoolWithContext(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	t.Cleanup(cancel)

	jobQueue := harmony.WorkerPoolWithContext(ctx, 100, func(n int) {
		time.Sleep(10 * time.Millisecond)
	})

	var done int32
	go func() {
		defer func() {
			atomic.StoreInt32(&done, 1)
		}()
		for i := 0; i < 1000; i++ {
			jobQueue <- i
		}
	}()

	// 100 goroutines will "sleep" (their work function) for 10 milliseconds each
	// in total we have 1000 jobs which is approx 10 jobs per goroutine
	// (100 milliseconds in total)

	<-ctx.Done()
	if atomic.LoadInt32(&done) == 0 {
		t.Error("gorotines didn't make it to finish")
	}
}

func ExampleWorkerPoolWithContext() {
	// Search for all possible primes within short period of time.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	var (
		primesCh = make(chan uint64)
		isPrime  = func(n uint64) bool {
			for i := uint64(2); i < (n/2)+1; i++ {
				if n%i == 0 {
					return false
				}
			}
			return true
		}
		totalWorkers = runtime.NumCPU() - 1
	)

	jobsQueue := harmony.WorkerPoolWithContext(ctx, totalWorkers, func(n uint64) {
		if !isPrime(n) {
			return
		}

		primesCh <- n
	})

	go func() {
		for i := uint64(0); i < math.MaxUint64; i++ {
			jobsQueue <- i
		}
	}()

	var results []uint64
	var mu sync.RWMutex
	go func() {
		for n := range primesCh {
			mu.Lock()
			results = append(results, n)
			mu.Unlock()
		}
	}()

	<-ctx.Done()
	close(primesCh)

	mu.RLock()
	fmt.Println(results)
	mu.RUnlock()
}
