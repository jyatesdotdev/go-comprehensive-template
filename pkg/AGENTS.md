# AGENTS.md — pkg/

Public, general-purpose packages. Unlike `internal/`, code here **can be imported by other
modules**, so its exported API is a contract.

Read the [root AGENTS.md](../AGENTS.md) first for module-wide rules.

## Packages

| Package | Purpose | Deps |
|---------|---------|------|
| [`cloudnative/`](cloudnative/AGENTS.md) | Env config, health checks, slog logging, request-logging middleware | stdlib only |
| [`systems/`](systems/AGENTS.md) | Atomic file I/O, line streaming, TCP/UDP networking, OS info | stdlib only |

Both are self-contained, depend only on the standard library, and (like everything else) do not
import any `internal/*` or each other.

## API-stability rules (this is what makes `pkg/` different)

- **Treat exported identifiers as a public contract.** Renaming or changing the signature of an
  exported func/type/method, or removing one, is a **breaking change**. Avoid it; when genuinely
  necessary, do it deliberately and call it out.
- **Additive change is preferred.** New exported functions/fields are fine; repurposing existing
  ones is not.
- **Document every exported identifier** with a full doc comment (name-first sentence). These
  packages are meant to read well via `go doc` / pkg.go.dev.
- Keep the dependency footprint at zero third-party packages — a big part of these packages' value
  is being liftable into a standalone module without untangling imports.

## Go best practices

- Accept interfaces / `context.Context`, return concrete types.
- Make zero values useful where possible, and provide `New*` constructors when initialization is
  required (e.g. maps must be made).
- Propagate errors with `%w`; never `log.Fatal` or `os.Exit` inside library code — return the error
  and let the caller decide.
