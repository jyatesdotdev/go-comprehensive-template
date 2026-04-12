# Design Patterns in Go

Go favors simplicity and composition. These patterns are idiomatic to Go and avoid the complexity of classical OOP patterns.

## Functional Options

The most popular Go pattern for configurable constructors. Instead of large config structs or telescoping parameters, use variadic option functions:

```go
type Option func(*Server)

func WithPort(p int) Option { return func(s *Server) { s.Port = p } }

func NewServer(addr string, opts ...Option) *Server {
    s := &Server{Addr: addr, Port: 8080} // sensible defaults
    for _, o := range opts { o(s) }
    return s
}

// Usage: clean, self-documenting, extensible
srv := NewServer("0.0.0.0", WithPort(9090), WithMaxConns(500))
```

Benefits: backwards-compatible API evolution, self-documenting call sites, sensible defaults.

## Interfaces

Go interfaces are satisfied implicitly — no `implements` keyword. Keep them small:

```go
type Notifier interface {
    Notify(msg string) error
}
```

Guidelines:
- **Accept interfaces, return structs** — callers define the contract they need
- **Small interfaces** — `io.Reader` (1 method) is the gold standard
- **Compose interfaces** — `ReadWriter` embeds `Reader` + `Writer`
- **Define interfaces at the consumer**, not the provider

## Composition over Inheritance

Go has no inheritance. Use struct embedding to share behavior:

```go
type Base struct {
    ID        string
    CreatedAt time.Time
}
func (b Base) Age() time.Duration { return time.Since(b.CreatedAt) }

type User struct {
    Base  // promoted fields and methods
    Name  string
}
// u.Age() works — delegated to embedded Base
```

Embedding is not inheritance — the outer type does not "override" methods, it shadows them.

## Strategy Pattern

Swap behavior at runtime via interfaces:

```go
type MultiNotifier []Notifier

func (mn MultiNotifier) Notify(msg string) error {
    var errs []error
    for _, n := range mn { errs = append(errs, n.Notify(msg)) }
    return errors.Join(errs...)
}
```

This is also the decorator/composite pattern — `MultiNotifier` itself satisfies `Notifier`.

## Custom Error Types

Three levels of error sophistication:

1. **Sentinel errors** — `var ErrNotFound = errors.New("not found")` — compare with `errors.Is`
2. **Wrapped errors** — `fmt.Errorf("fetch user: %w", err)` — preserves chain for `errors.Is`
3. **Typed errors** — struct implementing `error` — extract with `errors.As`

```go
type ValidationError struct {
    Field, Message string
}
func (e *ValidationError) Error() string { ... }

// Caller:
var ve *ValidationError
if errors.As(err, &ve) {
    log.Printf("field %s: %s", ve.Field, ve.Message)
}
```

## Anti-Patterns to Avoid

- **God interfaces** — interfaces with 10+ methods are a code smell
- **Premature abstraction** — don't create interfaces until you have 2+ implementations
- **Deep embedding** — more than 2 levels of embedding becomes confusing
- **Naked returns in complex functions** — use named returns only for documentation
- **init() abuse** — prefer explicit initialization over package-level `init()`

## Running the Example

```bash
go run examples/patterns/main.go
```

## See Also

- [Best Practices](best-practices.md) — Go idioms and package design
- [Concurrency](concurrency.md) — Patterns used alongside design patterns
- [EXTENDING.md](EXTENDING.md) — Adding new packages that use these patterns
