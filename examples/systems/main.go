// Example systems demonstrates TCP/UDP networking, atomic file I/O, and system info.
package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/example/go-template/pkg/systems"
)

func main() {
	fmt.Println("=== System Info ===")
	for k, v := range systems.SystemInfo() {
		fmt.Printf("  %s: %s\n", k, v)
	}

	demoAtomicWrite()
	demoReadLines()
	demoTCPEcho()
	demoUDPEcho()
}

func demoAtomicWrite() {
	fmt.Println("\n=== Atomic File Write ===")
	path := filepath.Join(os.TempDir(), "atomic-demo.txt")
	data := []byte("written atomically — no partial reads possible\n")
	if err := systems.AtomicWrite(path, data, 0o644); err != nil {
		fmt.Printf("  error: %v\n", err)
		return
	}
	content, _ := os.ReadFile(path) // #nosec G104,G304 -- demo code
	fmt.Printf("  wrote and read back: %s", content)
	_ = os.Remove(path) // #nosec G104 -- demo cleanup
}

func demoReadLines() {
	fmt.Println("=== Streaming Line Reader ===")
	path := filepath.Join(os.TempDir(), "lines-demo.txt")
	_ = os.WriteFile(path, []byte("line 1\nline 2\nline 3\n"), 0o600) // #nosec G104 -- demo code
	defer func() { _ = os.Remove(path) }()                          // #nosec G104 -- demo cleanup
	_ = systems.ReadLines(path, func(line string) error {            // #nosec G104 -- demo code
		fmt.Printf("  > %s\n", line)
		return nil
	})
}

func demoTCPEcho() {
	fmt.Println("\n=== TCP Echo Server ===")
	ctx, cancel := context.WithCancel(context.Background())
	ready := make(chan string, 1)

	// Start server on random port
	go func() {
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			fmt.Printf("  listen error: %v\n", err)
			return
		}
		ready <- ln.Addr().String()
		go func() { <-ctx.Done(); _ = ln.Close() }() // #nosec G104 -- shutdown signal
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func() {
				defer conn.Close() //nolint:errcheck // best-effort close
				_, _ = io.Copy(conn, conn) // #nosec G104 -- echo server demo
			}()
		}
	}()

	addr := <-ready
	resp, err := systems.TCPSend(addr, []byte("hello TCP"))
	if err != nil {
		fmt.Printf("  error: %v\n", err)
	} else {
		fmt.Printf("  sent 'hello TCP', got back: %s\n", resp)
	}
	cancel()
	time.Sleep(10 * time.Millisecond)
}

func demoUDPEcho() {
	fmt.Println("\n=== UDP Echo Server ===")
	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	// Start UDP echo server on random port
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		fmt.Printf("  listen error: %v\n", err)
		return
	}
	addr := pc.LocalAddr().String()
	go func() {
		defer pc.Close() //nolint:errcheck // best-effort close
		buf := make([]byte, 1024)
		for {
			n, remote, err := pc.ReadFrom(buf)
			if err != nil {
				return
			}
			_, _ = pc.WriteTo(buf[:n], remote) // #nosec G104 -- echo server demo
		}
	}()
	go func() { <-ctx.Done(); _ = pc.Close() }() // #nosec G104 -- shutdown signal

	resp, err := systems.UDPSend(addr, []byte("hello UDP"))
	if err != nil {
		fmt.Printf("  error: %v\n", err)
	} else {
		fmt.Printf("  sent 'hello UDP', got back: %s\n", resp)
	}
}
