package cloudnative

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoadConfig_Defaults(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("ENVIRONMENT", "")
	c := LoadConfig()
	if c.Port != "8080" || c.LogLevel != "info" || c.Environment != "development" {
		t.Fatalf("unexpected defaults: %+v", c)
	}
}

func TestLoadConfig_Custom(t *testing.T) {
	t.Setenv("PORT", "3000")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("ENVIRONMENT", "production")
	c := LoadConfig()
	if c.Port != "3000" || c.LogLevel != "debug" || c.Environment != "production" {
		t.Fatalf("unexpected config: %+v", c)
	}
}

func TestNewLogger_Levels(t *testing.T) {
	for _, level := range []string{"debug", "info", "warn", "error", "unknown"} {
		l := NewLogger(level)
		if l == nil {
			t.Fatalf("NewLogger(%q) returned nil", level)
		}
	}
}

func TestHealthChecker_Liveness(t *testing.T) {
	hc := NewHealthChecker()
	rec := httptest.NewRecorder()
	hc.LivenessHandler().ServeHTTP(rec, httptest.NewRequest("GET", "/livez", nil))
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body["status"] != "alive" {
		t.Fatalf("unexpected body: %v", body)
	}
}

func TestHealthChecker_Readiness_AllHealthy(t *testing.T) {
	hc := NewHealthChecker()
	hc.AddCheck("db", func() error { return nil })
	rec := httptest.NewRecorder()
	hc.ReadinessHandler().ServeHTTP(rec, httptest.NewRequest("GET", "/readyz", nil))
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body["ready"] != true {
		t.Fatalf("expected ready=true, got %v", body)
	}
}

func TestHealthChecker_Readiness_Unhealthy(t *testing.T) {
	hc := NewHealthChecker()
	hc.AddCheck("cache", func() error { return errors.New("down") })
	rec := httptest.NewRecorder()
	hc.ReadinessHandler().ServeHTTP(rec, httptest.NewRequest("GET", "/readyz", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
	var body map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body["ready"] != false {
		t.Fatalf("expected ready=false, got %v", body)
	}
}

func TestRequestLogger(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})
	handler := RequestLogger(logger, inner)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest("GET", "/test", nil))
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
}
