// Package main demonstrates Go error handling patterns.
package main

import (
	"errors"
	"fmt"
	"log"
)

// Sentinel errors — package-level, checked with errors.Is.
var (
	ErrNotFound = errors.New("not found")
	ErrDenied   = errors.New("access denied")
)

// ValidationError is a custom error type — checked with errors.As.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation: %s — %s", e.Field, e.Message)
}

// findUser demonstrates wrapping and sentinel errors.
func findUser(id int) (string, error) {
	if id <= 0 {
		return "", &ValidationError{Field: "id", Message: "must be positive"}
	}
	if id == 42 {
		return "", fmt.Errorf("findUser(%d): %w", id, ErrNotFound)
	}
	return fmt.Sprintf("user-%d", id), nil
}

func main() {
	for _, id := range []int{-1, 42, 7} {
		name, err := findUser(id)
		switch {
		case err == nil:
			log.Printf("found: %s", name)

		case errors.Is(err, ErrNotFound):
			log.Printf("id %d: %v", id, err)

		default:
			var ve *ValidationError
			if errors.As(err, &ve) {
				log.Printf("bad input on field %q: %s", ve.Field, ve.Message)
			} else {
				log.Printf("unexpected: %v", err)
			}
		}
	}
}
