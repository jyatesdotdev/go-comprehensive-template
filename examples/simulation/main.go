// Example simulation demonstrates concurrent Monte Carlo, numerical methods, and parallel simulations.
package main

import (
	"context"
	"fmt"
	"math"
	"math/rand/v2"
	"time"

	"github.com/example/go-template/internal/simulation"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. Monte Carlo estimation of Pi
	fmt.Println("=== Monte Carlo Pi Estimation ===")
	pi := simulation.MonteCarlo(ctx, 1_000_000, 8, func() float64 {
		x, y := rand.Float64(), rand.Float64()
		if x*x+y*y <= 1 {
			return 4.0
		}
		return 0.0
	})
	fmt.Printf("Estimated Pi: %.6f (error: %.6f)\n\n", pi, math.Abs(pi-math.Pi))

	// 2. Concurrent simulations
	fmt.Println("=== Concurrent Simulations ===")
	results := simulation.RunConcurrent(ctx, map[string]func(context.Context) (float64, error){
		"pi_monte_carlo": func(ctx context.Context) (float64, error) {
			v := simulation.MonteCarlo(ctx, 500_000, 4, func() float64 {
				x, y := rand.Float64(), rand.Float64()
				if x*x+y*y <= 1 {
					return 4.0
				}
				return 0.0
			})
			return v, nil
		},
		"integral_sin": func(_ context.Context) (float64, error) {
			// ∫sin(x)dx from 0 to Pi = 2.0
			return simulation.Trapezoid(math.Sin, 0, math.Pi, 10_000), nil
		},
		"sqrt2_newton": func(_ context.Context) (float64, error) {
			// Find root of x²-2 = 0 → x = √2
			return simulation.Newton(
				func(x float64) float64 { return x*x - 2 },
				func(x float64) float64 { return 2 * x },
				1.0, 1e-12, 100,
			), nil
		},
	})
	for _, r := range results {
		fmt.Printf("  %-20s = %.10f\n", r.Name, r.Value)
	}

	// 3. Newton's method
	fmt.Println("\n=== Newton's Method ===")
	root := simulation.Newton(
		func(x float64) float64 { return x*x*x - x - 2 },
		func(x float64) float64 { return 3*x*x - 1 },
		1.5, 1e-12, 50,
	)
	fmt.Printf("Root of x³-x-2: %.12f\n", root)

	// 4. Numerical integration
	fmt.Println("\n=== Trapezoidal Integration ===")
	integral := simulation.Trapezoid(func(x float64) float64 { return x * x }, 0, 1, 10_000)
	fmt.Printf("∫x²dx from 0 to 1 = %.10f (exact: 0.3333333333)\n", integral)
}
