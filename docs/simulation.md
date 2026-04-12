# Simulation & Numerical Computing

Concurrent simulations and numerical methods using Go's concurrency primitives.

## Monte Carlo Estimation

Monte Carlo methods estimate values through random sampling. Go's goroutines make it natural to parallelize trials:

```go
pi := simulation.MonteCarlo(ctx, 1_000_000, 8, func() float64 {
    x, y := rand.Float64(), rand.Float64()
    if x*x+y*y <= 1 {
        return 4.0
    }
    return 0.0
})
```

The `MonteCarlo` function distributes `n` trials across `numWorkers` goroutines and returns the mean of all samples. The trial function is stateless — each goroutine calls it independently.

### When to Use Monte Carlo

- Estimating probabilities or expected values
- Problems with high-dimensional integrals
- Risk analysis and financial modeling
- Physics simulations (particle transport, thermodynamics)

## Concurrent Simulations

`RunConcurrent` executes multiple independent simulations in parallel:

```go
results := simulation.RunConcurrent(ctx, map[string]func(context.Context) (float64, error){
    "sim_a": func(ctx context.Context) (float64, error) { ... },
    "sim_b": func(ctx context.Context) (float64, error) { ... },
})
```

Each simulation runs in its own goroutine. Results are collected with a mutex-protected slice. Context cancellation propagates to all simulations.

## Numerical Methods

### Newton's Method (Root Finding)

Iteratively finds roots of `f(x) = 0` using the derivative:

```go
root := simulation.Newton(f, fPrime, x0, tolerance, maxIterations)
```

Converges quadratically for well-behaved functions near the root. Watch out for:
- Division by zero when `f'(x) ≈ 0`
- Divergence with poor initial guesses
- Oscillation near inflection points

### Trapezoidal Integration

Approximates definite integrals by summing trapezoids:

```go
result := simulation.Trapezoid(math.Sin, 0, math.Pi, 10_000) // ≈ 2.0
```

Error is O(h²) where h = (b-a)/n. For higher accuracy, increase n or use adaptive methods.

## Concurrency Patterns for Simulations

### Worker Pool for Trials

The Monte Carlo implementation uses a semaphore pattern (buffered channel) to limit goroutine count while processing all trials concurrently.

### Independent Parallel Runs

When simulations are independent, launch each in a goroutine and collect results:

```go
var wg sync.WaitGroup
for _, sim := range simulations {
    wg.Add(1)
    go func() {
        defer wg.Done()
        sim.Run(ctx)
    }()
}
wg.Wait()
```

### Context for Cancellation

Always pass `context.Context` to long-running simulations. This enables:
- Timeout-based cancellation
- Graceful shutdown
- Resource cleanup

## Running the Example

```bash
go run ./examples/simulation/
```

## Performance Tips

- Use `math/rand/v2` (Go 1.22+) — no global mutex, safe for concurrent use
- Pre-allocate result slices when trial count is known
- For CPU-bound simulations, set workers to `runtime.NumCPU()`
- Profile with `go test -bench` to find the optimal worker count

## See Also

- [Concurrency](concurrency.md) — Parallel simulation patterns
- [Performance](performance.md) — Profiling and benchmarking simulations
