package container

import (
	"context"
	"fmt"

	"github.com/OmkarLande/notification-worker/internal/cache"
	"github.com/OmkarLande/notification-worker/internal/config"
	"github.com/OmkarLande/notification-worker/internal/database"
	"github.com/OmkarLande/notification-worker/internal/interfaces"
	"github.com/OmkarLande/notification-worker/internal/logger"
	"github.com/OmkarLande/notification-worker/internal/pipeline"
	"github.com/OmkarLande/notification-worker/internal/providers"
	"github.com/OmkarLande/notification-worker/internal/services"
	"github.com/OmkarLande/notification-worker/internal/workers"
)

// Repositories groups all repository implementations.
type Repositories struct {
	Jobs         interfaces.JobRepository
	Tasks        interfaces.TaskRepository
	TaskLogs     interfaces.TaskLogRepository
	Apps         interfaces.AppRepository
	Channels     interfaces.ChannelRepository
	JobChannels  interfaces.JobChannelRepository
	ChannelTasks interfaces.ChannelTaskRepository
}

// Container holds every dependency used by the application. Dependencies are
// resolved through this struct; global state and service locators are avoided.
type Container struct {
	Config      *config.AppConfig
	Logger      logger.Logger
	DB          *database.Database
	StatusCache *cache.StatusCache

	ProviderFactory *providers.Factory
	Pipeline        *pipeline.Pipeline

	Repos Repositories

	JobService          *services.JobService
	JobExecutionService *services.JobExecutionService

	Dispatcher *workers.TaskDispatcher
	TaskWorker *workers.TaskWorker
}

// New constructs a Container from fully initialized dependencies.
// All arguments are required; passing nil for any will return an error.
func New(
	cfg *config.AppConfig,
	log logger.Logger,
	db *database.Database,
	statusCache *cache.StatusCache,
	factory *providers.Factory,
	execPipeline *pipeline.Pipeline,
	repos Repositories,
	jobService *services.JobService,
	jobExecService *services.JobExecutionService,
	dispatcher *workers.TaskDispatcher,
	taskWorker *workers.TaskWorker,
) (*Container, error) {
	if cfg == nil {
		return nil, fmt.Errorf("container: config must not be nil")
	}
	if log == nil {
		return nil, fmt.Errorf("container: logger must not be nil")
	}
	if db == nil {
		return nil, fmt.Errorf("container: database must not be nil")
	}

	return &Container{
		Config:              cfg,
		Logger:              log,
		DB:                  db,
		StatusCache:         statusCache,
		ProviderFactory:     factory,
		Pipeline:            execPipeline,
		Repos:               repos,
		JobService:          jobService,
		JobExecutionService: jobExecService,
		Dispatcher:          dispatcher,
		TaskWorker:          taskWorker,
	}, nil
}

// Health performs a liveness check on all critical infrastructure components.
func (c *Container) Health(ctx context.Context) error {
	if err := c.DB.Health(ctx); err != nil {
		return fmt.Errorf("container health: %w", err)
	}
	return nil
}
