# New Developer Tutorial

A hands-on walkthrough from cloning the repo to building a feature. Each section builds on the previous one.

> **Prerequisites:** Go 1.22+, Make, Git. See [TOOLCHAIN.md](TOOLCHAIN.md) for install instructions.

## 1. Clone and Set Up

```bash
git clone https://github.com/example/go-template.git
cd go-template
go mod download
```

Verify your Go version matches the requirement in `go.mod`:

```bash
go version
```

## 2. Build

```bash
make build
```

This compiles `cmd/server/main.go` into `bin/server` with version metadata embedded via ldflags. You should see no output on success.

Inspect the binary:

```bash
ls -lh bin/server
```

## 3. Run the Server

```bash
make run
# Or directly:
./bin/server
```

The server starts on port 8080 (override with `PORT=3000 ./bin/server`). Test it:

```bash
curl http://localhost:8080/health
# {"status":"ok"}
```

Press `Ctrl+C` to trigger graceful shutdown.

## 4. Run Tests

```bash
# Unit tests with race detector
make test

# Integration tests (tagged files)
make test-integration

# Benchmarks
make bench

# Coverage report (opens HTML)
make cover
```

All tests use `go test -race -count=1` to catch data races and prevent caching.

## 5. Run Examples

Each `examples/` subdirectory is a standalone program demonstrating a package:

```bash
# RESTful API — starts a server, runs client requests, then exits
go run ./examples/api

# Concurrency patterns — goroutines, channels, worker pools
go run ./examples/concurrency

# ETL pipeline — fan-out/fan-in data processing
go run ./examples/pipeline

# Design patterns — observer, strategy, builder
go run ./examples/patterns

# Database — migrations, CRUD, repository pattern
go run ./examples/database

# Performance — benchmarking and optimization demos
go run ./examples/performance

# Systems programming — low-level OS interaction
go run ./examples/systems

# Simulation — numerical computing
go run ./examples/simulation

# Cloud-native — health checks, graceful shutdown
go run ./examples/cloudnative

# CLI — Cobra/Viper command-line app
go run ./examples/cli greet --name World
go run ./examples/cli list --output json
go run ./examples/cli config show
```

## 6. Run Linters

```bash
# Quick: go vet + staticcheck + golangci-lint
make lint

# Format code
make fmt
```

The project uses `.golangci.yml` to configure linters including `errcheck`, `govet`, `staticcheck`, `gosec`, `gocritic`, and `revive`. See [TOOLCHAIN.md](TOOLCHAIN.md) for installing these tools.

## 7. Run Security Scans

```bash
# All scanners at once
make security

# Or individually:
make security-govulncheck   # Go vulnerability database
make security-gosec          # Static security analysis
make security-nancy          # Dependency CVE scanning
make security-trivy          # Filesystem vulnerability scan
```

See [security-scanning.md](security-scanning.md) for details on each scanner.

## 8. Build a Docker Image

```bash
make docker
```

This runs a multi-stage build: compiles in `golang:1.22-alpine`, copies the binary into a `scratch` image. The final image contains only the binary and CA certificates.

```bash
docker run -p 8080:8080 server:dev
curl http://localhost:8080/health
```

## 9. Cross-Compile

```bash
make cross
ls bin/
# server-linux-amd64  server-linux-arm64  server-darwin-amd64  server-darwin-arm64  server-windows-amd64.exe
```

## 10. Add a Feature (Walkthrough)

Let's add a `/version` endpoint to the server as a practical exercise.

### Step 1: Edit `cmd/server/main.go`

Add a handler to the mux:

```go
mux.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, `{"version":"%s","commit":"%s"}`, version, commit)
})
```

### Step 2: Build and test

```bash
make build
./bin/server &
curl http://localhost:8080/version
kill %1
```

### Step 3: Run the full check suite

```bash
make test
make lint
make security-govulncheck
```

### Step 4: Commit

```bash
git add -A
git commit -m "feat: add /version endpoint"
```

For more complex additions (new commands, packages, dependencies), see [EXTENDING.md](EXTENDING.md).

## 11. Profile Performance

```bash
# CPU profile — runs benchmarks, opens pprof
make pprof-cpu

# Memory profile
make pprof-mem
```

See [performance.md](performance.md) for optimization techniques.

## Next Steps

- [ARCHITECTURE.md](ARCHITECTURE.md) — understand the project layout and design decisions
- [TOOLCHAIN.md](TOOLCHAIN.md) — full tool install guide and editor setup
- [EXTENDING.md](EXTENDING.md) — adding commands, packages, and dependencies
- [best-practices.md](best-practices.md) — Go idioms and error handling
- [concurrency.md](concurrency.md) — goroutine patterns and pitfalls
