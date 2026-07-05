# AGENTS.md — internal/concurrency

Reusable, generic concurrency primitives: worker pool, fan-in, or-done, semaphore, and a
concurrent-safe map. Stdlib only (`context`, `sync`).

Read [internal/AGENTS.md](../AGENTS.md) and the [root](../../AGENTS.md) first.

## Exports

- **`WorkerPool[T, R](ctx, n, jobs, fn) <-chan R`** — `n` workers consume `jobs`, send `fn` results
  to the returned channel, which closes after all workers finish.
- **`FanIn[T](ctx, chans...) <-chan T`** — merges many channels into one; closes when all inputs do.
- **`OrDone[T](ctx, src) <-chan T`** — mirrors `src` but stops on `ctx` cancellation (encapsulates
  the `select { case <-ctx.Done(); case v, ok := <-src }` boilerplate).
- **`Semaphore`** — counting semaphore over a buffered channel. `Acquire(ctx)` blocks until a slot
  frees **or ctx cancels** (returns `ctx.Err()`); `Release()` frees one slot.
- **`SafeMap[K comparable, V]`** — `RWMutex`-guarded map with `Set`, `Get` (comma-ok), `Len`.

## The channel discipline (the core lesson of this package — do not break it)

Every goroutine here follows the same rules; new code must too:

1. **Always `select` on `ctx.Done()`** on *both* the receive and the send side inside loops.
   A naked `out <- v` can block forever if the consumer stops reading → goroutine leak.
2. **Close each output channel exactly once**, from the producer side. The pattern is a
   `go func() { wg.Wait(); close(out) }()` closer (WorkerPool/FanIn) or `defer close(out)`
   (OrDone) — never both, never from a consumer.
3. **Return channels are unbuffered.** Back-pressure is intentional.
4. Callers own the *input* channel and must close it to signal "no more work"; these primitives
   never close a channel they did not create.

For `Semaphore`, every `Acquire` that returns `nil` must be paired with exactly one `Release`
(use `defer`). Do not `Release` without a successful `Acquire`.

## Go best practices for concurrency

- Use `errgroup`-style structured concurrency only if you add the dep; here, `sync.WaitGroup` +
  context is the deliberate stdlib approach.
- Never guard a map with a plain `sync.Mutex` for read-heavy workloads when `RWMutex` fits —
  `SafeMap` uses `RLock` for `Get`/`Len` on purpose.
- Prefer passing `context.Context` explicitly; never store it in a struct.
- Size worker counts from the workload, commonly `runtime.NumCPU()` for CPU-bound work.

## Testing

`make test` runs with `-race` — mandatory here; a race-free run is the bar for any change.
Godoc `Example*` tests in `example_test.go` (`ExampleSafeMap`, `ExampleSemaphore`,
`ExampleWorkerPool`) double as documentation with `// Output:` assertions — keep them passing.
When testing pools/pipelines, buffer and pre-close the `jobs` channel so tests are deterministic.
See [`examples/concurrency`](../../examples/AGENTS.md) for a full runnable tour.
