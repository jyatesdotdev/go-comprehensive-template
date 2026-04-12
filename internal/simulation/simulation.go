// Package simulation provides concurrent simulation and numerical computing primitives.
//
// Includes Monte Carlo estimation, concurrent simulation runners, and basic
// numerical methods (root finding, integration).
package simulation

import (
	"context"
	"math"
	"sync"

)

// Result holds the outcome of a simulation run.
type Result struct {
	// Name identifies the simulation.
	Name string
	// Value is the computed result.
	Value float64
	// Err is non-nil if the simulation failed.
	Err error
}

// MonteCarlo runs n trials across numWorkers goroutines. Each trial calls trialFn
// which returns a sample value. The returned mean is the average of all samples.
func MonteCarlo(ctx context.Context, n, numWorkers int, trialFn func() float64) float64 {
	var mu sync.Mutex
	var total float64

	var wg sync.WaitGroup
	work := make(chan struct{}, numWorkers)

	for range n {
		if ctx.Err() != nil {
			break
		}
		wg.Add(1)
		work <- struct{}{}
		go func() {
			defer wg.Done()
			v := trialFn()
			mu.Lock()
			total += v
			mu.Unlock()
			<-work
		}()
	}
	wg.Wait()
	return total / float64(n)
}

// RunConcurrent executes multiple named simulations concurrently and collects results.
func RunConcurrent(ctx context.Context, sims map[string]func(context.Context) (float64, error)) []Result {
	results := make([]Result, 0, len(sims))
	var mu sync.Mutex
	var wg sync.WaitGroup

	for name, fn := range sims {
		wg.Add(1)
		go func() {
			defer wg.Done()
			v, err := fn(ctx)
			mu.Lock()
			results = append(results, Result{Name: name, Value: v, Err: err})
			mu.Unlock()
		}()
	}
	wg.Wait()
	return results
}

// Newton finds a root of f near x0 using Newton's method.
// deriv is the derivative of f. Iterates up to maxIter times or until |f(x)| < tol.
func Newton(f, deriv func(float64) float64, x0, tol float64, maxIter int) float64 {
	x := x0
	for range maxIter {
		fx := f(x)
		if math.Abs(fx) < tol {
			return x
		}
		x -= fx / deriv(x)
	}
	return x
}

// Trapezoid approximates the integral of f from a to b using n trapezoids.
func Trapezoid(f func(float64) float64, a, b float64, n int) float64 {
	h := (b - a) / float64(n)
	sum := (f(a) + f(b)) / 2.0
	for i := 1; i < n; i++ {
		sum += f(a + float64(i)*h)
	}
	return sum * h
}
