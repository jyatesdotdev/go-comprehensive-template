# Testing in Go

Go has first-class testing support via the `testing` package and `go test` command.

## Running Tests

```bash
go test ./...                          # all unit tests
go test -v ./internal/api/...          # verbose, single package
go test -run TestItemHandler ./...     # filter by name
go test -race ./...                    # enable race detector
go test -tags=integration ./tests/...  # integration tests
go test -bench=. ./...                 # run benchmarks
go test -cover -coverprofile=cover.out ./...  # coverage
go tool cover -html=cover.out          # view coverage in browser
```

## Unit Tests

Test files live alongside the code they test (`foo_test.go` next to `foo.go`).

```go
func TestAdd(t *testing.T) {
    got := Add(2, 3)
    if got != 5 {
        t.Errorf("Add(2,3) = %d, want 5", got)
    }
}
```

## Table-Driven Tests

The standard Go pattern for testing multiple cases:

```go
func TestValidate(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"empty", "", true},
        {"valid", "Alice", false},
        {"too long", strings.Repeat("x", 51), true},
    }
    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            err := Validate(tc.input)
            if (err != nil) != tc.wantErr {
                t.Errorf("Validate(%q) err=%v, wantErr=%v", tc.input, err, tc.wantErr)
            }
        })
    }
}
```

## Subtests

`t.Run` creates named subtests that can be filtered with `-run`:

```bash
go test -run TestValidate/empty ./...
```

## HTTP Testing

Use `net/http/httptest` for testing handlers without a real server:

```go
func TestHandler(t *testing.T) {
    req := httptest.NewRequest("GET", "/items", nil)
    rec := httptest.NewRecorder()
    handler.ServeHTTP(rec, req)
    if rec.Code != http.StatusOK {
        t.Errorf("status = %d, want 200", rec.Code)
    }
}
```

For full integration tests, use `httptest.NewServer`:

```go
srv := httptest.NewServer(handler)
defer srv.Close()
resp, err := http.Get(srv.URL + "/items")
```

## Mocking via Interfaces

Go favors mocking through interfaces rather than frameworks:

```go
type Notifier interface {
    Notify(msg string) error
}

type mockNotifier struct {
    calls []string
    err   error
}

func (m *mockNotifier) Notify(msg string) error {
    m.calls = append(m.calls, msg)
    return m.err
}
```

Accept interfaces, return structs. This makes code naturally testable.

## Test Helpers

Create reusable helpers with `t.Helper()` so failures report the caller's line:

```go
func Equal[T comparable](t *testing.T, got, want T) {
    t.Helper()
    if got != want {
        t.Errorf("got %v, want %v", got, want)
    }
}
```

See `internal/testutil/` for the project's shared helpers.

## Benchmarks

Benchmark functions start with `Benchmark` and use `*testing.B`:

```go
func BenchmarkProcess(b *testing.B) {
    for b.Loop() {
        Process(data)
    }
}
```

Run with `go test -bench=. -benchmem ./...`. Use `b.RunParallel` for concurrent benchmarks.

## Integration Tests

Use build tags to separate integration tests from unit tests:

```go
//go:build integration

package tests

func TestDatabase_Integration(t *testing.T) {
    // requires running database
}
```

Run with `go test -tags=integration ./tests/...`.

## Race Detection

Always run tests with `-race` in CI:

```bash
go test -race ./...
```

This detects data races at runtime. Write concurrent tests to exercise shared state.

## Project Test Files

| File | Demonstrates |
|------|-------------|
| `internal/api/api_test.go` | Table-driven HTTP tests, httptest |
| `internal/concurrency/concurrency_test.go` | Race-safe tests, context cancellation, parallel benchmarks |
| `internal/patterns/patterns_test.go` | errors.Is/As, mocking via interfaces |
| `internal/pipeline/pipeline_test.go` | Channel-based pipeline tests, benchmarks |
| `internal/perf/perf_test.go` | Benchmark comparisons, table-driven benchmarks |
| `tests/integration_test.go` | Build-tagged integration test |
| `internal/testutil/testutil.go` | Shared test helpers |

## See Also

- [Best Practices](best-practices.md) — Error handling and interface design
- [Performance](performance.md) — Benchmarks and profiling
- [TUTORIAL.md](TUTORIAL.md) — Running tests as part of the developer workflow
