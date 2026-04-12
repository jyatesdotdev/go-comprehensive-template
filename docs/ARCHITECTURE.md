# Architecture

This document describes the project layout, package relationships, build system, and key design decisions.

## Project Layout

```
go-template/
├── cmd/server/           Application entry point
├── internal/             Private packages (not importable by external modules)
│   ├── api/              RESTful handlers, middleware, JSON helpers
│   ├── db/               Database connection pooling, transactions, migrations
│   ├── pipeline/         ETL, MapReduce, streaming data pipelines (generics)
│   ├── concurrency/      Worker pools, fan-out/fan-in, semaphores, safe maps
│   ├── patterns/         Functional options, composition, error types
│   ├── simulation/       Monte Carlo, numerical methods, concurrent runners
│   ├── perf/             Object pooling, pre-allocation helpers
│   ├── cli/              Config loading (Viper), output formatting (table/JSON/YAML)
│   └── testutil/         Generic test helpers, HTTP test utilities, mocks
├── pkg/                  Public packages (importable by external modules)
│   ├── systems/          File I/O, networking, OS interaction
│   └── cloudnative/      Config loading, health checks, structured logging
├── examples/             Standalone runnable examples (one per package)
├── tests/                Integration and end-to-end tests
├── docs/                 Documentation guides
├── .github/workflows/    CI: security scanning pipeline
├── Makefile              Build, test, lint, security, profiling targets
├── Dockerfile            Multi-stage build (golang:alpine → scratch)
└── .golangci.yml         Linter configuration
```

## Package Dependency Graph

The packages are intentionally independent — no cross-dependencies between internal packages. Each package depends only on the standard library (with two exceptions noted below).

```
cmd/server ──→ stdlib only (net/http, os, context)

internal/api        ──→ stdlib (net/http, encoding/json, sync)
internal/db         ──→ stdlib (database/sql, context)
internal/pipeline   ──→ stdlib (context, sync)
internal/concurrency──→ stdlib (context, sync)
internal/patterns   ──→ stdlib (errors, fmt, time)
internal/simulation ──→ stdlib (context, math, sync)
internal/perf       ──→ stdlib (sync)
internal/cli        ──→ github.com/spf13/cobra, github.com/spf13/viper
internal/testutil   ──→ stdlib (testing, net/http/httptest)

pkg/systems         ──→ stdlib (net, os, runtime, io, bufio)
pkg/cloudnative     ──→ stdlib (log/slog, net/http, encoding/json)
```

Test dependencies:
- `internal/api/api_test.go` imports `internal/testutil` for HTTP test helpers.
- `tests/integration_test.go` imports `internal/api` for end-to-end API testing.

Each `examples/<name>/main.go` imports its corresponding package:

| Example | Imports |
|---------|---------|
| `examples/api` | `internal/api` |
| `examples/database` | `internal/db` |
| `examples/pipeline` | `internal/pipeline` |
| `examples/concurrency` | `internal/concurrency` |
| `examples/patterns` | `internal/patterns` |
| `examples/simulation` | `internal/simulation` |
| `examples/performance` | `internal/perf` |
| `examples/cli` | `internal/cli` |
| `examples/cloudnative` | `pkg/cloudnative` |
| `examples/systems` | `pkg/systems` |

## Makefile Targets

### Build & Run

| Target | Description |
|--------|-------------|
| `make build` | Compiles `cmd/server` to `bin/server` with version/commit/date embedded via `-ldflags` |
| `make run` | Builds then runs the binary |
| `make cross` | Cross-compiles for linux/darwin (amd64+arm64) and windows (amd64) |
| `make docker` | Multi-stage Docker build — compiles in `golang:alpine`, runs from `scratch` |
| `make clean` | Removes `bin/` and coverage output |

Version metadata is injected at build time:
```
-ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"
```

### Testing & Quality

| Target | Description |
|--------|-------------|
| `make test` | Runs all unit tests with `-race -count=1` |
| `make test-integration` | Runs tests tagged `integration` |
| `make bench` | Runs benchmarks with `-benchmem` |
| `make cover` | Generates HTML coverage report |
| `make lint` | Runs `go vet`, `staticcheck`, and `golangci-lint` |
| `make fmt` | Formats code with `gofmt` and `goimports` |

### Profiling

| Target | Description |
|--------|-------------|
| `make pprof-cpu` | Generates and opens a CPU profile |
| `make pprof-mem` | Generates and opens a memory profile |

### Security Scanning

| Target | Description |
|--------|-------------|
| `make security` | Runs all four scanners below |
| `make security-govulncheck` | Checks against the Go vulnerability database |
| `make security-gosec` | Static security analysis of Go source |
| `make security-nancy` | Scans dependencies for known CVEs |
| `make security-trivy` | Filesystem vulnerability scan (HIGH/CRITICAL) |

The same scanners run in CI via `.github/workflows/security.yml` on push, PR, and weekly schedule. A `security-gate` job gates merges on all scanners passing.

## Design Decisions

### `internal/` vs `pkg/`

The Go compiler enforces that `internal/` packages can only be imported by code within the same module. This template uses that distinction deliberately:

- **`internal/`** — Application-specific logic that is tightly coupled to this project's domain. These packages demonstrate patterns (API handlers, DB access, pipelines) but are not designed as reusable libraries. Keeping them internal prevents accidental external coupling.

- **`pkg/`** — General-purpose utilities (`systems`, `cloudnative`) that could be extracted into standalone modules. These have no dependencies on other project packages and maintain stable, documented APIs.

### Standalone `cmd/server`

The server entry point (`cmd/server/main.go`) deliberately avoids importing any internal packages. It is a minimal HTTP server with graceful shutdown — a skeleton that developers extend by wiring in `internal/api` handlers and middleware. This keeps the entry point simple and makes the internal packages independently testable.

### Independent Packages

No internal package imports another internal package (except `testutil` in tests). This means:
- Each package can be understood, tested, and modified in isolation.
- There are no circular dependency risks.
- Packages can be extracted to separate modules without untangling imports.

### Generics for Reusable Patterns

The `pipeline`, `concurrency`, and `perf` packages use Go generics to provide type-safe, reusable primitives (`Stage[In, Out]`, `WorkerPool[T, R]`, `Pool[T]`) without requiring interface boxing or type assertions.

### Build Tags

- Default build excludes integration tests (`// +build !integration` on `cmd/server`).
- `make test-integration` uses `-tags=integration` to include them.
- This keeps `make test` fast for development while integration tests run in CI.

### Docker: Scratch Base Image

The Dockerfile uses a multi-stage build: compile with `golang:alpine`, deploy from `scratch`. The final image contains only the static binary and CA certificates — no shell, no package manager, minimal attack surface. It runs as non-root (UID 65534).

### Linting Strategy

`.golangci.yml` enables linters in three categories:
- **Correctness**: `errcheck`, `govet`, `staticcheck`, `unused`, `gosimple`, `ineffassign`
- **Style & quality**: `gocritic`, `revive`, `misspell`, `prealloc`
- **Security**: `gosec`, `bodyclose`, `sqlclosecheck`

Security linters are configured with `severity: error` to fail CI on findings.

## Related Documentation

- [Best Practices](best-practices.md) — Go idioms, error handling, package design
- [Concurrency](concurrency.md) — Goroutine patterns and pitfalls
- [Performance](performance.md) — Profiling and optimization
- [Cloud-Native](cloud-native.md) — Docker, health checks, observability
- [Security Scanning](security-scanning.md) — Scanner details and CI integration
- [Testing](testing.md) — Test strategy and helpers
- [TOOLCHAIN.md](TOOLCHAIN.md) — Required tools and installation
- [EXTENDING.md](EXTENDING.md) — How to add new packages and commands
- [TUTORIAL.md](TUTORIAL.md) — New developer walkthrough
