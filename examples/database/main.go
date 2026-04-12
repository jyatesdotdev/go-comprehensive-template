// Example database demonstrates the internal/db package patterns:
// connection config, migrations, repository, and transactions.
//
// This example shows the setup and types without requiring a live database.
// To run against a real database, import a driver (e.g., github.com/mattn/go-sqlite3)
// and call db.Open(cfg).
//
// Run: go run ./examples/database
package main

import (
	"fmt"
	"time"

	"github.com/example/go-template/internal/db"
)

// User is a domain entity.
type User struct {
	ID    int64
	Name  string
	Email string
}

// scanUser maps a database row to a User (used with Repository[User]).
var _ = scanUser

func scanUser(s db.Scanner) (User, error) {
	var u User
	err := s.Scan(&u.ID, &u.Name, &u.Email)
	return u, err
}

func main() {
	// 1. Connection pool configuration
	cfg := db.DefaultConfig("postgres", "postgres://user:pass@localhost:5432/mydb?sslmode=disable")
	fmt.Printf("Pool config: MaxOpen=%d, MaxIdle=%d, MaxLifetime=%v\n",
		cfg.MaxOpenConns, cfg.MaxIdleConns, cfg.ConnMaxLifetime)

	// 2. Custom config
	cfg = db.Config{
		Driver:          "sqlite3",
		DSN:             ":memory:",
		MaxOpenConns:    10,
		MaxIdleConns:    2,
		ConnMaxLifetime: 3 * time.Minute,
		ConnMaxIdleTime: 30 * time.Second,
	}
	fmt.Printf("Custom config: Driver=%s, MaxOpen=%d\n", cfg.Driver, cfg.MaxOpenConns)

	// 3. Migration definitions
	migrations := []db.Migration{
		{
			Version:     1,
			Description: "create users table",
			Up:          `CREATE TABLE users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, email TEXT UNIQUE NOT NULL)`,
			Down:        `DROP TABLE users`,
		},
		{
			Version:     2,
			Description: "add created_at column",
			Up:          `ALTER TABLE users ADD COLUMN created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP`,
			Down:        `ALTER TABLE users DROP COLUMN created_at`,
		},
	}
	fmt.Printf("Defined %d migrations\n", len(migrations))

	// 4. Repository pattern (requires a *sql.DB connection)
	// repo := &db.Repository[User]{DB: conn, Table: "users", Scan: scanUser}
	// user, err := repo.FindByID(ctx, "id", 1)
	// users, err := repo.FindAll(ctx)
	// id, err := repo.ExecInsert(ctx, "INSERT INTO users (name, email) VALUES (?, ?)", "Alice", "alice@example.com")
	fmt.Println("Repository[User] ready for: FindByID, FindAll, ExecInsert")

	// 5. Transaction helper (requires a *sql.DB connection)
	// err := db.InTx(ctx, conn, nil, func(tx *sql.Tx) error {
	//     _, err := tx.ExecContext(ctx, "INSERT INTO users (name, email) VALUES (?, ?)", "Bob", "bob@example.com")
	//     return err
	// })
	fmt.Println("InTx: auto-commit on success, rollback on error/panic")

	// 6. Health check (requires a *sql.DB connection)
	// err := db.HealthCheck(ctx, conn)
	fmt.Println("HealthCheck: PingContext with 2s timeout")

	fmt.Println("\n--- To run with a real database ---")
	fmt.Println("1. Import a driver: _ \"github.com/mattn/go-sqlite3\"")
	fmt.Println("2. conn, err := db.Open(cfg)")
	fmt.Println("3. db.NewMigrator(conn, migrations).Up(ctx)")
	fmt.Println("4. Use Repository[T] and InTx for CRUD and transactions")
}
