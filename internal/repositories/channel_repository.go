package repositories

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	entities "github.com/OmkarLande/notification-worker/internal/entites"
)

// PostgresChannelRepository implements interfaces.ChannelRepository.
type PostgresChannelRepository struct {
	pool *pgxpool.Pool
}

// NewChannelRepository constructs a PostgresChannelRepository.
func NewChannelRepository(pool *pgxpool.Pool) *PostgresChannelRepository {
	return &PostgresChannelRepository{pool: pool}
}

// GetAll returns all active channels ordered by ID.
func (r *PostgresChannelRepository) GetAll(ctx context.Context) ([]entities.Channel, error) {
	const q = `
		SELECT id, name, is_active, created_at, updated_at
		FROM channels WHERE is_active = TRUE ORDER BY id`

	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("channel repository: GetAll: %w", err)
	}
	defer rows.Close()

	var channels []entities.Channel
	for rows.Next() {
		var c entities.Channel
		if err := rows.Scan(&c.ID, &c.Name, &c.IsActive, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("channel repository: GetAll scan: %w", err)
		}
		channels = append(channels, c)
	}
	return channels, rows.Err()
}
