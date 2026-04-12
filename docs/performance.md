# Performance & Profiling

## Profiling with pprof

### HTTP Endpoint (runtime)

Import the blank identifier to register handlers:

```go
import _ "net/http/pprof"

go http.ListenAndServe("localhost:6060", nil)
```

Available profiles at `/debug/pprof/`:
- `heap` — memory allocations
- `goroutine` — all goroutine stacks
- `profile` — CPU profile (30s default)
- `trace` — execution trace

### Analyzing Profiles

```bash
# CPU profile
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=10

# Heap profile
go tool pprof http://localhost:6060/debug/pprof/heap

# Inside pprof interactive mode:
#   top10        — hottest functions
#   list FuncName — annotated source
#   web          — open graph in browser (requires graphviz)
```

### From Tests

```bash
go test -cpuprofile=cpu.prof -memprofile=mem.prof -bench=. ./internal/perf/
go tool pprof cpu.prof
```

## Benchmarking

### Writing Benchmarks

```go
func BenchmarkFoo(b *testing.B) {
    for b.Loop() {   // Go 1.24+ loop form
        foo()
    }
}
```

### Table-Driven Benchmarks

```go
func BenchmarkSizes(b *testing.B) {
    for _, n := range []int{100, 1000, 10000} {
        b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
            data := make([]int, n)
            b.ResetTimer()
            for b.Loop() {
                process(data)
            }
        })
    }
}
```

### Running Benchmarks

```bash
go test -bench=. -benchmem ./...          # all benchmarks with alloc stats
go test -bench=BenchmarkPool -count=5 ./internal/perf/  # repeated for benchstat
```

### Comparing Results with benchstat

```bash
go test -bench=. -count=6 ./... > old.txt
# make changes
go test -bench=. -count=6 ./... > new.txt
benchstat old.txt new.txt
```

## Memory Optimization

### sync.Pool for Object Reuse

Reuse short-lived objects to reduce GC pressure:

```go
pool := perf.NewPool(func() *bytes.Buffer { return new(bytes.Buffer) })

buf := pool.Get()
defer func() { buf.Reset(); pool.Put(buf) }()
buf.WriteString("data")
```

### Pre-allocation

Always pre-allocate slices and maps when the size is known:

```go
// Good: one allocation
out := make([]T, 0, len(src))

// Bad: repeated grow-and-copy
var out []T
```

### Reducing Allocations

- Return values instead of pointers when possible (keeps data on stack)
- Use `strings.Builder` instead of `fmt.Sprintf` in hot paths
- Prefer `[]byte` over `string` for mutable data
- Use `sync.Pool` for frequently allocated temporary objects

## Escape Analysis

The compiler decides whether variables live on the stack or heap. Inspect with:

```bash
go build -gcflags='-m' ./internal/perf/
```

Common causes of heap escape:
- Returning a pointer to a local variable
- Storing a local in an interface value
- Closures capturing local variables
- Slices that grow beyond their initial capacity

```go
// Stack-allocated (no escape)
func sum(nums []int) int {
    total := 0
    for _, n := range nums { total += n }
    return total
}

// Heap-allocated (escapes)
func sum(nums []int) *int {
    total := 0
    for _, n := range nums { total += n }
    return &total  // &total escapes
}
```

## GC Tuning

### GOGC

Controls GC frequency. Default is `100` (GC when heap doubles).

```bash
GOGC=50  ./server   # more frequent GC, lower peak memory
GOGC=200 ./server   # less frequent GC, higher throughput
GOGC=off ./server   # disable GC (use with GOMEMLIMIT)
```

### GOMEMLIMIT (Go 1.19+)

Soft memory limit. The GC becomes more aggressive as the limit approaches:

```bash
GOMEMLIMIT=512MiB ./server
```

Best practice: set `GOGC=off` with `GOMEMLIMIT` for predictable memory usage in containers.

### Runtime Memory Stats

```go
var m runtime.MemStats
runtime.ReadMemStats(&m)
fmt.Printf("HeapAlloc: %d MB\n", m.HeapAlloc/1024/1024)
fmt.Printf("NumGC: %d\n", m.NumGC)
```

## Makefile Targets

The project Makefile includes profiling targets:

```bash
make bench          # run all benchmarks with -benchmem
make pprof-cpu      # generate and open CPU profile
make pprof-mem      # generate and open memory profile
```

## Project Files

| File | Description |
|------|-------------|
| `internal/perf/perf.go` | Typed sync.Pool, escape analysis demos, pre-allocation helpers |
| `internal/perf/perf_test.go` | Benchmarks: pool vs alloc, prealloc vs naive, table-driven |
| `examples/performance/main.go` | Runnable demo: pprof, pooling, memstats, GC tuning |

## See Also

- [Concurrency](concurrency.md) — Goroutine patterns and pitfalls
- [Cloud-Native](cloud-native.md) — Production deployment and observability
- [Testing](testing.md) — Benchmarks and profiling in tests
