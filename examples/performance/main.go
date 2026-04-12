// Example performance demonstrates pprof integration, memory profiling,
// and allocation-conscious patterns.
package main

import (
	"bytes"
	"fmt"
	"net/http"
	_ "net/http/pprof" // #nosec G108 -- intentional pprof exposure for demo/profiling
	"runtime"
	"time"

	"github.com/example/go-template/internal/perf"
)

func main() {
	fmt.Println("=== 1. pprof HTTP Endpoint ===")
	pprofDemo()

	fmt.Println("\n=== 2. sync.Pool Object Reuse ===")
	poolDemo()

	fmt.Println("\n=== 3. Pre-allocation vs Naive Append ===")
	preallocDemo()

	fmt.Println("\n=== 4. Runtime Memory Stats ===")
	memStatsDemo()

	fmt.Println("\n=== 5. GC Tuning Notes ===")
	gcNotes()
}

func pprofDemo() {
	// In production, expose pprof on a separate internal port.
	go func() {
		fmt.Println("  pprof available at http://localhost:6060/debug/pprof/")
		_ = http.ListenAndServe("localhost:6060", nil) // #nosec G114 -- demo pprof server, timeouts not needed
	}()
	time.Sleep(10 * time.Millisecond) // let server start
	fmt.Println("  Endpoints: /debug/pprof/{heap,goroutine,profile,trace}")
	fmt.Println("  Usage:  go tool pprof http://localhost:6060/debug/pprof/heap")
}

func poolDemo() {
	pool := perf.NewPool(func() *bytes.Buffer { return new(bytes.Buffer) })

	for i := range 5 {
		buf := pool.Get()
		fmt.Fprintf(buf, "request-%d", i)
		fmt.Printf("  Reused buffer: %s\n", buf.String())
		buf.Reset()
		pool.Put(buf)
	}
}

func preallocDemo() {
	src := make([]int, 100_000)
	for i := range src {
		src[i] = i
	}
	double := func(n int) int { return n * 2 }

	start := time.Now()
	_ = perf.CollectWithPrealloc(src, double)
	preallocDur := time.Since(start)

	start = time.Now()
	_ = perf.CollectNaive(src, double)
	naiveDur := time.Since(start)

	fmt.Printf("  Prealloc: %v\n  Naive:    %v\n", preallocDur, naiveDur)
}

func memStatsDemo() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("  HeapAlloc:  %d KB\n", m.HeapAlloc/1024)
	fmt.Printf("  HeapSys:    %d KB\n", m.HeapSys/1024)
	fmt.Printf("  NumGC:      %d\n", m.NumGC)
	fmt.Printf("  GoroutineCount: %d\n", runtime.NumGoroutine())
}

func gcNotes() {
	fmt.Println("  GOGC=100 (default) — GC triggers when heap doubles")
	fmt.Println("  GOGC=50  — more frequent GC, lower memory")
	fmt.Println("  GOGC=200 — less frequent GC, higher throughput")
	fmt.Println("  GOMEMLIMIT=512MiB — soft memory limit (Go 1.19+)")
	fmt.Println("  Escape analysis: go build -gcflags='-m' ./...")
}
