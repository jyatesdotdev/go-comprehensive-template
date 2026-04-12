// Package testutil provides test helpers for assertions, HTTP testing, and mocking.
package testutil

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Equal fails the test if got != want.
func Equal[T comparable](t *testing.T, got, want T) {
	t.Helper()
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

// NoError fails the test if err is not nil.
func NoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// HasError fails the test if err is nil.
func HasError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// HTTPResult holds a decoded HTTP response for test assertions.
type HTTPResult struct {
	// Code is the HTTP status code.
	Code int
	// Body is the response body as a string.
	Body string
}

// DoRequest sends a request to the handler and returns the result.
func DoRequest(t *testing.T, handler http.Handler, method, path, body string) HTTPResult {
	t.Helper()
	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, bodyReader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return HTTPResult{Code: rec.Code, Body: rec.Body.String()}
}

// DecodeJSON unmarshals the response body into v.
func DecodeJSON(t *testing.T, body string, v any) {
	t.Helper()
	if err := json.Unmarshal([]byte(body), v); err != nil {
		t.Fatalf("decode JSON: %v\nbody: %s", err, body)
	}
}

// MockNotifier records calls to Notify for verification in tests.
type MockNotifier struct {
	// Calls records each message passed to Notify, in order.
	Calls []string
	// Err is the error returned by Notify on every call.
	Err error
}

// Notify appends msg to Calls and returns m.Err.
func (m *MockNotifier) Notify(msg string) error {
	m.Calls = append(m.Calls, msg)
	return m.Err
}
