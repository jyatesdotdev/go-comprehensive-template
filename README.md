# Go Comprehensive Template

A production-ready Go project template demonstrating idiomatic patterns for concurrency, cloud-native development, and high-performance computing.

## Quick Start

```bash
make build    # Build the binary
make run      # Build and run
make test     # Run tests with race detector
make lint     # Run vet + staticcheck
```

## Project Structure

```
cmd/server/          Entry point with graceful shutdown
internal/
  api/               RESTful API handlers and middleware
  db/                Database interaction (sql, sqlx, GORM)
  pipeline/          ETL/MapReduce and worker pool patterns
  cli/               CLI helpers (output formatting, config loading)
  concurrency/       Concurrency patterns and examples
  patterns/          Go-idiomatic design patterns
  simulation/        Numerical computing and simulations
pkg/systems/         Systems programming (CGO, syscalls, networking)
docs/                Guides: best practices, performance, cloud-native
examples/            Standalone runnable examples
tests/               Integration and end-to-end tests
```

## Build

```bash
make build              # Local build with version info
make cross              # Cross-compile for linux/darwin/windows
make docker             # Docker multi-stage build (scratch base)
```

Build tags control compilation:
- Default: excludes integration tests
- `integration`: includes integration test files
- `cgo`: enables CGO-dependent code

## Testing

```bash
make test               # Unit tests with -race
make test-integration   # Integration tests
make bench              # Benchmarks with memory stats
make cover              # Coverage report (HTML)
```

## Profiling

```bash
make pprof-cpu          # CPU profile
make pprof-mem          # Memory profile
```

## CLI Example

```bash
go run ./examples/cli greet --name World        # Greeting with flag
go run ./examples/cli list --filter go           # Filtered list output
go run ./examples/cli list --output json         # JSON output mode
go run ./examples/cli config show                # Show resolved config
go run ./examples/cli completion bash            # Generate shell completions
```

The CLI example uses Cobra for command structure, Viper for config/env binding, and `internal/cli` for reusable output formatting (tables, JSON, YAML, colored status).

## Documentation

### Getting Started

- [Tutorial](docs/TUTORIAL.md) — New developer walkthrough: clone, build, test, add a feature
- [Toolchain](docs/TOOLCHAIN.md) — Required tools, install instructions, editor setup
- [Architecture](docs/ARCHITECTURE.md) — Project layout, dependency graph, design decisions
- [Extending](docs/EXTENDING.md) — Adding commands, packages, dependencies, linter rules

### Package Guides

- [RESTful API](docs/restful-api.md) — Handlers, middleware, JSON helpers
- [Database](docs/database.md) — sql, sqlx, GORM, repository pattern
- [Concurrency](docs/concurrency.md) — Goroutines, channels, patterns, pitfalls
- [ETL Pipelines](docs/etl-pipelines.md) — MapReduce, worker pools, batching
- [Design Patterns](docs/design-patterns.md) — Options, observer, strategy, factory
- [Simulation](docs/simulation.md) — Monte Carlo, numerical computing
- [Systems Programming](docs/systems-programming.md) — CGO, syscalls, networking
- [Cloud-Native](docs/cloud-native.md) — Docker, health checks, observability
- [CLI Development](docs/cli.md) — Cobra, Viper, output formatting, shell completion

### Practices & Quality

- [Best Practices](docs/best-practices.md) — Idioms, error handling, package design
- [Performance](docs/performance.md) — Profiling, optimization, GC tuning
- [Testing](docs/testing.md) — Test strategies, table-driven tests, benchmarks
- [Security Scanning](docs/security-scanning.md) — govulncheck, gosec, trivy, nancy
- [Third-Party Libraries](docs/third-party.md) — Popular packages, dependency management
- [Documentation Guide](docs/documentation-guide.md) — godoc, examples, API docs
