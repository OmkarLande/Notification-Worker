// Package repositories provides PostgreSQL implementations of every repository
// interface defined in internal/interfaces. All queries use raw pgx — no ORM.
// Repositories contain no business logic; they only perform CRUD operations.
package repositories

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	entities "github.com/OmkarLande/notification-worker/internal/entites"
)

// PostgresJobRepository implements interfaces.JobRepository.
type PostgresJobRepository struct {
	pool *pgxpool.Pool
}

// NewJobRepository constructs a PostgresJobRepository backed by the given pool.
func NewJobRepository(pool *pgxpool.Pool) *PostgresJobRepository {
	return &PostgresJobRepository{pool: pool}
}

// GetByID loads a job by its primary key.
func (r *PostgresJobRepository) GetByID(ctx context.Context, id int) (*entities.Job, error) {
	const q = `
		SELECT id, app_id, name, COALESCE(description,''), status_id,
		       max_thread_count, max_retry_count, COALESCE(arguments,'{}')::text, created_at, updated_at
		FROM jobs WHERE id = $1`

	row := r.pool.QueryRow(ctx, q, id)
	j := &entities.Job{}
	var args []byte
	if err := row.Scan(
		&j.ID, &j.AppID, &j.Name, &j.Description,
		&j.StatusID, &j.MaxThreadCount, &j.MaxRetryCount, &args,
		&j.CreatedAt, &j.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("job repository: GetByID(%d): %w", id, err)
	}
	j.Arguments = args
	return j, nil
}

// GetActiveJobs returns all jobs with an Active status.
func (r *PostgresJobRepository) GetActiveJobs(ctx context.Context) ([]entities.Job, error) {
	const q = `
		SELECT j.id, j.app_id, j.name, COALESCE(j.description,''), j.status_id,
		       j.max_thread_count, j.max_retry_count, COALESCE(j.arguments,'{}')::text, j.created_at, j.updated_at
		FROM jobs j
		JOIN job_status js ON js.id = j.status_id
		WHERE js.name = 'Active'
		ORDER BY j.id`

	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("job repository: GetActiveJobs: %w", err)
	}
	defer rows.Close()

	var jobs []entities.Job
	for rows.Next() {
		var j entities.Job
		var args []byte
		if err := rows.Scan(
			&j.ID, &j.AppID, &j.Name, &j.Description,
			&j.StatusID, &j.MaxThreadCount, &j.MaxRetryCount, &args,
			&j.CreatedAt, &j.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("job repository: GetActiveJobs scan: %w", err)
		}
		j.Arguments = args
		jobs = append(jobs, j)
	}
	return jobs, rows.Err()
}