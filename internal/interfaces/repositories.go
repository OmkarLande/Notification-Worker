package interfaces

import (
	"context"
	"time"

	entities "github.com/OmkarLande/notification-worker/internal/entites"
)

// JobRepository provides read access to the jobs table.
type JobRepository interface {
	GetByID(ctx context.Context, id int) (*entities.Job, error)
	GetActiveJobs(ctx context.Context) ([]entities.Job, error)
}

// TaskRepository manages task lifecycle persistence.
type TaskRepository interface {
	Create(ctx context.Context, task *entities.Task) (*entities.Task, error)
	GetByStatus(ctx context.Context, statusID, limit int) ([]entities.Task, error)
	UpdateStatus(ctx context.Context, taskID, statusID int) error
	UpdateStartTime(ctx context.Context, taskID int, t time.Time) error
	UpdateEndTime(ctx context.Context, taskID int, t time.Time) error
}

// TaskLogRepository persists execution failure logs.
type TaskLogRepository interface {
	Create(ctx context.Context, log *entities.TaskLog) error
}

// AppRepository provides read access to the apps table.
type AppRepository interface {
	GetByID(ctx context.Context, id int) (*entities.App, error)
}

// ChannelRepository provides read access to the channels table.
type ChannelRepository interface {
	GetAll(ctx context.Context) ([]entities.Channel, error)
}

// JobChannelRepository resolves which channels are assigned to a job.
type JobChannelRepository interface {
	GetByJobID(ctx context.Context, jobID int) ([]entities.Channel, error)
}

// ChannelTaskRepository manages channel_tasks rows.
type ChannelTaskRepository interface {
	Create(ctx context.Context, taskID, channelID, statusID int) error
	GetChannelsByTaskID(ctx context.Context, taskID int) ([]entities.Channel, error)
}
