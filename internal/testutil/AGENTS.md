# AGENTS.md — internal/testutil

Shared test helpers: generic assertions, an HTTP handler test driver, JSON decode, and a mock
notifier. Stdlib only (`testing`, `net/http/httptest`, `encoding/json`).

Read [internal/AGENTS.md](../AGENTS.md) and the [root](../../AGENTS.md) first.

## Special status

This is the **one `internal/` package other packages' tests may import** (e.g.
`internal/api/api_test.go`). It is the sanctioned exception to the "no cross-imports" rule —
*because* it exists, production code never needs to depend on another internal package for testing.

- Keep it **dependency-light** (stdlib only) and **test-support only** — no production logic.
- It is imported widely, so treat its exported signatures as a stable mini-API: changing them
  ripples into many test files. Add helpers rather than repurposing existing ones.

## Exports

- **`Equal[T comparable](t, got, want)`** — fails (non-fatal `Errorf`) on inequality.
- **`NoError(t, err)`** — fatal (`Fatalf`) if `err != nil`.
- **`HasError(t, err)`** — fatal if `err == nil`.
- **`DoRequest(t, handler, method, path, body) HTTPResult`** — drives an `http.Handler` via
  `httptest.NewRequest`/`NewRecorder`; sets `Content-Type: application/json` when `body` is non-empty;
  returns `HTTPResult{Code, Body}`.
- **`DecodeJSON(t, body, v)`** — unmarshals, failing the test with the offending body on error.
- **`MockNotifier`** — records each `Notify` message in `Calls` and returns the configured `Err`;
  satisfies `patterns.Notifier`'s method shape for verification.

## Conventions (follow these when adding helpers)

- **Every helper calls `t.Helper()`** as its first statement, so failures point at the caller's
  line, not here. New helpers must do the same.
- Distinguish **fatal vs non-fatal**: use `t.Fatalf` when continuing is pointless (a nil error you
  were about to dereference), `t.Errorf` when the test can keep gathering failures.
- Take `t *testing.T` as the first arg; keep helpers generic (`[T comparable]`) where it avoids
  per-type duplication.
- No global state; helpers must be safe under `t.Parallel()`.

## Testing

`testutil_test.go` tests the helpers themselves (a `*testing.T` shim / behavioral checks). Because
everything else depends on these, keep them correct and green — a bug here can mask real failures
elsewhere.
