package repositories

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	entities "github.com/OmkarLande/notification-worker/internal/entites"
)

// PostgresJobChannelRepository implements interfaces.JobChannelRepository.
type PostgresJobChannelRepository struct {
	pool *pgxpool.Pool
}

// NewJobChannelRepository constructs a PostgresJobChannelRepository.
func NewJobChannelRepository(pool *pgxpool.Pool) *PostgresJobChannelRepository {
	return &PostgresJobChannelRepository{pool: pool}
}

// GetByJobID returns all channels linked to the given job.
func (r *PostgresJobChannelRepository) GetByJobID(ctx context.Context, jobID int) ([]entities.Channel, error) {
	const q = `
		SELECT c.id, c.name, c.is_active, c.created_at, c.updated_at
		FROM channels c
		JOIN job_channels jc ON jc.channel_id = c.id
		WHERE jc.job_id = $1
		ORDER BY c.id`

	rows, err := r.pool.Query(ctx, q, jobID)
	if err != nil {
		return nil, fmt.Errorf("job channel repository: GetByJobID(%d): %w", jobID, err)
	}
	defer rows.Close()

	var channels []entities.Channel
	for rows.Next() {
		var c entities.Channel
		if err := rows.Scan(&c.ID, &c.Name, &c.IsActive, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("job channel repository: GetByJobID scan: %w", err)
		}
		channels = append(channels, c)
	}
	return channels, rows.Err()
}
