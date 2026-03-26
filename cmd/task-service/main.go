package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fadeedan/pipo_hse_2026/internal/db/migrations"
	"github.com/fadeedan/pipo_hse_2026/internal/platform/config"
	"github.com/fadeedan/pipo_hse_2026/internal/platform/httpx"
	"github.com/fadeedan/pipo_hse_2026/internal/platform/jwt"
	"github.com/fadeedan/pipo_hse_2026/internal/platform/logging"
	"github.com/fadeedan/pipo_hse_2026/internal/platform/metrics"
	"github.com/fadeedan/pipo_hse_2026/internal/platform/migrate"
	"github.com/fadeedan/pipo_hse_2026/internal/platform/postgres"
	"github.com/fadeedan/pipo_hse_2026/internal/tasks"
)

func main() {
	cfg := config.LoadTaskService()
	logger := logging.New(cfg.LogLevel)

	ctx := context.Background()
	pool, err := postgres.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("database connection failed", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := migrate.Apply(ctx, pool, migrations.Files); err != nil {
		logger.Error("migration failed", "err", err)
		os.Exit(1)
	}

	tokens := jwt.NewManager(cfg.JWTSecret, "auth-service", 24*time.Hour)
	repo := tasks.NewPostgresRepository(pool)
	service := tasks.NewService(repo)
	h := tasks.NewHandler(service, tokens, logger)

	m := metrics.New(cfg.AppName)
	router := h.Router(m.Handler())
	router = m.Middleware(router)
	router = httpx.RequestLogger(logger, cfg.AppName)(router)
	router = httpx.Recover(logger)(router)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("task service started", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", "err", err)
			os.Exit(1)
		}
	}()

	shutdownWithSignal(logger, srv)
}

func shutdownWithSignal(logger interface {
	Info(string, ...any)
	Error(string, ...any)
}, srv *http.Server) {
	sigCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	<-sigCtx.Done()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed", "err", err)
		return
	}
	logger.Info("task service stopped")
}
