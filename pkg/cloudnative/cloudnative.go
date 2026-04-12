// Package cloudnative provides cloud-native primitives: config loading,
// health checks, structured logging, and observability middleware.
package cloudnative

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"
)

// --- Config ---

// Config holds application configuration loaded from environment variables.
type Config struct {
	// Port is the TCP port the server listens on (default "8080").
	Port string
	// LogLevel controls structured log verbosity: "debug", "info", "warn", or "error".
	LogLevel string
	// Environment is the deployment environment name (e.g. "development", "production").
	Environment string
}

// LoadConfig reads configuration from environment with sensible defaults.
func LoadConfig() Config {
	return Config{
		Port:        envOr("PORT", "8080"),
		LogLevel:    envOr("LOG_LEVEL", "info"),
		Environment: envOr("ENVIRONMENT", "development"),
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// --- Structured Logging ---

// NewLogger returns a structured JSON logger at the given level.
func NewLogger(level string) *slog.Logger {
	var lvl slog.Level
	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl}))
}

// --- Health Checks ---

// CheckFunc is a dependency health check that returns nil if healthy.
type CheckFunc func() error

// HealthChecker manages liveness and readiness probes.
type HealthChecker struct {
	mu     sync.RWMutex
	checks map[string]CheckFunc
}

// NewHealthChecker creates a HealthChecker.
func NewHealthChecker() *HealthChecker {
	return &HealthChecker{checks: make(map[string]CheckFunc)}
}

// AddCheck registers a named readiness check.
func (h *HealthChecker) AddCheck(name string, fn CheckFunc) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.checks[name] = fn
}

// LivenessHandler returns 200 if the process is alive.
func (h *HealthChecker) LivenessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "alive"}) // #nosec G104 -- best-effort HTTP response write
	}
}

// ReadinessHandler returns 200 only if all registered checks pass.
func (h *HealthChecker) ReadinessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.mu.RLock()
		defer h.mu.RUnlock()

		results := make(map[string]string, len(h.checks))
		healthy := true
		for name, fn := range h.checks {
			if err := fn(); err != nil {
				results[name] = err.Error()
				healthy = false
			} else {
				results[name] = "ok"
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if !healthy {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ready": healthy, "checks": results}) // #nosec G104 -- best-effort HTTP response write
	}
}

// --- Observability Middleware ---

// RequestLogger wraps an http.Handler with structured request logging.
func RequestLogger(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		logger.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.status,
			"duration_ms", fmt.Sprintf("%.2f", float64(time.Since(start).Microseconds())/1000),
		)
	})
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}
