package harmony_test

import (
	"context"
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/butuzov/harmony"
)

func ExampleOrContextDone() {
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

func TestOrDone(t *testing.T) {
	var (
		done     = make(chan struct{})
		incoming = make(chan int)
		outgoing = harmony.OrDone(done, incoming)
		limit    = rand.Intn(25_000) // rand to make it shorter.
	)

	// producer
	go func() {
		defer close(incoming)
		for i := 1; i <= limit; i++ {
			incoming <- i
		}
	}()

	// consumer
	var lastRead int
	for val := range outgoing {
		lastRead = val
	}

	if lastRead != limit {
		t.Errorf("last(%d) vs want(%d)", lastRead, limit)
	}
}

func TestOrDone_WithDone(t *testing.T) {
	var (
		done     = make(chan struct{})
		incoming = make(chan int)
		outgoing = harmony.OrDone(done, incoming)
		limit    = 25_000
	)

	// producer
	go func() {
		defer close(incoming)
		for i := 1; i <= limit; i++ {
			incoming <- i
		}
	}()

	// close done after 0.001 sec... (approx ~1k channel messages)
	timer := time.AfterFunc(time.Millisecond, func() { close(done) })
	t.Cleanup(func() { timer.Stop() })

	var lastRead int
	for val := range outgoing {
		lastRead = val
	}

	if lastRead == limit { // Last Read is limit, if we not close done.
		t.Errorf("last(%d) vs want(%d)", lastRead, limit)
	}
}

func TestOrContextDone(t *testing.T) {
	incoming := make(chan int)
	go func() {
		defer close(incoming)
		for i := 0; i < 10; i++ {
			incoming <- i
		}
	}()

	// Creating cancelable pipeline for incoming chan.
	chOut := harmony.OrContextDone(context.Background(), incoming)
	var res []int
	want := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	for val := range chOut {
		res = append(res, val)
	}

	if !reflect.DeepEqual(res, want) {
		t.Errorf("want %v vs got %v", want, res)
	}
}

func TestOrContextDone_WithContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	var (
		incoming = make(chan int)
		outgoing = harmony.OrContextDone(ctx, incoming)
		limit    = 25_000
	)

	// producer
	go func() {
		defer close(incoming)
		for i := 1; i <= limit; i++ {
			incoming <- i
		}
	}()

	// close done after 0.001 sec... (approx ~1k channel messages)
	timer := time.AfterFunc(time.Millisecond, func() { cancel() })
	t.Cleanup(func() { timer.Stop() })

	var lastRead int
	for val := range outgoing {
		lastRead = val
	}

	if lastRead == limit { // Last Read is limit, if we not close done.
		t.Errorf("last(%d) vs want(%d)", lastRead, limit)
	}
}
