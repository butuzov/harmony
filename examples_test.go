//go:generate go run ./cmd/internal/gendone/
//go:generate make doc
package harmony_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/butuzov/harmony"
)

// --- Bridge Pattern Example --------------------------------------------------

// --- Fan-in Pattern Example --------------------------------------------------

// --- Future Pattern Example --------------------------------------------------

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

	a := harmony.FututeWithContext(context.Background(), func() string {
		return getRandomDogPicture()
	})

	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()
	b := harmony.FututeWithContext(ctx, func() string {
		return getRandomDogPicture()
	})
	fmt.Printf("Rate My Dog: \n\ta) %s\n\tb) %s\n", <-a, <-b)
}

func ExampleFututeWithContext() {
	// Requests random dogs picture from dog.ceo (dog as service)
	ctx := context.Background()
	a := harmony.FututeWithContext(ctx, func() int { return 1 })
	b := harmony.FututeWithContext(ctx, func() int { return 0 })
	fmt.Println(<-a, <-b)
	// Output: 1 0
}

// --- OrDone Pattern Example --------------------------------------------------

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

	for val := range harmony.OrWithContext(ctx, incoming) {
		results = append(results, val)
	}

	fmt.Println(results)
}

// --- Pipeline Pattern Example ------------------------------------------------

// --- Queue Pattern Example ---------------------------------------------------

// --- Tee Pattern Example -----------------------------------------------------

func ExampleTeeWithDone() {
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

	go consumer(ch1, &wg, func(i int) { sum += i })
	go consumer(ch2, &wg, func(i int) { prod *= i })

	wg.Wait()
	fmt.Printf("Sequence sum is %d. Sequence product is %d", sum, prod)
	// Output: Sequence sum is 55. Sequence product is 3628800
}

// --- WorkerPool Pattern  Example ---------------------------------------------
