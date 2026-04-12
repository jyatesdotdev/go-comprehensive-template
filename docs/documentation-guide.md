# Documentation Guide

Go's documentation philosophy is simple: documentation lives in the code, close to what it describes.

## godoc Conventions

### Package Comments

Every package should have a package comment. For multi-file packages, place it in `doc.go`:

```go
// Package pipeline provides ETL, MapReduce, and streaming data pipeline patterns.
//
// Pipelines are built by composing Stages — functions that consume an input channel
// and produce an output channel. Generic helpers (Map, Filter, Reduce, Batch, FlatMap)
// create stages for common transformations.
package pipeline
```

Rules:
- Start with `// Package <name>` followed by a description
- Use blank comment lines (`//`) to separate paragraphs
- Keep the first sentence concise — it appears in package listings

### Exported Symbols

Every exported type, function, method, and constant needs a comment:

```go
// WorkerPool runs fn concurrently across n workers, processing items from jobs.
// Results are sent to the returned channel, which closes when all work is done.
func WorkerPool[T, R any](ctx context.Context, n int, jobs <-chan T, fn func(context.Context, T) R) <-chan R {
```

Rules:
- Start with the symbol name: `// WorkerPool runs...` not `// This function runs...`
- Describe behavior, parameters, and return values
- Document panics, goroutine safety, and channel ownership
- Use `// Deprecated:` prefix for deprecated symbols

### Grouping with Headings

Use `//` comments with `# Heading` syntax (Go 1.19+) to organize doc sections:

```go
// # Functional Options
//
// The options pattern provides flexible, extensible configuration...
```

### Links and References

Reference other symbols with bracket syntax:

```go
// NewServer creates a [Server] with the given [Option] values applied.
// See [WithPort] and [WithMaxConns] for available options.
```

## Example Functions

Example functions are the most powerful Go documentation feature. They are compiled, tested, and rendered in godoc.

### Naming Convention

```
Example()                  — package-level example
ExampleFoo()               — example for function Foo
ExampleFoo_bar()           — named variant "bar" for Foo
ExampleT_Method()          — example for method T.Method
ExampleT_Method_suffix()   — named variant for T.Method
```

### Writing Examples

Place examples in `example_test.go` files. Use `// Output:` comments to make them testable:

```go
package pipeline_test  // use _test suffix for external test package

import (
    "context"
    "fmt"
    "github.com/example/go-template/internal/pipeline"
)

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
```

Key points:
- Use the `_test` package suffix to show the public API as users see it
- `// Output:` makes the example a test — `go test` verifies the output
- `// Unordered output:` for concurrent results where order varies
- Examples without `// Output:` are compiled but not executed during tests

### This Template's Examples

This project includes testable examples in `example_test.go` files:
- `internal/concurrency/example_test.go` — SafeMap, Semaphore
- `internal/pipeline/example_test.go` — Map, Filter, Reduce, Generator
- `internal/patterns/example_test.go` — functional options, errors, validation

Run them with:
```bash
go test ./internal/... -run Example -v
```

## API Documentation

### Generating Docs Locally

```bash
# Install pkgsite (official godoc server)
go install golang.org/x/pkgsite/cmd/pkgsite@latest

# Serve docs locally
pkgsite -http=:6060
# Visit http://localhost:6060/github.com/example/go-template
```

Or use the simpler `go doc` CLI:

```bash
go doc ./internal/pipeline          # Package overview
go doc ./internal/pipeline.Map      # Specific function
go doc -all ./internal/concurrency  # Full package docs
```

### Structuring API Docs

Organize packages so godoc renders well:
- One concept per package — `pipeline`, `concurrency`, `patterns`
- Group related types and functions together in the source
- Use `doc.go` for long package descriptions
- Keep internal implementation in unexported types

## Project Documentation

### README.md

The README is the entry point. Structure it as:
1. One-line description
2. Quick start (build, run, test)
3. Project structure overview
4. Build and test commands
5. Links to detailed guides

### docs/ Directory

Organize guides by topic, not by package:
- `best-practices.md` — idioms, error handling, formatting
- `concurrency.md` — patterns, pitfalls, context usage
- `performance.md` — profiling, benchmarks, optimization
- `cloud-native.md` — Docker, health checks, observability
- `testing.md` — strategies, table-driven tests, mocking
- `documentation-guide.md` — this file

### Code Comments vs. Docs

| Use code comments for | Use docs/ guides for |
|---|---|
| What a function does | Why a pattern exists |
| Parameter constraints | Architecture decisions |
| Concurrency guarantees | Getting started tutorials |
| Deprecation notices | Comparison of approaches |

## Checklist

When documenting a new package:
- [ ] Package comment starting with `// Package <name>`
- [ ] All exported symbols have comments starting with the symbol name
- [ ] At least one `Example` function in `example_test.go`
- [ ] Examples use `// Output:` for testability
- [ ] `go doc` output reads well
- [ ] Complex packages have a `doc.go` with extended description

## See Also

- [Best Practices](best-practices.md) — Go idioms including naming and comments
- [EXTENDING.md](EXTENDING.md) — Documenting new packages and commands
