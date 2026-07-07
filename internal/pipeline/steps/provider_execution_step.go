package steps

import (
	"context"
	"fmt"

	entities "github.com/OmkarLande/notification-worker/internal/entites"
	"github.com/OmkarLande/notification-worker/internal/pipeline"
)

// ProviderExecutionStep invokes the application-specific provider logic.
// It maps the generic pipeline context into the strongly-typed entities context
// that the provider expects, preventing circular dependencies.
type ProviderExecutionStep struct{}

func NewProviderExecutionStep() *ProviderExecutionStep {
	return &ProviderExecutionStep{}
}

func (s *ProviderExecutionStep) Name() string {
	return "ProviderExecutionStep"
}

func (s *ProviderExecutionStep) Order() int {
	return 20
}

func (s *ProviderExecutionStep) Execute(ctx context.Context, execution *pipeline.ExecutionContext) error {
	ec := entities.ExecutionContext{
		Task: execution.Task,
		Job:  execution.Job,
		App:  execution.App,
	}

	output, err := execution.Provider.Execute(ctx, ec)
	if err != nil {
		return fmt.Errorf("provider execution failed: %w", err)
	}

	// Store the provider's ExecutionOutput in the shared data map
	// so future steps (like templates or delivery) can consume it.
	execution.Data[pipeline.ContextExecutionOutput] = output

	return nil
}
