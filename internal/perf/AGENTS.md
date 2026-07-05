# AGENTS.md — internal/perf

Performance / allocation-conscious utilities: a generic `sync.Pool` wrapper, pre-allocation
helpers, and escape-analysis demonstration functions. Stdlib only (`sync`).

Read [internal/AGENTS.md](../AGENTS.md) and the [root](../../AGENTS.md) first.

## Exports

- **`Pool[T]`** — typed wrapper over `sync.Pool`. `NewPool(newFn func() T)`; `Get() T` type-asserts
  the stored `any`; `Put(v T)` returns it. Contract, per `sync.Pool`:
  - Objects may be evicted at any GC; `Get` may return a brand-new object or a reused one.
  - **`Get` gives no cleanliness guarantee — reset state before use or before `Put`** (the test
    resets the buffer before returning it). Store pointer-like types (e.g. `*bytes.Buffer`) so
    reuse actually avoids allocation.
- **`SumNoEscape` / `SumEscapes`** — paired demo of stack vs heap: returning a value keeps it on the
  stack; returning `*int` forces a heap escape. Both are marked `//go:noinline` **so the escape
  behavior is observable** — keep that directive or the demo collapses.
- **`CollectWithPrealloc` vs `CollectNaive`** — pre-sized `make([]T, 0, len(src))` versus growing
  from nil. `CollectNaive` is intentionally un-optimized for benchmark contrast and carries
  `//nolint:prealloc // intentionally naive for benchmarking comparison` — do not "fix" it.

## Go best practices for performance work

- **Measure, don't guess.** Change perf code only with a benchmark (`make bench`, `-benchmem`)
  showing before/after `ns/op`, `B/op`, `allocs/op`.
- Pre-allocate slices/maps when the final size is known (`make([]T, 0, n)` / `make(map[K]V, n)`).
- Inspect escapes with `go build -gcflags='-m'`; use `sync.Pool` only for genuinely hot,
  short-lived, reusable objects (it is not a general cache).
- Prefer readability by default — reach for these patterns where profiling shows a real hot path.

## Testing & benchmarks

`perf_test.go` holds both correctness tests and `Benchmark*` functions. Benchmarks are the
point of this package: keep the prealloc-vs-naive and escape pairs so the comparison stays
demonstrable. See [`examples/performance`](../../examples/AGENTS.md) for a pprof-integrated demo.
