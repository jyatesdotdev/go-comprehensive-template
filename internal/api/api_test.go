package api

import (
	"net/http"
	"testing"

	"github.com/example/go-template/internal/testutil"
)

// TestItemHandler_CRUD uses table-driven subtests to exercise all CRUD routes.
func TestItemHandler_CRUD(t *testing.T) {
	store := NewStore()
	handler := ItemHandler(store)

	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		wantCode   int
		wantSubstr string
	}{
		{"list empty", "GET", "/items", "", http.StatusOK, "[]"},
		{"create item", "POST", "/items", `{"id":"1","name":"Widget"}`, http.StatusCreated, "Widget"},
		{"get item", "GET", "/items/1", "", http.StatusOK, "Widget"},
		{"get missing", "GET", "/items/999", "", http.StatusNotFound, "not found"},
		{"create no id", "POST", "/items", `{"name":"X"}`, http.StatusBadRequest, "id required"},
		{"delete item", "DELETE", "/items/1", "", http.StatusOK, "deleted"},
		{"delete missing", "DELETE", "/items/999", "", http.StatusNotFound, "not found"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := testutil.DoRequest(t, handler, tc.method, tc.path, tc.body)
			testutil.Equal(t, res.Code, tc.wantCode)
			if tc.wantSubstr != "" && !contains(res.Body, tc.wantSubstr) {
				t.Errorf("body %q missing %q", res.Body, tc.wantSubstr)
			}
		})
	}
}

// TestJSON verifies that the JSON helper writes a valid response envelope.
func TestJSON(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		JSON(w, http.StatusOK, map[string]string{"key": "val"})
	})
	res := testutil.DoRequest(t, handler, "GET", "/", "")
	testutil.Equal(t, res.Code, http.StatusOK)

	var resp Response
	testutil.DecodeJSON(t, res.Body, &resp)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || findSubstr(s, substr))
}

func findSubstr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
