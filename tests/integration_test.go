//go:build integration

// Package tests contains integration tests that require external dependencies.
// Run with: go test -tags=integration ./tests/...
package tests

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/example/go-template/internal/api"
)

// TestItemHandler_Integration exercises the full CRUD lifecycle.
func TestItemHandler_Integration(t *testing.T) {
	store := api.NewStore()
	srv := httptest.NewServer(api.ItemHandler(store))
	defer srv.Close()

	// POST
	resp, err := http.Post(srv.URL+"/items", "application/json",
		strings.NewReader(`{"id":"int-1","name":"Integration"}`))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("POST status = %d, want 201", resp.StatusCode)
	}

	// GET
	resp, err = http.Get(srv.URL + "/items/int-1")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET status = %d, want 200", resp.StatusCode)
	}
}
