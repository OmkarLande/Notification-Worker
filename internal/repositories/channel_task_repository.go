package repositories

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	entities "github.com/OmkarLande/notification-worker/internal/entites"
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

// GetChannelsByTaskID returns all active channels assigned to a given task.
func (r *PostgresChannelTaskRepository) GetChannelsByTaskID(ctx context.Context, taskID int) ([]entities.Channel, error) {
	const q = `
		SELECT c.id, c.name, c.is_active, c.created_at, c.updated_at
		FROM channels c
		JOIN channel_tasks ct ON ct.channel_id = c.id
		WHERE ct.task_id = $1 AND c.is_active = true
		ORDER BY c.id`

	rows, err := r.pool.Query(ctx, q, taskID)
	if err != nil {
		return nil, fmt.Errorf("channel task repository: GetChannelsByTaskID(%d): %w", taskID, err)
	}
	defer rows.Close()

	var channels []entities.Channel
	for rows.Next() {
		var c entities.Channel
		if err := rows.Scan(&c.ID, &c.Name, &c.IsActive, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("channel task repository: GetChannelsByTaskID scan: %w", err)
		}
		channels = append(channels, c)
	}
	return channels, rows.Err()
}
