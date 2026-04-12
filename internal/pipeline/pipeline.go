// Package pipeline provides ETL, MapReduce, and streaming data pipeline patterns.
//
// Pipelines are built by composing Stages — functions that consume an input channel
// and produce an output channel. Generic helpers (Map, Filter, Reduce, Batch, FlatMap)
// create stages for common transformations. Chain connects stages into a pipeline.
package pipeline

import (
	"context"
	"sync"
)

// Stage transforms an input channel into an output channel.
type Stage[In, Out any] func(ctx context.Context, in <-chan In) <-chan Out

// Generator produces items and sends them to the returned channel.
func Generator[T any](ctx context.Context, items ...T) <-chan T {
	out := make(chan T)
	go func() {
		defer close(out)
		for _, item := range items {
			select {
			case out <- item:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

// Map applies fn to each item, outputting transformed values.
func Map[In, Out any](fn func(In) Out) Stage[In, Out] {
	return func(ctx context.Context, in <-chan In) <-chan Out {
		out := make(chan Out)
		go func() {
			defer close(out)
			for {
				select {
				case <-ctx.Done():
					return
				case v, ok := <-in:
					if !ok {
						return
					}
					select {
					case out <- fn(v):
					case <-ctx.Done():
						return
					}
				}
			}
		}()
		return out
	}
}

// Filter passes through only items where predicate returns true.
func Filter[T any](pred func(T) bool) Stage[T, T] {
	return func(ctx context.Context, in <-chan T) <-chan T {
		out := make(chan T)
		go func() {
			defer close(out)
			for {
				select {
				case <-ctx.Done():
					return
				case v, ok := <-in:
					if !ok {
						return
					}
					if pred(v) {
						select {
						case out <- v:
						case <-ctx.Done():
							return
						}
					}
				}
			}
		}()
		return out
	}
}

// FlatMap applies fn to each item, flattening the resulting slices into the output.
func FlatMap[In, Out any](fn func(In) []Out) Stage[In, Out] {
	return func(ctx context.Context, in <-chan In) <-chan Out {
		out := make(chan Out)
		go func() {
			defer close(out)
			for {
				select {
				case <-ctx.Done():
					return
				case v, ok := <-in:
					if !ok {
						return
					}
					for _, r := range fn(v) {
						select {
						case out <- r:
						case <-ctx.Done():
							return
						}
					}
				}
			}
		}()
		return out
	}
}

// Batch collects up to size items into slices before sending them downstream.
// The final batch may be smaller than size.
func Batch[T any](size int) Stage[T, []T] {
	return func(ctx context.Context, in <-chan T) <-chan []T {
		out := make(chan []T)
		go func() {
			defer close(out)
			buf := make([]T, 0, size)
			flush := func() {
				if len(buf) > 0 {
					select {
					case out <- buf:
					case <-ctx.Done():
					}
					buf = make([]T, 0, size)
				}
			}
			for {
				select {
				case <-ctx.Done():
					return
				case v, ok := <-in:
					if !ok {
						flush()
						return
					}
					buf = append(buf, v)
					if len(buf) >= size {
						flush()
					}
				}
			}
		}()
		return out
	}
}

// Reduce consumes all items from in, combining them with fn starting from initial.
func Reduce[T, R any](ctx context.Context, in <-chan T, initial R, fn func(R, T) R) R {
	acc := initial
	for {
		select {
		case <-ctx.Done():
			return acc
		case v, ok := <-in:
			if !ok {
				return acc
			}
			acc = fn(acc, v)
		}
	}
}

// FanOut distributes items from in across n copies of stage, merging results.
func FanOut[In, Out any](n int, stage Stage[In, Out]) Stage[In, Out] {
	return func(ctx context.Context, in <-chan In) <-chan Out {
		outs := make([]<-chan Out, n)
		// Each worker gets items from the shared input channel
		for i := range n {
			// Wrap in a per-worker stage that reads from the shared channel
			outs[i] = stage(ctx, in)
		}
		// Merge all outputs
		merged := make(chan Out)
		var wg sync.WaitGroup
		for _, ch := range outs {
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
						case merged <- v:
						case <-ctx.Done():
							return
						}
					}
				}
			}()
		}
		go func() { wg.Wait(); close(merged) }()
		return merged
	}
}

// MapReduce runs a parallel map phase across n workers, then reduces results.
func MapReduce[In, Mid, Out any](
	ctx context.Context,
	in <-chan In,
	n int,
	mapFn func(In) Mid,
	reduceFn func(Out, Mid) Out,
	initial Out,
) Out {
	mapped := FanOut[In, Mid](n, Map(mapFn))(ctx, in)
	return Reduce(ctx, mapped, initial, reduceFn)
}
