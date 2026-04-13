package api

import (
	"net/http"
	"testing"

	"github.com/example/go-template/internal/testutil"
)

func TestChain(t *testing.T) {
	var order []string
	mw := func(tag string) Middleware {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, tag)
				next.ServeHTTP(w, r)
			})
		}
	}
	inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		order = append(order, "handler")
		w.WriteHeader(http.StatusOK)
	})
	h := Chain(inner, mw("a"), mw("b"))
	testutil.DoRequest(t, h, "GET", "/", "")
	if len(order) != 3 || order[0] != "a" || order[1] != "b" || order[2] != "handler" {
		t.Errorf("unexpected order: %v", order)
	}
}

func TestLogging(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})
	res := testutil.DoRequest(t, Logging(inner), "GET", "/test", "")
	testutil.Equal(t, res.Code, http.StatusTeapot)
}

func TestRecovery(t *testing.T) {
	inner := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		panic("boom")
	})
	res := testutil.DoRequest(t, Recovery(inner), "GET", "/", "")
	testutil.Equal(t, res.Code, http.StatusInternalServerError)
}

func TestCORS(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		JSON(w, http.StatusOK, "ok")
	})
	h := CORS(inner)

	// Normal request
	res := testutil.DoRequest(t, h, "GET", "/", "")
	testutil.Equal(t, res.Code, http.StatusOK)

	// OPTIONS preflight — need to check status directly via recorder
	res = testutil.DoRequest(t, h, "OPTIONS", "/", "")
	testutil.Equal(t, res.Code, http.StatusNoContent)
}
