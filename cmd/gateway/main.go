package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/fadeedan/pipo_hse_2026/internal/platform/config"
	"github.com/fadeedan/pipo_hse_2026/internal/platform/httpx"
	"github.com/fadeedan/pipo_hse_2026/internal/platform/logging"
	"github.com/fadeedan/pipo_hse_2026/internal/platform/metrics"
)

func main() {
	cfg := config.LoadGateway()
	logger := logging.New(cfg.LogLevel)

	authProxy, err := newProxy(cfg.AuthServiceURL, "/api/auth")
	if err != nil {
		logger.Error("auth proxy init failed", "err", err)
		os.Exit(1)
	}
	taskProxy, err := newProxy(cfg.TaskServiceURL, "/api/tasks")
	if err != nil {
		logger.Error("task proxy init failed", "err", err)
		os.Exit(1)
	}

	webDir := getenv("WEB_DIR", "web")
	apiSpecPath := getenv("OPENAPI_PATH", filepath.Join("api", "openapi.yaml"))

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.Handle("/api/auth/", authProxy)
	mux.Handle("/api/tasks/", taskProxy)

	mux.HandleFunc("/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, apiSpecPath)
	})
	mux.HandleFunc("/docs", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(webDir, "docs.html"))
	})

	fs := http.FileServer(http.Dir(webDir))
	mux.Handle("/", fs)

	m := metrics.New(cfg.AppName)
	router := m.Middleware(mux)
	router = httpx.RequestLogger(logger, cfg.AppName)(router)
	router = httpx.Recover(logger)(router)

	rootMux := http.NewServeMux()
	rootMux.Handle("/metrics", m.Handler())
	rootMux.Handle("/", router)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           rootMux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("gateway started", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", "err", err)
			os.Exit(1)
		}
	}()

	shutdownWithSignal(logger, srv)
}

func newProxy(targetURL, prefix string) (http.Handler, error) {
	target, err := url.Parse(targetURL)
	if err != nil {
		return nil, fmt.Errorf("parse target url %q: %w", targetURL, err)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	defaultDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		defaultDirector(req)
		trimmedPath := strings.TrimPrefix(req.URL.Path, prefix)
		if trimmedPath == "" {
			trimmedPath = "/"
		}
		req.URL.Path = joinPath(target.Path, trimmedPath)
		req.Host = target.Host
	}
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		httpx.WriteError(w, http.StatusBadGateway, "upstream is unavailable")
	}
	return proxy, nil
}

func joinPath(base, req string) string {
	switch {
	case strings.HasSuffix(base, "/") && strings.HasPrefix(req, "/"):
		return base + strings.TrimPrefix(req, "/")
	case !strings.HasSuffix(base, "/") && !strings.HasPrefix(req, "/"):
		return base + "/" + req
	default:
		return base + req
	}
}

func shutdownWithSignal(logger *slog.Logger, srv *http.Server) {
	sigCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	<-sigCtx.Done()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed", "err", err)
		return
	}
	logger.Info("gateway stopped")
}

func getenv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return def
}
