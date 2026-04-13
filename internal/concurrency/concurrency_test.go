package concurrency

import (
	"context"
	"sync"
	"testing"
)

func TestWorkerPool(t *testing.T) {
	ctx := context.Background()
	jobs := make(chan int, 5)
	for i := 1; i <= 5; i++ {
		jobs <- i
	}
	close(jobs)

	results := WorkerPool(ctx, 3, jobs, func(_ context.Context, n int) int { return n * 2 })

	var sum int
	for r := range results {
		sum += r
	}
	if sum != 30 { // 2+4+6+8+10
		t.Errorf("got sum %d, want 30", sum)
	}
}

func TestWorkerPool_ContextCancel(t *testing.T) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	jobs := make(chan int)
	results := WorkerPool(ctx, 2, jobs, func(_ context.Context, n int) int { return n })
	cancel()
	// Results channel should close after cancellation.
	for range results { //nolint:revive // drain channel
	}
}

func TestSafeMap(t *testing.T) {
	m := NewSafeMap[string, int]()
	m.Set("a", 1)
	v, ok := m.Get("a")
	if !ok || v != 1 {
		t.Errorf("Get(a) = %d, %v; want 1, true", v, ok)
	}
	_, ok = m.Get("missing")
	if ok {
		t.Error("Get(missing) should return false")
	}
	if m.Len() != 1 {
		t.Errorf("Len() = %d, want 1", m.Len())
	}
}

func TestSafeMap_Concurrent(t *testing.T) {
	m := NewSafeMap[int, int]()
	var wg sync.WaitGroup
	for i := range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m.Set(i, i*2)
			m.Get(i)
		}()
	}
	wg.Wait()
	if m.Len() != 100 {
		t.Errorf("Len() = %d, want 100", m.Len())
	}
}

func TestSemaphore(t *testing.T) {
	sem := NewSemaphore(2)
	ctx := context.Background()
	if err := sem.Acquire(ctx); err != nil {
		t.Fatal(err)
	}
	if err := sem.Acquire(ctx); err != nil {
		t.Fatal(err)
	}
	// Third acquire with cancelled context should fail.
	ctx2, cancel := context.WithCancel(ctx)
	cancel()
	if err := sem.Acquire(ctx2); err == nil {
		t.Error("expected error from cancelled context")
	}
	sem.Release()
}

func TestFanIn(t *testing.T) {
	ctx := context.Background()
	ch1 := make(chan int, 1)
	ch2 := make(chan int, 1)
	ch1 <- 1
	ch2 <- 2
	close(ch1)
	close(ch2)

	merged := FanIn(ctx, ch1, ch2)
	sum := 0
	for v := range merged {
		sum += v
	}
	if sum != 3 {
		t.Errorf("FanIn sum = %d, want 3", sum)
	}
}

func TestOrDone(t *testing.T) {
	ctx := context.Background()
	src := make(chan int, 3)
	src <- 1
	src <- 2
	src <- 3
	close(src)

	out := OrDone(ctx, src)
	var got []int
	for v := range out {
		got = append(got, v)
	}
	if len(got) != 3 || got[0]+got[1]+got[2] != 6 {
		t.Errorf("OrDone got %v, want [1 2 3]", got)
	}
}

func TestOrDone_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	src := make(chan int) // unbuffered, will block
	out := OrDone(ctx, src)
	cancel()
	for range out { //nolint:revive
	}
}

// BenchmarkSafeMap measures concurrent Set/Get throughput.
func BenchmarkSafeMap(b *testing.B) {
	m := NewSafeMap[int, int]()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			m.Set(i, i)
			m.Get(i)
			i++
		}
	})
}
