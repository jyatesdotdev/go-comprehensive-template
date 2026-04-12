// Package perf provides performance utilities: object pooling, pre-allocation
// helpers, and examples of allocation-conscious patterns.
package perf

import (
	"sync"
)

// Pool is a typed wrapper around sync.Pool for object reuse.
type Pool[T any] struct {
	pool sync.Pool
}

// NewPool creates a Pool with the given constructor.
func NewPool[T any](newFn func() T) *Pool[T] {
	return &Pool[T]{
		pool: sync.Pool{New: func() any { return newFn() }},
	}
}

// Get retrieves an object from the pool.
func (p *Pool[T]) Get() T { return p.pool.Get().(T) }

// Put returns an object to the pool.
func (p *Pool[T]) Put(v T) { p.pool.Put(v) }

// --- Escape analysis demonstration functions ---

// SumNoEscape computes a sum without heap allocation (value stays on stack).
//
//go:noinline
func SumNoEscape(nums []int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

// SumEscapes returns a pointer, forcing the result to escape to the heap.
//
//go:noinline
func SumEscapes(nums []int) *int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return &total
}

// --- Pre-allocation patterns ---

// CollectWithPrealloc builds a result slice with a pre-allocated capacity,
// avoiding repeated grow-and-copy during append.
func CollectWithPrealloc[T any](src []T, fn func(T) T) []T {
	out := make([]T, 0, len(src))
	for _, v := range src {
		out = append(out, fn(v))
	}
	return out
}

// CollectNaive builds a result slice without pre-allocation (for benchmarking comparison).
func CollectNaive[T any](src []T, fn func(T) T) []T {
	var out []T //nolint:prealloc // intentionally naive for benchmarking comparison
	for _, v := range src {
		out = append(out, fn(v))
	}
	return out
}
