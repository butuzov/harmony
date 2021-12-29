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
- `OrDone` / `OrContextDone` 
- `Pipeline` 
- `Queue` 
- `Tee` 
- `WorkerPool`

### Index

- [func BridgeWithContext[T any](ctx context.Context, incoming <-chan (<-chan T)) <-chan T](<#func-bridgewithcontext>)
- [func FanIn[T any](ch1, ch2 <-chan T, channels ...<-chan T) <-chan T](<#func-fanin>)
- [func FanInWithContext[T any](ctx context.Context, ch1, ch2 <-chan T, channels ...<-chan T) <-chan T](<#func-faninwithcontext>)
- [func Futute[T any](futureFn func() T) <-chan T](<#func-futute>)
- [func FututeWithContext[T any](ctx context.Context, futureFn func() T) <-chan T](<#func-fututewithcontext>)
- [func OrContextDone[T any](ctx context.Context, incoming <-chan T) <-chan T](<#func-orcontextdone>)
- [func OrDone[T any](done chan struct{}, incoming <-chan T) <-chan T](<#func-ordone>)
- [func PipelineWithContext[T any](ctx context.Context, incoming <-chan T, totalWorkers int, workerFn func(T) T) <-chan T](<#func-pipelinewithcontext>)
- [func PipelineWithDone[T any](done chan struct{}, incoming <-chan T, totalWorkers int, workerFn func(T) T) <-chan T](<#func-pipelinewithdone>)
- [func Queue[T any](genFn func() T) <-chan T](<#func-queue>)
- [func QueueWithContext[T any](ctx context.Context, genFn func() T) <-chan T](<#func-queuewithcontext>)
- [func QueueWithDone[T any](done chan struct{}, genFn func() T) <-chan T](<#func-queuewithdone>)
- [func TeeWithContext[T any](ctx context.Context, incoming <-chan T) (<-chan T, <-chan T)](<#func-teewithcontext>)
- [func TeeWithDone[T any](done chan struct{}, incoming <-chan T) (<-chan T, <-chan T)](<#func-teewithdone>)
- [func WorkerPoolWithContext[T any](ctx context.Context, totalWorkers int, workerFn func(T)) chan<- T](<#func-workerpoolwithcontext>)
- [func WorkerPoolWithDone[T any](done chan struct{}, totalWorkers int, workerFn func(T)) chan<- T](<#func-workerpoolwithdone>)


### func BridgeWithContext

```go
func BridgeWithContext[T any](ctx context.Context, incoming <-chan (<-chan T)) <-chan T
```

BridgeWithContext will return chan of generic type `T` used a pipe for the values received from the sequence of channels. Close channel .received from `incoming`. in order to  switch for a new one. Goroutines exists on close of `incoming` or context canceled.

### func FanIn

```go
func FanIn[T any](ch1, ch2 <-chan T, channels ...<-chan T) <-chan T
```

FanIn returns unbuffered channel of generic type `T` which is serving as a delivery pipeline for the values received from at least 2 incoming channels, it is closed once all of the incoming channels are closed.

### func FanInWithContext

```go
func FanInWithContext[T any](ctx context.Context, ch1, ch2 <-chan T, channels ...<-chan T) <-chan T
```

FanInWithContext returns unbuffered channel of generic type `T` which serves as delivery pipeline for the values received from at least 2 incoming channels, its closed once all of the incoming channels closed or context cancelled.

<details><summary>Example</summary>
<p>

```go
package main

import (
	"context"
	"fmt"
	"github.com/butuzov/harmony"
	"sync"
	"time"
)

func main() {
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
}
```

#### Output

```
[0 1 2 3 4 5 6]
```

</p>
</details>

### func Futute

```go
func Futute[T any](futureFn func() T) <-chan T
```

Future will return buffered channel of size 1 and generic type `T` which eventually will contain the result of the execution futureFn

### func FututeWithContext

```go
func FututeWithContext[T any](ctx context.Context, futureFn func() T) <-chan T
```

FututeWithContext.T any. will return buffered channel of size 1 and generic type `T`, which will eventually contain the results of the execution futureFn, or be closed in case if context cancelled.

### func OrContextDone

```go
func OrContextDone[T any](ctx context.Context, incoming <-chan T) <-chan T
```

OrContextDone will return a new unbuffered channel of type `T` that serves as a pipeline for the incoming channel. Channel is closed once the context is canceled or the incoming channel is closed. This is variation or the pattern that usually called `OrDone` or`Cancel`.

<details><summary>Example</summary>
<p>

```go
package main

import (
	"fmt"
	"github.com/butuzov/harmony"
	"time"
)

func main() {
	var (
		done     = make(chan struct{})
		incoming = make(chan int)
		outgoing = harmony.OrDone(done, incoming)
		results  []int
	)

	time.AfterFunc(5*time.Nanosecond, func() { close(done) })

	// producer
	go func() {
		defer close(incoming)
		for i := 1; i < 100; i++ {
			time.Sleep(time.Millisecond)
			incoming <- i
		}
	}()

	// consumer
	for val := range outgoing {
		results = append(results, val)
		// We going to cancel execution once we reach any number devisable by 7
	}

	<-done

	fmt.Println(results)
}
```

</p>
</details>

### func OrDone

```go
func OrDone[T any](done chan struct{}, incoming <-chan T) <-chan T
```

OrDone will return a new unbuffered channel of type `T` that serves as a pipeline for the values from the incoming channel. Channel is closed once the done chan is closed or the incoming channel is closed. This pattern usually called `OrDone` or `Cancel`.

### func PipelineWithContext

```go
func PipelineWithContext[T any](ctx context.Context, incoming <-chan T, totalWorkers int, workerFn func(T) T) <-chan T
```

PipelineWithContext returns the channel of generic type `T` that can serve as a pipeline for the next stage. It's implemented in same manner as a `WorkerPool` and allows to specify number of workers that going to proseed values received from the incoming channel. Outgoing channel is going to be closed once the incoming chan is closed or context canceld.

<details><summary>Example</summary>
<p>

```go
package main

import (
	"context"
	"fmt"
	"github.com/butuzov/harmony"
	"sort"
)

func main() {
	lim := 10

	seedCh := func() <-chan int {
		ch := make(chan int)

		go func() {
			defer close(ch)

			for i := 0; i <= lim; i++ {
				ch <- i
			}
		}()

		return ch
	}()

	ctx := context.Background()

	var pipe <-chan int

	// we starting many subWorkers because we don't care about order of results
	// read from the pipe.
	pipe = seedCh
	// does nothing
	pipe = harmony.PipelineWithContext(ctx, pipe, 2, func(n int) int { return n })
	// does multiplication
	pipe = harmony.PipelineWithContext(ctx, pipe, 3, func(n int) int { return n * 2 })
	// does addition
	pipe = harmony.PipelineWithContext(ctx, pipe, 3, func(n int) int { return n + 2 })

	result := []int{}
	for val := range pipe {
		result = append(result, val)
	}

	sort.Ints(result)
	fmt.Printf("%v", result)
}
```

#### Output

```
[2 4 6 8 10 12 14 16 18 20 22]
```

</p>
</details>

### func PipelineWithDone

```go
func PipelineWithDone[T any](done chan struct{}, incoming <-chan T, totalWorkers int, workerFn func(T) T) <-chan T
```

PipelineWithDone returns the channel of generic type `T` that can serve as a pipeline for a next stage. It's implemented in same manner as a `WorkerPool` and allows to specify number of workers that going to proseed values received from the incoming channel. The returned channel is going to be closed once one of incoming or done channels are closed.

<details><summary>Example</summary>
<p>

```go
package main

import (
	"fmt"
	"github.com/butuzov/harmony"
	"sort"
)

func main() {
	lim := 10

	seedCh := func() <-chan int {
		ch := make(chan int)

		go func() {
			defer close(ch)

			for i := 0; i <= lim; i++ {
				ch <- i
			}
		}()

		return ch
	}()

	done := make(chan struct{})
	defer close(done)

	var pipe <-chan int

	// we starting many subWorkers because we don't care about order of results
	// read from the pipe.
	pipe = seedCh
	// does nothing
	pipe = harmony.PipelineWithDone(done, pipe, 2, func(n int) int { return n })
	// does multiplication
	pipe = harmony.PipelineWithDone(done, pipe, 3, func(n int) int { return n * 2 })
	// does addition
	pipe = harmony.PipelineWithDone(done, pipe, 3, func(n int) int { return n + 2 })

	result := []int{}
	for val := range pipe {
		result = append(result, val)
	}

	sort.Ints(result)
	fmt.Printf("%v", result)
}
```

#### Output

```
[2 4 6 8 10 12 14 16 18 20 22]
```

</p>
</details>

### func Queue

```go
func Queue[T any](genFn func() T) <-chan T
```

Queue returns an unbuffered channel that is populated by func `genFn`. Chan is closed once context is Done. It's similar to `Future` pattern, but doesn't have a limit to just one result. Also: it's leaking gorotine.

### func QueueWithContext

```go
func QueueWithContext[T any](ctx context.Context, genFn func() T) <-chan T
```

QueueWithContext returns an unbuffered channel that is populated by func `genFn`. Chan is closed once context is Done. It's similar to `Future` pattern, but doesn't have a limit to just one result.

### func QueueWithDone

```go
func QueueWithDone[T any](done chan struct{}, genFn func() T) <-chan T
```

QueueWithContext returns an unbuffered channel that is populated by the func `genFn`. Chan is closed once done chan is closed. It's similar to `Future` pattern, but doesn't have a limit to just one result.

<details><summary>Example</summary>
<p>

```go
package main

import (
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
	incoming := harmony.Queue(fib(10))
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

### func TeeWithContext

```go
func TeeWithContext[T any](ctx context.Context, incoming <-chan T) (<-chan T, <-chan T)
```

TeeWithContext will return two channels of generic type `T` used to fan
-out data from the incoming channel. Channels needs to be read in order next iteration over incoming chanel happen.

### func TeeWithDone

```go
func TeeWithDone[T any](done chan struct{}, incoming <-chan T) (<-chan T, <-chan T)
```

TeeWithContext will return two channels of generic type `T` used to fan
-out data from the incoming channel. Channels needs to be read in order next iteration over incoming chanel happen.

### func WorkerPoolWithContext

```go
func WorkerPoolWithContext[T any](ctx context.Context, totalWorkers int, workerFn func(T)) chan<- T
```

WorkerPoolWithContext returns channel of generic type `T` which excepts jobs of the same type for some number of workers that do workerFn. If you want to stop WorkerPool, close the jobQueue channel or cancel the context.

<details><summary>Example</summary>
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
```

</p>
</details>

### func WorkerPoolWithDone

```go
func WorkerPoolWithDone[T any](done chan struct{}, totalWorkers int, workerFn func(T)) chan<- T
```

WorkerPoolWithDone returns channel of generic type `T` which excepts jobs of the same type for some number of workers that do workerFn. If you want to stop WorkerPool, close the jobQueue channel or close a `done` chan.

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
