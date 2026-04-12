package perf

import (
	"bytes"
	"testing"
)

// --- Benchmark: sync.Pool vs fresh allocation ---

func BenchmarkBufferPool(b *testing.B) {
	pool := NewPool(func() *bytes.Buffer { return new(bytes.Buffer) })
	b.ResetTimer()
	for b.Loop() {
		buf := pool.Get()
		buf.WriteString("hello world")
		buf.Reset()
		pool.Put(buf)
	}
}

func BenchmarkBufferAlloc(b *testing.B) {
	for b.Loop() {
		buf := new(bytes.Buffer)
		buf.WriteString("hello world")
		buf.Reset()
	}
}

// --- Benchmark: pre-allocation vs naive append ---

func BenchmarkCollectPrealloc(b *testing.B) {
	src := make([]int, 10_000)
	for i := range src {
		src[i] = i
	}
	double := func(n int) int { return n * 2 }
	b.ResetTimer()
	for b.Loop() {
		_ = CollectWithPrealloc(src, double)
	}
}

func BenchmarkCollectNaive(b *testing.B) {
	src := make([]int, 10_000)
	for i := range src {
		src[i] = i
	}
	double := func(n int) int { return n * 2 }
	b.ResetTimer()
	for b.Loop() {
		_ = CollectNaive(src, double)
	}
}

// --- Benchmark: escape vs no-escape ---

func BenchmarkSumNoEscape(b *testing.B) {
	nums := make([]int, 1000)
	for i := range nums {
		nums[i] = i
	}
	b.ResetTimer()
	for b.Loop() {
		_ = SumNoEscape(nums)
	}
}

func BenchmarkSumEscapes(b *testing.B) {
	nums := make([]int, 1000)
	for i := range nums {
		nums[i] = i
	}
	b.ResetTimer()
	for b.Loop() {
		_ = SumEscapes(nums)
	}
}

// --- Table-driven benchmark pattern ---

func BenchmarkCollectSizes(b *testing.B) {
	sizes := []struct {
		name string
		n    int
	}{
		{"100", 100},
		{"1K", 1_000},
		{"10K", 10_000},
	}
	double := func(n int) int { return n * 2 }
	for _, tc := range sizes {
		src := make([]int, tc.n)
		for i := range src {
			src[i] = i
		}
		b.Run("Prealloc/"+tc.name, func(b *testing.B) {
			for b.Loop() {
				_ = CollectWithPrealloc(src, double)
			}
		})
		b.Run("Naive/"+tc.name, func(b *testing.B) {
			for b.Loop() {
				_ = CollectNaive(src, double)
			}
		})
	}
}
