package reliability

import (
	"context"
	"fmt"
	"time"

	"github.com/OmkarLande/notification-worker/internal/cache"
	entities "github.com/OmkarLande/notification-worker/internal/entites"
	"github.com/OmkarLande/notification-worker/internal/interfaces"
	"github.com/OmkarLande/notification-worker/internal/logger"
	"github.com/OmkarLande/notification-worker/internal/pipeline"
)

// RetryService is responsible for duplicating failed tasks for a retry
// while preserving complete execution lineage.
type RetryService interface {
	ScheduleRetry(ctx context.Context, execution *pipeline.ExecutionContext, failure FailureContext) error
}

type retryServiceImpl struct {
	taskRepo        interfaces.TaskRepository
	channelTaskRepo interfaces.ChannelTaskRepository
	statusCache     *cache.StatusCache
	log             logger.Logger
}

// NewRetryService constructs a RetryService.
func NewRetryService(
	taskRepo interfaces.TaskRepository,
	channelTaskRepo interfaces.ChannelTaskRepository,
	statusCache *cache.StatusCache,
	log logger.Logger,
) RetryService {
	return &retryServiceImpl{
		taskRepo:        taskRepo,
		channelTaskRepo: channelTaskRepo,
		statusCache:     statusCache,
		log:             log,
	}
}

func (s *retryServiceImpl) ScheduleRetry(ctx context.Context, execution *pipeline.ExecutionContext, failure FailureContext) error {
	job := execution.Job
	task := execution.Task

	if task.CurrentRetryCount >= job.MaxRetryCount {
		s.log.Warn("Retry limit reached", "task_id", task.ID, "job_id", job.ID, "max_retries", job.MaxRetryCount)
		return nil // Not an error, just no retry scheduled.
	}

	needToPickStatus := s.statusCache.TaskStatusID("NeedToPick")

	// Create a new Task row pointing to the old task as its parent.
	newTask := &entities.Task{
		JobID:             job.ID,
		ParentTaskID:      &task.ID,
		StatusID:          needToPickStatus,
		Arguments:         task.Arguments, // Exact copy
		TaskTriggerTime:   time.Now(),
		CurrentRetryCount: task.CurrentRetryCount + 1,
	}

	createdTask, err := s.taskRepo.Create(ctx, newTask)
	if err != nil {
		return fmt.Errorf("retry service: failed to create retry task: %w", err)
	}

	// Recreate ChannelTasks for the new Task to ensure delivery channels are attached.
	// Currently we blindly recreate ALL active channels that were attached to the original task.
	// Future phase: only attach channels that failed.
	channels, err := s.channelTaskRepo.GetChannelsByTaskID(ctx, task.ID)
	if err != nil {
		return fmt.Errorf("retry service: failed to fetch original channels: %w", err)
	}

	for _, ch := range channels {
		if err := s.channelTaskRepo.Create(ctx, createdTask.ID, ch.ID, needToPickStatus); err != nil {
			s.log.Error("Retry service: failed to attach channel to retry task", 
				"task_id", createdTask.ID, 
				"channel_id", ch.ID, 
				"error", err,
			)
			// Proceeding with other channels despite error
		}
	}

	s.log.Info("Retry Scheduled",
		"original_task_id", task.ID,
		"new_task_id", createdTask.ID,
		"job_id", job.ID,
		"retry_count", createdTask.CurrentRetryCount,
	)

	return nil
}
