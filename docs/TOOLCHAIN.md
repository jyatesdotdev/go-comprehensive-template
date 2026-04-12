# Toolchain

Tools and setup required to develop with this template.

## Required Tools

| Tool | Version | Purpose |
|------|---------|---------|
| [Go](https://go.dev/dl/) | 1.26.1+ (see `go.mod`) | Compiler, test runner, profiler |
| [GNU Make](https://www.gnu.org/software/make/) | 3.81+ | Build orchestration (`Makefile`) |
| [Git](https://git-scm.com/) | 2.x+ | Version control, build metadata |

## Linting & Formatting Tools

| Tool | Purpose | Makefile target |
|------|---------|-----------------|
| [golangci-lint](https://golangci-lint.run/) | Meta-linter (runs errcheck, govet, gosec, etc.) | `make lint` |
| [staticcheck](https://staticcheck.dev/) | Additional static analysis | `make lint` |
| [goimports](https://pkg.go.dev/golang.org/x/tools/cmd/goimports) | Import sorting + formatting | `make fmt` |

The linter configuration lives in `.golangci.yml`. See [ARCHITECTURE.md](ARCHITECTURE.md) for details on the linting strategy.

## Security Scanning Tools

| Tool | Purpose | Makefile target |
|------|---------|-----------------|
| [govulncheck](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck) | Go vulnerability database | `make security-govulncheck` |
| [gosec](https://github.com/securego/gosec) | Static security analysis | `make security-gosec` |
| [nancy](https://github.com/sonatype-nexus-community/nancy) | Dependency CVE scanning | `make security-nancy` |
| [trivy](https://github.com/aquasecurity/trivy) | Filesystem vulnerability scanning | `make security-trivy` |

Run all at once with `make security`.

## Optional Tools

| Tool | Purpose |
|------|---------|
| [Docker](https://www.docker.com/) | Container builds (`make docker`) |

## Installation

### macOS (Homebrew)

```bash
# Go
brew install go

# Build tools
brew install make git

# Linting & formatting
brew install golangci-lint
go install honnef.co/go/tools/cmd/staticcheck@latest
go install golang.org/x/tools/cmd/goimports@latest

# Security scanners
go install golang.org/x/vuln/cmd/govulncheck@latest
go install github.com/securego/gosec/v2/cmd/gosec@latest
brew install sonatype-nexus-community/nancy-tap/nancy
brew install trivy

# Optional
brew install --cask docker
```

### Linux (Debian/Ubuntu)

```bash
# Go — download from https://go.dev/dl/
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.26.1.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin

# Build tools
sudo apt-get install -y make git

# Linting & formatting
# golangci-lint — see https://golangci-lint.run/welcome/install/
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin
go install honnef.co/go/tools/cmd/staticcheck@latest
go install golang.org/x/tools/cmd/goimports@latest

# Security scanners
go install golang.org/x/vuln/cmd/govulncheck@latest
go install github.com/securego/gosec/v2/cmd/gosec@latest
# nancy — download binary from https://github.com/sonatype-nexus-community/nancy/releases
# trivy — see https://aquasecurity.github.io/trivy/latest/getting-started/installation/
sudo apt-get install -y trivy
```

### Windows

```powershell
# Go — download MSI from https://go.dev/dl/
# Git — download from https://git-scm.com/download/win
# Make — install via chocolatey or use Git Bash
choco install make

# All go install commands work the same:
go install honnef.co/go/tools/cmd/staticcheck@latest
go install golang.org/x/tools/cmd/goimports@latest
go install golang.org/x/vuln/cmd/govulncheck@latest
go install github.com/securego/gosec/v2/cmd/gosec@latest

# golangci-lint — download binary from https://golangci-lint.run/welcome/install/#windows
# trivy — download from https://github.com/aquasecurity/trivy/releases
```

### Verify Installation

```bash
go version                # Go 1.26.1+
golangci-lint --version   # any recent version
staticcheck -version
govulncheck -version
gosec --version
trivy --version
make --version
```

## Editor Setup

### VS Code

1. Install the [Go extension](https://marketplace.visualstudio.com/items?itemName=golang.Go) (includes gopls).
2. Add to `.vscode/settings.json`:

```json
{
  "go.lintTool": "golangci-lint",
  "go.lintFlags": ["--fast"],
  "go.formatTool": "goimports",
  "go.testFlags": ["-race", "-count=1"],
  "go.buildTags": "integration",
  "gopls": {
    "ui.semanticTokens": true,
    "ui.diagnostic.analyses": {
      "unusedparams": true,
      "shadow": true
    }
  }
}
```

This gives you:
- Lint-on-save via golangci-lint (uses `.golangci.yml`)
- Auto-format with goimports
- Race detection in tests
- Integration test tag support in the editor

### GoLand / IntelliJ

1. Go to **Settings → Go → Go Modules** — ensure module integration is enabled.
2. Go to **Settings → Tools → File Watchers** — add `golangci-lint` and `goimports` watchers.
3. Go to **Settings → Go → Build Tags** — add `integration` to see integration test files.
4. The built-in Go plugin handles gopls, formatting, and test running automatically.

## See Also

- [ARCHITECTURE.md](ARCHITECTURE.md) — project structure and design decisions
- [EXTENDING.md](EXTENDING.md) — how to add commands, packages, and linter rules
- [TUTORIAL.md](TUTORIAL.md) — new developer walkthrough
- [security-scanning.md](security-scanning.md) — detailed security scanning guide
