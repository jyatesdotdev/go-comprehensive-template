package db

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"time"
)

// Migration represents a single schema migration.
type Migration struct {
	// Version is the unique, monotonically increasing migration number.
	Version int
	// Description is a short human-readable label for the migration.
	Description string
	// Up is the SQL executed when applying the migration.
	Up string
	// Down is the SQL executed when rolling back the migration.
	Down string
}

// Migrator runs schema migrations against a database.
type Migrator struct {
	db         *sql.DB
	table      string
	migrations []Migration
}

// NewMigrator creates a migrator that tracks applied versions in the given table.
func NewMigrator(db *sql.DB, migrations []Migration) *Migrator {
	return &Migrator{db: db, table: "schema_migrations", migrations: migrations}
}

// Up applies all pending migrations in version order.
func (m *Migrator) Up(ctx context.Context) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	applied, err := m.applied(ctx)
	if err != nil {
		return err
	}

	sorted := make([]Migration, len(m.migrations))
	copy(sorted, m.migrations)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Version < sorted[j].Version })

	for _, mg := range sorted {
		if applied[mg.Version] {
			continue
		}
		if err := InTx(ctx, m.db, nil, func(tx *sql.Tx) error {
			if _, err := tx.ExecContext(ctx, mg.Up); err != nil {
				return fmt.Errorf("migration %d (%s): %w", mg.Version, mg.Description, err)
			}
			_, err := tx.ExecContext(ctx,
				fmt.Sprintf("INSERT INTO %s (version, description, applied_at) VALUES (?, ?, ?)", m.table),
				mg.Version, mg.Description, time.Now().UTC(),
			)
			return err
		}); err != nil {
			return err
		}
	}
	return nil
}

func (m *Migrator) ensureTable(ctx context.Context) error {
	q := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
		version INTEGER PRIMARY KEY,
		description TEXT NOT NULL,
		applied_at TIMESTAMP NOT NULL
	)`, m.table)
	_, err := m.db.ExecContext(ctx, q)
	return err
}

func (m *Migrator) applied(ctx context.Context) (map[int]bool, error) {
	rows, err := m.db.QueryContext(ctx, fmt.Sprintf("SELECT version FROM %s", m.table))
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck // best-effort close
	out := make(map[int]bool)
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		out[v] = true
	}
	return out, rows.Err()
}
