# AGENTS.md — pkg/systems

Systems-programming utilities: OS/process info, atomic file writes, line streaming, and TCP/UDP
networking. Public API — stdlib only (`net`, `os`, `bufio`, `io`, `runtime`, `filepath`, `sync`).

Read [pkg/AGENTS.md](../AGENTS.md) and the [root](../../AGENTS.md) first. **Exported signatures are
a stable contract.**

## Exports & contracts

- **`SystemInfo() map[string]string`** — hostname, pid, uid, GOOS, GOARCH, CPU count, cwd.
- **`AtomicWrite(path, data, perm)`** — write-to-temp-then-rename for crash-safe file updates:
  creates a temp file **in the same directory** (so `rename` is atomic on one filesystem), writes,
  `Chmod`, `Sync` (flush to disk), `Close`, then `os.Rename`. On any error it removes the temp file.
  **Keep the full sequence** — dropping `Sync` or renaming across filesystems breaks atomicity.
- **`ReadLines(path, fn)`** — streams a file line by line via `bufio.Scanner`, calling `fn(line)`;
  stops early if `fn` returns an error. Constant memory — do not replace with `io.ReadAll` for large
  files. Note: default `Scanner` has a max token size (~64KB/line); raise the buffer if lines can be longer.
- **`TCPServer(ctx, addr, handler)`** — accept loop; a goroutine closes the listener on
  `ctx.Done()`; each connection is handled in its own goroutine and closed after `handler` returns;
  waits for in-flight connections via `WaitGroup` before returning.
- **`TCPSend(addr, data)`** — dials, writes, **half-closes with `CloseWrite`** (signals EOF so the
  peer's `io.ReadAll` returns), then reads the full response. Preserve the `CloseWrite`.
- **`UDPServer(ctx, addr, handler)` / `UDPSend(addr, data)`** — datagram equivalents; both use a
  **65535-byte** buffer (max UDP payload) and copy each packet before handing it to `handler`
  (the shared read buffer is reused).

## Business rules & safety

- **Atomicity depends on same-filesystem rename** — the temp file is created in `filepath.Dir(path)`
  precisely for this reason. Don't relocate it to `os.TempDir()`.
- **Copy UDP packets before use.** `UDPServer` copies `buf[:n]` into a fresh slice because the read
  buffer is overwritten on the next `ReadFromUDP`. Any new packet handling must copy too.
- Networking calls that block or spawn goroutines are `ctx`-scoped; cancellation closes the
  listener/socket to unblock `Accept`/`ReadFromUDP`.
- Justified suppressions here reflect design intent: `// #nosec G304 -- path is caller-controlled
  by design` (this is a file utility — callers pass the path) and `// #nosec G104` best-effort
  closes/cleanup. Do not route untrusted input into `path` without validating upstream.

## Go best practices for systems code

- Always `defer` close of files/connections; treat close errors on read paths as best-effort but
  **check close/rename errors on write paths** (data loss risk) — this package returns them.
- Stream, don't slurp: prefer `bufio.Scanner`/`io.Copy` over reading whole files into memory.
- Make network servers cancellable and drain in-flight work before returning.

## Testing

`systems_test.go` exercises round trips on ephemeral ports (`127.0.0.1:0`) and temp files/dirs
(`t.TempDir()`). Bind to port 0 and read back the assigned address to avoid flaky fixed ports; use
a `context` you cancel to stop servers. Runnable demo: [`examples/systems`](../../examples/AGENTS.md).
