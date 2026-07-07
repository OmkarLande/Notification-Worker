package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	entities "github.com/OmkarLande/notification-worker/internal/entites"
)

// PostgresTaskRepository implements interfaces.TaskRepository.
type PostgresTaskRepository struct {
	pool *pgxpool.Pool
}

// NewTaskRepository constructs a PostgresTaskRepository.
func NewTaskRepository(pool *pgxpool.Pool) *PostgresTaskRepository {
	return &PostgresTaskRepository{pool: pool}
}

// Create inserts a new task row and returns the persisted entity with its ID.
func (r *PostgresTaskRepository) Create(ctx context.Context, t *entities.Task) (*entities.Task, error) {
	const q = `
		INSERT INTO tasks (job_id, parent_task_id, status_id, arguments, task_trigger_time, current_retry_count)
		VALUES ($1, $2, $3, $4, $5, 0)
		RETURNING id, created_at, updated_at`

	row := r.pool.QueryRow(ctx, q,
		t.JobID, t.ParentTaskID, t.StatusID, t.Arguments, t.TaskTriggerTime,
	)
	if err := row.Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt); err != nil {
		return nil, fmt.Errorf("task repository: Create: %w", err)
	}
	return t, nil
}

// GetByStatus returns up to limit tasks with the given statusID, oldest first.
func (r *PostgresTaskRepository) GetByStatus(ctx context.Context, statusID, limit int) ([]entities.Task, error) {
	const q = `
		SELECT id, job_id, parent_task_id, status_id,
		       COALESCE(arguments,'{}')::text,
		       task_trigger_time, task_start_time, task_end_time,
		       current_retry_count, created_at, updated_at
		FROM tasks
		WHERE status_id = $1
		ORDER BY task_trigger_time ASC
		LIMIT $2`

	rows, err := r.pool.Query(ctx, q, statusID, limit)
	if err != nil {
		return nil, fmt.Errorf("task repository: GetByStatus: %w", err)
	}
	defer rows.Close()

	var tasks []entities.Task
	for rows.Next() {
		var t entities.Task
		var args []byte
		if err := rows.Scan(
			&t.ID, &t.JobID, &t.ParentTaskID, &t.StatusID, &args,
			&t.TaskTriggerTime, &t.TaskStartTime, &t.TaskEndTime,
			&t.CurrentRetryCount, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("task repository: GetByStatus scan: %w", err)
		}
		t.Arguments = args
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

// UpdateStatus sets the status_id of a task.
func (r *PostgresTaskRepository) UpdateStatus(ctx context.Context, taskID, statusID int) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE tasks SET status_id = $1, updated_at = NOW() WHERE id = $2`,
		statusID, taskID,
	)
	if err != nil {
		return fmt.Errorf("task repository: UpdateStatus(task=%d, status=%d): %w", taskID, statusID, err)
	}
	return nil
}

// UpdateStartTime records when processing of a task began.
func (r *PostgresTaskRepository) UpdateStartTime(ctx context.Context, taskID int, t time.Time) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE tasks SET task_start_time = $1, updated_at = NOW() WHERE id = $2`,
		t, taskID,
	)
	if err != nil {
		return fmt.Errorf("task repository: UpdateStartTime(task=%d): %w", taskID, err)
	}
	return nil
}

// UpdateEndTime records when processing of a task finished.
func (r *PostgresTaskRepository) UpdateEndTime(ctx context.Context, taskID int, t time.Time) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE tasks SET task_end_time = $1, updated_at = NOW() WHERE id = $2`,
		t, taskID,
	)
	if err != nil {
		return fmt.Errorf("task repository: UpdateEndTime(task=%d): %w", taskID, err)
	}
	return nil
}