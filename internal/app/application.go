package app

import (
	"context"
	"fmt"

	"github.com/OmkarLande/notification-worker/internal/cache"
	"github.com/OmkarLande/notification-worker/internal/config"
	"github.com/OmkarLande/notification-worker/internal/container"
	"github.com/OmkarLande/notification-worker/internal/database"
	"github.com/OmkarLande/notification-worker/internal/logger"
	"github.com/OmkarLande/notification-worker/internal/pipeline"
	"github.com/OmkarLande/notification-worker/internal/pipeline/steps"
	"github.com/OmkarLande/notification-worker/internal/providers"
	"github.com/OmkarLande/notification-worker/internal/providers/expense"
	"github.com/OmkarLande/notification-worker/internal/providers/stackday"
	"github.com/OmkarLande/notification-worker/internal/repositories"
	"github.com/OmkarLande/notification-worker/internal/services"
	"github.com/OmkarLande/notification-worker/internal/workers"
)

// Application owns the complete lifecycle of the Notification Worker.
type Application struct {
	container *container.Container
	cfg       *config.AppConfig
}

// New wires all infrastructure and application components in dependency order
// and returns a ready-to-start Application.
func New(cfg *config.AppConfig) (*Application, error) {
	// 1. Logger.
	log := logger.New(cfg.Worker.AppEnv)
	log.Info("Logger initialized", "env", cfg.Worker.AppEnv, "worker_id", cfg.Worker.WorkerID)

	// 2. Database.
	pool, err := database.NewPool(cfg.Database, log)
	if err != nil {
		return nil, fmt.Errorf("app: %w", err)
	}
	db := database.New(pool)

	// 3. Status cache — loaded once, read-only during execution.
	statusCache, err := cache.Load(context.Background(), db.Pool)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("app: status cache: %w", err)
	}
	log.Info("Status cache loaded",
		"task_statuses", len(statusCache.TaskStatuses()),
	)

	// 4. Repositories.
	jobRepo := repositories.NewJobRepository(db.Pool)
	taskRepo := repositories.NewTaskRepository(db.Pool)
	taskLogRepo := repositories.NewTaskLogRepository(db.Pool)
	appRepo := repositories.NewAppRepository(db.Pool)
	channelRepo := repositories.NewChannelRepository(db.Pool)
	jobChannelRepo := repositories.NewJobChannelRepository(db.Pool)
	channelTaskRepo := repositories.NewChannelTaskRepository(db.Pool)

	repos := container.Repositories{
		Jobs:         jobRepo,
		Tasks:        taskRepo,
		TaskLogs:     taskLogRepo,
		Apps:         appRepo,
		Channels:     channelRepo,
		JobChannels:  jobChannelRepo,
		ChannelTasks: channelTaskRepo,
	}

	// 5. Provider factory.
	registry := providers.NewRegistry()
	factory := providers.NewFactory(registry)

	// 6. Register provider implementations.
	stackdayProvider, err := stackday.New(cfg.Database.URL, log)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("app: stackday provider: %w", err)
	}
	if err := factory.Register("stackday", stackdayProvider); err != nil {
		db.Close()
		return nil, fmt.Errorf("app: register stackday: %w", err)
	}

	expenseProvider := expense.New(cfg.Firebase, log)
	if err := factory.Register("expense", expenseProvider); err != nil {
		db.Close()
		return nil, fmt.Errorf("app: register expense: %w", err)
	}

	log.Info("Provider factory initialized", "registered", factory.RegisteredNames())

	// 7. Execution Pipeline
	execPipeline := pipeline.NewPipeline(log)
	execPipeline.AddStep(steps.NewValidateContextStep())
	execPipeline.AddStep(steps.NewProviderExecutionStep())
	execPipeline.AddStep(steps.NewFinalizeExecutionStep())

	log.Info("Execution pipeline initialized", "steps", 3)

	// 8. Services.
	jobService := services.NewJobService(jobRepo, appRepo, statusCache, log)
	jobExecService := services.NewJobExecutionService(
		jobService, taskRepo, jobChannelRepo, channelTaskRepo, factory, statusCache, log,
	)

	// 9. Workers.
	taskWorker := workers.NewTaskWorker(taskRepo, taskLogRepo, factory, statusCache, execPipeline, log)
	dispatcher := workers.NewTaskDispatcher(taskRepo, jobRepo, appRepo, taskWorker, statusCache, log)

	// 10. Container.
	c, err := container.New(
		cfg, log, db, statusCache, factory, execPipeline,
		repos, jobService, jobExecService, dispatcher, taskWorker,
	)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("app: container: %w", err)
	}

	log.Info("Dependency container built successfully")
	return &Application{container: c, cfg: cfg}, nil
}

// Start logs the startup banner.
func (a *Application) Start() {
	a.container.Logger.Info(
		"🚀 Notification Worker started",
		"worker_id", a.cfg.Worker.WorkerID,
		"env", a.cfg.Worker.AppEnv,
		"port", a.cfg.Worker.Port,
		"shutdown_timeout", a.cfg.Worker.ShutdownTimeout,
	)
	a.container.Logger.Info("Press Ctrl+C to stop.")
}

// Shutdown performs an orderly teardown of all resources.
func (a *Application) Shutdown(ctx context.Context) {
	log := a.container.Logger
	log.Info("Shutting down Notification Worker gracefully...", "worker_id", a.cfg.Worker.WorkerID)
	a.container.DB.Close()
	log.Info("Shutdown complete. Goodbye.")
}

// Container returns the application's dependency container.
func (a *Application) Container() *container.Container {
	return a.container
}
