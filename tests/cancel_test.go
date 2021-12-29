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

func ExampleOrWithDone() {
	var (
		done     = make(chan struct{})
		incoming = make(chan int)
		outgoing = harmony.OrWithDone(done, incoming)
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

	for val := range outgoing {
		results = append(results, val)
	}

	<-done

	fmt.Println(results)
}

func TestOrWithDone(t *testing.T) {
	var (
		done     = make(chan struct{})
		incoming = make(chan int)
		outgoing = harmony.OrWithDone(done, incoming)
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

func TestOrWithDone_WithDone(t *testing.T) {
	var (
		done     = make(chan struct{})
		incoming = make(chan int)
		outgoing = harmony.OrWithDone(done, incoming)
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

func TestOrWithContext(t *testing.T) {
	incoming := make(chan int)
	go func() {
		defer close(incoming)
		for i := 0; i < 10; i++ {
			incoming <- i
		}
	}()

	// Creating cancelable pipeline for incoming chan.
	chOut := harmony.OrWithContext(context.Background(), incoming)
	var res []int
	want := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	for val := range chOut {
		res = append(res, val)
	}

	if !reflect.DeepEqual(res, want) {
		t.Errorf("want %v vs got %v", want, res)
	}
}

func TestOrWithContext_WithContext(t *testing.T) {
	// close done after 0.001 sec... (approx ~1k channel messages)
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	t.Cleanup(cancel)

	var (
		incoming = make(chan int)
		outgoing = harmony.OrWithContext(ctx, incoming)
		limit    = 25_000
	)

	// producer
	go func() {
		defer close(incoming)
		for i := 1; i <= limit; i++ {
			incoming <- i
		}
	}()

	var lastRead int
	for val := range outgoing {
		lastRead = val
	}

	if lastRead == limit { // Last Read is limit, if we not close done.
		t.Errorf("last(%d) vs want(%d)", lastRead, limit)
	}
}
