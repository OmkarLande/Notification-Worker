package config

import (
	"fmt"
	"time"
)

// DatabaseConfig holds PostgreSQL connection and connection-pool settings.
type DatabaseConfig struct {
	// URL is the full PostgreSQL connection string.
	// Example: postgres://user:pass@host:5432/dbname?sslmode=disable
	URL string

	// MaxConns is the maximum number of connections in the pool.
	MaxConns int32

	// MinConns is the minimum number of idle connections maintained in the pool.
	MinConns int32

	// MaxConnLifetime is the maximum duration a connection may be reused.
	MaxConnLifetime time.Duration

	// HealthTimeout is the timeout used when pinging the database during startup.
	HealthTimeout time.Duration
}

// loadDatabaseConfig reads database settings from environment variables.
func loadDatabaseConfig() (DatabaseConfig, error) {
	url := getEnv("DB_URL", "")
	if url == "" {
		return DatabaseConfig{}, fmt.Errorf("DB_URL is required but not set")
	}

	maxConns := int32(getEnvInt("DB_MAX_CONNS", 10))
	minConns := int32(getEnvInt("DB_MIN_CONNS", 2))
	lifetimeMin := getEnvInt("DB_MAX_CONN_LIFETIME_MINUTES", 60)
	healthTimeoutSec := getEnvInt("DB_HEALTH_TIMEOUT_SECONDS", 5)

	if maxConns < 1 {
		return DatabaseConfig{}, fmt.Errorf("DB_MAX_CONNS must be >= 1, got %d", maxConns)
	}
	if minConns < 0 || minConns > maxConns {
		return DatabaseConfig{}, fmt.Errorf("DB_MIN_CONNS must be between 0 and DB_MAX_CONNS (%d), got %d", maxConns, minConns)
	}

	return DatabaseConfig{
		URL:             url,
		MaxConns:        maxConns,
		MinConns:        minConns,
		MaxConnLifetime: time.Duration(lifetimeMin) * time.Minute,
		HealthTimeout:   time.Duration(healthTimeoutSec) * time.Second,
	}, nil
}
