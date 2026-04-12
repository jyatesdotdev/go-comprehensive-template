// Example concurrency demonstrates goroutines, channels, select, context, and sync patterns.
package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/example/go-template/internal/concurrency"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	workerPoolExample(ctx)
	fanInExample(ctx)
	selectExample()
	contextCancellation()
	syncOnceExample()
	semaphoreExample(ctx)
	safeMapExample()
}

// workerPoolExample shows a pool of workers processing jobs concurrently.
func workerPoolExample(ctx context.Context) {
	fmt.Println("=== Worker Pool ===")
	jobs := make(chan int, 10)
	go func() {
		for i := range 10 {
			jobs <- i
		}
		close(jobs)
	}()

	results := concurrency.WorkerPool(ctx, 3, jobs, func(_ context.Context, n int) int {
		time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond) // #nosec G404 -- non-cryptographic jitter for demo
		return n * n
	})

	for r := range results {
		fmt.Printf("  result: %d\n", r)
	}
}

// fanInExample merges output from multiple goroutines.
func fanInExample(ctx context.Context) {
	fmt.Println("=== Fan-In ===")
	gen := func(prefix string, count int) <-chan string {
		ch := make(chan string)
		go func() {
			defer close(ch)
			for i := range count {
				ch <- fmt.Sprintf("%s-%d", prefix, i)
			}
		}()
		return ch
	}

	merged := concurrency.FanIn(ctx, gen("A", 3), gen("B", 3))
	for v := range merged {
		fmt.Printf("  %s\n", v)
	}
}

// selectExample demonstrates multiplexing with select.
func selectExample() {
	fmt.Println("=== Select ===")
	ch1 := make(chan string, 1)
	ch2 := make(chan string, 1)
	ch1 <- "from ch1"
	ch2 <- "from ch2"

	// Non-blocking select with default
	select {
	case v := <-ch1:
		fmt.Printf("  received: %s\n", v)
	case v := <-ch2:
		fmt.Printf("  received: %s\n", v)
	default:
		fmt.Println("  no data ready")
	}

	// Timeout pattern
	slow := make(chan string)
	go func() {
		time.Sleep(100 * time.Millisecond)
		slow <- "slow result"
	}()
	select {
	case v := <-slow:
		fmt.Printf("  got: %s\n", v)
	case <-time.After(50 * time.Millisecond):
		fmt.Println("  timed out waiting for slow channel")
	}
}

// contextCancellation shows parent-child context cancellation propagation.
func contextCancellation() {
	fmt.Println("=== Context Cancellation ===")
	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		select {
		case <-ctx.Done():
			fmt.Printf("  worker stopped: %v\n", ctx.Err())
		case <-time.After(5 * time.Second):
			fmt.Println("  worker finished")
		}
	}()

	cancel() // signal the worker to stop
	wg.Wait()
}

// syncOnceExample shows one-time initialization.
func syncOnceExample() {
	fmt.Println("=== sync.Once ===")
	var once sync.Once
	var wg sync.WaitGroup
	for range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			once.Do(func() { fmt.Println("  initialized (runs once)") })
		}()
	}
	wg.Wait()
}

// semaphoreExample limits concurrency to 2 at a time.
func semaphoreExample(ctx context.Context) {
	fmt.Println("=== Semaphore ===")
	sem := concurrency.NewSemaphore(2)
	var wg sync.WaitGroup
	for i := range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := sem.Acquire(ctx); err != nil {
				return
			}
			defer sem.Release()
			fmt.Printf("  worker %d running\n", i)
			time.Sleep(30 * time.Millisecond)
		}()
	}
	wg.Wait()
}

// safeMapExample demonstrates concurrent map access.
func safeMapExample() {
	fmt.Println("=== SafeMap ===")
	m := concurrency.NewSafeMap[string, int]()
	var wg sync.WaitGroup
	for i := range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m.Set(fmt.Sprintf("key-%d", i), i*10)
		}()
	}
	wg.Wait()
	fmt.Printf("  map size: %d\n", m.Len())
}
