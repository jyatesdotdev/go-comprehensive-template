// Package db provides database interaction patterns using database/sql.
// It demonstrates connection pooling, repository pattern, migrations,
// and transaction helpers that work with any database/sql driver.
package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Config holds database connection and pool settings.
type Config struct {
	// Driver is the database/sql driver name (e.g. "postgres", "sqlite3").
	Driver string
	// DSN is the data source name / connection string.
	DSN string
	// MaxOpenConns is the maximum number of open connections to the database.
	MaxOpenConns int
	// MaxIdleConns is the maximum number of idle connections in the pool.
	MaxIdleConns int
	// ConnMaxLifetime is the maximum time a connection may be reused.
	ConnMaxLifetime time.Duration
	// ConnMaxIdleTime is the maximum time a connection may sit idle.
	ConnMaxIdleTime time.Duration
}

// DefaultConfig returns sensible pool defaults.
func DefaultConfig(driver, dsn string) Config {
	return Config{
		Driver:          driver,
		DSN:             dsn,
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 1 * time.Minute,
	}
}

// Open creates a *sql.DB with connection pool settings applied.
func Open(cfg Config) (*sql.DB, error) {
	db, err := sql.Open(cfg.Driver, cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("db open: %w", err)
	}
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	if err := db.Ping(); err != nil {
		_ = db.Close() // #nosec G104 -- best-effort cleanup on ping failure
		return nil, fmt.Errorf("db ping: %w", err)
	}
	return db, nil
}

// HealthCheck verifies the database connection with a timeout.
func HealthCheck(ctx context.Context, db *sql.DB) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return db.PingContext(ctx)
}

// InTx executes fn within a transaction, committing on success or
// rolling back on error/panic.
func InTx(ctx context.Context, db *sql.DB, opts *sql.TxOptions, fn func(*sql.Tx) error) (err error) {
	tx, err := db.BeginTx(ctx, opts)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	if err = fn(tx); err != nil { //nolint:gocritic // named return used in defer for rollback
		return err
	}
	return tx.Commit()
}
