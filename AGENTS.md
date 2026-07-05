# AGENTS.md — go-template (root index)

Production-ready Go project template demonstrating idiomatic patterns for concurrency,
cloud-native development, high-performance computing, APIs, databases, and CLIs. Every
package is a self-contained, well-documented reference implementation.

- **Module:** `github.com/example/go-template`
- **Go version:** 1.26 (see `go.mod` — do not downgrade the language version)
- **Nature:** a *template*. Code here is meant to be read, copied, and adapted. Prefer
  clarity and idiomatic style over cleverness. Keep every package independently understandable.

This file is the index. Each directory that contains source has its own `AGENTS.md` with
package-specific rules. **Before editing any file, read the `AGENTS.md` in its directory
and the ones above it** — nested files add to (and override) the ones above.

## Directory index

| Location | Purpose | AGENTS.md |
|----------|---------|-----------|
| `cmd/server/` | HTTP server entry point (graceful shutdown skeleton) | [cmd/server/AGENTS.md](cmd/server/AGENTS.md) |
| `internal/` | Private packages (app-specific, not externally importable) | [internal/AGENTS.md](internal/AGENTS.md) |
| `internal/api/` | RESTful handlers, middleware, JSON helpers | [internal/api/AGENTS.md](internal/api/AGENTS.md) |
| `internal/cli/` | Cobra/Viper config loading + output formatting | [internal/cli/AGENTS.md](internal/cli/AGENTS.md) |
| `internal/concurrency/` | Worker pools, fan-in, semaphores, safe maps | [internal/concurrency/AGENTS.md](internal/concurrency/AGENTS.md) |
| `internal/db/` | `database/sql` pooling, repository, migrations, tx | [internal/db/AGENTS.md](internal/db/AGENTS.md) |
| `internal/patterns/` | Functional options, interfaces, errors, embedding | [internal/patterns/AGENTS.md](internal/patterns/AGENTS.md) |
| `internal/perf/` | Object pooling, pre-allocation, escape analysis | [internal/perf/AGENTS.md](internal/perf/AGENTS.md) |
| `internal/pipeline/` | ETL / MapReduce / streaming stages (generics) | [internal/pipeline/AGENTS.md](internal/pipeline/AGENTS.md) |
| `internal/simulation/` | Monte Carlo, numerical methods, concurrent runners | [internal/simulation/AGENTS.md](internal/simulation/AGENTS.md) |
| `internal/testutil/` | Generic test assertions, HTTP test helpers, mocks | [internal/testutil/AGENTS.md](internal/testutil/AGENTS.md) |
| `pkg/` | Public packages (stable, externally importable) | [pkg/AGENTS.md](pkg/AGENTS.md) |
| `pkg/cloudnative/` | Config, health checks, slog logging, observability | [pkg/cloudnative/AGENTS.md](pkg/cloudnative/AGENTS.md) |
| `pkg/systems/` | File I/O, TCP/UDP networking, OS interaction | [pkg/systems/AGENTS.md](pkg/systems/AGENTS.md) |
| `examples/` | Standalone runnable demos (one per package) | [examples/AGENTS.md](examples/AGENTS.md) |
| `tests/` | Integration / end-to-end tests (build-tagged) | [tests/AGENTS.md](tests/AGENTS.md) |

`docs/` holds long-form guides (Markdown, no code); `.github/workflows/` holds CI. Neither
has an `AGENTS.md`. When you change behavior documented in `docs/`, update the matching guide.

## Architecture rules (apply everywhere)

1. **Packages are independent.** No `internal/*` package imports another `internal/*` package.
   The only exception is that `*_test.go` files may import `internal/testutil`. Do not introduce
   cross-package dependencies — if you need shared logic, it probably belongs in the caller.
   See `docs/ARCHITECTURE.md` for the dependency graph.
2. **`internal/` vs `pkg/`.** `internal/` is app-specific and compiler-enforced private. `pkg/`
   is general-purpose and importable by other modules — treat its exported API as stable
   (see [pkg/AGENTS.md](pkg/AGENTS.md)).
3. **`cmd/server` imports no internal packages.** It is a minimal, extensible skeleton by design.
4. **Stdlib-first.** Most packages depend only on the standard library. The only production
   third-party dependencies are Cobra + Viper (used only by `internal/cli`) and the YAML encoder;
   `internal/db` tests use `go-sqlmock`. Do not add dependencies without a strong reason.
5. **Generics for reusable primitives.** `pipeline`, `concurrency`, `perf`, and `db` use type
   parameters (`Stage[In, Out]`, `WorkerPool[T, R]`, `Pool[T]`, `Repository[T]`) instead of
   `interface{}` boxing. Follow that style when extending them.
6. **Context first.** Any function that blocks, does I/O, or spawns goroutines takes
   `ctx context.Context` as its first parameter and honors cancellation.

## Go conventions (apply everywhere)

- **Errors:** wrap with `fmt.Errorf("context: %w", err)`. Compare sentinels with `errors.Is`,
  extract typed errors with `errors.As`. Return errors; never `panic` in library code (the API
  `Recovery` middleware is the one deliberate recover site).
- **Doc comments:** every exported identifier has a doc comment starting with its name. Match the
  existing density and voice when adding code.
- **Concurrency:** goroutines must have a clear exit path — always `select` on `ctx.Done()` in
  loops, and close each output channel exactly once (via `defer` or a single `wg.Wait()` closer).
  Run `make test` (which enables `-race`) after any concurrency change.
- **Formatting:** `gofmt -s` + `goimports` (`make fmt`). Tabs, not spaces.

### Security & linter suppression comments — required format

Suppressions are used deliberately and **must carry a justification**:

- `// #nosec G### -- reason` for gosec (e.g. `// #nosec G104 -- best-effort HTTP response write`).
- `//nolint:linter // reason` for golangci-lint (e.g. `//nolint:errcheck // best-effort close`).

Never add a bare suppression. Never suppress to hide a real issue. gosec runs at
`severity: error` in CI — a new unjustified finding fails the build. Note the standing
invariant in `internal/db`: SQL identifiers (table/column names) are developer-controlled
struct fields, never user input — that is why its `fmt.Sprintf` queries carry `#nosec G201`.
Do not route user input into those positions.

## Commands (Makefile)

| Command | What it does |
|---------|--------------|
| `make build` | Build `cmd/server` → `bin/server` with version/commit/date via `-ldflags` |
| `make run` | Build then run |
| `make test` | Unit tests, `-race -count=1` (run this after every change) |
| `make test-integration` | Tests tagged `integration` (`tests/`, build tag) |
| `make bench` | Benchmarks with `-benchmem` |
| `make cover` | HTML coverage report |
| `make lint` | `go vet` + `staticcheck` + `golangci-lint` |
| `make fmt` | `gofmt -s -w` + `goimports -w` |
| `make security` | `govulncheck`, `gosec`, `nancy`, `trivy` (also gated in CI) |
| `make cross` | Cross-compile linux/darwin/windows |
| `make docker` | Multi-stage build → `scratch` (non-root UID 65534) |

## Build tags

- Default build **excludes** integration tests. `cmd/server/main.go` carries `//go:build !integration`.
- `tests/` carries `//go:build integration`; include with `-tags=integration`.
- `cgo` gates CGO-dependent code where present.

## Definition of done for any change

1. `make fmt` — formatted.
2. `make test` — unit tests pass with the race detector.
3. `make lint` — no new linter findings (add justified suppressions only when truly warranted).
4. Exported identifiers documented; the package's `AGENTS.md` still accurate (update it if not).
5. If you touched a `pkg/` exported signature, confirm it is a deliberate, backward-compatible change.
