package perf

import (
	"bytes"
	"testing"
)

func TestPool(t *testing.T) {
	pool := NewPool(func() *bytes.Buffer { return new(bytes.Buffer) })
	buf := pool.Get()
	buf.WriteString("hello")
	if buf.String() != "hello" {
		t.Fatal("unexpected buffer content")
	}
	buf.Reset()
	pool.Put(buf)
	buf2 := pool.Get()
	if buf2.Len() != 0 {
		t.Fatal("expected empty buffer from pool")
	}
}

func TestSumNoEscape(t *testing.T) {
	if got := SumNoEscape([]int{1, 2, 3}); got != 6 {
		t.Fatalf("got %d, want 6", got)
	}
}

func TestSumEscapes(t *testing.T) {
	got := SumEscapes([]int{4, 5, 6})
	if *got != 15 {
		t.Fatalf("got %d, want 15", *got)
	}
}

func TestCollectWithPrealloc(t *testing.T) {
	src := []int{1, 2, 3}
	got := CollectWithPrealloc(src, func(n int) int { return n * 2 })
	if len(got) != 3 || got[0] != 2 || got[1] != 4 || got[2] != 6 {
		t.Fatalf("got %v", got)
	}
}

func TestCollectNaive(t *testing.T) {
	src := []int{1, 2, 3}
	got := CollectNaive(src, func(n int) int { return n + 10 })
	if len(got) != 3 || got[0] != 11 || got[1] != 12 || got[2] != 13 {
		t.Fatalf("got %v", got)
	}
}

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
