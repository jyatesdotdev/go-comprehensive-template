package db

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestMigrator_Up(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	migrations := []Migration{
		{Version: 1, Description: "create users", Up: "CREATE TABLE users (id INT)", Down: "DROP TABLE users"},
	}
	m := NewMigrator(db, migrations)

	// ensureTable
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS schema_migrations").
		WillReturnResult(sqlmock.NewResult(0, 0))
	// applied query returns empty
	mock.ExpectQuery("SELECT version FROM schema_migrations").
		WillReturnRows(sqlmock.NewRows([]string{"version"}))
	// transaction for migration 1
	mock.ExpectBegin()
	mock.ExpectExec("CREATE TABLE users").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO schema_migrations").
		WithArgs(1, "create users", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := m.Up(context.Background()); err != nil {
		t.Fatalf("Up() error = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMigrator_Up_SkipsApplied(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	migrations := []Migration{
		{Version: 1, Description: "create users", Up: "CREATE TABLE users (id INT)"},
	}
	m := NewMigrator(db, migrations)

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS schema_migrations").
		WillReturnResult(sqlmock.NewResult(0, 0))
	// version 1 already applied
	mock.ExpectQuery("SELECT version FROM schema_migrations").
		WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow(1))

	if err := m.Up(context.Background()); err != nil {
		t.Fatalf("Up() error = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMigrator_Up_EnsureTableError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	m := NewMigrator(db, nil)

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS schema_migrations").
		WillReturnError(errors.New("create fail"))

	if err := m.Up(context.Background()); err == nil {
		t.Fatal("expected error")
	}
}

func TestMigrator_Up_AppliedQueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	m := NewMigrator(db, nil)

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS schema_migrations").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("SELECT version FROM schema_migrations").
		WillReturnError(errors.New("query fail"))

	if err := m.Up(context.Background()); err == nil {
		t.Fatal("expected error")
	}
}

func TestMigrator_Up_MigrationError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	migrations := []Migration{
		{Version: 1, Description: "bad", Up: "INVALID SQL"},
	}
	m := NewMigrator(db, migrations)

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS schema_migrations").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("SELECT version FROM schema_migrations").
		WillReturnRows(sqlmock.NewRows([]string{"version"}))
	mock.ExpectBegin()
	mock.ExpectExec("INVALID SQL").
		WillReturnError(errors.New("syntax error"))
	mock.ExpectRollback()

	if err := m.Up(context.Background()); err == nil {
		t.Fatal("expected error")
	}
}
