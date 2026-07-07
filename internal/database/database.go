package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Database wraps a pgxpool.Pool and provides a stable interface for database
// operations. Callers depend on *Database rather than *pgxpool.Pool directly,
// which allows Health(), Close(), Stats(), and future transactional helpers to
// be added without modifying any call sites.
type Database struct {
	// Pool is the underlying connection pool. Exposed for use by repositories
	// that need direct query access.
	Pool *pgxpool.Pool
}

// New wraps an already-created pool in a Database instance.
func New(pool *pgxpool.Pool) *Database {
	return &Database{Pool: pool}
}

// Health pings the database and returns an error if the database is
// unreachable. It is safe to call from a health-check endpoint.
func (d *Database) Health(ctx context.Context) error {
	if err := d.Pool.Ping(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}
	return nil
}

// Close releases all connections in the pool. It should be called during
// graceful shutdown after all in-flight queries have completed.
func (d *Database) Close() {
	d.Pool.Close()
}

// Stats returns a snapshot of the pool's current connection statistics.
func (d *Database) Stats() *pgxpool.Stat {
	return d.Pool.Stat()
}
