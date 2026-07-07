// Package app owns the application lifecycle for the Notification Worker.
// It is responsible for wiring all infrastructure dependencies and orchestrating
// the startup and shutdown sequences. Keeping lifecycle logic here — rather than
// in main — makes the application testable and keeps main.go minimal.
package app

import (
	"context"
	"fmt"

	"github.com/OmkarLande/notification-worker/internal/config"
	"github.com/OmkarLande/notification-worker/internal/container"
	"github.com/OmkarLande/notification-worker/internal/database"
	"github.com/OmkarLande/notification-worker/internal/logger"
	"github.com/OmkarLande/notification-worker/internal/providers"
)

// Application owns the complete lifecycle of the Notification Worker.
// It holds the fully wired container and exposes Start / Shutdown methods
// that main.go calls in sequence.
type Application struct {
	container *container.Container
	cfg       *config.AppConfig
}

// New wires all infrastructure components in the correct dependency order and
// returns a ready-to-start Application. It returns a descriptive error if any
// component fails to initialize so that main.go can exit immediately with a
// useful message.
//
// Startup order:
//  1. Logger
//  2. Database connection pool
//  3. Database wrapper
//  4. Provider registry + factory
//  5. Container
func New(cfg *config.AppConfig) (*Application, error) {
	// 1. Logger — must be first so every subsequent step can log.
	log := logger.New(cfg.Worker.AppEnv)
	log.Info("Logger initialized", "env", cfg.Worker.AppEnv, "worker_id", cfg.Worker.WorkerID)

	// 2. Database connection pool.
	pool, err := database.NewPool(cfg.Database, log)
	if err != nil {
		return nil, fmt.Errorf("app: %w", err)
	}

	// 3. Wrap the pool in the Database abstraction.
	db := database.New(pool)

	// 4. Provider registry and factory (empty — implementations registered in
	//    Phase 3 when individual providers are built).
	registry := providers.NewRegistry()
	factory := providers.NewFactory(registry)

	log.Info("Provider factory initialized", "registered_providers", len(factory.RegisteredNames()))

	// 5. Assemble the DI container.
	c, err := container.New(cfg, log, db, factory)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("app: failed to build container: %w", err)
	}

	log.Info("Dependency container built successfully")

	return &Application{
		container: c,
		cfg:       cfg,
	}, nil
}

// Start logs the startup banner. In future phases this will also start the
// job scheduler, task consumers, and HTTP health server.
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

// Shutdown performs an orderly teardown of all resources. It should be called
// after the OS signal has been received and before the process exits.
func (a *Application) Shutdown(ctx context.Context) {
	log := a.container.Logger
	log.Info("Shutting down Notification Worker gracefully...",
		"worker_id", a.cfg.Worker.WorkerID,
	)

	a.container.DB.Close()

	log.Info("Shutdown complete. Goodbye.")
}

// Container returns the application's dependency container.
// Intended for use in integration tests and health checks.
func (a *Application) Container() *container.Container {
	return a.container
}
