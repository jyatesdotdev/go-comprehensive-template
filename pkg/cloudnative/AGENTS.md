# AGENTS.md — pkg/cloudnative

Cloud-native primitives for 12-factor services: environment config, structured logging (`slog`),
liveness/readiness health checks, and request-logging middleware. Public API — stdlib only.

Read [pkg/AGENTS.md](../AGENTS.md) and the [root](../../AGENTS.md) first. **Exported signatures are
a stable contract** (see pkg rules).

## Exports & contracts

- **`Config{ Port, LogLevel, Environment }`** + **`LoadConfig()`** — reads `PORT` (`8080`),
  `LOG_LEVEL` (`info`), `ENVIRONMENT` (`development`) from the environment with those defaults
  (12-factor: config comes from the environment, not files here).
- **`NewLogger(level) *slog.Logger`** — JSON handler to `os.Stdout`; maps
  `debug|info|warn|error` → `slog.Level`, **defaulting unknown levels to `info`**. Use structured
  key/value logging (`logger.Info("msg", "k", v)`), never `fmt`-built log strings.
- **`HealthChecker`** — `RWMutex`-guarded set of named `CheckFunc` (`func() error`):
  - `AddCheck(name, fn)` registers a **readiness** dependency check.
  - `LivenessHandler()` — always `200 {"status":"alive"}` (process is up). Liveness must **not**
    depend on downstream systems, or Kubernetes will kill a healthy pod during a dependency blip.
  - `ReadinessHandler()` — runs all checks; **any failure → `503`** with a per-check report
    (`ok` or the error string). This is what gates traffic.
- **`RequestLogger(logger, next)`** — middleware logging `method`, `path`, `status`, `duration_ms`;
  wraps the `ResponseWriter` (`responseWriter`) to capture the status code, defaulting to `200`.

## Business rules (preserve these semantics)

- **Liveness ≠ readiness.** Keep liveness dependency-free and readiness dependency-aware. Do not
  merge them.
- Readiness must return **`503` (ServiceUnavailable)**, not `500`, when a check fails.
- Health/logging response-write errors are best-effort (`// #nosec G104`), consistent with the API package.

## Go best practices for cloud-native code

- Prefer `slog` structured fields over string interpolation so logs are queryable.
- Register checks with short internal timeouts (a slow dependency shouldn't hang the probe); the
  `CheckFunc` is responsible for bounding its own work.
- Read config once at startup from the environment; don't scatter `os.Getenv` through the codebase.
- Pair this with `cmd/server`'s graceful shutdown for a complete deployable service.

## Testing

`cloudnative_test.go` drives handlers with `httptest` and asserts status codes and JSON bodies
(e.g. a failing readiness check → `503`). Set env vars to test `LoadConfig` defaults/overrides.
Runnable demo: [`examples/cloudnative`](../../examples/AGENTS.md).
