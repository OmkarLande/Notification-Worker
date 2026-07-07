package steps

import (
	"context"

	"github.com/OmkarLande/notification-worker/internal/pipeline"
)

// FinalizeExecutionStep handles final cleanup, metrics, or metadata persistence.
// In Phase 4 it acts as a placeholder at the end of the pipeline.
// Logging is entirely owned by the Pipeline runner.
type FinalizeExecutionStep struct{}

func NewFinalizeExecutionStep() *FinalizeExecutionStep {
	return &FinalizeExecutionStep{}
}

func (s *FinalizeExecutionStep) Name() string {
	return "FinalizeExecutionStep"
}

func (s *FinalizeExecutionStep) Order() int {
	return 100
}

func (s *FinalizeExecutionStep) Execute(_ context.Context, _ *pipeline.ExecutionContext) error {
	// Future phases will extend this step to persist insights, PDFs,
	// or specific metadata generated during execution.
	return nil
}
