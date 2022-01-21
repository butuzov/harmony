# harmony  [![Coverage Status](https://coveralls.io/repos/github/butuzov/harmony/badge.svg?t=1njyDt)](https://coveralls.io/github/butuzov/harmony) [![build status](https://github.com/butuzov/harmony/actions/workflows/main.yaml/badge.svg?branch=main)]() [![MIT License](http://img.shields.io/badge/license-MIT-blue.svg)](http://www.opensource.org/licenses/MIT)

Generic Concurrency Patterns Library

## Reference (generated by [gomarkdoc](<https://github.com/princjef/gomarkdoc>))

<!-- You can Edit Content above this comment --->
<!-- Start --->
```go
import "github.com/butuzov/harmony"
```

Package `harmony` provides generic concurrency patterns library, created for educational proposes by it's author. It provides next patterns: 
- `Bridge` 
- `FanIn` 
- `Feature` 
- `OrWithDone` / `OrWithDone` 
- `Pipeline` 
- `Queue` 
- `Tee` 
- `WorkerPool`

Code generated by cmd/internal/gendone. DO NOT EDIT.n

<details><summary>Example (Fastest Sqrt)</summary>
<p>

Example.fastestSqrt dmonstrates combination of few techniques

```go
package main

import (
	"context"
	"fmt"
	"github.com/butuzov/harmony"
	"math/rand"
	"time"
)

func main() {
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
				// fmt.Printf("y=%f, x=%f.", y, x)
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

	ch := harmony.FututeWithContext(ctx, func() uint64 {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		v := r.Uint64()
		fmt.Printf("Initial number: %d.", v)
		return v
	})

	ch1, ch2 := harmony.TeeWithContext(ctx, ch)

	out := harmony.FanInWithContext(ctx,
		harmony.OrWithContext(ctx, harmony.PipelineWithContext(ctx, ch1, 1, sqrtBabylonian)),
		harmony.OrWithContext(ctx, harmony.PipelineWithContext(ctx, ch2, 1, sqrtBakhshali)),
	)
	fmt.Printf("Result is :%v", <-out)
}
```

</p>
</details>

### Index

- [func BridgeWithContext[T any](ctx context.Context, incoming <-chan (<-chan T)) <-chan T](<#func-bridgewithcontext>)
- [func BridgeWithDone[T any](done <-chan struct{}, incoming <-chan (<-chan T)) <-chan T](<#func-bridgewithdone>)
- [func FanInWithContext[T any](ctx context.Context, ch1, ch2 <-chan T, channels ...<-chan T) <-chan T](<#func-faninwithcontext>)
- [func FanInWithDone[T any](done <-chan struct{}, ch1, ch2 <-chan T, channels ...<-chan T) <-chan T](<#func-faninwithdone>)
- [func FututeWithContext[T any](ctx context.Context, futureFn func() T) <-chan T](<#func-fututewithcontext>)
- [func FututeWithDone[T any](done <-chan struct{}, futureFn func() T) <-chan T](<#func-fututewithdone>)
- [func OrWithContext[T any](ctx context.Context, incoming <-chan T) <-chan T](<#func-orwithcontext>)
- [func OrWithDone[T any](done <-chan struct{}, incoming <-chan T) <-chan T](<#func-orwithdone>)
- [func PipelineWithContext[T1, T2 any](ctx context.Context, incoming <-chan T1, totalWorkers int, workerFn func(T1) T2) <-chan T2](<#func-pipelinewithcontext>)
- [func PipelineWithDone[T1, T2 any](done <-chan struct{}, incoming <-chan T1, totalWorkers int, workerFn func(T1) T2) <-chan T2](<#func-pipelinewithdone>)
- [func QueueWithContext[T any](ctx context.Context, genFn func() T) <-chan T](<#func-queuewithcontext>)
- [func QueueWithDone[T any](done <-chan struct{}, genFn func() T) <-chan T](<#func-queuewithdone>)
- [func TeeWithContext[T any](ctx context.Context, incoming <-chan T) (<-chan T, <-chan T)](<#func-teewithcontext>)
- [func TeeWithDone[T any](done <-chan struct{}, incoming <-chan T) (<-chan T, <-chan T)](<#func-teewithdone>)
- [func WorkerPoolWithContext[T any](ctx context.Context, jobQueue chan T, maxWorkers int, workFunc func(T))](<#func-workerpoolwithcontext>)
- [func WorkerPoolWithDone[T any](done <-chan struct{}, jobQueue chan T, maxWorkers int, workFunc func(T))](<#func-workerpoolwithdone>)


### func BridgeWithContext

```go
func BridgeWithContext[T any](ctx context.Context, incoming <-chan (<-chan T)) <-chan T
```

BridgeWithContext will return chan of generic type `T` used a pipe for the values received from the sequence of channels. Close channel .received from `incoming`. in order to  switch for a new one. Goroutines exists on close of `incoming` or context canceled.

### func BridgeWithDone

```go
func BridgeWithDone[T any](done <-chan struct{}, incoming <-chan (<-chan T)) <-chan T
```

BridgeWithDone will return chan of generic type `T` used a pipe for the values received from the sequence of channels. Close channel .received from `incoming`. in order to  switch for a new one. Goroutines exists on close of `incoming` or done chan closed.

### func FanInWithContext

```go
func FanInWithContext[T any](ctx context.Context, ch1, ch2 <-chan T, channels ...<-chan T) <-chan T
```

FanInWithContext returns unbuffered channel of generic type `T` which serves as delivery pipeline for the values received from at least 2 incoming channels, its closed once all of the incoming channels closed or context cancelled.

<details><summary>Example</summary>
<p>


-
-
- Fan
-in Pattern Example 
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-

```go
package main

import (
	"context"
	"fmt"
	"github.com/butuzov/harmony"
)

func main() {
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
	for val := range harmony.FanInWithContext(ctx, ch1, ch2, ch3) {
		fmt.Println(val)
	}
}
```

</p>
</details>

### func FanInWithDone

```go
func FanInWithDone[T any](done <-chan struct{}, ch1, ch2 <-chan T, channels ...<-chan T) <-chan T
```

FanInWithDone returns unbuffered channel of generic type `T` which serves as delivery pipeline for the values received from at least 2 incoming channels, its closed once all of the incoming channels closed or done is closed

### func FututeWithContext

```go
func FututeWithContext[T any](ctx context.Context, futureFn func() T) <-chan T
```

FututeWithContext.T any. will return buffered channel of size 1 and generic type `T`, which will eventually contain the results of the execution `futureFn``, or be closed in case if context cancelled.

<details><summary>Example</summary>
<p>

```go
package main

import (
	"context"
	"fmt"
	"github.com/butuzov/harmony"
)

func main() {
	// Requests random dogs picture from dog.ceo (dog as service)
	ctx := context.Background()
	a := harmony.FututeWithContext(ctx, func() int { return 1 })
	b := harmony.FututeWithContext(ctx, func() int { return 0 })
	fmt.Println(<-a, <-b)
}
```

#### Output

```
1 0
```

</p>
</details>

<details><summary>Example (Dogs_as_service)</summary>
<p>

FututeWithContext is shows creation of two "futures" that are used in our "rate our dogs" startup.

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/butuzov/harmony"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func main() {
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

	a := harmony.FututeWithContext(context.Background(), func() string {
		return getRandomDogPicture()
	})

	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()
	b := harmony.FututeWithContext(ctx, func() string {
		return getRandomDogPicture()
	})
	fmt.Printf("Rate My Dog: ..a) %s..b) %s.", <-a, <-b)
}
```

</p>
</details>

### func FututeWithDone

```go
func FututeWithDone[T any](done <-chan struct{}, futureFn func() T) <-chan T
```

FututeWithDone.T any. will return buffered channel of size 1 and generic type `T`, which will eventually contain the results of the execution `futureFn``, or be closed in case if context cancelled.

### func OrWithContext

```go
func OrWithContext[T any](ctx context.Context, incoming <-chan T) <-chan T
```

OrWithDone will return a new unbuffered channel of type `T` that serves as a pipeline for the incoming channel. Channel is closed once the context is canceled or the incoming channel is closed. This is variation or the pattern that usually called `OrWithDone` or`Cancel`.

<details><summary>Example (&ibonacci)</summary>
<p>

ExampleOrWithContext 
- shows how many fibonacci sequence numbers we can generate in one millisecond.

```go
package main

import (
	"context"
	"fmt"
	"github.com/butuzov/harmony"
	"time"
)

func main() {
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

	for val := range harmony.OrWithContext(ctx, incoming) {
		results = append(results, val)
	}

	fmt.Println(results)
}
```

</p>
</details>

### func OrWithDone

```go
func OrWithDone[T any](done <-chan struct{}, incoming <-chan T) <-chan T
```

OrWithDone will return a new unbuffered channel of type `T` that serves as a pipeline for the incoming channel. Channel is closed once the context is canceled or the incoming channel is closed. This is variation or the pattern that usually called `OrWithDone` or`Cancel`.

### func PipelineWithContext

```go
func PipelineWithContext[T1, T2 any](ctx context.Context, incoming <-chan T1, totalWorkers int, workerFn func(T1) T2) <-chan T2
```

PipelineWithContext returns the channel of generic type `T2` that can serve as a pipeline for the next stage. It's implemented in almost same manner as a `WorkerPool` and allows to specify number of workers that going to proseed values received from the incoming channel. Outgoing channel is going to be closed once the incoming chan is closed or context canceld.

<details><summary>Example (0rimes)</summary>
<p>


-
-
- Pipeline Pattern Example 
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-
-

```go
package main

import (
	"context"
	"fmt"
	"github.com/butuzov/harmony"
	"math"
	"time"
)

func main() {
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

	for val := range harmony.PipelineWithContext(ctx, incomingCh, 100, workerFunc) {
		if val == 0 {
			continue
		}

		results = append(results, val)
	}

	fmt.Println(results)
}
```

</p>
</details>

### func PipelineWithDone

```go
func PipelineWithDone[T1, T2 any](done <-chan struct{}, incoming <-chan T1, totalWorkers int, workerFn func(T1) T2) <-chan T2
```

PipelineWithDone returns the channel of generic type `T2` that can serve as a pipeline for the next stage. It's implemented in almost same manner as a `WorkerPool` and allows to specify number of workers that going to proseed values received from the incoming channel. Outgoing channel is going to be closed once the incoming chan is closed or context canceld.

### func QueueWithContext

```go
func QueueWithContext[T any](ctx context.Context, genFn func() T) <-chan T
```

QueueWithContext returns an unbuffered channel that is populated by func `genFn`. Chan is closed once context is Done. It's similar to `Future` pattern, but doesn't have a limit to just one result.

<details><summary>Example</summary>
<p>

Generate fibonacci sequence

```go
package main

import (
	"context"
	"fmt"
	"github.com/butuzov/harmony"
)

func main() {
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
}
```

#### Output

```
[1 1 2 3 5 8 13 21 34 55]
```

</p>
</details>

### func QueueWithDone

```go
func QueueWithDone[T any](done <-chan struct{}, genFn func() T) <-chan T
```

QueueWithDone returns an unbuffered channel that is populated by func `genFn`. Chan is closed once context is Done. It's similar to `Future` pattern, but doesn't have a limit to just one result.

### func TeeWithContext

```go
func TeeWithContext[T any](ctx context.Context, incoming <-chan T) (<-chan T, <-chan T)
```

TeeWithContext will return two channels of generic type `T` used to fan
-out data from the incoming channel. Channels needs to be read in order next iteration over incoming chanel happen.

### func TeeWithDone

```go
func TeeWithDone[T any](done <-chan struct{}, incoming <-chan T) (<-chan T, <-chan T)
```

TeeWithDone will return two channels of generic type `T` used to fan
-out data from the incoming channel. Channels needs to be read in order next iteration over incoming chanel happen.

<details><summary>Example</summary>
<p>

```go
package main

import (
	"fmt"
	"github.com/butuzov/harmony"
	"sync"
)

func main() {
	done := make(chan struct{})
	pipe := make(chan int)

	ch1, ch2 := harmony.TeeWithDone(done, pipe)

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
}
```

#### Output

```
Sequence sum is 55. Sequence product is 3628800
```

</p>
</details>

### func WorkerPoolWithContext

```go
func WorkerPoolWithContext[T any](ctx context.Context, jobQueue chan T, maxWorkers int, workFunc func(T))
```

WorkerPoolWithContext accepts channel of generic type `T` which is used to serve jobs to max workersTotal workers. Goroutines stop: Once channel closed and drained, or context cancelled.

<details><summary>Example (0rimes)</summary>
<p>

```go
package main

import (
	"context"
	"fmt"
	"github.com/butuzov/harmony"
	"math"
	"runtime"
	"sync"
	"time"
)

func main() {
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
```

</p>
</details>

### func WorkerPoolWithDone

```go
func WorkerPoolWithDone[T any](done <-chan struct{}, jobQueue chan T, maxWorkers int, workFunc func(T))
```

WorkerPoolWithDone accepts channel of generic type `T` which is used to serve jobs to max workersTotal workers. Goroutines stop: Once channel closed and drained, or done is closed

<!-- End --->
<!-- You can Edit Content under this comment --->

## Resources

* `talk` [Bryan C. Mills - Rethinking Classical Concurrency Patterns](https://www.youtube.com/watch?v=5zXAHh5tJqQ) + [`slides`](https://drive.google.com/file/d/1nPdvhB0PutEJzdCq5ms6UI58dp50fcAN/view) + [`comments+notes`](https://github.com/sourcegraph/gophercon-2018-liveblog/issues/35)
* `book` [Katherine Cox-Buday - Concurrency In Go](https://www.oreilly.com/library/view/concurrency-in-go/9781491941294/)
* `talk` [Rob Pike - Concurrency is not Parallelism](https://www.youtube.com/watch?v=oV9rvDllKEg) + [`slides`](https://go.dev/talks/2012/waza.slide)
* `blog` [Go Concurrency Patterns: Context](https://go.dev/blog/context)
* `blog` [Go Concurrency Patterns: Pipelines and cancellation](https://go.dev/blog/pipelines)
* `talk` [Sameer Ajmani  - Advanced Go Concurrency Patterns](https://www.youtube.com/watch?v=QDDwwePbDtw) + [`slides`](https://talks.golang.org/2013/advconc.slide)
* `talk` [Rob Pike - Go Concurrency Patterns](https://www.youtube.com/watch?v=f6kdp27TYZs) + [`slides`](https://talks.golang.org/2012/concurrency.slide)
* `blog` [Go Concurrency Patterns: Timing out, moving on](https://go.dev/blog/concurrency-timeouts)
