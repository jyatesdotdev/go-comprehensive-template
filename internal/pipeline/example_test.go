package pipeline_test

import (
	"context"
	"fmt"

	"github.com/example/go-template/internal/pipeline"
)

func ExampleGenerator() {
	ctx := context.Background()
	ch := pipeline.Generator(ctx, "a", "b", "c")
	for v := range ch {
		fmt.Println(v)
	}
	// Output:
	// a
	// b
	// c
}

func ExampleMap() {
	ctx := context.Background()
	src := pipeline.Generator(ctx, 1, 2, 3)
	doubled := pipeline.Map(func(n int) int { return n * 2 })(ctx, src)
	for v := range doubled {
		fmt.Println(v)
	}
	// Output:
	// 2
	// 4
	// 6
}

func ExampleFilter() {
	ctx := context.Background()
	src := pipeline.Generator(ctx, 1, 2, 3, 4, 5, 6)
	evens := pipeline.Filter(func(n int) bool { return n%2 == 0 })(ctx, src)
	for v := range evens {
		fmt.Println(v)
	}
	// Output:
	// 2
	// 4
	// 6
}

func ExampleReduce() {
	ctx := context.Background()
	src := pipeline.Generator(ctx, 1, 2, 3, 4, 5)
	sum := pipeline.Reduce(ctx, src, 0, func(acc, v int) int { return acc + v })
	fmt.Println(sum)
	// Output:
	// 15
}

func ExampleBatch() {
	ctx := context.Background()
	src := pipeline.Generator(ctx, 1, 2, 3, 4, 5)
	batches := pipeline.Batch[int](2)(ctx, src)
	for b := range batches {
		fmt.Println(b)
	}
	// Output:
	// [1 2]
	// [3 4]
	// [5]
}
