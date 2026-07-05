# AGENTS.md — internal/pipeline

Generic, channel-based data pipelines: ETL, MapReduce, and streaming. Stages compose into
pipelines; every stage is context-cancellable. Stdlib only (`context`, `sync`).

Read [internal/AGENTS.md](../AGENTS.md) and the [root](../../AGENTS.md) first.

## The core type

```go
type Stage[In, Out any] func(ctx context.Context, in <-chan In) <-chan Out
```

A stage reads an input channel and returns a new output channel. Compose stages by feeding one's
output into the next. `Generator` seeds a pipeline from a slice.

## Exports

- **`Generator[T](ctx, items...)`** — source stage emitting each item.
- **`Map[In,Out](fn)`**, **`Filter[T](pred)`**, **`FlatMap[In,Out](fn)`** — element transforms.
- **`Batch[T](size)`** — groups items into `[]T` of up to `size`; **the final batch may be smaller**.
- **`Reduce[T,R](ctx, in, initial, fn)`** — terminal fold, drains `in` to a single value.
- **`FanOut[In,Out](n, stage)`** — runs `n` copies of `stage` over the **shared** input channel and
  merges outputs (work-distribution, not duplication).
- **`MapReduce[In,Mid,Out](ctx, in, n, mapFn, reduceFn, initial)`** — parallel map across `n`
  workers, then reduce.

## The channel discipline (identical to `internal/concurrency` — do not break it)

Every stage goroutine must:

1. **`defer close(out)`** — each stage owns and closes exactly the channel it created.
2. **`select` on `ctx.Done()` on both receive and send.** Never write `out <- v` unguarded; a
   stalled consumer would leak the goroutine. Follow the existing `select { case out <- v: case
   <-ctx.Done(): return }` shape.
3. Propagate "no more data" by ranging until the input channel is closed (`v, ok := <-in; if !ok`),
   then returning. `Batch` additionally **flushes its buffer on close** before returning.
4. `Batch` reallocates its buffer after each flush (`buf = make([]T, 0, size)`) so a sent slice is
   never aliased/overwritten — preserve that when editing.

Cancellation is cooperative: cancel the `ctx` (or let upstream close) to tear the whole pipeline
down. There are no buffered channels — back-pressure flows naturally.

## Go best practices

- Keep stage functions pure and side-effect-free where possible; put orchestration in the caller.
- Choose `FanOut`/`MapReduce` for CPU-bound element work; size `n` to `runtime.NumCPU()`.
- Reductions are inherently serial — do the heavy per-item work in `Map`, keep `reduceFn` cheap.

## Testing

`make test` runs `-race` — required, this is all goroutines and channels. `pipeline_test.go`
covers stages; `example_test.go` provides godoc `Example*` runnable docs with `// Output:` blocks.
Always drive tests with a cancellable `context` and either close the source or cancel to avoid
hanging tests. Full tour in [`examples/pipeline`](../../examples/AGENTS.md).
