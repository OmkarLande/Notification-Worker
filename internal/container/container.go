// Package container provides the dependency injection container for the
// Notification Worker. The container wires all infrastructure components
// together and acts as the single resolution point for the application.
// The rest of the application must never construct dependencies directly;
// they are always obtained through the container.
package container

import (
	"context"
	"fmt"

	"github.com/OmkarLande/notification-worker/internal/config"
	"github.com/OmkarLande/notification-worker/internal/database"
	"github.com/OmkarLande/notification-worker/internal/logger"
	"github.com/OmkarLande/notification-worker/internal/providers"
)

// Container holds every infrastructure dependency used by the application.
// Add new fields here as additional layers (repositories, services, channels)
// are implemented in subsequent phases.
type Container struct {
	// Config is the fully loaded and validated application configuration.
	Config *config.AppConfig

	// Logger is the structured logger used throughout the application.
	Logger logger.Logger

	// DB is the wrapped PostgreSQL connection pool.
	DB *database.Database

	// ProviderFactory resolves app-specific notification providers by name.
	ProviderFactory *providers.Factory

	// --- Phase 3: Repositories ---
	// Repositories *repositories.Registry

	// --- Phase 4: Services ---
	// Services *services.Registry

	// --- Phase 5: Channels ---
	// Channels *channels.Registry
}

// New constructs a Container from already-initialized dependencies.
// All arguments are required; passing nil for any will return an error.
func New(
	cfg *config.AppConfig,
	log logger.Logger,
	db *database.Database,
	factory *providers.Factory,
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
	if factory == nil {
		return nil, fmt.Errorf("container: provider factory must not be nil")
	}

	return &Container{
		Config:          cfg,
		Logger:          log,
		DB:              db,
		ProviderFactory: factory,
	}, nil
}

// Health performs a liveness check on all critical infrastructure components.
// It is intended to back a future /health HTTP endpoint.
func (c *Container) Health(ctx context.Context) error {
	if err := c.DB.Health(ctx); err != nil {
		return fmt.Errorf("container health: %w", err)
	}
	return nil
}
