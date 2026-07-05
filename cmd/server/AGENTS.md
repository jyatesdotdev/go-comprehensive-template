# AGENTS.md — cmd/server

The application entry point: a minimal HTTP server with a `/health` endpoint and graceful
shutdown. `package main`, one file (`main.go`).

Read the [root AGENTS.md](../../AGENTS.md) first for module-wide rules.

## What this is (and is not)

This is a **deliberately minimal skeleton**, not a finished service. Developers extend it by
wiring in handlers and middleware from `internal/api`, health checks from `pkg/cloudnative`, etc.

- **It imports no `internal/*` or `pkg/*` packages** — only the standard library. This is an
  intentional architecture decision (keeps the entry point trivial and the packages independently
  testable). When you extend it, importing `internal/api` / `pkg/cloudnative` is expected and fine,
  but keep `main` thin: wiring only, no business logic.

## Business rules / invariants

- **Build tag:** the file starts with `//go:build !integration` so it is excluded from
  integration builds. Preserve this tag.
- **Port:** listens on `$PORT`, default `8080` (via the local `envOr` helper).
- **Server timeouts** (keep these — they prevent slow-loris and hung connections):
  `ReadTimeout 5s`, `WriteTimeout 10s`, `IdleTimeout 120s`.
- **Graceful shutdown:** on `SIGINT`/`SIGTERM`, `srv.Shutdown` runs with a **30s** timeout in a
  background goroutine. Do not remove signal handling when adding features.
- **Startup error handling:** `ListenAndServe` returning `http.ErrServerClosed` is the *normal*
  shutdown path and must not be treated as fatal; any other error is `log.Fatal`.
- **Health endpoint:** `/health` returns `200` with `{"status":"ok"}`. If you add real readiness
  logic, prefer `pkg/cloudnative`'s `HealthChecker` (liveness vs readiness) rather than hand-rolling.

## Version metadata (known gap to fill when needed)

`make build` injects `-X main.version=… -X main.commit=… -X main.date=…`, but `main.go` does
**not yet declare** those package-level `var`s, so the linker flags are currently no-ops. If you
add a `--version` flag or version logging, first declare:

```go
var (
    version = "dev"
    commit  = "none"
    date    = "unknown"
)
```

so the `-ldflags` values actually bind.

## Go best practices for entry points

- Keep `main` short: parse config/env, construct dependencies, wire routes, start the server,
  block on shutdown. Push logic into packages that can be unit-tested.
- Return non-zero exit codes on fatal startup failure (`log.Fatal` / `os.Exit(1)`).
- Always give `Shutdown` a bounded context so a hung request cannot block termination forever.
- Prefer `http.ServeMux` method+pattern routing (Go 1.22+), matching `internal/api`.

## Testing

No `main_test.go` here — logic lives in packages. End-to-end HTTP behavior is covered by
`tests/` (integration, build-tagged). If you add non-trivial wiring, cover it there.
