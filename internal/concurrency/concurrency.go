// Package concurrency provides reusable concurrency patterns for Go applications.
//
// Patterns included: worker pools, fan-out/fan-in, or-done channels,
// semaphore-based throttling, and safe map access.
package concurrency

import (
	"context"
	"sync"
)

// WorkerPool runs fn concurrently across n workers, processing items from jobs.
// Results are sent to the returned channel, which closes when all work is done.
func WorkerPool[T, R any](ctx context.Context, n int, jobs <-chan T, fn func(context.Context, T) R) <-chan R {
	out := make(chan R)
	var wg sync.WaitGroup
	for range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case j, ok := <-jobs:
					if !ok {
						return
					}
					select {
					case out <- fn(ctx, j):
					case <-ctx.Done():
						return
					}
				}
			}
		}()
	}
	go func() { wg.Wait(); close(out) }()
	return out
}

// FanIn merges multiple channels into a single channel.
func FanIn[T any](ctx context.Context, channels ...<-chan T) <-chan T {
	out := make(chan T)
	var wg sync.WaitGroup
	for _, ch := range channels {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case v, ok := <-ch:
					if !ok {
						return
					}
					select {
					case out <- v:
					case <-ctx.Done():
						return
					}
				}
			}
		}()
	}
	go func() { wg.Wait(); close(out) }()
	return out
}

// OrDone returns a channel that mirrors src but respects ctx cancellation.
func OrDone[T any](ctx context.Context, src <-chan T) <-chan T {
	out := make(chan T)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-src:
				if !ok {
					return
				}
				select {
				case out <- v:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}

// Semaphore limits concurrent access to a resource.
type Semaphore struct {
	ch chan struct{}
}

// NewSemaphore creates a semaphore with the given max concurrency.
func NewSemaphore(n int) *Semaphore {
	return &Semaphore{ch: make(chan struct{}, n)}
}

// Acquire blocks until a slot is available or ctx is cancelled.
func (s *Semaphore) Acquire(ctx context.Context) error {
	select {
	case s.ch <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Release frees a slot.
func (s *Semaphore) Release() { <-s.ch }

// SafeMap is a generic concurrent-safe map.
type SafeMap[K comparable, V any] struct {
	mu sync.RWMutex
	m  map[K]V
}

// NewSafeMap creates an initialized SafeMap.
func NewSafeMap[K comparable, V any]() *SafeMap[K, V] {
	return &SafeMap[K, V]{m: make(map[K]V)}
}

// Set stores a key-value pair.
func (s *SafeMap[K, V]) Set(k K, v V) {
	s.mu.Lock()
	s.m[k] = v
	s.mu.Unlock()
}

// Get retrieves a value by key.
func (s *SafeMap[K, V]) Get(k K) (V, bool) {
	s.mu.RLock()
	v, ok := s.m[k]
	s.mu.RUnlock()
	return v, ok
}

// Len returns the number of entries.
func (s *SafeMap[K, V]) Len() int {
	s.mu.RLock()
	n := len(s.m)
	s.mu.RUnlock()
	return n
}
