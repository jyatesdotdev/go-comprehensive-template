package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// ErrNotFound is returned when a query finds no matching row.
var ErrNotFound = errors.New("not found")

// Scanner abstracts row scanning (works with *sql.Row and *sql.Rows).
type Scanner interface {
	Scan(dest ...any) error
}

// Repository provides generic CRUD operations for a table.
// The scan function maps a row to the entity type T.
type Repository[T any] struct {
	// DB is the database connection pool.
	DB *sql.DB
	// Table is the SQL table name.
	Table string
	// Scan converts a row into an entity of type T.
	Scan func(Scanner) (T, error)
}

// FindByID retrieves a single row by its primary key column.
func (r *Repository[T]) FindByID(ctx context.Context, idCol string, id any) (T, error) {
	q := fmt.Sprintf("SELECT * FROM %s WHERE %s = ?", r.Table, idCol) // #nosec G201 -- table/column names are developer-controlled struct fields
	row := r.DB.QueryRowContext(ctx, q, id)
	entity, err := r.Scan(row)
	if errors.Is(err, sql.ErrNoRows) {
		var zero T
		return zero, ErrNotFound
	}
	return entity, err
}

// FindAll retrieves all rows from the table.
func (r *Repository[T]) FindAll(ctx context.Context) ([]T, error) {
	q := fmt.Sprintf("SELECT * FROM %s", r.Table) // #nosec G201 -- table name is a developer-controlled struct field
	rows, err := r.DB.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck // best-effort close

	var results []T
	for rows.Next() {
		entity, err := r.Scan(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, entity)
	}
	return results, rows.Err()
}

// ExecInsert runs an INSERT statement and returns the last insert ID.
func (r *Repository[T]) ExecInsert(ctx context.Context, query string, args ...any) (int64, error) {
	result, err := r.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}
