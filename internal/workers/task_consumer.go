package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/OmkarLande/notification-worker/internal/cache"
	entities "github.com/OmkarLande/notification-worker/internal/entites"
	"github.com/OmkarLande/notification-worker/internal/interfaces"
	"github.com/OmkarLande/notification-worker/internal/logger"
	"github.com/OmkarLande/notification-worker/internal/providers"
)

// TaskWorker executes a single task. It is fully stateless — it receives a
// complete ExecutionContext from the Dispatcher so it never queries repositories
// during execution. App-specific behavior lives entirely in the provider.
type TaskWorker struct {
	taskRepo    interfaces.TaskRepository
	taskLogRepo interfaces.TaskLogRepository
	factory     *providers.Factory
	statusCache *cache.StatusCache
	log         logger.Logger
}

// NewTaskWorker constructs a TaskWorker.
func NewTaskWorker(
	taskRepo interfaces.TaskRepository,
	taskLogRepo interfaces.TaskLogRepository,
	factory *providers.Factory,
	statusCache *cache.StatusCache,
	log logger.Logger,
) *TaskWorker {
	return &TaskWorker{
		taskRepo:    taskRepo,
		taskLogRepo: taskLogRepo,
		factory:     factory,
		statusCache: statusCache,
		log:         log,
	}
}

// Process executes one task through its complete lifecycle:
//
//	NeedToPick → Picked → Processing → Completed
//	                              └──────────────→ Failed (on error)
//
// On failure, an error log is persisted to task_logs.
func (w *TaskWorker) Process(ctx context.Context, ec entities.ExecutionContext) error {
	taskID := ec.Task.ID

	// 1. Mark as Picked.
	if err := w.updateStatus(ctx, taskID, "Picked"); err != nil {
		return err
	}
	w.log.Info("Task Picked", "task_id", taskID, "job", ec.Job.Name, "app", ec.App.Name)

	// 2. Resolve provider (already initialized at startup).
	provider, err := w.factory.Get(ec.App.Name)
	if err != nil {
		return w.failTask(ctx, ec, "resolve_provider", err)
	}

	// 3. Mark as Processing and record start time.
	if err := w.updateStatus(ctx, taskID, "Processing"); err != nil {
		return err
	}
	startTime := time.Now()
	if err := w.taskRepo.UpdateStartTime(ctx, taskID, startTime); err != nil {
		w.log.Warn("Task: failed to record start time", "task_id", taskID, "error", err)
	}

	// Extract user_id from task arguments for logging.
	var args struct {
		UserID int `json:"user_id"`
	}
	_ = json.Unmarshal(ec.Task.Arguments, &args)

	w.log.Info("Task Processing",
		"task_id", taskID,
		"job", ec.Job.Name,
		"provider", ec.App.Name,
		"user_id", args.UserID,
	)

	// 4. Execute via provider — no app-specific logic here.
	output, err := provider.Execute(ctx, ec)
	endTime := time.Now()
	_ = w.taskRepo.UpdateEndTime(ctx, taskID, endTime)

	if err != nil {
		return w.failTask(ctx, ec, "provider_execute", err)
	}

	// 5. Mark as Completed.
	if err := w.updateStatus(ctx, taskID, "Completed"); err != nil {
		return err
	}

	w.log.Info("Task Completed",
		"task_id", taskID,
		"job", ec.Job.Name,
		"provider", ec.App.Name,
		"user_id", args.UserID,
		"duration", output.Duration,
	)

	return nil
}

// updateStatus is a helper that updates task status by name via the cache.
func (w *TaskWorker) updateStatus(ctx context.Context, taskID int, statusName string) error {
	statusID := w.statusCache.TaskStatusID(statusName)
	if err := w.taskRepo.UpdateStatus(ctx, taskID, statusID); err != nil {
		return fmt.Errorf("task worker: %w", err)
	}
	return nil
}

// failTask marks a task as Failed, persists an error log, and returns the
// original error wrapped with context.
func (w *TaskWorker) failTask(ctx context.Context, ec entities.ExecutionContext, step string, cause error) error {
	taskID := ec.Task.ID

	w.log.Error("Task Failed",
		"task_id", taskID, "step", step,
		"job", ec.Job.Name, "error", cause)

	_ = w.updateStatus(ctx, taskID, "Failed")

	errPayload, _ := json.Marshal(map[string]any{
		"step":    step,
		"message": cause.Error(),
	})
	_ = w.taskLogRepo.Create(ctx, &entities.TaskLog{
		TaskID:   taskID,
		JobID:    ec.Job.ID,
		StepName: step,
		ErrorLog: errPayload,
	})

	return fmt.Errorf("task worker [task=%d step=%s]: %w", taskID, step, cause)
}