# AGENTS.md — internal/

Private, application-specific packages. The Go compiler enforces that everything under
`internal/` can only be imported from within this module — never by external code.

Read the [root AGENTS.md](../AGENTS.md) first for module-wide rules.

## The independence rule (most important)

**No `internal/*` package imports another `internal/*` package.** Each package here is a
standalone reference implementation that can be read, tested, and lifted in isolation. The
only exception: `*_test.go` files may import [`testutil`](testutil/AGENTS.md).

If you feel tempted to import one internal package from another, that is a signal the shared
code belongs elsewhere (the caller, `pkg/`, or the standard library). Preserving independence
keeps the dependency graph acyclic and every package extractable.

## Packages

| Package | One-liner | External deps |
|---------|-----------|---------------|
| [`api/`](api/AGENTS.md) | RESTful handlers, middleware chain, JSON envelope | stdlib only |
| [`cli/`](cli/AGENTS.md) | Viper config loading + table/JSON/YAML output | cobra, viper, yaml |
| [`concurrency/`](concurrency/AGENTS.md) | Worker pools, fan-in, or-done, semaphore, safe map | stdlib only |
| [`db/`](db/AGENTS.md) | `database/sql` pooling, generic repository, migrations, tx | stdlib (+ sqlmock in tests) |
| [`patterns/`](patterns/AGENTS.md) | Functional options, interfaces, embedding, error types | stdlib only |
| [`perf/`](perf/AGENTS.md) | Generic object pool, pre-allocation, escape analysis demos | stdlib only |
| [`pipeline/`](pipeline/AGENTS.md) | Generic ETL / MapReduce / streaming stages | stdlib only |
| [`simulation/`](simulation/AGENTS.md) | Monte Carlo, numerical methods, concurrent runners | stdlib only |
| [`testutil/`](testutil/AGENTS.md) | Generic assertions, HTTP test helpers, mocks | stdlib only (test support) |

## Shared conventions

- **Stdlib-first, generics for reuse.** Only `cli` pulls in third-party deps (Cobra/Viper/YAML).
  The rest are pure standard library. Reusable primitives use type parameters, not `interface{}`.
- **Context-aware concurrency.** `concurrency`, `pipeline`, and `simulation` all follow the same
  discipline: `select` on `ctx.Done()` in every channel loop; close each output channel exactly
  once. Verify with `make test` (race detector on).
- **Justified suppressions only.** `// #nosec G### -- reason` / `//nolint:linter // reason`. See root.
- Every exported identifier carries a doc comment; each package has a runnable demo under
  [`examples/`](../examples/AGENTS.md) and godoc `Example*` tests where illustrative.
