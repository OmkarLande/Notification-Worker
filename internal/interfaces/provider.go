package interfaces

import (
	"context"

	entities "github.com/OmkarLande/notification-worker/internal/entites"
	"github.com/OmkarLande/notification-worker/internal/providers/dto"
)

// Provider is the contract every app-specific notification provider must satisfy.
// Providers are initialized once in their constructor and registered with the
// ProviderFactory at startup. factory.Get() always returns a live, ready provider.
type Provider interface {
	// Name returns the unique app identifier used as the factory registry key.
	// Must match the value in the apps.name column (e.g. "stackday", "expense").
	Name() string

	// GetNotificationEnabledUsers returns the list of users for whom tasks
	// should be created. Called once per job trigger by JobExecutionService.
	GetNotificationEnabledUsers(ctx context.Context) ([]dto.User, error)

	// Execute performs the complete execution for a single task.
	// The ExecutionContext carries the Task (with its Arguments JSON), the parent
	// Job, and the owning App so the provider never needs to query the worker DB.
	Execute(ctx context.Context, ec entities.ExecutionContext) (*dto.ExecutionOutput, error)

	// Health checks provider connectivity. Called by container health checks.
	Health(ctx context.Context) error
}