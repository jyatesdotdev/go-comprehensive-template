// Package api provides RESTful API handlers, middleware, and JSON helpers.
package api

import (
	"encoding/json"
	"net/http"
	"sync"
)

// Response is a standard JSON response envelope.
type Response struct {
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

// JSON writes a JSON response with the given status code.
func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Response{Data: data}) // #nosec G104 -- best-effort HTTP response write
}

// Error writes a JSON error response.
func Error(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Response{Error: msg}) // #nosec G104 -- best-effort HTTP response write
}

// Decode reads JSON from the request body into v.
func Decode(r *http.Request, v any) error {
	defer r.Body.Close() //nolint:errcheck // best-effort close
	return json.NewDecoder(r.Body).Decode(v)
}

// Item is a sample resource for CRUD demonstrations.
type Item struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Store is a thread-safe in-memory store for Items.
type Store struct {
	mu    sync.RWMutex
	items map[string]Item
}

// NewStore creates an empty Store.
func NewStore() *Store {
	return &Store{items: make(map[string]Item)}
}

// ItemHandler returns an http.Handler for /items CRUD.
// Routes: GET /items, GET /items/{id}, POST /items, DELETE /items/{id}
func ItemHandler(s *Store) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /items", func(w http.ResponseWriter, _ *http.Request) {
		s.mu.RLock()
		defer s.mu.RUnlock()
		items := make([]Item, 0, len(s.items))
		for _, it := range s.items {
			items = append(items, it)
		}
		JSON(w, http.StatusOK, items)
	})
	mux.HandleFunc("GET /items/{id}", func(w http.ResponseWriter, r *http.Request) {
		s.mu.RLock()
		defer s.mu.RUnlock()
		it, ok := s.items[r.PathValue("id")]
		if !ok {
			Error(w, http.StatusNotFound, "not found")
			return
		}
		JSON(w, http.StatusOK, it)
	})
	mux.HandleFunc("POST /items", func(w http.ResponseWriter, r *http.Request) {
		var it Item
		if err := Decode(r, &it); err != nil {
			Error(w, http.StatusBadRequest, err.Error())
			return
		}
		if it.ID == "" {
			Error(w, http.StatusBadRequest, "id required")
			return
		}
		s.mu.Lock()
		s.items[it.ID] = it
		s.mu.Unlock()
		JSON(w, http.StatusCreated, it)
	})
	mux.HandleFunc("DELETE /items/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		s.mu.Lock()
		defer s.mu.Unlock()
		if _, ok := s.items[id]; !ok {
			Error(w, http.StatusNotFound, "not found")
			return
		}
		delete(s.items, id)
		JSON(w, http.StatusOK, map[string]string{"deleted": id})
	})
	return mux
}
