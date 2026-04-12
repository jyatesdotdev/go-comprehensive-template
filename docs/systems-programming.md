# Systems Programming in Go

## System Calls & OS Interaction

Go provides OS interaction through `os`, `os/exec`, `syscall`, and `runtime` packages.

```go
// Process and environment info
os.Getpid()          // current PID
os.Hostname()         // machine hostname
os.Getenv("HOME")     // environment variable
runtime.GOOS          // "linux", "darwin", "windows"
runtime.NumCPU()      // available CPUs

// Process execution
cmd := exec.CommandContext(ctx, "ls", "-la")
output, err := cmd.CombinedOutput()

// Signals
sigCh := make(chan os.Signal, 1)
signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
<-sigCh // blocks until signal received
```

## CGO Basics

CGO enables calling C code from Go. Enable with `import "C"` preceded by a comment block.

```go
package main

/*
#include <stdlib.h>
#include <math.h>
*/
import "C"
import "fmt"

func main() {
    // Call C's sqrt
    result := C.sqrt(C.double(144))
    fmt.Println(float64(result)) // 12

    // Allocate/free C memory
    cstr := C.CString("hello")
    defer C.free(unsafe.Pointer(cstr))
}
```

**CGO guidelines:**
- Adds build complexity (requires C compiler)
- Disables cross-compilation by default
- Function calls across the boundary are ~100x slower than pure Go calls
- Use build tags to provide pure-Go fallbacks: `//go:build !cgo`
- Prefer pure Go when performance is acceptable

## Network Programming

### TCP Server/Client

```go
// Server: accept connections, handle concurrently
ln, _ := net.Listen("tcp", ":8080")
for {
    conn, _ := ln.Accept()
    go handleConn(conn) // one goroutine per connection
}

// Client: dial, write, read
conn, _ := net.Dial("tcp", "localhost:8080")
conn.Write([]byte("request"))
response, _ := io.ReadAll(conn)
```

Key patterns:
- Use `context.Context` for timeouts and cancellation
- Set deadlines with `conn.SetDeadline(time.Now().Add(5*time.Second))`
- Always close connections in defer
- Use `net.ListenConfig` for server-side context support

### UDP

```go
// Server
conn, _ := net.ListenPacket("udp", ":9090")
buf := make([]byte, 65535)
n, addr, _ := conn.ReadFrom(buf)
conn.WriteTo(buf[:n], addr) // echo back

// Client
conn, _ := net.Dial("udp", "localhost:9090")
conn.Write([]byte("ping"))
```

UDP vs TCP:
- UDP: connectionless, no delivery guarantee, lower overhead
- TCP: connection-oriented, reliable, ordered delivery
- Use UDP for: DNS, metrics, real-time data where loss is acceptable
- Use TCP for: HTTP, database connections, file transfer

## File I/O

### Basic Operations

```go
// Read entire file
data, err := os.ReadFile("config.json")

// Write entire file (creates or truncates)
os.WriteFile("output.txt", data, 0644)

// Streaming read (memory efficient for large files)
f, _ := os.Open("large.csv")
defer f.Close()
scanner := bufio.NewScanner(f)
for scanner.Scan() {
    process(scanner.Text())
}
```

### Atomic Writes

Prevent partial reads by writing to a temp file then renaming:

```go
func AtomicWrite(path string, data []byte, perm os.FileMode) error {
    tmp, _ := os.CreateTemp(filepath.Dir(path), ".tmp-*")
    tmp.Write(data)
    tmp.Chmod(perm)
    tmp.Sync()  // flush to disk
    tmp.Close()
    return os.Rename(tmp.Name(), path) // atomic on same filesystem
}
```

### File Permissions

```go
os.Chmod("file.txt", 0644)     // rw-r--r--
os.Chown("file.txt", uid, gid) // change owner
info, _ := os.Stat("file.txt")
info.Mode()                     // permission bits
info.Size()                     // file size
info.ModTime()                  // last modification
```

## Best Practices

1. **Always set timeouts** on network operations — unbounded reads/writes leak goroutines
2. **Use `io.LimitReader`** to prevent memory exhaustion from untrusted input
3. **Prefer `bufio`** for line-oriented I/O over manual byte slicing
4. **Atomic writes** for config files and any data read by other processes
5. **Avoid CGO** unless wrapping an existing C library with no Go alternative
6. **Use `os.CreateTemp`** instead of constructing temp paths manually
7. **Handle `EINTR`** — syscalls can be interrupted by signals on Unix

## See Also

- `pkg/systems/` — TCP/UDP helpers, atomic write, system info
- `examples/systems/` — Runnable demos for all patterns above
- [Concurrency](concurrency.md) — Goroutine patterns for network servers
- [Performance](performance.md) — Profiling system-level code
