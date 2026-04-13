package simulation

import (
	"context"
	"fmt"
	"math"
	"testing"
)

func TestMonteCarlo(t *testing.T) {
	// Constant trial function → mean should equal the constant.
	got := MonteCarlo(context.Background(), 1000, 4, func() float64 { return 3.0 })
	if math.Abs(got-3.0) > 0.01 {
		t.Fatalf("expected ~3.0, got %f", got)
	}
}

func TestMonteCarlo_CancelledContext(_ *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	// Should still return without hanging; value is degenerate but shouldn't panic.
	_ = MonteCarlo(ctx, 100, 2, func() float64 { return 1.0 })
}

func TestRunConcurrent(t *testing.T) {
	sims := map[string]func(context.Context) (float64, error){
		"ok":  func(_ context.Context) (float64, error) { return 42, nil },
		"err": func(_ context.Context) (float64, error) { return 0, fmt.Errorf("fail") },
	}
	results := RunConcurrent(context.Background(), sims)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		switch r.Name {
		case "ok":
			if r.Value != 42 || r.Err != nil {
				t.Errorf("unexpected ok result: %+v", r)
			}
		case "err":
			if r.Err == nil {
				t.Errorf("expected error for 'err' sim")
			}
		}
	}
}

func TestNewton(t *testing.T) {
	// Find root of x^2 - 4 (root at x=2) starting from x0=3.
	f := func(x float64) float64 { return x*x - 4 }
	df := func(x float64) float64 { return 2 * x }
	got := Newton(f, df, 3.0, 1e-9, 100)
	if math.Abs(got-2.0) > 1e-6 {
		t.Fatalf("expected ~2.0, got %f", got)
	}
}

func TestTrapezoid(t *testing.T) {
	// Integral of x from 0 to 1 = 0.5
	got := Trapezoid(func(x float64) float64 { return x }, 0, 1, 1000)
	if math.Abs(got-0.5) > 1e-6 {
		t.Fatalf("expected ~0.5, got %f", got)
	}
}
