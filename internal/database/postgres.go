// Package database provides the PostgreSQL connection infrastructure for the
// Notification Worker. It wraps pgxpool.Pool in a Database struct so callers
// can rely on a stable, versioned interface rather than the pool directly.
package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/OmkarLande/notification-worker/internal/config"
	"github.com/OmkarLande/notification-worker/internal/logger"
)

// NewPool creates and validates a pgxpool.Pool using the provided
// DatabaseConfig. It applies all pool-tuning parameters from config and pings
// the database to confirm connectivity before returning. If the ping fails the
// pool is closed and an error is returned so the application fails fast.
func NewPool(cfg config.DatabaseConfig, log logger.Logger) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("database: failed to parse connection string: %w", err)
	}

	poolCfg.MaxConns = cfg.MaxConns
	poolCfg.MinConns = cfg.MinConns
	poolCfg.MaxConnLifetime = cfg.MaxConnLifetime

	log.Info("Connecting to PostgreSQL...",
		"max_conns", cfg.MaxConns,
		"min_conns", cfg.MinConns,
		"max_conn_lifetime", cfg.MaxConnLifetime,
	)

	pool, err := pgxpool.NewWithConfig(context.Background(), poolCfg)
	if err != nil {
		return nil, fmt.Errorf("database: failed to create connection pool: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.HealthTimeout)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("database: ping failed (timeout=%s): %w", cfg.HealthTimeout, err)
	}

	log.Info("PostgreSQL connected successfully",
		"max_conns", cfg.MaxConns,
	)

	return pool, nil
}