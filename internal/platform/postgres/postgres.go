package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	maxConnectAttempts = 30
	retryDelay         = 2 * time.Second
	pingTimeout        = 5 * time.Second
)

func Open(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	var lastErr error
	for attempt := 1; attempt <= maxConnectAttempts; attempt++ {
		pool, err := openOnce(ctx, dsn)
		if err == nil {
			return pool, nil
		}
		lastErr = err

		if attempt == maxConnectAttempts {
			break
		}
		if err := sleepWithContext(ctx, retryDelay); err != nil {
			return nil, fmt.Errorf("database connect canceled after %d attempt(s): %w", attempt, err)
		}
	}

	return nil, fmt.Errorf("database connection failed after %d attempt(s): %w", maxConnectAttempts, lastErr)
}

func openOnce(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}
	cfg.MaxConns = 10
	cfg.MinConns = 1
	cfg.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, pingTimeout)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return pool, nil
}

func sleepWithContext(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
