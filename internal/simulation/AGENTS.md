# AGENTS.md — internal/simulation

Concurrent simulation and numerical-computing primitives: Monte Carlo estimation, a concurrent
simulation runner, and basic numerical methods. Stdlib only (`context`, `math`, `sync`).

Read [internal/AGENTS.md](../AGENTS.md) and the [root](../../AGENTS.md) first.

## Exports

- **`MonteCarlo(ctx, n, numWorkers, trialFn) float64`** — runs `n` trials, at most `numWorkers`
  concurrently (bounded by a buffered "slots" channel), and returns the **mean** of the samples.
  Accumulation is guarded by a `sync.Mutex`.
- **`RunConcurrent(ctx, sims) []Result`** — runs a `map[string]func(ctx)(float64,error)` of named
  simulations concurrently; each outcome becomes a `Result{Name, Value, Err}` (append is mutex-guarded).
- **`Newton(f, deriv, x0, tol, maxIter) float64`** — Newton–Raphson root finding; stops when
  `|f(x)| < tol` or after `maxIter` iterations.
- **`Trapezoid(f, a, b, n) float64`** — composite trapezoidal integral over `n` subintervals.

## Business rules & numerical subtleties (know these before editing)

- **`MonteCarlo` divides by `n`, always.** If `ctx` is cancelled partway, the loop stops spawning
  trials but the result is still `total / n` — i.e. a *partial* mean skewed toward zero, not the
  mean of completed trials. Treat a cancelled Monte Carlo result as invalid; don't rely on it.
- **Randomness is the caller's job.** `trialFn func() float64` supplies the sampling; this package
  seeds nothing. For reproducible tests, pass a deterministic or fixed-seed `trialFn`. If a trial
  uses shared RNG state, it must be safe for `numWorkers` concurrent calls.
- **`RunConcurrent` result order is non-deterministic** (map iteration + goroutine scheduling).
  Callers that need order must sort by `Name`.
- **`Newton` assumes `deriv(x) != 0`** near the root — a zero derivative divides by zero
  (`±Inf`/`NaN`). It does not converge for all inputs; `maxIter` is the safety bound.
- Numerics are `float64`; expect floating-point error. Tests must compare with a tolerance
  (`math.Abs(got-want) < eps`), never `==`.

## Go best practices

- Same concurrency discipline as `concurrency`/`pipeline`: honor `ctx`, guard shared accumulators
  with a mutex (or use per-worker partials + a final combine to reduce contention).
- Keep numerical methods allocation-free in hot loops; pass functions, don't box values.
- Bound iteration counts (`maxIter`, `n`) so a bad input can't loop forever.

## Testing

`make test` runs `-race` (mutex-guarded accumulation must stay race-free). Use fixed seeds and
tolerance-based assertions; for Monte Carlo, a larger `n` tightens the estimate — pick `n` big
enough for a stable test but small enough to stay fast. See
[`examples/simulation`](../../examples/AGENTS.md).
