# Concurrency in Go

Go's concurrency model is built on goroutines and channels, following the CSP (Communicating Sequential Processes) paradigm. This guide covers patterns, primitives, and pitfalls.

## Goroutines

Goroutines are lightweight threads managed by the Go runtime. They start with ~2KB stack that grows as needed.

```go
go func() {
    // runs concurrently
}()
```

Always ensure goroutines can exit. A goroutine that blocks forever is a **goroutine leak**.

## Channels

Channels are typed conduits for communication between goroutines.

```go
ch := make(chan int)      // unbuffered — send blocks until receive
ch := make(chan int, 10)  // buffered — send blocks when full
```

**Directional channels** restrict usage at the type level:

```go
func producer(out chan<- int) { out <- 42 }
func consumer(in <-chan int)  { v := <-in }
```

**Closing channels** signals no more values. Only the sender should close. Receiving from a closed channel yields the zero value.

```go
close(ch)
for v := range ch { /* iterates until closed */ }
```

## Select Statement

`select` multiplexes channel operations:

```go
select {
case v := <-ch1:
    // ch1 ready
case ch2 <- val:
    // sent to ch2
case <-time.After(1 * time.Second):
    // timeout
default:
    // non-blocking
}
```

When multiple cases are ready, one is chosen **at random**.

## Context Package

`context.Context` carries deadlines, cancellation signals, and request-scoped values.

```go
// Cancellation
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

// Timeout
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)

// Check in long-running work
select {
case <-ctx.Done():
    return ctx.Err()
case result <- doWork():
}
```

**Rules:**
- Pass `context.Context` as the first parameter
- Never store contexts in structs
- Use `context.TODO()` only as a placeholder during refactoring

## Sync Primitives

### sync.WaitGroup

Wait for a collection of goroutines to finish:

```go
var wg sync.WaitGroup
for range 10 {
    wg.Add(1)
    go func() {
        defer wg.Done()
        // work
    }()
}
wg.Wait()
```

### sync.Mutex / sync.RWMutex

Protect shared state:

```go
var mu sync.Mutex
mu.Lock()
// critical section
mu.Unlock()
```

Use `RWMutex` when reads vastly outnumber writes — multiple readers can hold the lock simultaneously.

### sync.Once

Execute initialization exactly once, even across goroutines:

```go
var once sync.Once
once.Do(func() { /* runs once */ })
```

### sync.Pool

Reuse temporary objects to reduce GC pressure:

```go
pool := &sync.Pool{
    New: func() any { return make([]byte, 1024) },
}
buf := pool.Get().([]byte)
defer pool.Put(buf)
```

## Concurrency Patterns

### Worker Pool

Distribute work across a fixed number of goroutines. See `internal/concurrency.WorkerPool`.

```
jobs channel → [worker 1] → results channel
               [worker 2] →
               [worker 3] →
```

### Fan-Out / Fan-In

**Fan-out**: Multiple goroutines read from the same channel.
**Fan-in**: Merge multiple channels into one. See `internal/concurrency.FanIn`.

### Pipeline

Chain stages where each stage is a goroutine reading from an input channel and writing to an output channel:

```go
func stage(in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for v := range in {
            out <- transform(v)
        }
    }()
    return out
}
```

### Or-Done Channel

Wrap a channel read with context cancellation. See `internal/concurrency.OrDone`.

### Semaphore

Limit concurrent access using a buffered channel. See `internal/concurrency.Semaphore`.

## Common Pitfalls

### 1. Goroutine Leaks

A goroutine that blocks on a channel forever leaks memory. Always provide a cancellation path:

```go
// BAD: leaks if nobody reads from ch
go func() { ch <- result }()

// GOOD: respects cancellation
go func() {
    select {
    case ch <- result:
    case <-ctx.Done():
    }
}()
```

### 2. Race Conditions

Concurrent access to shared state without synchronization. Detect with:

```bash
go test -race ./...
go run -race main.go
```

### 3. Channel Deadlocks

Sending on an unbuffered channel with no receiver, or waiting on a channel that's never closed.

### 4. Closing a Channel Twice

Causes a panic. Only the sender should close, and only once.

### 5. Loop Variable Capture (pre-Go 1.22)

In Go < 1.22, loop variables were shared across iterations. Go 1.22+ creates a new variable per iteration, fixing this. If targeting older versions:

```go
for _, v := range items {
    v := v // shadow the variable
    go func() { use(v) }()
}
```

## Running the Examples

```bash
go run ./examples/concurrency/
```

## Further Reading

- [Go Concurrency Patterns (Rob Pike)](https://go.dev/talks/2012/concurrency.slide)
- [Advanced Concurrency Patterns](https://go.dev/blog/pipelines)
- [The Go Memory Model](https://go.dev/ref/mem)

## See Also

- [ETL Pipelines](etl-pipelines.md) — Worker pools and pipeline patterns
- [Performance](performance.md) — Profiling concurrent code
- [Testing](testing.md) — Testing concurrent code with the race detector
