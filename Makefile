BINARY   := server
PKG      := github.com/example/go-template
CMD      := ./cmd/server

# Build metadata
VERSION  ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT   ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE     ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS  := -ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

.PHONY: build run test lint clean docker cross \
       security security-govulncheck security-gosec security-nancy security-trivy

build:
	go build $(LDFLAGS) -o bin/$(BINARY) $(CMD)

run: build
	./bin/$(BINARY)

test:
	go test -race -count=1 ./...

test-integration:
	go test -race -tags=integration -count=1 ./...

bench:
	go test -bench=. -benchmem ./...

lint:
	go vet ./...
	@command -v staticcheck >/dev/null 2>&1 && staticcheck ./... || echo "staticcheck not installed"
	@command -v golangci-lint >/dev/null 2>&1 && golangci-lint run ./... || echo "golangci-lint not installed"

fmt:
	gofmt -s -w .
	goimports -w .

clean:
	rm -rf bin/ cover.out

cover:
	go test -coverprofile=cover.out ./...
	go tool cover -html=cover.out

# Cross-compilation targets
cross:
	GOOS=linux   GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY)-linux-amd64   $(CMD)
	GOOS=linux   GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY)-linux-arm64   $(CMD)
	GOOS=darwin  GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY)-darwin-amd64  $(CMD)
	GOOS=darwin  GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY)-darwin-arm64  $(CMD)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY)-windows-amd64.exe $(CMD)

docker:
	docker build -t $(BINARY):$(VERSION) .

# Profile targets
pprof-cpu:
	go test -cpuprofile=cpu.prof -bench=. ./...
	go tool pprof cpu.prof

pprof-mem:
	go test -memprofile=mem.prof -bench=. ./...
	go tool pprof mem.prof

# --- Security scanning targets ---

security: security-govulncheck security-gosec security-nancy security-trivy
	@echo "=== All security scans complete ==="

security-govulncheck:
	@echo "=== govulncheck: checking Go vulnerability database ==="
	govulncheck ./...

security-gosec:
	@echo "=== gosec: static security analysis ==="
	gosec ./...

security-nancy:
	@echo "=== nancy: dependency CVE scanning ==="
	go list -json -deps ./... | nancy sleuth

security-trivy:
	@echo "=== trivy: filesystem vulnerability scanning ==="
	trivy fs --severity HIGH,CRITICAL --exit-code 1 .
