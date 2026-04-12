# Third-Party Integration Guide

## Dependency Management with Go Modules

### Module Basics

```bash
# Initialize a new module
go mod init github.com/yourorg/yourproject

# Add a dependency (automatically updates go.mod and go.sum)
go get github.com/gin-gonic/gin@latest

# Add a specific version
go get github.com/lib/pq@v1.10.9

# Update all dependencies
go get -u ./...

# Update a specific dependency
go get -u github.com/gin-gonic/gin

# Remove unused dependencies
go mod tidy

# Verify dependency integrity
go mod verify
```

### go.mod File

```go
module github.com/yourorg/yourproject

go 1.22

require (
    github.com/gin-gonic/gin v1.9.1
    github.com/jmoiron/sqlx v1.3.5
    go.uber.org/zap v1.27.0
)

require (
    // indirect dependencies managed automatically
    github.com/mattn/go-isatty v0.0.20 // indirect
)
```

### go.sum and Security

`go.sum` contains cryptographic checksums for every dependency version. Always commit it to version control. Go verifies these checksums on every build to detect tampering.

```bash
# Verify all dependencies match go.sum
go mod verify

# Download dependencies to local cache
go mod download
```

## Vendoring

Vendoring copies all dependencies into a `vendor/` directory within your project, enabling fully offline and reproducible builds.

```bash
# Create/update vendor directory
go mod vendor

# Build using vendored dependencies
go build -mod=vendor ./...

# Run tests with vendored dependencies
go test -mod=vendor ./...
```

When to vendor:
- CI/CD environments without internet access
- Regulatory requirements for auditable dependencies
- Protection against upstream repository deletion
- Air-gapped deployments

Add to `.gitignore` or commit `vendor/` based on your team's policy. Committing ensures reproducibility; ignoring keeps the repo smaller.

## Module Proxies

Go modules use a proxy to fetch dependencies. The default is `proxy.golang.org`.

```bash
# Default proxy chain (check proxy, then direct)
export GOPROXY=https://proxy.golang.org,direct

# Use a private proxy for internal modules
export GOPROXY=https://goproxy.mycompany.com,https://proxy.golang.org,direct

# Skip proxy for private modules
export GOPRIVATE=github.com/mycompany/*,gitlab.internal.com/*

# Disable proxy entirely (fetch directly from VCS)
export GOPROXY=direct

# Disable checksum verification for private modules
export GONOSUMCHECK=github.com/mycompany/*
```

### Private Module Configuration

For private repositories, configure Git authentication:

```bash
# Use SSH for private GitHub repos
git config --global url."git@github.com:mycompany/".insteadOf "https://github.com/mycompany/"

# Or use a personal access token
git config --global url."https://${GITHUB_TOKEN}@github.com/".insteadOf "https://github.com/mycompany/"
```

### Running a Private Proxy

Popular options:
- **Athens** (`github.com/gomods/athens`) — open-source Go module proxy with storage backends (disk, S3, GCS)
- **Artifactory** / **Nexus** — enterprise artifact managers with Go module support
- **GOPROXY** (`github.com/goproxy/goproxy`) — minimalist proxy library

## Popular Libraries by Category

### Web Frameworks
| Library | Use Case | Import Path |
|---------|----------|-------------|
| Gin | High-performance HTTP framework | `github.com/gin-gonic/gin` |
| Echo | Minimalist, extensible framework | `github.com/labstack/echo/v4` |
| Chi | Lightweight, idiomatic router | `github.com/go-chi/chi/v5` |
| Fiber | Express-inspired, fasthttp-based | `github.com/gofiber/fiber/v2` |

### Database
| Library | Use Case | Import Path |
|---------|----------|-------------|
| sqlx | Extensions to database/sql | `github.com/jmoiron/sqlx` |
| GORM | Full-featured ORM | `gorm.io/gorm` |
| pgx | PostgreSQL driver (pure Go) | `github.com/jackc/pgx/v5` |
| go-sqlite3 | SQLite driver (CGO) | `github.com/mattn/go-sqlite3` |
| golang-migrate | Database migrations | `github.com/golang-migrate/migrate/v4` |

### Logging
| Library | Use Case | Import Path |
|---------|----------|-------------|
| slog | Structured logging (stdlib, Go 1.21+) | `log/slog` |
| zap | High-performance structured logging | `go.uber.org/zap` |
| zerolog | Zero-allocation JSON logging | `github.com/rs/zerolog` |

### Configuration
| Library | Use Case | Import Path |
|---------|----------|-------------|
| Viper | Config files, env vars, flags | `github.com/spf13/viper` |
| envconfig | Struct-based env var parsing | `github.com/kelseyhightower/envconfig` |
| koanf | Lightweight, extensible config | `github.com/knadh/koanf/v2` |

### Testing
| Library | Use Case | Import Path |
|---------|----------|-------------|
| testify | Assertions and mocking | `github.com/stretchr/testify` |
| gomock | Interface mocking (code gen) | `go.uber.org/mock/gomock` |
| httptest | HTTP testing (stdlib) | `net/http/httptest` |
| testcontainers | Docker-based integration tests | `github.com/testcontainers/testcontainers-go` |

### Observability
| Library | Use Case | Import Path |
|---------|----------|-------------|
| OpenTelemetry | Traces, metrics, logs | `go.opentelemetry.io/otel` |
| Prometheus client | Metrics exposition | `github.com/prometheus/client_golang` |
| pprof | Profiling (stdlib) | `net/http/pprof` |

### CLI
| Library | Use Case | Import Path |
|---------|----------|-------------|
| Cobra | CLI framework | `github.com/spf13/cobra` |
| urfave/cli | Simple CLI apps | `github.com/urfave/cli/v2` |

### Concurrency & Async
| Library | Use Case | Import Path |
|---------|----------|-------------|
| errgroup | Goroutine error propagation | `golang.org/x/sync/errgroup` |
| semaphore | Weighted semaphore | `golang.org/x/sync/semaphore` |
| singleflight | Duplicate call suppression | `golang.org/x/sync/singleflight` |

## Integration Patterns

### Wrapping Third-Party Libraries

Always wrap third-party dependencies behind your own interfaces. This enables testing, swapping implementations, and controlling the API surface.

```go
// Define your interface
type Logger interface {
    Info(msg string, args ...any)
    Error(msg string, args ...any)
}

// Wrap the third-party implementation
type ZapLogger struct {
    sugar *zap.SugaredLogger
}

func (z *ZapLogger) Info(msg string, args ...any)  { z.sugar.Infow(msg, args...) }
func (z *ZapLogger) Error(msg string, args ...any) { z.sugar.Errorw(msg, args...) }
```

### Adapter Pattern for Swappable Backends

```go
// Storage interface — your code depends on this, not on any specific library
type Storage interface {
    Get(ctx context.Context, key string) ([]byte, error)
    Set(ctx context.Context, key string, val []byte, ttl time.Duration) error
}

// Redis adapter
type RedisStorage struct { client *redis.Client }

// In-memory adapter (for testing or development)
type MemStorage struct {
    mu    sync.RWMutex
    items map[string][]byte
}
```

### Dependency Injection

```go
type Server struct {
    logger  Logger
    storage Storage
    db      *sql.DB
}

func NewServer(logger Logger, storage Storage, db *sql.DB) *Server {
    return &Server{logger: logger, storage: storage, db: db}
}
```

This pattern makes testing straightforward — inject mocks for any dependency.

## Evaluating Third-Party Libraries

Before adding a dependency, consider:

1. **Maintenance**: Is it actively maintained? Check commit frequency and issue response times
2. **License**: Is the license compatible with your project? (MIT, Apache 2.0, BSD are generally safe)
3. **Dependencies**: How many transitive dependencies does it pull in? (`go mod graph`)
4. **Alternatives**: Can the standard library solve this? Go's stdlib is extensive
5. **Size**: What's the binary size impact?
6. **Security**: Check `govulncheck` for known vulnerabilities

```bash
# Inspect dependency graph
go mod graph

# Check for vulnerabilities
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...

# See why a dependency is included
go mod why github.com/some/package

# List all dependencies with their licenses
go list -m -json all
```

## Minimal Dependency Philosophy

Go's standard library is unusually comprehensive. Prefer it when possible:

| Need | Standard Library | Third-Party (if needed) |
|------|-----------------|------------------------|
| HTTP server | `net/http` | Gin, Echo, Chi |
| JSON | `encoding/json` | `github.com/json-iterator/go` |
| Logging | `log/slog` | zap, zerolog |
| Testing | `testing` | testify |
| Context | `context` | — |
| Crypto | `crypto/*` | — |
| Templates | `text/template`, `html/template` | — |

Add third-party libraries when they provide significant value: better performance, complex functionality you'd otherwise reimplement, or battle-tested solutions to hard problems.

## See Also

- [EXTENDING.md](EXTENDING.md) — Adding and wrapping third-party dependencies
- [Best Practices](best-practices.md) — Dependency management and interface design
