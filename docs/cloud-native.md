# Cloud-Native Development in Go

## Configuration Management

Load config from environment variables with sensible defaults. Avoid config files in containers — the environment is the config source (12-factor app).

```go
cfg := cloudnative.LoadConfig()
// PORT, LOG_LEVEL, ENVIRONMENT read from env with defaults
```

For complex apps, add fields to `Config` and corresponding `envOr` calls. Keep config flat and explicit — avoid deeply nested YAML/TOML in cloud deployments.

## Health Checks

Kubernetes and load balancers need two probes:

- **Liveness** (`/healthz`): Is the process alive? Returns 200 always. If this fails, the container is restarted.
- **Readiness** (`/readyz`): Can the process serve traffic? Checks dependencies (DB, cache, etc.). If this fails, traffic is removed but the container stays running.

```go
health := cloudnative.NewHealthChecker()
health.AddCheck("database", func() error {
    return db.Ping()
})
mux.HandleFunc("GET /healthz", health.LivenessHandler())
mux.HandleFunc("GET /readyz", health.ReadinessHandler())
```

## Graceful Shutdown

Handle `SIGINT`/`SIGTERM` to drain in-flight requests before exiting. Use `signal.NotifyContext` for clean integration with `context`:

```go
ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
defer stop()
// ... start server in goroutine ...
<-ctx.Done()
srv.Shutdown(context.WithTimeout(context.Background(), 30*time.Second))
```

Key points:
- Set a shutdown timeout (30s is typical for Kubernetes)
- Kubernetes sends SIGTERM, then waits `terminationGracePeriodSeconds` before SIGKILL
- Close database connections and flush logs in the shutdown path

## Structured Logging

Use `log/slog` (Go 1.21+) for structured JSON logging. JSON logs are parseable by CloudWatch, Datadog, ELK, etc.

```go
logger := cloudnative.NewLogger("info") // JSON output to stdout
logger.Info("request", "method", "GET", "path", "/api", "status", 200)
```

Output:
```json
{"time":"2024-01-01T00:00:00Z","level":"INFO","msg":"request","method":"GET","path":"/api","status":200}
```

Log to stdout/stderr — let the container runtime handle log collection.

## Observability Middleware

Wrap your handler with `RequestLogger` to log every request with method, path, status, and duration:

```go
handler := cloudnative.RequestLogger(logger, mux)
```

For production, add metrics (Prometheus) and tracing (OpenTelemetry) as separate middleware layers.

## Dockerfile Best Practices

```dockerfile
FROM golang:1.22-alpine AS builder
COPY go.mod go.sum ./
RUN go mod download          # Cache dependencies
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /bin/server ./cmd/server

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /bin/server /server
USER 65534:65534             # Non-root (nobody)
EXPOSE 8080
ENTRYPOINT ["/server"]
```

Key practices:
- **Multi-stage build**: Builder stage compiles, final stage is `scratch` (zero attack surface)
- **Static binary**: `CGO_ENABLED=0` produces a fully static binary
- **Strip symbols**: `-ldflags="-s -w"` reduces binary size ~30%
- **Non-root user**: Run as UID 65534 (nobody) for security
- **Cache deps**: Copy `go.mod`/`go.sum` first so `go mod download` is cached
- **OCI labels**: Add `org.opencontainers.image.source` for registry linking

## Kubernetes Deployment Pattern

```yaml
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: app
        image: app:latest
        ports:
        - containerPort: 8080
        env:
        - name: ENVIRONMENT
          value: production
        - name: LOG_LEVEL
          value: info
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /readyz
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        resources:
          requests:
            memory: "64Mi"
            cpu: "100m"
          limits:
            memory: "128Mi"
            cpu: "500m"
```

## Running the Example

```bash
# Default config
go run ./examples/cloudnative

# Custom config
PORT=9090 LOG_LEVEL=debug ENVIRONMENT=staging go run ./examples/cloudnative

# Test endpoints
curl localhost:8080/healthz   # Liveness
curl localhost:8080/readyz    # Readiness
curl localhost:8080/          # App endpoint
```

## See Also

- [TOOLCHAIN.md](TOOLCHAIN.md) — Docker and tool installation
- [Security Scanning](security-scanning.md) — Container and dependency scanning
- [Performance](performance.md) — Profiling and optimization
