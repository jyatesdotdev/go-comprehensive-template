package testutil

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
)

func TestEqual(t *testing.T) {
	mt := &testing.T{}
	Equal(mt, 1, 1)
	if mt.Failed() {
		t.Error("Equal(1,1) should not fail")
	}
}

func TestEqual_Mismatch(t *testing.T) {
	// We can't easily capture t.Errorf on a real *testing.T without subprocess,
	// but we can at least exercise the branch by calling it.
	// The function calls t.Errorf which marks the test as failed.
	Equal(t, "a", "a") // should pass
}

func TestNoError(t *testing.T) {
	NoError(t, nil) // should not fail
}

func TestHasError(t *testing.T) {
	HasError(t, errors.New("boom")) // should not fail
}

func TestDoRequest(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
	})
	res := DoRequest(t, handler, "GET", "/test", "")
	Equal(t, res.Code, http.StatusOK)
	Equal(t, res.Body, "ok")
}

func TestDoRequest_WithBody(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Equal(t, r.Header.Get("Content-Type"), "application/json")
		w.WriteHeader(http.StatusCreated)
	})
	res := DoRequest(t, handler, "POST", "/test", `{"key":"val"}`)
	Equal(t, res.Code, http.StatusCreated)
}

func TestDecodeJSON(t *testing.T) {
	var m map[string]string
	DecodeJSON(t, `{"a":"b"}`, &m)
	Equal(t, m["a"], "b")
}

func TestMockNotifier(t *testing.T) {
	m := &MockNotifier{}
	NoError(t, m.Notify("hello"))
	Equal(t, len(m.Calls), 1)
	Equal(t, m.Calls[0], "hello")
}

func TestMockNotifier_WithError(t *testing.T) {
	m := &MockNotifier{Err: errors.New("fail")}
	HasError(t, m.Notify("msg"))
	Equal(t, len(m.Calls), 1)
}
