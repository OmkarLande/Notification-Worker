// Package workers contains the task dispatch and execution infrastructure.
// Workers are generic — they contain no app-specific logic.
package workers

import (
	"context"
	"fmt"
	"sync"

	"github.com/OmkarLande/notification-worker/internal/cache"
	entities "github.com/OmkarLande/notification-worker/internal/entites"
	"github.com/OmkarLande/notification-worker/internal/interfaces"
	"github.com/OmkarLande/notification-worker/internal/logger"
)

// TaskDispatcher picks NeedToPick tasks, builds ExecutionContexts (loading Job
// and App once per distinct ID within a dispatch cycle), and fans them out to
// TaskWorker goroutines up to maxWorkers concurrently.
//
// The Dispatcher has no knowledge of job semantics. It only orchestrates the
// lifecycle handoff from the database queue to the worker pool.
type TaskDispatcher struct {
	taskRepo    interfaces.TaskRepository
	jobRepo     interfaces.JobRepository
	appRepo     interfaces.AppRepository
	taskWorker  *TaskWorker
	statusCache *cache.StatusCache
	log         logger.Logger
}

// NewTaskDispatcher constructs a TaskDispatcher.
func NewTaskDispatcher(
	taskRepo interfaces.TaskRepository,
	jobRepo interfaces.JobRepository,
	appRepo interfaces.AppRepository,
	taskWorker *TaskWorker,
	statusCache *cache.StatusCache,
	log logger.Logger,
) *TaskDispatcher {
	return &TaskDispatcher{
		taskRepo:    taskRepo,
		jobRepo:     jobRepo,
		appRepo:     appRepo,
		taskWorker:  taskWorker,
		statusCache: statusCache,
		log:         log,
	}
}

// Run performs one dispatch cycle: fetches NeedToPick tasks, resolves their
// context (Job + App), and processes them concurrently up to maxWorkers.
// It is named Run to signal that it will become a continuous polling loop in
// a future phase.
func (d *TaskDispatcher) Run(ctx context.Context, maxWorkers int) error {
	if maxWorkers < 1 {
		maxWorkers = 1
	}

	statusID := d.statusCache.TaskStatusID("NeedToPick")
	tasks, err := d.taskRepo.GetByStatus(ctx, statusID, maxWorkers*2)
	if err != nil {
		return fmt.Errorf("dispatcher: fetch tasks: %w", err)
	}

	if len(tasks) == 0 {
		d.log.Debug("Dispatcher: no tasks to process")
		return nil
	}

	d.log.Info("Dispatcher: tasks fetched", "count", len(tasks))

	// Build ExecutionContexts — load each distinct Job and App only once.
	jobCache := make(map[int]*entities.Job)
	appCache := make(map[int]*entities.App)

	contexts := make([]entities.ExecutionContext, 0, len(tasks))
	for i := range tasks {
		job, ok := jobCache[tasks[i].JobID]
		if !ok {
			job, err = d.jobRepo.GetByID(ctx, tasks[i].JobID)
			if err != nil {
				d.log.Error("Dispatcher: failed to load job", "job_id", tasks[i].JobID, "error", err)
				continue
			}
			jobCache[tasks[i].JobID] = job
		}

		app, ok := appCache[job.AppID]
		if !ok {
			app, err = d.appRepo.GetByID(ctx, job.AppID)
			if err != nil {
				d.log.Error("Dispatcher: failed to load app", "app_id", job.AppID, "error", err)
				continue
			}
			appCache[job.AppID] = app
		}

		contexts = append(contexts, entities.ExecutionContext{
			Task: &tasks[i],
			Job:  job,
			App:  app,
		})
	}

	// Fan out to TaskWorker goroutines using a semaphore for concurrency control.
	sem := make(chan struct{}, maxWorkers)
	var wg sync.WaitGroup

	for _, ec := range contexts {
		wg.Add(1)
		go func(ec entities.ExecutionContext) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			if err := d.taskWorker.Process(ctx, ec); err != nil {
				d.log.Error("Dispatcher: task processing error",
					"task_id", ec.Task.ID, "error", err)
			}
		}(ec)
	}

	wg.Wait()
	d.log.Info("Dispatcher: cycle complete", "processed", len(contexts))
	return nil
}
