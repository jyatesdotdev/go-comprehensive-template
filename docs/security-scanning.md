# Security Scanning

This project integrates five security scanning tools into the build process and CI pipeline.

## Tools

| Tool | Purpose | Scope |
|------|---------|-------|
| [govulncheck](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck) | Go vulnerability database | Known CVEs in Go stdlib and dependencies |
| [gosec](https://github.com/securego/gosec) | Static analysis | Hardcoded creds, SQL injection, weak crypto, etc. |
| [golangci-lint](https://golangci-lint.run/) | Linter aggregator | Runs gosec + bodyclose + sqlclosecheck via `.golangci.yml` |
| [nancy](https://github.com/sonatype-nexus-community/nancy) | Dependency CVE scan | Sonatype OSS Index lookup on `go.mod` deps |
| [trivy](https://github.com/aquasecurity/trivy) | Filesystem/container scan | HIGH and CRITICAL severity findings |

## Installation

```bash
# govulncheck
go install golang.org/x/vuln/cmd/govulncheck@latest

# gosec
go install github.com/securego/gosec/v2/cmd/gosec@latest

# golangci-lint (macOS)
brew install golangci-lint
# or: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# nancy
go install github.com/sonatype-nexus-community/nancy@latest

# trivy (macOS)
brew install trivy
```

## Usage

Run all scanners:

```bash
make security
```

Run individually:

```bash
make security-govulncheck   # Go vuln database check
make security-gosec          # Static security analysis
make security-nancy          # Dependency CVE scan
make security-trivy          # Filesystem scan (HIGH,CRITICAL)
```

golangci-lint (which includes gosec, bodyclose, sqlclosecheck) runs via:

```bash
make lint
```

## CI/CD

The GitHub Actions workflow at `.github/workflows/security.yml` runs all five scanners in parallel on:

- Push to `main`
- Pull requests targeting `main`
- Weekly schedule (Monday 6am UTC) to catch newly disclosed CVEs

A `security-gate` job aggregates results — if any scanner fails, the gate fails. Set `security-gate` as a required status check in your branch protection rules:

> Settings → Branches → Branch protection rules → Require status checks → `Security Gate`

gosec and trivy upload SARIF results to GitHub's Security tab (requires `security-events: write` permission).

## Suppression Patterns

When a finding is a false positive or accepted risk, suppress it per-tool:

### gosec (inline)

```go
// #nosec G104 -- error intentionally ignored in cleanup path
_ = file.Close()
```

Or via golangci-lint's nolint directive:

```go
//nolint:gosec // G104: accepted risk
_ = file.Close()
```

### gosec (rule exclusion in `.golangci.yml`)

```yaml
linters-settings:
  gosec:
    excludes:
      - G104  # unhandled errors (covered by errcheck)
```

### nancy (`.nancy-ignore`)

Create `.nancy-ignore` in the project root:

```
# CVE accepted — no fix available, mitigated by network policy
CVE-2023-XXXXX
```

### trivy (`.trivyignore`)

Create `.trivyignore` in the project root:

```
# Accepted risk — mitigated at infrastructure level
CVE-2023-XXXXX
```

### golangci-lint (inline)

```go
//nolint:bodyclose // response body closed in deferred cleanup
//nolint:sqlclosecheck // rows closed via helper function
```

## Configuration

### `.golangci.yml` Security Linters

The config enables three security-focused linters:

- `gosec` — findings treated as `severity: error` (not just warnings)
- `bodyclose` — ensures HTTP response bodies are closed
- `sqlclosecheck` — ensures SQL rows and statements are closed

Thresholds are set to `severity: medium` and `confidence: medium` to balance signal vs noise.

### Trivy Severity Filter

The Makefile and CI only fail on `HIGH` and `CRITICAL` findings. To include medium:

```bash
trivy fs --severity MEDIUM,HIGH,CRITICAL --exit-code 1 .
```

## Failure Behavior

All tools exit non-zero on findings, which fails `make security` and CI:

| Tool | Fails on |
|------|----------|
| govulncheck | Any known vulnerability in call graph |
| gosec | Any finding at configured severity/confidence |
| golangci-lint | Any linter violation |
| nancy | Any CVE in dependencies |
| trivy | HIGH or CRITICAL findings |

## See Also

- [TOOLCHAIN.md](TOOLCHAIN.md) — Installing security tools
- [Cloud-Native](cloud-native.md) — Container scanning in CI/CD
