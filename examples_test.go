//go:generate go run ./cmd/internal/gendone/
//go:generate make doc
package harmony_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/butuzov/harmony"
)

func ExampleFanInWithContext() {
	// return channel that generate
	filler := func(start, stop int) chan int {
		ch := make(chan int)

		go func() {
			defer close(ch)
			for i := start; i <= stop; i++ {
				ch <- i
			}
		}()

		return ch
	}

	ch1 := filler(10, 12)
	ch2 := filler(12, 14)
	ch3 := filler(15, 16)

	ctx := context.Background()
	if ch, err := harmony.FanInWithContext(ctx, ch1, ch2, ch3); err != nil {
		for val := range ch {
			fmt.Println(val)
		}
	}
}

// FututeWithContext is shows creation of two "futures" that are used in our
// "rate our dogs" startup.
func ExampleFututeWithContext_dogs_as_service() {
	// Requests random dogs picture from dog.ceo (dog as service)
	getRandomDogPicture := func() string {
		var data struct {
			Message string "json:'message'"
		}

		const API_URL = "https://dog.ceo/api/breeds/image/random"
		ctx := context.Background()

		if req, err := http.NewRequestWithContext(ctx, http.MethodGet, API_URL, nil); err != nil {
			log.Println(fmt.Errorf("request: %w", err))
			return ""
		} else if res, err := http.DefaultClient.Do(req); err != nil {
			log.Println(fmt.Errorf("request: %w", err))
			return ""
		} else {
			defer res.Body.Close()

			if body, err := ioutil.ReadAll(res.Body); err != nil {
				log.Println(fmt.Errorf("reading body: %w", err))
				return ""
			} else if err := json.Unmarshal(body, &data); err != nil {
				log.Println(fmt.Errorf("unmarshal: %w", err))
				return ""
			}
		}

		return data.Message
	}

	a, _ := harmony.FututeWithContext(context.Background(), func() string {
		return getRandomDogPicture()
	})

	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()
	b, _ := harmony.FututeWithContext(ctx, func() string {
		return getRandomDogPicture()
	})
	fmt.Printf("Rate My Dog: \n\ta) %s\n\tb) %s\n", <-a, <-b)
}

func ExampleFututeWithContext() {
	// Requests random dogs picture from dog.ceo (dog as service)
	ctx := context.Background()
	a, _ := harmony.FututeWithContext(ctx, func() int { return 1 })
	b, _ := harmony.FututeWithContext(ctx, func() int { return 0 })
	fmt.Println(<-a, <-b)
	// Output: 1 0
}

// ExampleOrWithContext - shows how many fibonacci sequence numbers we can
// generate in one millisecond.
func ExampleOrWithContext_Fibonacci() {
	var (
		incoming = make(chan int)
		results  []int
		fib      func(int) int
	)

	fib = func(n int) int {
		if n < 2 {
			return n
		}
		return fib(n-2) + fib(n-1)
	}

	// producer
	go func() {
		i := 0
		for {
			i++
			incoming <- fib(i)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	if ch, err := harmony.OrWithContext(ctx, incoming); err == nil {
		for val := range ch {
			results = append(results, val)
		}
		fmt.Println(results)
	} else {
		log.Printf("Error: %v", err)
	}
}

func ExamplePipelineWithContext_Primes() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	var (
		incomingCh = make(chan uint64)
		isPrime    = func(n uint64) bool {
			for i := uint64(2); i < (n/2)+1; i++ {
				if n%i == 0 {
					return false
				}
			}
			return true
		}
	)

	var results []uint64
	workerFunc := func(n uint64) uint64 {
		if isPrime(n) {
			return n
		}
		return 0
	}

	// Producer: Initial numbers
	go func() {
		for i := uint64(0); i < math.MaxUint64; i++ {
			incomingCh <- i
		}
	}()

	if ch, err := harmony.PipelineWithContext(ctx, incomingCh, 100, workerFunc); err != nil {
		log.Printf("Error: %v", err)
	} else {
		for val := range ch {
			if val == 0 {
				continue
			}
			results = append(results, val)
		}
		fmt.Println(results)
	}
}

// Generate fibonacci sequence
func ExampleQueueWithContext() {
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
	incoming, err := harmony.QueueWithContext(context.Background(), fib(10))
	if err != nil {
		log.Printf("err: %v", err)
		return
	}

	for i := 0; i < cap(first10FibNumbers); i++ {
		first10FibNumbers[i] = <-incoming
	}

	fmt.Println(first10FibNumbers)
	// Output: [1 1 2 3 5 8 13 21 34 55]
}

func ExampleTeeWithDone() {
	done := make(chan struct{})
	pipe := make(chan int)

	ch1, ch2, _ := harmony.TeeWithDone(done, pipe)

	// generator
	go func() {
		defer close(pipe)
		for i := 1; i <= 10; i++ {
			pipe <- i
		}
	}()

	// consumers
	consumer := func(ch <-chan int, wg *sync.WaitGroup, consum func(i int)) {
		defer wg.Done()

		for k := range ch {
			consum(k)
		}
	}

	sum, prod := 0, 1
	wg := sync.WaitGroup{}
	wg.Add(2)

	go consumer(ch1, &wg, func(i int) { sum += i })  // Sum
	go consumer(ch2, &wg, func(i int) { prod *= i }) // Product/Factorial

	wg.Wait()
	fmt.Printf("Sequence sum is %d. Sequence product is %d", sum, prod)
	// Output: Sequence sum is 55. Sequence product is 3628800
}

func ExampleWorkerPoolWithContext_Primes() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	var (
		primesCh   = make(chan uint64)
		incomingCh = make(chan uint64)
		isPrime    = func(n uint64) bool {
			for i := uint64(2); i < (n/2)+1; i++ {
				if n%i == 0 {
					return false
				}
			}
			return true
		}
		totalWorkers = runtime.NumCPU() - 1
	)

	// Producer: Initial numbers
	go func() {
		for i := uint64(0); i < math.MaxUint64; i++ {
			incomingCh <- i
		}
	}()

	// Consumers Worker Pool: checking primes of incoming numbers.
	harmony.WorkerPoolWithContext(ctx, incomingCh, totalWorkers, func(n uint64) {
		if !isPrime(n) {
			return
		}
		primesCh <- n
	})

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

	mu.RLock()
	fmt.Println(results)
	mu.RUnlock()
}

// Other Examples

// What SQRT funtion is faster? Complex example that shows the combination of
// few patterns Queue, Tee, FanIn patterns.
func Example_fastestSqrt() {
	// the fastert square root cracker....
	type Report struct {
		Method string
		Value  uint64
	}

	var (
		// Babylonian method
		sqrtBabylonian = func(n uint64) Report {
			var (
				o = float64(n) // Original value as float64
				x = float64(n) // x of binary search
				y = 1.0        // y of binary search
				e = 1e-5       // error
			)

			for x-y > e {
				x = (x + y) / 2
				y = o / x
				// fmt.Printf("y=%f, x=%f\n", y, x)
			}

			return Report{"Babylonian", uint64(x)}
		}

		// Bakhshali method
		sqrtBakhshali = func(n uint64) Report {
			iterate := func(x float64) float64 {
				a := (float64(n) - x*x) / (2 * x)
				xa := x + a
				return xa - ((a * a) / (2 * xa))
			}

			var (
				o = float64(n)
				x = float64(n) / 2.0
				e = 1e-5
			)

			for x*x-o > e {
				x = iterate(x)
			}
			return Report{"Bakhshali", uint64(x)}
		}
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch, _ := harmony.FututeWithContext(ctx, func() uint64 {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		v := r.Uint64()
		fmt.Printf("Initial number: %d\n", v)
		return v
	})

	if ch1, ch2, err := harmony.TeeWithContext(ctx, ch); err == nil {
		log.Printf("err: %v", err)
		return
	} else {
		chRep1, _ := harmony.PipelineWithContext(ctx, ch1, 1, sqrtBabylonian)
		chRep2, _ := harmony.PipelineWithContext(ctx, ch2, 1, sqrtBakhshali)

		chRep1, _ = harmony.OrWithContext(ctx, chRep1)
		chRep2, _ = harmony.OrWithContext(ctx, chRep2)

		out, _ := harmony.FanInWithContext(ctx, chRep1, chRep2)
		fmt.Printf("Result is :%v", <-out)
	}
}
