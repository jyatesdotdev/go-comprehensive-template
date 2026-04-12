# syntax=docker/dockerfile:1

# ── Build stage ──────────────────────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache ca-certificates git

WORKDIR /src

# Layer cache: dependencies first
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Build args for version injection (match Makefile ldflags)
ARG VERSION=dev
ARG COMMIT=none
ARG DATE=unknown

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}" \
    -o /bin/server ./cmd/server

# ── Final stage ──────────────────────────────────────────────────────────────
FROM scratch

LABEL org.opencontainers.image.source="https://github.com/example/go-template" \
      org.opencontainers.image.description="Go template server" \
      org.opencontainers.image.licenses="MIT"

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /bin/server /server

EXPOSE 8080

# Non-root: nobody user
USER 65534:65534

# Health checks: use orchestrator probes (k8s, ECS, etc.) against GET /health
# scratch has no shell or utilities for Docker HEALTHCHECK.

ENTRYPOINT ["/server"]
