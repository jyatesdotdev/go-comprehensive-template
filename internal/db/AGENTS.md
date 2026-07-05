# AGENTS.md — internal/db

`database/sql` patterns: connection-pool configuration, a transaction helper, a generic
repository, and a minimal migration runner. **Driver-agnostic** — no SQL driver is imported here.

Read [internal/AGENTS.md](../AGENTS.md) and the [root](../../AGENTS.md) first.

## Files

- `db.go` — `Config`, `DefaultConfig`, `Open`, `HealthCheck`, `InTx`.
- `repository.go` — `ErrNotFound`, `Scanner`, generic `Repository[T]` (`FindByID`, `FindAll`, `ExecInsert`).
- `migrations.go` — `Migration`, `Migrator` (`Up`, plus internal `ensureTable`/`applied`).
- `*_test.go` — use `github.com/DATA-DOG/go-sqlmock` (the only non-stdlib dep, test-only).

## Connection & pooling (`db.go`)

- The **caller registers the driver** (blank-import e.g. `_ "github.com/lib/pq"`); this package
  only calls `sql.Open(cfg.Driver, cfg.DSN)`. Do not add a driver dependency here.
- `DefaultConfig` pool settings (tune per workload, but these are the sane defaults):
  `MaxOpenConns 25`, `MaxIdleConns 5`, `ConnMaxLifetime 5m`, `ConnMaxIdleTime 1m`.
- **`Open` pings** and closes the handle if the ping fails — a returned `*sql.DB` is verified live.
- **`HealthCheck`** wraps `PingContext` with a **2s** timeout.

## Transactions (`InTx`) — invariant

`InTx(ctx, db, opts, fn)` runs `fn` in a transaction and:

- **commits** if `fn` returns nil,
- **rolls back** if `fn` returns an error, and
- on **panic**, rolls back and **re-panics** (so a panic is never silently swallowed).

It uses a **named return `err`** so the `defer` can see `fn`'s error — do not "clean up" that named
return; it is load-bearing (hence the `//nolint:gocritic`). Always route multi-statement writes
(including the migrator's per-migration work) through `InTx`.

## Repository[T] (`repository.go`)

- Generic over the entity type; you supply `Scan func(Scanner) (T, error)`. `Scanner` abstracts
  `*sql.Row` and `*sql.Rows` so one scan func serves single-row and multi-row reads.
- `FindByID` translates `sql.ErrNoRows` → the package sentinel **`ErrNotFound`** (check with
  `errors.Is`). Preserve that mapping.
- `FindAll` closes `rows` (best-effort `//nolint:errcheck`) and returns `rows.Err()` after the loop.

## SQL safety (critical, read before touching any query string)

Queries use `fmt.Sprintf` to inject **table/column names**, carrying `// #nosec G201 -- table/
column names are developer-controlled struct fields`. This is safe **only** because those
identifiers come from `Repository.Table` / method args set in code, never from user input.

- **Never** pass user-supplied values as table or column names.
- **Always** pass *values* as bound parameters (`?` / `args...`), never via string formatting.
- Placeholders here are `?` (SQLite/MySQL style). For PostgreSQL (`$1, $2, …`) adapt the
  placeholder style in the query you write — the driver does not translate it.

## Migrations (`migrations.go`)

- State table defaults to `schema_migrations` (`version` PK, `description`, `applied_at`).
- `Up` is **idempotent**: it ensures the table, reads applied versions, sorts pending migrations by
  `Version` ascending, and applies each **in its own transaction**, recording `applied_at` in UTC.
- Give each migration a **unique, monotonically increasing `Version`**. `Down` is defined on the
  struct but the runner currently only applies `Up` — add a `Down`/rollback path deliberately if needed.

## Testing

Use `sqlmock`: `sqlmock.New(sqlmock.MonitorPingsOption(true))`, set expectations
(`ExpectPing`, `ExpectQuery`, `ExpectExec`, `ExpectBegin/Commit/Rollback`), then assert
`mock.ExpectationsMet()`. No real database required — keep it that way so `make test` stays
hermetic. `sqlclosecheck` + `rows.Close()` guard against leaked result sets.
