package steps

import (
	"context"
	"fmt"

	"github.com/OmkarLande/notification-worker/internal/pipeline"
)

// ValidateContextStep verifies that all required pointers on the ExecutionContext
// are present before processing begins. It contains no business logic.
type ValidateContextStep struct{}

func NewValidateContextStep() *ValidateContextStep {
	return &ValidateContextStep{}
}

func (s *ValidateContextStep) Name() string {
	return "ValidateContextStep"
}

func (s *ValidateContextStep) Order() int {
	return 10
}

func (s *ValidateContextStep) Execute(_ context.Context, execution *pipeline.ExecutionContext) error {
	if execution.Task == nil {
		return fmt.Errorf("task is nil")
	}
	if execution.Job == nil {
		return fmt.Errorf("job is nil")
	}
	if execution.App == nil {
		return fmt.Errorf("app is nil")
	}
	if execution.Provider == nil {
		return fmt.Errorf("provider is nil")
	}
	return nil
}
