# AGENTS.md — examples/

Standalone, runnable demonstrations — one directory per topic. Each is `package main` with a
`main.go` you can run directly:

```bash
go run ./examples/<name>
```

Read the [root AGENTS.md](../AGENTS.md) first. This file covers **all** example directories; there
is no separate `AGENTS.md` per example because they share the same conventions.

## The examples and what they demonstrate

| Directory | Demonstrates | Imports |
|-----------|--------------|---------|
| `api/` | RESTful CRUD server + HTTP client round-trip | `internal/api` |
| `cli/` | Cobra subcommands, Viper config/env, table/JSON/YAML output (has `main_test.go`) | `internal/cli` |
| `cloudnative/` | Env config, health checks, structured logging, graceful shutdown | `pkg/cloudnative` |
| `concurrency/` | Goroutines, channels, select, context, sync patterns | `internal/concurrency` |
| `database/` | Connection config, migrations, repository, transactions | `internal/db` |
| `errors/` | Sentinel errors (`errors.Is`), typed errors (`errors.As`), wrapping | stdlib only |
| `patterns/` | Functional options, interfaces, embedding, error types | `internal/patterns` |
| `performance/` | pprof integration, memory profiling, allocation-conscious code | `internal/perf` |
| `pipeline/` | ETL, MapReduce, streaming stages | `internal/pipeline` |
| `simulation/` | Concurrent Monte Carlo, numerical methods, parallel sims | `internal/simulation` |
| `systems/` | TCP/UDP networking, atomic file I/O, system info | `pkg/systems` |
| `thirdparty/` | Wrapping deps behind interfaces, adapter pattern, dependency injection | stdlib only |

`errors/` and `thirdparty/` have **no backing package** — they are pure-stdlib pattern demos and
own their illustrative types inline.

## Conventions & rules

- **Each example is `package main` and self-contained.** It imports at most its one backing package
  (plus stdlib / Cobra+Viper for `cli`). Do not import one example from another; do not add
  cross-example shared code.
- **Examples must stay runnable.** If you change a backing package's exported API, update every
  example that uses it so `go run ./examples/<name>` and `go build ./...` still succeed.
- **Examples are excluded from linting and formatting.** `.golangci.yml` lists `examples$` under
  both linter and formatter exclusions, so relaxed error handling for brevity (`resp, _ := ...`) is
  acceptable *here* to keep demos readable — but `go vet` and `go build` still apply. Do **not**
  copy that looseness back into `internal/` or `pkg/`.
- **Optimize for teaching.** Prefer clear, linear, well-commented `main` functions with printed
  output that shows what happened. This is documentation-as-code.
- **Examples may have tests** (`cli/main_test.go` does). Add a `main_test.go` when a demo has logic
  worth pinning; otherwise `go build ./examples/...` is the smoke test.

## Relationship to godoc examples

These runnable `main` programs are distinct from the `Example*` functions in the packages'
`example_test.go` files (which appear on pkg.go.dev and run under `go test`). When you add a
feature, consider updating **both**: the godoc `Example*` for API-level docs and the `examples/`
demo for an end-to-end runnable tour.
