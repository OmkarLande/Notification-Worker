package services

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

// taskArgs is the JSON payload written into task.Arguments.
// The provider deserializes this at execution time.
type taskArgs struct {
	UserID int `json:"user_id"`
}

// JobExecutionService orchestrates the full job trigger pipeline:
// validate → resolve provider → fetch users → create tasks + channel tasks.
// It intentionally knows nothing about how tasks are executed.
type JobExecutionService struct {
	jobService      *JobService
	taskRepo        interfaces.TaskRepository
	jobChannelRepo  interfaces.JobChannelRepository
	channelTaskRepo interfaces.ChannelTaskRepository
	factory         *providers.Factory
	statusCache     *cache.StatusCache
	log             logger.Logger
}

// NewJobExecutionService constructs a JobExecutionService.
func NewJobExecutionService(
	jobService *JobService,
	taskRepo interfaces.TaskRepository,
	jobChannelRepo interfaces.JobChannelRepository,
	channelTaskRepo interfaces.ChannelTaskRepository,
	factory *providers.Factory,
	statusCache *cache.StatusCache,
	log logger.Logger,
) *JobExecutionService {
	return &JobExecutionService{
		jobService:      jobService,
		taskRepo:        taskRepo,
		jobChannelRepo:  jobChannelRepo,
		channelTaskRepo: channelTaskRepo,
		factory:         factory,
		statusCache:     statusCache,
		log:             log,
	}
}

// TriggerJob validates the job, fetches notification-enabled users from the
// provider, and creates one Task (with associated ChannelTasks) per user.
// It does not execute the tasks — that is the Dispatcher's responsibility.
func (s *JobExecutionService) TriggerJob(ctx context.Context, jobID int) error {
	s.log.Info("Triggering job", "job_id", jobID)

	// 1. Validate job and app.
	job, app, err := s.jobService.GetValidatedJob(ctx, jobID)
	if err != nil {
		return fmt.Errorf("job execution service: %w", err)
	}

	// 2. Resolve the already-initialized provider.
	provider, err := s.factory.Get(app.Name)
	if err != nil {
		return fmt.Errorf("job execution service: %w", err)
	}

	// 3. Fetch notification-enabled users from the provider.
	users, err := provider.GetNotificationEnabledUsers(ctx)
	if err != nil {
		return fmt.Errorf("job execution service: GetNotificationEnabledUsers: %w", err)
	}
	s.log.Info("Users loaded", "job_id", jobID, "count", len(users))

	// 4. Resolve channels assigned to this job.
	channels, err := s.jobChannelRepo.GetByJobID(ctx, jobID)
	if err != nil {
		return fmt.Errorf("job execution service: GetByJobID: %w", err)
	}

	needToPickID := s.statusCache.TaskStatusID("NeedToPick")
	pendingID := s.statusCache.ChannelTaskStatusID("Pending")

	// 5. Create one task per user, plus channel_task rows for each channel.
	for _, user := range users {
		args, err := json.Marshal(taskArgs{UserID: user.ID})
		if err != nil {
			return fmt.Errorf("job execution service: marshal args for user %d: %w", user.ID, err)
		}

		task, err := s.taskRepo.Create(ctx, &entities.Task{
			JobID:           job.ID,
			StatusID:        needToPickID,
			Arguments:       args,
			TaskTriggerTime: time.Now(),
		})
		if err != nil {
			return fmt.Errorf("job execution service: Create task for user %d: %w", user.ID, err)
		}

		s.log.Info("Task created", "task_id", task.ID, "job_id", job.ID, "user_id", user.ID)

		for _, ch := range channels {
			if err := s.channelTaskRepo.Create(ctx, task.ID, ch.ID, pendingID); err != nil {
				s.log.Warn("Failed to create channel task",
					"task_id", task.ID, "channel", ch.Name, "error", err)
			}
		}
	}

	s.log.Info("Job triggered successfully", "job_id", jobID, "tasks_created", len(users))
	return nil
}
