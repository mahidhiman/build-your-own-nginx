package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"example.com/tinyproxy/internal/config"
	"example.com/tinyproxy/internal/limiter"
	"example.com/tinyproxy/internal/proxy"
)

func main() {
	configPath := flag.String("config", "config.json", "path to JSON config")
	flag.Parse()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg, err := config.Load(*configPath)
	if err != nil {
		logger.Error("invalid configuration", "error", err)
		os.Exit(2)
	}
	handler, err := proxy.NewHandler(cfg.Upstreams, cfg.UpstreamTimeout())
	if err != nil {
		logger.Error("create proxy", "error", err)
		os.Exit(2)
	}
	limit, err := limiter.New(cfg.MaxInflight)
	if err != nil {
		logger.Error("create limiter", "error", err)
		os.Exit(2)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), cfg.UpstreamTimeout())
		defer cancel()
		if err := limit.Acquire(ctx); err != nil {
			http.Error(w, "proxy overloaded", http.StatusServiceUnavailable)
			return
		}
		defer limit.Release()
		start := time.Now()
		wrapped := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		handler.ServeHTTP(wrapped, r.WithContext(ctx))
		logger.Info("request completed", "method", r.Method, "path", r.URL.Path, "status", wrapped.status, "duration_ms", time.Since(start).Milliseconds(), "inflight", limit.Inflight(), "remote", r.RemoteAddr)
	}))

	srv := &http.Server{Addr: cfg.Listen, Handler: mux, ReadHeaderTimeout: cfg.ReadHeaderTimeout(), IdleTimeout: cfg.IdleTimeout(), MaxHeaderBytes: 1 << 20}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	errCh := make(chan error, 1)
	go func() {
		logger.Info("proxy listening", "addr", cfg.Listen, "upstreams", len(cfg.Upstreams))
		errCh <- srv.ListenAndServe()
	}()
	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Warn("graceful shutdown failed; closing", "error", err)
			_ = srv.Close()
		}
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server stopped", "error", err)
			os.Exit(1)
		}
	}
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// Unwrap allows http.ResponseController, including ReverseProxy's streaming
// logic, to reach optional interfaces implemented by the underlying writer.
func (w *statusWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }
