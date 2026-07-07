// Package cache provides in-memory caches populated once at application startup.
// Caching lookup tables like status IDs avoids repeated database queries during
// high-frequency task execution.
package cache

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// StatusCache holds the ID mappings for every status lookup table.
// It is loaded once during startup and is read-only thereafter — safe for
// concurrent access without locks.
type StatusCache struct {
	taskStatus        map[string]int
	jobStatus         map[string]int
	channelTaskStatus map[string]int
}

// Load queries the three status tables and populates a StatusCache.
// Returns an error if any table cannot be read.
func Load(ctx context.Context, pool *pgxpool.Pool) (*StatusCache, error) {
	c := &StatusCache{
		taskStatus:        make(map[string]int),
		jobStatus:         make(map[string]int),
		channelTaskStatus: make(map[string]int),
	}

	if err := loadTable(ctx, pool, "task_status", c.taskStatus); err != nil {
		return nil, fmt.Errorf("cache: task_status: %w", err)
	}
	if err := loadTable(ctx, pool, "job_status", c.jobStatus); err != nil {
		return nil, fmt.Errorf("cache: job_status: %w", err)
	}
	if err := loadTable(ctx, pool, "channel_task_status", c.channelTaskStatus); err != nil {
		return nil, fmt.Errorf("cache: channel_task_status: %w", err)
	}

	return c, nil
}

// loadTable reads all rows from a status table into the given map.
func loadTable(ctx context.Context, pool *pgxpool.Pool, table string, dest map[string]int) error {
	rows, err := pool.Query(ctx, "SELECT id, name FROM "+table)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return err
		}
		dest[name] = id
	}
	return rows.Err()
}

// TaskStatusID returns the ID for the given task status name.
// Panics if the name is not in the cache — this indicates a startup bug,
// not a runtime error.
func (c *StatusCache) TaskStatusID(name string) int {
	id, ok := c.taskStatus[name]
	if !ok {
		panic(fmt.Sprintf("cache: unknown task status %q — check seed data", name))
	}
	return id
}

// JobStatusID returns the ID for the given job status name.
func (c *StatusCache) JobStatusID(name string) int {
	id, ok := c.jobStatus[name]
	if !ok {
		panic(fmt.Sprintf("cache: unknown job status %q — check seed data", name))
	}
	return id
}

// ChannelTaskStatusID returns the ID for the given channel task status name.
func (c *StatusCache) ChannelTaskStatusID(name string) int {
	id, ok := c.channelTaskStatus[name]
	if !ok {
		panic(fmt.Sprintf("cache: unknown channel task status %q — check seed data", name))
	}
	return id
}

// TaskStatuses returns a copy of the full task status map for logging.
func (c *StatusCache) TaskStatuses() map[string]int {
	out := make(map[string]int, len(c.taskStatus))
	for k, v := range c.taskStatus {
		out[k] = v
	}
	return out
}
