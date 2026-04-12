// Package main demonstrates third-party integration patterns:
// wrapping dependencies behind interfaces, adapter pattern, and dependency injection.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// --- Interface Wrapping ---

// Logger defines the application's logging contract.
// Third-party loggers (zap, zerolog) are wrapped behind this interface.
type Logger interface {
	Info(msg string, args ...any)
	Error(msg string, args ...any)
}

// SlogAdapter wraps the stdlib slog.Logger behind our Logger interface.
type SlogAdapter struct{ l *slog.Logger }

// Info logs an informational message via slog.
func (s *SlogAdapter) Info(msg string, args ...any) { s.l.Info(msg, args...) }

// Error logs an error message via slog.
func (s *SlogAdapter) Error(msg string, args ...any) { s.l.Error(msg, args...) }

// --- Adapter Pattern: Swappable Storage ---

// Storage defines a key-value store contract.
type Storage interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, val []byte, ttl time.Duration) error
}

// MemStorage is an in-memory Storage adapter (swap for Redis, DynamoDB, etc.).
type MemStorage struct {
	mu    sync.RWMutex
	items map[string][]byte
}

// NewMemStorage returns an initialized in-memory Storage.
func NewMemStorage() *MemStorage {
	return &MemStorage{items: make(map[string][]byte)}
}

// Get retrieves the value for key from the in-memory store.
func (m *MemStorage) Get(_ context.Context, key string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok := m.items[key]
	if !ok {
		return nil, fmt.Errorf("key %q not found", key)
	}
	return v, nil
}

// Set stores a key-value pair in the in-memory store. The ttl parameter is ignored.
func (m *MemStorage) Set(_ context.Context, key string, val []byte, _ time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items[key] = val
	return nil
}

// --- Dependency Injection ---

// App composes dependencies via interfaces, not concrete types.
type App struct {
	logger  Logger
	storage Storage
}

// NewApp creates an App with the given dependencies.
func NewApp(logger Logger, storage Storage) *App {
	return &App{logger: logger, storage: storage}
}

// Run stores and retrieves a greeting to demonstrate the wired dependencies.
func (a *App) Run(ctx context.Context) {
	a.logger.Info("storing value", "key", "greeting")
	_ = a.storage.Set(ctx, "greeting", []byte("Hello, Go!"), 0)

	val, err := a.storage.Get(ctx, "greeting")
	if err != nil {
		a.logger.Error("get failed", "error", err)
		return
	}
	a.logger.Info("retrieved value", "key", "greeting", "value", string(val))
}

func main() {
	fmt.Println("=== Third-Party Integration Patterns ===")
	fmt.Println()

	// Wire dependencies — swap implementations without changing App
	logger := &SlogAdapter{l: slog.Default()}
	storage := NewMemStorage()
	app := NewApp(logger, storage)

	app.Run(context.Background())

	fmt.Println("\nKey takeaway: depend on interfaces, not concrete third-party types.")
	fmt.Println("This lets you swap implementations (e.g., MemStorage → Redis) with zero changes to business logic.")
}
