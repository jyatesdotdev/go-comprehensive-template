package concurrency_test

import (
	"context"
	"fmt"

	"github.com/example/go-template/internal/concurrency"
)

func ExampleSafeMap() {
	m := concurrency.NewSafeMap[string, int]()
	m.Set("x", 42)
	v, ok := m.Get("x")
	fmt.Println(v, ok)
	fmt.Println(m.Len())
	// Output:
	// 42 true
	// 1
}

func ExampleSemaphore() {
	sem := concurrency.NewSemaphore(2)
	ctx := context.Background()

	_ = sem.Acquire(ctx)
	_ = sem.Acquire(ctx)
	fmt.Println("acquired 2 slots")

	sem.Release()
	fmt.Println("released 1 slot")

	_ = sem.Acquire(ctx)
	fmt.Println("re-acquired")
	// Output:
	// acquired 2 slots
	// released 1 slot
	// re-acquired
}

func ExampleWorkerPool() {
	ctx := context.Background()
	jobs := make(chan int, 3)
	for _, v := range []int{10, 20, 30} {
		jobs <- v
	}
	close(jobs)

	results := concurrency.WorkerPool(ctx, 2, jobs, func(_ context.Context, n int) int {
		return n * 2
	})

	sum := 0
	for r := range results {
		sum += r
	}
	fmt.Println(sum)
	// Output:
	// 120
}
