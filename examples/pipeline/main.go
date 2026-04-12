// Example pipeline demonstrates ETL, MapReduce, and streaming patterns.
package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/example/go-template/internal/pipeline"
)

func main() {
	ctx := context.Background()

	fmt.Println("=== 1. ETL Pipeline: Extract → Transform → Load ===")
	etlDemo(ctx)

	fmt.Println("\n=== 2. MapReduce: Parallel Word Count ===")
	mapReduceDemo(ctx)

	fmt.Println("\n=== 3. Streaming with Batch ===")
	batchDemo(ctx)

	fmt.Println("\n=== 4. FlatMap Pipeline ===")
	flatMapDemo(ctx)
}

// etlDemo shows a classic Extract → Transform → Load pipeline.
func etlDemo(ctx context.Context) {
	// Extract: raw records
	type Record struct{ Name, Email string }
	raw := pipeline.Generator(ctx,
		Record{"alice", "ALICE@EXAMPLE.COM"},
		Record{"bob", "BOB@EXAMPLE.COM"},
		Record{"", "EMPTY@EXAMPLE.COM"},
		Record{"charlie", "CHARLIE@EXAMPLE.COM"},
	)

	// Transform: filter invalid, normalize
	valid := pipeline.Filter(func(r Record) bool { return r.Name != "" })(ctx, raw)
	normalized := pipeline.Map(func(r Record) Record {
		return Record{Name: strings.Title(r.Name), Email: strings.ToLower(r.Email)}
	})(ctx, valid)

	// Load: consume results
	for r := range normalized {
		fmt.Printf("  Loaded: %s <%s>\n", r.Name, r.Email)
	}
}

// mapReduceDemo counts words in parallel using MapReduce.
func mapReduceDemo(ctx context.Context) {
	lines := pipeline.Generator(ctx,
		"the quick brown fox",
		"the fox jumped over the lazy dog",
		"the dog barked at the fox",
	)

	result := pipeline.MapReduce(
		ctx, lines, 3,
		// Map: line → word count map
		func(line string) map[string]int {
			counts := make(map[string]int)
			for _, w := range strings.Fields(line) {
				counts[strings.ToLower(w)]++
			}
			return counts
		},
		// Reduce: merge maps
		func(acc map[string]int, m map[string]int) map[string]int {
			for k, v := range m {
				acc[k] += v
			}
			return acc
		},
		make(map[string]int),
	)

	for word, count := range result {
		fmt.Printf("  %s: %d\n", word, count)
	}
}

// batchDemo collects items into batches for bulk processing.
func batchDemo(ctx context.Context) {
	numbers := pipeline.Generator(ctx, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10)
	batches := pipeline.Batch[int](3)(ctx, numbers)

	for batch := range batches {
		fmt.Printf("  Batch: %v (sum=%d)\n", batch, sum(batch))
	}
}

func sum(nums []int) int {
	s := 0
	for _, n := range nums {
		s += n
	}
	return s
}

// flatMapDemo splits lines into words using FlatMap.
func flatMapDemo(ctx context.Context) {
	lines := pipeline.Generator(ctx, "hello world", "go is great")
	words := pipeline.FlatMap(func(line string) []string {
		return strings.Fields(line)
	})(ctx, lines)
	upper := pipeline.Map(strings.ToUpper)(ctx, words)

	for w := range upper {
		fmt.Printf("  %s\n", w)
	}
}
