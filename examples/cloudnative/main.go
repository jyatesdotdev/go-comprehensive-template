// Example cloudnative demonstrates config, health checks, structured logging, and graceful shutdown.
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/example/go-template/pkg/cloudnative"
)

func main() {
	cfg := cloudnative.LoadConfig()
	logger := cloudnative.NewLogger(cfg.LogLevel)

	// Health checks
	health := cloudnative.NewHealthChecker()
	health.AddCheck("self", func() error { return nil })

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", health.LivenessHandler())
	mux.HandleFunc("GET /readyz", health.ReadinessHandler())
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "env=%s\n", cfg.Environment)
	})

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      cloudnative.RequestLogger(logger, mux),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("server starting", "port", cfg.Port, "env", cfg.Environment)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			logger.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down")
	shutCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	srv.Shutdown(shutCtx)
	logger.Info("stopped")
}
