# ETL/MapReduce Pipelines

Go's channels and goroutines make it a natural fit for data pipelines. This guide covers the patterns in `internal/pipeline`.

## Core Concept: Stages

A **Stage** is a function that takes an input channel and returns an output channel. Stages run in their own goroutine and respect `context.Context` for cancellation.

```go
type Stage[In, Out any] func(ctx context.Context, in <-chan In) <-chan Out
```

Stages compose naturally — the output of one feeds the input of the next.

## Pipeline Primitives

### Generator
Creates a channel from a slice of values. Starting point for most pipelines.

```go
ch := pipeline.Generator(ctx, "a", "b", "c")
```

### Map
Transforms each item. Preserves order (single goroutine).

```go
upper := pipeline.Map(strings.ToUpper)(ctx, input)
```

### Filter
Passes through items matching a predicate.

```go
valid := pipeline.Filter(func(r Record) bool { return r.Name != "" })(ctx, input)
```

### FlatMap
Maps each item to a slice, flattening results into the output stream.

```go
words := pipeline.FlatMap(func(line string) []string {
    return strings.Fields(line)
})(ctx, lines)
```

### Batch
Collects items into fixed-size slices for bulk processing. The final batch may be smaller.

```go
batches := pipeline.Batch[int](100)(ctx, numbers) // <-chan []int
```

### Reduce
Terminal operation — consumes all items, folding them into a single result.

```go
total := pipeline.Reduce(ctx, numbers, 0, func(acc, n int) int { return acc + n })
```

## Parallel Processing

### FanOut
Distributes work from a single input channel across `n` parallel copies of a stage, then merges results. Order is not preserved.

```go
processed := pipeline.FanOut(4, pipeline.Map(expensiveTransform))(ctx, input)
```

### MapReduce
Combines parallel mapping with reduction. Useful for aggregation workloads.

```go
wordCounts := pipeline.MapReduce(ctx, lines, 4,
    func(line string) map[string]int { /* count words */ },
    func(acc, m map[string]int) map[string]int { /* merge */ },
    make(map[string]int),
)
```

## ETL Pattern

A typical Extract → Transform → Load pipeline chains stages:

```go
raw := pipeline.Generator(ctx, records...)                          // Extract
valid := pipeline.Filter(isValid)(ctx, raw)                         // Transform
normalized := pipeline.Map(normalize)(ctx, valid)                   // Transform
batches := pipeline.Batch[Record](100)(ctx, normalized)             // Batch for bulk load
for batch := range batches { db.BulkInsert(batch) }                 // Load
```

## Design Principles

1. **Channels as streams** — each stage owns its output channel and closes it when done
2. **Context propagation** — every stage checks `ctx.Done()` for cancellation
3. **Backpressure** — unbuffered channels naturally apply backpressure between stages
4. **Generics** — stages are type-safe via Go generics, no `interface{}` casting
5. **Composition** — complex pipelines are built from simple, testable stages

## Anti-Patterns to Avoid

- **Goroutine leaks**: Always check `ctx.Done()` in select statements
- **Unclosed channels**: The producer (stage) must close its output channel
- **Shared mutable state**: Pass data through channels, not shared variables
- **Unbounded fan-out**: Use `FanOut` with a fixed worker count, not one goroutine per item

## Running the Example

```bash
go run ./examples/pipeline/
```

Demonstrates ETL, MapReduce word count, batching, and FlatMap pipelines.

## See Also

- [Concurrency](concurrency.md) — Goroutine patterns used in pipelines
- [Performance](performance.md) — Profiling and optimizing pipeline throughput
