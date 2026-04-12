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
	if err := systems.AtomicWrite(path, data, 0644); err != nil {
		fmt.Printf("  error: %v\n", err)
		return
	}
	content, _ := os.ReadFile(path)
	fmt.Printf("  wrote and read back: %s", content)
	os.Remove(path)
}

func demoReadLines() {
	fmt.Println("=== Streaming Line Reader ===")
	path := filepath.Join(os.TempDir(), "lines-demo.txt")
	os.WriteFile(path, []byte("line 1\nline 2\nline 3\n"), 0644)
	defer os.Remove(path)
	systems.ReadLines(path, func(line string) error {
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
		go func() { <-ctx.Done(); ln.Close() }()
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func() {
				defer conn.Close()
				io.Copy(conn, conn) // echo
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
		defer pc.Close()
		buf := make([]byte, 1024)
		for {
			n, remote, err := pc.ReadFrom(buf)
			if err != nil {
				return
			}
			pc.WriteTo(buf[:n], remote)
		}
	}()
	go func() { <-ctx.Done(); pc.Close() }()

	resp, err := systems.UDPSend(addr, []byte("hello UDP"))
	if err != nil {
		fmt.Printf("  error: %v\n", err)
	} else {
		fmt.Printf("  sent 'hello UDP', got back: %s\n", resp)
	}
}
