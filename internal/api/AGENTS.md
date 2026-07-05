# AGENTS.md — internal/api

RESTful API building blocks: a JSON response envelope, request helpers, composable middleware,
and a thread-safe in-memory CRUD store used for demonstration.

Read [internal/AGENTS.md](../AGENTS.md) and the [root](../../AGENTS.md) first.

## Files

- `api.go` — `Response` envelope, `JSON` / `Error` / `Decode` helpers, `Item`, `Store`, `ItemHandler`.
- `middleware.go` — `Middleware` type, `Chain`, `Logging`, `Recovery`, `CORS`, `statusWriter`.
- `api_test.go`, `middleware_test.go` — table-driven tests (import `internal/testutil`).

Depends only on the standard library (`net/http`, `encoding/json`, `sync`, `log`, `time`).

## Key exports & contracts

- **`Response{ Data, Error }`** — the standard envelope; both fields `omitempty`. Success responses
  set `Data`, failures set `Error`. Keep all responses going through this shape.
- **`JSON(w, status, data)` / `Error(w, status, msg)`** — set `Content-Type: application/json`,
  write the status, then encode. Response-write errors are intentionally ignored
  (`// #nosec G104 -- best-effort HTTP response write`).
- **`Decode(r, v)`** — decodes the request body and closes it. Returns the decode error to the caller.
- **`Store`** — in-memory `map[string]Item` guarded by a `sync.RWMutex`. Reads take `RLock`,
  writes take `Lock`. **Any new store method must hold the appropriate lock.** `NewStore()` must be
  used (the map has to be initialized).
- **`ItemHandler(s) http.Handler`** — routes via Go 1.22+ method+pattern muxing.

## Business rules (HTTP contract — preserve these)

| Route | Rule | Status |
|-------|------|--------|
| `GET /items` | list all (empty list is `[]`, not null — pre-sized slice) | 200 |
| `GET /items/{id}` | not in store → `"not found"` | 404 |
| `POST /items` | body must decode; **`id` is required** → else `"id required"` | 400 on bad input, 201 on success |
| `DELETE /items/{id}` | missing id → `"not found"`; else returns `{"deleted": id}` | 404 / 200 |

Read path values with `r.PathValue("id")` (matches the `{id}` wildcard).

## Middleware rules

- **`Chain(h, mws...)`** applies middleware so the **first listed is outermost**. A typical stack
  (see `examples/api`) is `Chain(handler, Recovery, Logging, CORS)` — Recovery outermost so it
  catches panics from everything inside.
- **`Recovery`** is the one sanctioned `recover()` site in the codebase: it logs the panic and
  returns `500 "internal error"`. Do not add `recover()` elsewhere.
- **`Logging`** wraps the writer in `statusWriter` to capture the status code. If you add
  middleware that needs the final status, reuse this wrapping pattern rather than re-reading it.
- **`CORS`** is intentionally permissive (`*`) for template/demo use. **Tighten the allowed origin
  before any production use.**

## Go best practices for this package

- Prefer stdlib `net/http` routing (method + `{wildcard}` patterns) over a third-party router.
- Return early on error; always write exactly one status code per request.
- Keep handlers small; validate input, then act. Guard shared state with the mutex, and hold the
  lock for the shortest span possible (copy out, unlock, then serialize).

## Testing

Table-driven subtests via `testutil.DoRequest(handler, method, path, body)` →
`HTTPResult{Code, Body}`; assert with `testutil.Equal`. Full CRUD lifecycle is also covered
end-to-end in [`tests/`](../../tests/AGENTS.md). Run `make test` (race detector) after changes —
the store is concurrent.
