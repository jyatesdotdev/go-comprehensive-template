# Go Best Practices Guide

## Table of Contents

- [Go Idioms](#go-idioms)
- [Error Handling](#error-handling)
- [Package Design](#package-design)
- [Formatting & Linting](#formatting--linting)

---

## Go Idioms

### Naming Conventions

```go
// Use MixedCaps (exported) and mixedCaps (unexported). Never underscores.
type HTTPClient struct{}  // Acronyms are all-caps
func ServeHTTP()          // Not ServeHttp

// Single-letter receivers are idiomatic
func (s *Server) Start() error { ... }

// Interface names: single-method interfaces use -er suffix
type Reader interface { Read(p []byte) (n int, err error) }

// Getters omit "Get" prefix
func (u *User) Name() string    // Not GetName()
func (u *User) SetName(n string) // Setters keep "Set"
```

### Zero Values Are Useful

Design types so the zero value is immediately usable:

```go
// sync.Mutex is usable without initialization
var mu sync.Mutex
mu.Lock()

// bytes.Buffer works at zero value
var buf bytes.Buffer
buf.WriteString("hello")

// Your own types should follow this pattern
type Counter struct {
    n int64 // zero value = 0, ready to use
}
```

### Accept Interfaces, Return Structs

```go
// Good: accept the narrowest interface needed
func Process(r io.Reader) error { ... }

// Good: return concrete types so callers get full API
func NewServer(addr string) *Server { ... }
```

### Use `make` vs Composite Literals

```go
// Slices: use make when you know capacity, literal when you have values
s := make([]int, 0, 100)  // known capacity
s := []int{1, 2, 3}       // known values

// Maps: use make for empty maps
m := make(map[string]int, 100) // known size hint
m := map[string]int{"a": 1}   // known values
```

### Prefer `var` for Zero Values, `:=` for Non-Zero

```go
var s string          // zero value intent is clear
name := "default"     // non-zero, use short declaration
```

### Enums with `iota`

```go
type Status int

const (
    StatusPending Status = iota
    StatusActive
    StatusClosed
)

func (s Status) String() string {
    return [...]string{"pending", "active", "closed"}[s]
}
```

---

## Error Handling

### Wrap Errors with Context

```go
// Use fmt.Errorf with %w to wrap errors (preserves the chain)
if err := db.Query(ctx, q); err != nil {
    return fmt.Errorf("fetch user %d: %w", id, err)
}
```

### Sentinel Errors

```go
// Define at package level for errors callers need to check
var (
    ErrNotFound   = errors.New("not found")
    ErrConflict   = errors.New("conflict")
)

// Callers use errors.Is
if errors.Is(err, ErrNotFound) {
    http.Error(w, "not found", 404)
}
```

### Custom Error Types

```go
// Use when callers need structured error data
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation: %s — %s", e.Field, e.Message)
}

// Callers use errors.As
var ve *ValidationError
if errors.As(err, &ve) {
    log.Printf("bad field: %s", ve.Field)
}
```

### Don't Panic

```go
// Panics are for truly unrecoverable programmer errors.
// In libraries, ALWAYS return errors. Never panic.

// Acceptable: program setup that must succeed
func mustEnv(key string) string {
    v := os.Getenv(key)
    if v == "" {
        panic(fmt.Sprintf("required env var %s not set", key))
    }
    return v
}
```

### Handle Errors Once

```go
// Bad: logging AND returning
if err != nil {
    log.Println(err) // logged here
    return err        // and again by caller
}

// Good: return with context, let the caller decide
if err != nil {
    return fmt.Errorf("opening config: %w", err)
}
```

See [`examples/errors/`](../examples/errors/) for runnable examples.

---

## Package Design

### Package Naming

```
// Good: short, lowercase, singular nouns
package http
package user
package auth

// Bad
package httpUtils     // no mixedCaps
package common        // too vague
package models        // avoid plural
```

### Organize by Responsibility, Not Layer

```
// Prefer domain-oriented layout
internal/
  user/         # User domain: service, repo, handlers
  order/        # Order domain
  auth/         # Auth domain

// Avoid layer-oriented layout
internal/
  models/       # All models lumped together
  handlers/     # All handlers lumped together
  services/     # All services lumped together
```

### Minimize Exported API

```go
// Export only what consumers need. Start unexported, promote later.
// Unexported types can still satisfy exported interfaces.

type server struct{ ... }  // unexported

func New() *server { ... } // exported constructor, unexported return is fine
                            // (callers use the interface or returned pointer)
```

### Internal Packages

Use `internal/` to prevent external imports of implementation details:

```
mymodule/
  internal/db/     # Only mymodule can import this
  pkg/client/      # Anyone can import this
```

### Dependency Injection via Interfaces

```go
// Define interfaces where they're used, not where they're implemented
package order

type UserStore interface {
    Get(ctx context.Context, id string) (*User, error)
}

type Service struct {
    users UserStore
}

func NewService(users UserStore) *Service {
    return &Service{users: users}
}
```

---

## Formatting & Linting

### gofmt / goimports

All Go code must be formatted with `gofmt`. No exceptions, no debates.

```bash
# Format all files
gofmt -w .

# goimports also manages import grouping
go install golang.org/x/tools/cmd/goimports@latest
goimports -w .
```

Import grouping convention:

```go
import (
    // stdlib
    "context"
    "fmt"

    // third-party
    "github.com/gin-gonic/gin"

    // internal
    "github.com/example/go-template/internal/api"
)
```

### go vet

Built-in static analysis. Run it always:

```bash
go vet ./...
```

Catches: printf format mismatches, unreachable code, suspicious constructs, copy-lock violations.

### golangci-lint

The standard meta-linter. Runs 50+ linters in one pass:

```bash
# Install
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run
golangci-lint run ./...
```

Recommended `.golangci.yml`:

```yaml
linters:
  enable:
    - errcheck       # unchecked errors
    - govet          # go vet
    - staticcheck    # advanced static analysis
    - unused         # unused code
    - gosimple       # simplifications
    - ineffassign    # ineffectual assignments
    - gocritic       # opinionated style checks
    - revive         # flexible linter (replaces golint)
    - misspell       # spelling in comments/strings
    - prealloc       # slice preallocation hints
    - bodyclose      # unclosed HTTP response bodies
    - noctx          # HTTP requests without context

linters-settings:
  gocritic:
    enabled-tags:
      - diagnostic
      - style
      - performance

issues:
  exclude-use-default: false
  max-issues-per-linter: 0
  max-same-issues: 0
```

### staticcheck

The most precise Go static analyzer. Included in golangci-lint but also runs standalone:

```bash
go install honnef.co/go/tools/cmd/staticcheck@latest
staticcheck ./...
```

### CI Integration

Add to your CI pipeline (the Makefile already has `lint` and `fmt` targets):

```bash
make fmt       # format code
make lint      # run golangci-lint
go vet ./...   # built-in checks
```

---

## Quick Reference

| Principle | Do | Don't |
|---|---|---|
| Naming | `HTTPClient`, `userID` | `HttpClient`, `userId` |
| Errors | `return fmt.Errorf("x: %w", err)` | `log.Println(err); return err` |
| Packages | `package user` | `package userUtils` |
| Interfaces | Define where consumed | Define where implemented |
| Zero values | Design types to be useful at zero | Require constructor for basic use |
| Formatting | `gofmt -w .` always | Manual formatting |
| Panics | Only in `main` or `must*` helpers | In library code |

## See Also

- [Design Patterns](design-patterns.md) — Options, observer, strategy, factory patterns
- [Testing](testing.md) — Test strategies, table-driven tests, benchmarks
- [Documentation Guide](documentation-guide.md) — godoc conventions and examples
- [ARCHITECTURE.md](ARCHITECTURE.md) — Project layout and design decisions
