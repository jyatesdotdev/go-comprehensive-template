// Package main demonstrates Go-idiomatic design patterns.
package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/example/go-template/internal/patterns"
)

func main() {
	// 1. Functional Options
	fmt.Println("=== Functional Options ===")
	srv := patterns.NewServer("0.0.0.0",
		patterns.WithPort(9090),
		patterns.WithMaxConns(500),
		patterns.WithReadTimeout(30*time.Second),
	)
	fmt.Printf("Server: %s:%d (maxConns=%d, readTimeout=%v)\n\n",
		srv.Addr, srv.Port, srv.MaxConns, srv.ReadTimeout)

	// 2. Interfaces & Strategy Pattern
	fmt.Println("=== Strategy via Interfaces ===")
	notifiers := patterns.MultiNotifier{
		patterns.EmailNotifier{Addr: "<email>"},
		patterns.SlackNotifier{Webhook: "https://hooks.slack.example.com/xyz"},
	}
	_ = notifiers.Notify("deployment complete")
	fmt.Println()

	// 3. Embedding / Composition
	fmt.Println("=== Embedding ===")
	u := patterns.NewUser("Alice", "<email>")
	fmt.Printf("User: %s (id=%s, age=%v)\n\n", u.Name, u.ID, u.Age().Truncate(time.Microsecond))

	// 4. Custom Error Types
	fmt.Println("=== Error Types ===")
	// Sentinel errors
	err := patterns.Wrap(patterns.ErrNotFound, "fetch user")
	fmt.Printf("wrapped: %v | Is NotFound: %v\n", err, errors.Is(err, patterns.ErrNotFound))

	// Typed errors with errors.As
	if err := patterns.Validate(""); err != nil {
		var ve *patterns.ValidationError
		if errors.As(err, &ve) {
			fmt.Printf("validation failed: field=%s msg=%s\n", ve.Field, ve.Message)
		}
	}
	fmt.Println("valid:", patterns.Validate("Alice"))
}
