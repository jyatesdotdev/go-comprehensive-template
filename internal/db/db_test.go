package db

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig("postgres", "host=localhost")
	if cfg.Driver != "postgres" {
		t.Fatalf("Driver = %q, want postgres", cfg.Driver)
	}
	if cfg.DSN != "host=localhost" {
		t.Fatalf("DSN = %q, want host=localhost", cfg.DSN)
	}
	if cfg.MaxOpenConns != 25 {
		t.Fatalf("MaxOpenConns = %d, want 25", cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns != 5 {
		t.Fatalf("MaxIdleConns = %d, want 5", cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime != 5*time.Minute {
		t.Fatalf("ConnMaxLifetime = %v, want 5m", cfg.ConnMaxLifetime)
	}
	if cfg.ConnMaxIdleTime != 1*time.Minute {
		t.Fatalf("ConnMaxIdleTime = %v, want 1m", cfg.ConnMaxIdleTime)
	}
}

func TestHealthCheck(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectPing()
	if err := HealthCheck(context.Background(), db); err != nil {
		t.Fatalf("HealthCheck() error = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestHealthCheckError(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectPing().WillReturnError(errors.New("down"))
	if err := HealthCheck(context.Background(), db); err == nil {
		t.Fatal("expected error")
	}
}

func TestInTx_Commit(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectCommit()

	err = InTx(context.Background(), db, nil, func(_ *sql.Tx) error {
		return nil
	})
	if err != nil {
		t.Fatalf("InTx() error = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestInTx_Rollback(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectRollback()

	fnErr := errors.New("fail")
	err = InTx(context.Background(), db, nil, func(_ *sql.Tx) error {
		return fnErr
	})
	if !errors.Is(err, fnErr) {
		t.Fatalf("InTx() error = %v, want %v", err, fnErr)
	}
}

func TestInTx_Panic(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectRollback()

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic")
		}
		if r != "boom" {
			t.Fatalf("recover = %v, want boom", r)
		}
	}()

	_ = InTx(context.Background(), db, nil, func(_ *sql.Tx) error {
		panic("boom")
	})
}

func TestInTx_BeginError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectBegin().WillReturnError(errors.New("begin fail"))

	err = InTx(context.Background(), db, nil, func(_ *sql.Tx) error {
		return nil
	})
	if err == nil {
		t.Fatal("expected error")
	}
}
