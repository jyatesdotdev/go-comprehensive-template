package systems

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSystemInfo(t *testing.T) {
	info := SystemInfo()
	for _, key := range []string{"hostname", "pid", "uid", "os", "arch", "cpus", "cwd"} {
		if info[key] == "" {
			t.Errorf("SystemInfo()[%q] is empty", key)
		}
	}
}

func TestAtomicWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.txt")
	data := []byte("hello atomic")
	if err := AtomicWrite(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(data) {
		t.Errorf("got %q, want %q", got, data)
	}
}

func TestAtomicWrite_BadDir(t *testing.T) {
	err := AtomicWrite("/nonexistent/dir/file.txt", []byte("x"), 0o644)
	if err == nil {
		t.Fatal("expected error for bad directory")
	}
}

func TestReadLines(t *testing.T) {
	path := filepath.Join(t.TempDir(), "lines.txt")
	_ = os.WriteFile(path, []byte("a\nb\nc\n"), 0o644)
	var lines []string
	err := ReadLines(path, func(line string) error {
		lines = append(lines, line)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 3 || lines[0] != "a" || lines[2] != "c" {
		t.Errorf("got %v", lines)
	}
}

func TestReadLines_Error(t *testing.T) {
	err := ReadLines("/nonexistent/file", func(string) error { return nil })
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestReadLines_CallbackError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "lines.txt")
	_ = os.WriteFile(path, []byte("a\nb\n"), 0o644)
	stop := fmt.Errorf("stop")
	err := ReadLines(path, func(string) error { return stop })
	if err != stop {
		t.Errorf("got %v, want %v", err, stop)
	}
}

func TestTCPServerAndSend(t *testing.T) {
	// Use port 0 via a temporary listener to find a free port
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	ln.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = TCPServer(ctx, addr, func(conn net.Conn) {
			data, _ := io.ReadAll(conn)
			_, _ = conn.Write([]byte("echo:" + string(data)))
		})
	}()
	time.Sleep(100 * time.Millisecond)

	resp, err := TCPSend(addr, []byte("hello"))
	if err != nil {
		t.Fatal(err)
	}
	if string(resp) != "echo:hello" {
		t.Errorf("got %q", resp)
	}
	cancel()
}

func TestUDPServerAndSend(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	ln.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = UDPServer(ctx, addr, func(data []byte, from *net.UDPAddr, conn *net.UDPConn) {
			_, _ = conn.WriteToUDP(append([]byte("echo:"), data...), from)
		})
	}()
	time.Sleep(100 * time.Millisecond)

	resp, err := UDPSend(addr, []byte("hi"))
	if err != nil {
		t.Fatal(err)
	}
	if string(resp) != "echo:hi" {
		t.Errorf("got %q", resp)
	}
	cancel()
}
