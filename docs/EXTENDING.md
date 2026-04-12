# Extending the Template

How to add new components to this project. Each section includes the steps and a minimal working example.

> See also: [ARCHITECTURE.md](ARCHITECTURE.md) for project layout, [TOOLCHAIN.md](TOOLCHAIN.md) for required tools.

---

## Adding a New Command

Commands live under `cmd/`. Each command is a separate `main` package that compiles to its own binary.

### Steps

1. Create `cmd/<name>/main.go`
2. Add a build target to the `Makefile` (or use `go build ./cmd/<name>`)
3. Add a run target if desired

### Example: `cmd/worker/main.go`

```go
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	log.Println("worker started")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	fmt.Println("worker stopped")
}
```

### Makefile addition

```makefile
build-worker:
	go build $(LDFLAGS) -o bin/worker ./cmd/worker
```

---

## Adding a New Internal Package

Internal packages live under `internal/`. They are private to this module — other modules cannot import them.

### Steps

1. Create `internal/<name>/<name>.go`
2. Add a package doc comment as the first line
3. Add `_test.go` alongside
4. Optionally add an `example_test.go` for runnable godoc examples

### Example: `internal/cache/cache.go`

```go
// Package cache provides a simple in-memory TTL cache.
package cache

import (
	"sync"
	"time"
)

// Cache is a thread-safe key-value store with expiration.
type Cache struct {
	mu    sync.RWMutex
	items map[string]entry
}

type entry struct {
	value     string
	expiresAt time.Time
}

// New creates an empty Cache.
func New() *Cache {
	return &Cache{items: make(map[string]entry)}
}

// Set stores a value with the given TTL.
func (c *Cache) Set(key, value string, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = entry{value: value, expiresAt: time.Now().Add(ttl)}
}

// Get retrieves a value. Returns ("", false) if missing or expired.
func (c *Cache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.items[key]
	if !ok || time.Now().After(e.expiresAt) {
		return "", false
	}
	return e.value, true
}
```

### Test file: `internal/cache/cache_test.go`

```go
package cache

import (
	"testing"
	"time"
)

func TestCache_SetGet(t *testing.T) {
	c := New()
	c.Set("k", "v", time.Minute)
	got, ok := c.Get("k")
	if !ok || got != "v" {
		t.Fatalf("Get(k) = %q, %v; want %q, true", got, ok, "v")
	}
}
```

---

## Adding a New `pkg` Library

Packages under `pkg/` are public — they can be imported by other modules. Use `pkg/` for reusable, self-contained libraries with stable APIs.

### When to use `pkg/` vs `internal/`

| Use `pkg/`                          | Use `internal/`                     |
|-------------------------------------|-------------------------------------|
| Reusable across projects            | Specific to this application        |
| Stable, versioned API               | Free to change without notice       |
| No dependency on app internals      | May depend on other internal pkgs   |

### Steps

1. Create `pkg/<name>/<name>.go`
2. Add a package doc comment
3. Keep dependencies minimal (prefer stdlib-only)
4. Add an example under `examples/<name>/main.go`

### Example: `pkg/retry/retry.go`

```go
// Package retry provides configurable retry logic with exponential backoff.
package retry

import (
	"fmt"
	"time"
)

// Do calls fn up to maxAttempts times. Waits delay between attempts,
// doubling the delay each time. Returns the last error if all attempts fail.
func Do(maxAttempts int, delay time.Duration, fn func() error) error {
	var err error
	for i := range maxAttempts {
		if err = fn(); err == nil {
			return nil
		}
		if i < maxAttempts-1 {
			time.Sleep(delay)
			delay *= 2
		}
	}
	return fmt.Errorf("after %d attempts: %w", maxAttempts, err)
}
```

---

## Adding a Third-Party Dependency

### Steps

1. Add the dependency:
   ```bash
   go get github.com/some/package@latest
   ```
2. Import and use it in your code
3. Run `go mod tidy` to clean up `go.mod` / `go.sum`

### Best practice: wrap behind an interface

Don't scatter third-party types through your codebase. Define an interface in your code and write a thin adapter. This lets you swap implementations without changing business logic.

See `examples/thirdparty/main.go` for a full working example of this pattern. The key idea:

```go
// Your interface (in your code)
type Logger interface {
    Info(msg string, args ...any)
    Error(msg string, args ...any)
}

// Adapter (in your code) — wraps the third-party type
type SlogAdapter struct{ l *slog.Logger }

func (s *SlogAdapter) Info(msg string, args ...any)  { s.l.Info(msg, args...) }
func (s *SlogAdapter) Error(msg string, args ...any) { s.l.Error(msg, args...) }
```

Your application depends on `Logger`, not `*slog.Logger`. Swapping to zap or zerolog means writing a new adapter — zero changes to business logic.

### Current dependencies

This template uses three external dependencies (see `go.mod`):

- `github.com/spf13/cobra` — CLI command framework
- `github.com/spf13/viper` — configuration management
- `gopkg.in/yaml.v3` — YAML parsing

---

## Adding a New Linter Rule

Linter configuration lives in `.golangci.yml` at the project root.

### Steps

1. Find the linter name in the [golangci-lint docs](https://golangci-lint.run/usage/linters/)
2. Add it to the `linters.enable` list
3. Optionally configure it under `linters-settings`
4. Run `golangci-lint run ./...` to verify

### Example: enable `nilerr` (catches returning nil when err is non-nil)

```yaml
linters:
  enable:
    # ... existing linters ...
    - nilerr
```

### Example: add settings for an existing linter

```yaml
linters-settings:
  revive:
    rules:
      - name: exported
        severity: warning
```

### Currently enabled linters

| Category    | Linters                                          |
|-------------|--------------------------------------------------|
| Correctness | errcheck, govet, staticcheck, unused, gosimple, ineffassign |
| Style       | gocritic, revive, misspell, prealloc             |
| Security    | gosec, bodyclose, sqlclosecheck                  |

### Suppressing a false positive

Use a `//nolint` directive with the linter name:

```go
val := doSomething() //nolint:errcheck // intentionally ignoring error
```

---

## Checklist

After adding any new component:

- [ ] `go build ./...` compiles cleanly
- [ ] `go test ./...` passes
- [ ] `go vet ./...` reports no issues
- [ ] `golangci-lint run ./...` passes (if installed)
- [ ] New exported symbols have godoc comments starting with the symbol name
