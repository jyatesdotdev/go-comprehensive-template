// Package systems provides systems programming utilities: networking, file I/O, and OS interaction.
package systems

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

// SystemInfo returns basic OS/process information.
func SystemInfo() map[string]string {
	hostname, _ := os.Hostname()
	wd, _ := os.Getwd()
	return map[string]string{
		"hostname": hostname,
		"pid":      fmt.Sprintf("%d", os.Getpid()),
		"uid":      fmt.Sprintf("%d", os.Getuid()),
		"os":       runtime.GOOS,
		"arch":     runtime.GOARCH,
		"cpus":     fmt.Sprintf("%d", runtime.NumCPU()),
		"cwd":      wd,
	}
}

// AtomicWrite writes data to a file atomically by writing to a temp file then renaming.
func AtomicWrite(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}
	tmpName := tmp.Name()
	defer func() {
		if err != nil {
			os.Remove(tmpName)
		}
	}()
	if _, err = tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("write: %w", err)
	}
	if err = tmp.Chmod(perm); err != nil {
		tmp.Close()
		return fmt.Errorf("chmod: %w", err)
	}
	if err = tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("sync: %w", err)
	}
	if err = tmp.Close(); err != nil {
		return fmt.Errorf("close: %w", err)
	}
	return os.Rename(tmpName, path)
}

// ReadLines streams lines from a file, calling fn for each. Stops if fn returns an error.
func ReadLines(path string, fn func(line string) error) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if err := fn(scanner.Text()); err != nil {
			return err
		}
	}
	return scanner.Err()
}

// TCPServer listens on addr and handles each connection with handler.
// Shuts down when ctx is cancelled.
func TCPServer(ctx context.Context, addr string, handler func(net.Conn)) error {
	lc := net.ListenConfig{}
	ln, err := lc.Listen(ctx, "tcp", addr)
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	go func() {
		<-ctx.Done()
		ln.Close()
	}()
	for {
		conn, err := ln.Accept()
		if err != nil {
			if ctx.Err() != nil {
				break
			}
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer conn.Close()
			handler(conn)
		}()
	}
	wg.Wait()
	return nil
}

// TCPSend dials addr, writes data, and returns the response.
func TCPSend(addr string, data []byte) ([]byte, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	if _, err := conn.Write(data); err != nil {
		return nil, err
	}
	conn.(*net.TCPConn).CloseWrite()
	return io.ReadAll(conn)
}

// UDPServer listens on addr and calls handler for each packet.
func UDPServer(ctx context.Context, addr string, handler func(data []byte, from *net.UDPAddr, conn *net.UDPConn)) error {
	uaddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}
	conn, err := net.ListenUDP("udp", uaddr)
	if err != nil {
		return err
	}
	go func() {
		<-ctx.Done()
		conn.Close()
	}()
	buf := make([]byte, 65535)
	for {
		n, remote, err := conn.ReadFromUDP(buf)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			continue
		}
		pkt := make([]byte, n)
		copy(pkt, buf[:n])
		handler(pkt, remote, conn)
	}
}

// UDPSend sends a single UDP packet and returns the response.
func UDPSend(addr string, data []byte) ([]byte, error) {
	conn, err := net.Dial("udp", addr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	if _, err := conn.Write(data); err != nil {
		return nil, err
	}
	buf := make([]byte, 65535)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}
