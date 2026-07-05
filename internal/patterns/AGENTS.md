# AGENTS.md — internal/patterns

A catalog of idiomatic Go design patterns, meant to be read as reference. Stdlib only
(`errors`, `fmt`, `time`). One file, `patterns.go`, organized into labeled sections.

Read [internal/AGENTS.md](../AGENTS.md) and the [root](../../AGENTS.md) first.

## What's demonstrated (keep each idiom canonical)

- **Functional options** — `Server`, `Option func(*Server)`, `With*` constructors, `NewServer(addr, opts...)`.
  Defaults live in `NewServer`: port `8080`, read `5s`, write `10s`, max conns `100`. New options
  follow the `With<Field>(v) Option { return func(s *Server){ s.Field = v } }` shape.
- **Interfaces & composition** — minimal `Reader`/`Writer` and their `ReadWriter` composition
  (mirrors `io`); keep interfaces small.
- **Strategy / composite** — `Notifier` interface, `EmailNotifier`/`SlackNotifier` implementations,
  and `MultiNotifier []Notifier` that fans out and **joins errors with `errors.Join`**.
- **Embedding (composition over inheritance)** — `Base` embedded in `User`; `User` gains `Base`'s
  fields and `Age()` method. Use embedding, not "inheritance" framing.
- **Error types** — sentinels `ErrNotFound` / `ErrUnauthorized` (compare with `errors.Is`);
  `ValidationError` struct (extract with `errors.As`); `Wrap(err, msg)` preserving the chain via `%w`.

## Business rules

- `Validate(name)` returns a `*ValidationError` when `name` is empty (`required`) or longer than
  **50 chars** (`too long`); nil otherwise. Return typed errors, not bare strings.
- `Wrap(nil, …)` returns `nil` — wrapping a nil error is a no-op (preserve this).

## Go best practices reinforced here

- **Accept interfaces, return concrete types.** `NewServer`/`NewUser` return concrete values;
  functions consume the small interfaces.
- **Sentinel vs typed errors:** use a sentinel when callers only need identity (`errors.Is`); use a
  typed error when callers need structured fields (`errors.As`). This package shows both — mirror
  the matching one when adding errors elsewhere.
- **Options over telescoping constructors / config structs** when a type has many optional knobs.
- Value vs pointer receivers: note `ValidationError.Error()` is on `*ValidationError`, so construct
  and compare it as a pointer.

## Testing

`patterns_test.go` covers behavior; `example_test.go` holds godoc `Example*` functions with
`// Output:` blocks — these are compiled and run by `go test`, so keep their output current.
When adding a pattern, add a matching runnable slice to [`examples/patterns`](../../examples/AGENTS.md).
