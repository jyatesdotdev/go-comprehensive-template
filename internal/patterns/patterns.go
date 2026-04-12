// Package patterns demonstrates Go-idiomatic design patterns:
// functional options, interfaces, composition, embedding, and error types.
package patterns

import (
	"errors"
	"fmt"
	"time"
)

// --- Functional Options Pattern ---

// Server demonstrates the functional options pattern for flexible construction.
type Server struct {
	Addr         string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	MaxConns     int
}

// Option configures a Server.
type Option func(*Server)

// WithPort sets the server listen port.
func WithPort(p int) Option { return func(s *Server) { s.Port = p } }

// WithReadTimeout sets the read timeout duration.
func WithReadTimeout(d time.Duration) Option { return func(s *Server) { s.ReadTimeout = d } }

// WithWriteTimeout sets the write timeout duration.
func WithWriteTimeout(d time.Duration) Option { return func(s *Server) { s.WriteTimeout = d } }

// WithMaxConns sets the maximum number of concurrent connections.
func WithMaxConns(n int) Option { return func(s *Server) { s.MaxConns = n } }

// NewServer creates a Server with the given address and applies any options.
// Defaults: port 8080, read timeout 5s, write timeout 10s, max connections 100.
func NewServer(addr string, opts ...Option) *Server {
	s := &Server{Addr: addr, Port: 8080, ReadTimeout: 5 * time.Second, WriteTimeout: 10 * time.Second, MaxConns: 100}
	for _, o := range opts {
		o(s)
	}
	return s
}

// --- Interfaces & Composition ---

// Reader is a minimal read interface (mirrors io.Reader).
type Reader interface{ Read(p []byte) (int, error) }

// Writer is a minimal write interface (mirrors io.Writer).
type Writer interface{ Write(p []byte) (int, error) }

// ReadWriter combines Reader and Writer via interface composition.
type ReadWriter interface {
	Reader
	Writer
}

// Notifier demonstrates strategy pattern via interfaces.
type Notifier interface {
	Notify(msg string) error
}

// EmailNotifier sends notifications via email.
type EmailNotifier struct{ Addr string }

// SlackNotifier sends notifications via a Slack webhook.
type SlackNotifier struct{ Webhook string }

// Notify sends msg via email.
func (e EmailNotifier) Notify(msg string) error {
	fmt.Printf("[Email→%s] %s\n", e.Addr, msg)
	return nil
}

// Notify sends msg via Slack webhook.
func (s SlackNotifier) Notify(msg string) error {
	fmt.Printf("[Slack→%s] %s\n", s.Webhook, msg)
	return nil
}

// MultiNotifier composes multiple notifiers (decorator/composite pattern).
type MultiNotifier []Notifier

// Notify sends msg to all notifiers, joining any errors.
func (mn MultiNotifier) Notify(msg string) error {
	var errs []error
	for _, n := range mn {
		if err := n.Notify(msg); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// --- Embedding (Composition over Inheritance) ---

// Base provides common fields via embedding.
type Base struct {
	ID        string
	CreatedAt time.Time
}

// Age returns the duration since the entity was created.
func (b Base) Age() time.Duration { return time.Since(b.CreatedAt) }

// User embeds Base, gaining its fields and methods.
type User struct {
	Base
	Name  string
	Email string
}

// NewUser creates a User with a generated ID and the current timestamp.
func NewUser(name, email string) User {
	return User{Base: Base{ID: fmt.Sprintf("u_%d", time.Now().UnixNano()), CreatedAt: time.Now()}, Name: name, Email: email}
}

// --- Custom Error Types ---

// Sentinel errors for comparison with errors.Is.
var (
	// ErrNotFound indicates a requested resource does not exist.
	ErrNotFound = errors.New("not found")
	// ErrUnauthorized indicates the caller lacks permission.
	ErrUnauthorized = errors.New("unauthorized")
)

// ValidationError is a structured error supporting errors.As.
type ValidationError struct {
	Field   string
	Message string
}

// Error returns a human-readable validation error string.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation: %s — %s", e.Field, e.Message)
}

// Wrap adds context to errors while preserving the chain.
func Wrap(err error, msg string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", msg, err)
}

// Validate demonstrates returning typed errors.
func Validate(name string) error {
	if name == "" {
		return &ValidationError{Field: "name", Message: "required"}
	}
	if len(name) > 50 {
		return &ValidationError{Field: "name", Message: "too long"}
	}
	return nil
}
