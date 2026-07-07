package repositories

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	entities "github.com/OmkarLande/notification-worker/internal/entites"
)

// PostgresTaskLogRepository implements interfaces.TaskLogRepository.
type PostgresTaskLogRepository struct {
	pool *pgxpool.Pool
}

// NewTaskLogRepository constructs a PostgresTaskLogRepository.
func NewTaskLogRepository(pool *pgxpool.Pool) *PostgresTaskLogRepository {
	return &PostgresTaskLogRepository{pool: pool}
}

// Create inserts a task_log record. Called only when a task step fails.
func (r *PostgresTaskLogRepository) Create(ctx context.Context, log *entities.TaskLog) error {
	perfLog := log.PerformanceLog
	if perfLog == nil {
		perfLog, _ = json.Marshal(map[string]any{})
	}
	errLog := log.ErrorLog
	if errLog == nil {
		errLog, _ = json.Marshal(map[string]any{})
	}

	_, err := r.pool.Exec(ctx,
		`INSERT INTO task_logs (task_id, job_id, step_name, performance_log, error_log)
		 VALUES ($1, $2, $3, $4, $5)`,
		log.TaskID, log.JobID, log.StepName, perfLog, errLog,
	)
	if err != nil {
		return fmt.Errorf("task log repository: Create(task=%d): %w", log.TaskID, err)
	}
	return nil
}