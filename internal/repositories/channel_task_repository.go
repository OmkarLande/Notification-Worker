package repositories

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresChannelTaskRepository implements interfaces.ChannelTaskRepository.
type PostgresChannelTaskRepository struct {
	pool *pgxpool.Pool
}

// NewChannelTaskRepository constructs a PostgresChannelTaskRepository.
func NewChannelTaskRepository(pool *pgxpool.Pool) *PostgresChannelTaskRepository {
	return &PostgresChannelTaskRepository{pool: pool}
}

// Create inserts a channel_tasks row linking a task to a delivery channel with
// the given initial status. Called by JobExecutionService when creating tasks.
func (r *PostgresChannelTaskRepository) Create(ctx context.Context, taskID, channelID, statusID int) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO channel_tasks (task_id, channel_id, status_id) VALUES ($1, $2, $3)`,
		taskID, channelID, statusID,
	)
	if err != nil {
		return fmt.Errorf("channel task repository: Create(task=%d, channel=%d): %w", taskID, channelID, err)
	}
	return nil
}
