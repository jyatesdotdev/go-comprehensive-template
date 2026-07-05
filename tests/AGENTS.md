# AGENTS.md — tests/

Integration and end-to-end tests that exercise packages together (as opposed to the unit tests
that live beside each package's source).

Read the [root AGENTS.md](../AGENTS.md) first.

## Build tag & how to run

- Every file here starts with **`//go:build integration`** and declares `package tests`.
- These are **excluded from the default `make test`** so unit-test runs stay fast.
- Run them explicitly:

```bash
make test-integration        # go test -race -tags=integration -count=1 ./...
```

**Any new file in this directory must carry the `//go:build integration` tag**, or it will run in
the default suite (and may pull heavy setup into every `make test`).

## What belongs here

- Cross-package / end-to-end flows. The current `integration_test.go` spins up a real
  `httptest.NewServer(api.ItemHandler(...))` and drives the full CRUD lifecycle over HTTP against
  `internal/api`.
- Tests that need external resources (a real database, network services, containers). This tag is
  the seam that keeps such tests out of the fast inner loop and lets CI run them separately.

Unit tests do **not** belong here — put those in `*_test.go` next to the code they test.

## Conventions

- Import the packages under test (`internal/*`, `pkg/*`) and exercise them through their public
  surface, as a real consumer would. Assert on observable behavior (status codes, bodies, side
  effects), not internals.
- Prefer `httptest` servers and ephemeral ports / `t.TempDir()` over fixed ports and shared paths,
  so tests are hermetic and parallel-safe.
- `-race` is on — integration code spawns goroutines/servers, so keep it race-clean.
- Always clean up (`defer srv.Close()`, close response bodies) so a failing test doesn't leak
  resources into the next.

## Go best practices

- Gate anything slow or dependency-heavy behind the build tag; keep the default suite hermetic.
- Make integration tests idempotent and independent — no ordering assumptions, no shared mutable
  fixtures between tests.
- If you introduce a real external dependency, document how to provision it (and wire it into
  `.github/workflows/` CI) so `make test-integration` is reproducible.
