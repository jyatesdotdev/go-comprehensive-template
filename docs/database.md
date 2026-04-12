# Database Interaction in Go

## database/sql (Standard Library)

The `database/sql` package provides a generic interface around SQL databases. It handles connection pooling automatically.

### Connection Pooling

```go
db, _ := sql.Open("postgres", dsn)
db.SetMaxOpenConns(25)          // max simultaneous connections
db.SetMaxIdleConns(5)           // idle connections kept alive
db.SetConnMaxLifetime(5 * time.Minute)  // recycle connections
db.SetConnMaxIdleTime(1 * time.Minute)  // close idle connections
```

**Guidelines:**
- `MaxOpenConns` — set to your DB's max connections divided by app instances
- `MaxIdleConns` — keep low to avoid holding stale connections
- `ConnMaxLifetime` — prevents using connections the server has closed
- Always call `db.Ping()` after opening to verify connectivity

### Querying

```go
// Single row
var name string
err := db.QueryRowContext(ctx, "SELECT name FROM users WHERE id = ?", 1).Scan(&name)
if errors.Is(err, sql.ErrNoRows) { /* not found */ }

// Multiple rows — always close and check Err()
rows, err := db.QueryContext(ctx, "SELECT id, name FROM users")
if err != nil { return err }
defer rows.Close()
for rows.Next() {
    var id int; var name string
    rows.Scan(&id, &name)
}
if err := rows.Err(); err != nil { return err }
```

### Transactions

```go
tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
if err != nil { return err }
defer tx.Rollback() // no-op after commit

_, err = tx.ExecContext(ctx, "UPDATE accounts SET balance = balance - ? WHERE id = ?", amount, from)
if err != nil { return err }
_, err = tx.ExecContext(ctx, "UPDATE accounts SET balance = balance + ? WHERE id = ?", amount, to)
if err != nil { return err }

return tx.Commit()
```

See `internal/db.InTx()` for a panic-safe transaction helper.

### Prepared Statements

```go
stmt, err := db.PrepareContext(ctx, "INSERT INTO users (name) VALUES (?)")
defer stmt.Close()
for _, name := range names {
    stmt.ExecContext(ctx, name)
}
```

Use prepared statements when executing the same query many times in a loop.

## sqlx

[sqlx](https://github.com/jmoiron/sqlx) extends `database/sql` with struct scanning and named queries. It's a thin wrapper — no ORM overhead.

```go
import "github.com/jmoiron/sqlx"

type User struct {
    ID    int    `db:"id"`
    Name  string `db:"name"`
    Email string `db:"email"`
}

db := sqlx.MustConnect("postgres", dsn)

// Struct scanning
var user User
db.GetContext(ctx, &user, "SELECT * FROM users WHERE id = $1", 1)

// Slice scanning
var users []User
db.SelectContext(ctx, &users, "SELECT * FROM users ORDER BY name")

// Named queries
db.NamedExecContext(ctx, "INSERT INTO users (name, email) VALUES (:name, :email)", user)
```

**When to use sqlx:** You want less boilerplate than `database/sql` but don't want a full ORM.

## GORM

[GORM](https://gorm.io) is a full-featured ORM with auto-migrations, associations, hooks, and query building.

```go
import "gorm.io/gorm"
import "gorm.io/driver/postgres"

type User struct {
    gorm.Model
    Name  string
    Email string `gorm:"uniqueIndex"`
}

db, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{})
db.AutoMigrate(&User{})

// CRUD
db.Create(&User{Name: "Alice", Email: "alice@example.com"})
var user User
db.First(&user, 1)
db.Model(&user).Update("Name", "Bob")
db.Delete(&user)

// Querying
var users []User
db.Where("name LIKE ?", "%ali%").Find(&users)
```

**When to use GORM:** Rapid prototyping, complex associations, when you prefer convention over configuration. Avoid for performance-critical paths — use raw SQL or sqlx instead.

## Migrations

### Approach 1: Code-based (this template)

The `internal/db.Migrator` runs SQL migrations tracked in a `schema_migrations` table:

```go
migrations := []db.Migration{
    {Version: 1, Description: "create users", Up: "CREATE TABLE users (...)"},
    {Version: 2, Description: "add index", Up: "CREATE INDEX idx_email ON users(email)"},
}
db.NewMigrator(conn, migrations).Up(ctx)
```

### Approach 2: File-based tools

- [golang-migrate](https://github.com/golang-migrate/migrate) — CLI + library, supports many databases
- [goose](https://github.com/pressly/goose) — SQL or Go migrations
- [atlas](https://atlasgo.io) — declarative schema management

## Repository Pattern

The generic `Repository[T]` in `internal/db/repository.go` provides type-safe CRUD:

```go
repo := &db.Repository[User]{
    DB:    conn,
    Table: "users",
    Scan:  func(s db.Scanner) (User, error) {
        var u User
        err := s.Scan(&u.ID, &u.Name, &u.Email)
        return u, err
    },
}
user, err := repo.FindByID(ctx, "id", 1)
users, err := repo.FindAll(ctx)
```

For production, define domain-specific repository interfaces:

```go
type UserRepository interface {
    FindByID(ctx context.Context, id int64) (User, error)
    FindByEmail(ctx context.Context, email string) (User, error)
    Create(ctx context.Context, u *User) error
}
```

This decouples business logic from the storage layer and simplifies testing with mocks.

## Choosing a Library

| Feature | database/sql | sqlx | GORM |
|---------|-------------|------|------|
| Struct scanning | Manual | Automatic | Automatic |
| Query building | Raw SQL | Raw SQL | Chainable API |
| Migrations | External | External | AutoMigrate |
| Performance | Best | Near-native | Overhead |
| Learning curve | Low | Low | Medium |
| Dependencies | None | Minimal | Heavy |

**Recommendation:** Start with `database/sql` + the repository pattern. Adopt sqlx if scanning boilerplate becomes painful. Use GORM only when its features (associations, hooks, auto-migrate) justify the overhead.

## See Also

- [RESTful API](restful-api.md) — API handlers that use database layers
- [Testing](testing.md) — Testing database code with mocks and integration tests
- [Best Practices](best-practices.md) — Interface design and error handling
