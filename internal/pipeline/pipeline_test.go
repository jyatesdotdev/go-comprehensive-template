package pipeline

import (
	"context"
	"testing"
)

func TestMapFilter(t *testing.T) {
	ctx := context.Background()
	src := Generator(ctx, 1, 2, 3, 4, 5)
	doubled := Map(func(n int) int { return n * 2 })(ctx, src)
	evens := Filter(func(n int) bool { return n > 4 })(ctx, doubled)

	var got []int
	for v := range evens {
		got = append(got, v)
	}
	// doubled: 2,4,6,8,10 → >4: 6,8,10
	sum := 0
	for _, v := range got {
		sum += v
	}
	if sum != 24 {
		t.Errorf("sum = %d, want 24", sum)
	}
}

func TestReduce(t *testing.T) {
	ctx := context.Background()
	src := Generator(ctx, 1, 2, 3, 4)
	got := Reduce(ctx, src, 0, func(acc, v int) int { return acc + v })
	if got != 10 {
		t.Errorf("Reduce = %d, want 10", got)
	}
}

func TestBatch(t *testing.T) {
	ctx := context.Background()
	src := Generator(ctx, 1, 2, 3, 4, 5)
	batches := Batch[int](2)(ctx, src)

	var sizes []int
	for b := range batches {
		sizes = append(sizes, len(b))
	}
	// Expect batches of [2, 2, 1]
	if len(sizes) != 3 {
		t.Fatalf("got %d batches, want 3", len(sizes))
	}
	if sizes[0] != 2 || sizes[1] != 2 || sizes[2] != 1 {
		t.Errorf("batch sizes = %v, want [2 2 1]", sizes)
	}
}

func TestMapReduce(t *testing.T) {
	ctx := context.Background()
	src := Generator(ctx, "hello", "world", "go")
	got := MapReduce(ctx, src, 2,
		func(s string) int { return len(s) },
		func(acc, v int) int { return acc + v },
		0,
	)
	if got != 12 { // 5+5+2
		t.Errorf("MapReduce = %d, want 12", got)
	}
}

func BenchmarkMapFilter(b *testing.B) {
	items := make([]int, 1000)
	for i := range items {
		items[i] = i
	}
	b.ResetTimer()
	for b.Loop() {
		ctx := context.Background()
		src := Generator(ctx, items...)
		doubled := Map(func(n int) int { return n * 2 })(ctx, src)
		evens := Filter(func(n int) bool { return n%2 == 0 })(ctx, doubled)
		for range evens { //nolint:revive // drain channel
		}
	}
}
