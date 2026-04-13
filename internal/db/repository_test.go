package db

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

type user struct {
	ID   int
	Name string
}

func scanUser(s Scanner) (user, error) {
	var u user
	err := s.Scan(&u.ID, &u.Name)
	return u, err
}

func TestFindByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := &Repository[user]{DB: db, Table: "users", Scan: scanUser}

	mock.ExpectQuery("SELECT \\* FROM users WHERE id = \\?").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "Alice"))

	u, err := repo.FindByID(context.Background(), "id", 1)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if u.Name != "Alice" {
		t.Fatalf("Name = %q, want Alice", u.Name)
	}
}

func TestFindByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := &Repository[user]{DB: db, Table: "users", Scan: scanUser}

	mock.ExpectQuery("SELECT \\* FROM users WHERE id = \\?").
		WithArgs(99).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))

	_, err = repo.FindByID(context.Background(), "id", 99)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("FindByID() error = %v, want ErrNotFound", err)
	}
}

func TestFindAll(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := &Repository[user]{DB: db, Table: "users", Scan: scanUser}

	mock.ExpectQuery("SELECT \\* FROM users").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
			AddRow(1, "Alice").
			AddRow(2, "Bob"))

	users, err := repo.FindAll(context.Background())
	if err != nil {
		t.Fatalf("FindAll() error = %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("len = %d, want 2", len(users))
	}
}

func TestFindAll_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := &Repository[user]{DB: db, Table: "users", Scan: scanUser}

	mock.ExpectQuery("SELECT \\* FROM users").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))

	users, err := repo.FindAll(context.Background())
	if err != nil {
		t.Fatalf("FindAll() error = %v", err)
	}
	if len(users) != 0 {
		t.Fatalf("len = %d, want 0", len(users))
	}
}

func TestFindAll_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := &Repository[user]{DB: db, Table: "users", Scan: scanUser}

	mock.ExpectQuery("SELECT \\* FROM users").WillReturnError(errors.New("query fail"))

	_, err = repo.FindAll(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestExecInsert(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := &Repository[user]{DB: db, Table: "users", Scan: scanUser}

	mock.ExpectExec("INSERT INTO users").
		WithArgs("Alice").
		WillReturnResult(sqlmock.NewResult(42, 1))

	id, err := repo.ExecInsert(context.Background(), "INSERT INTO users (name) VALUES (?)", "Alice")
	if err != nil {
		t.Fatalf("ExecInsert() error = %v", err)
	}
	if id != 42 {
		t.Fatalf("id = %d, want 42", id)
	}
}

func TestExecInsert_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := &Repository[user]{DB: db, Table: "users", Scan: scanUser}

	mock.ExpectExec("INSERT INTO users").WillReturnError(errors.New("insert fail"))

	_, err = repo.ExecInsert(context.Background(), "INSERT INTO users (name) VALUES (?)", "Alice")
	if err == nil {
		t.Fatal("expected error")
	}
}
