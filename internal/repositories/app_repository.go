package repositories

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	entities "github.com/OmkarLande/notification-worker/internal/entites"
)

// PostgresAppRepository implements interfaces.AppRepository.
type PostgresAppRepository struct {
	pool *pgxpool.Pool
}

// NewAppRepository constructs a PostgresAppRepository.
func NewAppRepository(pool *pgxpool.Pool) *PostgresAppRepository {
	return &PostgresAppRepository{pool: pool}
}

// GetByID loads an app by its primary key.
func (r *PostgresAppRepository) GetByID(ctx context.Context, id int) (*entities.App, error) {
	const q = `
		SELECT id, name, base_url,
		       COALESCE(connection_string,''), COALESCE(database_name,''),
		       maintenance_mode, created_at, updated_at
		FROM apps WHERE id = $1`

	a := &entities.App{}
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&a.ID, &a.Name, &a.BaseURL,
		&a.ConnectionString, &a.DatabaseName,
		&a.MaintenanceMode, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("app repository: GetByID(%d): %w", id, err)
	}
	return a, nil
}
