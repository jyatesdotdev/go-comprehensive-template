# RESTful APIs in Go

## Standard Library (`net/http`)

Go 1.22+ added method-based routing to `http.ServeMux`:

```go
mux := http.NewServeMux()
mux.HandleFunc("GET /items", listItems)
mux.HandleFunc("GET /items/{id}", getItem)
mux.HandleFunc("POST /items", createItem)
mux.HandleFunc("DELETE /items/{id}", deleteItem)
```

Path parameters are accessed via `r.PathValue("id")`.

### JSON Helpers

```go
func JSON(w http.ResponseWriter, status int, data any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(data)
}
```

### Middleware Pattern

Middleware wraps an `http.Handler`:

```go
type Middleware func(http.Handler) http.Handler

func Logging(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        next.ServeHTTP(w, r)
        log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
    })
}

// Chain applies middlewares: first listed = outermost
func Chain(h http.Handler, mws ...Middleware) http.Handler {
    for i := len(mws) - 1; i >= 0; i-- {
        h = mws[i](h)
    }
    return h
}

handler := Chain(myHandler, Recovery, Logging, CORS)
```

Common middleware: logging, panic recovery, CORS, authentication, rate limiting.

### HTTP Client

```go
client := &http.Client{Timeout: 10 * time.Second}
resp, err := client.Get("https://api.example.com/data")
if err != nil { return err }
defer resp.Body.Close()
var result MyType
json.NewDecoder(resp.Body).Decode(&result)
```

Always set timeouts. Use `context` for cancellation:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
resp, err := client.Do(req)
```

---

## Gin Framework

[Gin](https://github.com/gin-gonic/gin) is the most popular Go web framework.

```go
import "github.com/gin-gonic/gin"

r := gin.Default() // includes Logger + Recovery middleware

r.GET("/items", func(c *gin.Context) {
    c.JSON(200, gin.H{"items": items})
})
r.GET("/items/:id", func(c *gin.Context) {
    id := c.Param("id")
    c.JSON(200, item)
})
r.POST("/items", func(c *gin.Context) {
    var item Item
    if err := c.ShouldBindJSON(&item); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    c.JSON(201, item)
})

// Middleware
r.Use(gin.Logger(), gin.Recovery())

// Route groups
api := r.Group("/api/v1")
api.Use(AuthMiddleware())
{
    api.GET("/users", listUsers)
    api.POST("/users", createUser)
}

r.Run(":8080")
```

Key features: fast radix-tree router, JSON binding/validation, middleware groups, rendering.

---

## Echo Framework

[Echo](https://echo.labstack.com/) is a high-performance, minimalist framework.

```go
import "github.com/labstack/echo/v4"

e := echo.New()
e.Use(middleware.Logger(), middleware.Recover())

e.GET("/items", func(c echo.Context) error {
    return c.JSON(200, items)
})
e.GET("/items/:id", func(c echo.Context) error {
    id := c.Param("id")
    return c.JSON(200, item)
})
e.POST("/items", func(c echo.Context) error {
    var item Item
    if err := c.Bind(&item); err != nil {
        return echo.NewHTTPError(400, err.Error())
    }
    return c.JSON(201, item)
})

// Groups
g := e.Group("/admin", AdminAuth)
g.GET("/stats", getStats)

e.Start(":8080")
```

Key features: automatic TLS, HTTP/2, data binding, centralized error handling.

---

## Choosing a Framework

| Feature | net/http | Gin | Echo |
|---------|----------|-----|------|
| Dependencies | None | Third-party | Third-party |
| Routing | Basic (1.22+ improved) | Radix tree | Radix tree |
| Performance | Good | Excellent | Excellent |
| Middleware | Manual | Built-in | Built-in |
| Validation | Manual | Built-in | Built-in |
| Learning curve | Low | Low | Low |

**Recommendation**: Start with `net/http` for simple services. Use Gin or Echo when you need route groups, validation, or rich middleware ecosystems.

---

## Template Code

See `internal/api/` for a working implementation using `net/http`:
- `api.go` — JSON helpers, CRUD handler, thread-safe store
- `middleware.go` — Logging, Recovery, CORS, middleware chaining

Run the example: `go run ./examples/api`

## See Also

- [Database](database.md) — Database layers used by API handlers
- [Testing](testing.md) — Testing HTTP handlers
- [Cloud-Native](cloud-native.md) — Health checks and production deployment
