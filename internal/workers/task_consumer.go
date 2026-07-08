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
	"github.com/OmkarLande/notification-worker/internal/pipeline"
	"github.com/OmkarLande/notification-worker/internal/providers"
)

// TaskWorker executes a single task by orchestrating the execution pipeline.
// It is fully stateless and generic — it receives a complete ExecutionContext
// from the Dispatcher, resolves the provider, builds the pipeline context,
// and delegates execution to the pipeline.
type TaskWorker struct {
	taskRepo    interfaces.TaskRepository
	factory     *providers.Factory
	statusCache *cache.StatusCache
	pipeline    *pipeline.Pipeline
	log         logger.Logger
}

// NewTaskWorker constructs a TaskWorker.
func NewTaskWorker(
	taskRepo interfaces.TaskRepository,
	factory *providers.Factory,
	statusCache *cache.StatusCache,
	execPipeline *pipeline.Pipeline,
	log logger.Logger,
) *TaskWorker {
	return &TaskWorker{
		taskRepo:    taskRepo,
		factory:     factory,
		statusCache: statusCache,
		pipeline:    execPipeline,
		log:         log,
	}
}

// Process executes one task through its complete lifecycle:
//
//	NeedToPick → Picked → Processing → Completed
//	                              └──────────────→ Failed (on error)
//
// Execution is entirely delegated to the pipeline.
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

	// Extract user_id from task arguments for logging context.
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

	// 4. Build Pipeline Context and Run Pipeline
	pCtx := &pipeline.ExecutionContext{
		Task:     ec.Task,
		Job:      ec.Job,
		App:      ec.App,
		Provider: provider,
		Metadata: make(map[string]any),
	}

	res, err := w.pipeline.Run(ctx, pCtx)
	endTime := time.Now()
	_ = w.taskRepo.UpdateEndTime(ctx, taskID, endTime)

	if err != nil || !res.Success {
		return w.failTask(ctx, ec, "pipeline_execution", err)
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
		"duration", res.Duration,
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

	return fmt.Errorf("task worker [task=%d step=%s]: %w", taskID, step, cause)
}