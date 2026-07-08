package steps

import (
	"context"

	"github.com/OmkarLande/notification-worker/internal/pipeline"
	"github.com/OmkarLande/notification-worker/internal/reliability"
)

// FinalizeExecutionStep handles final cleanup, metrics, or metadata persistence.
type FinalizeExecutionStep struct {
	reliabilityManager *reliability.Manager
}

func NewFinalizeExecutionStep(rm *reliability.Manager) *FinalizeExecutionStep {
	return &FinalizeExecutionStep{
		reliabilityManager: rm,
	}
}

func (s *FinalizeExecutionStep) Name() string {
	return "FinalizeExecutionStep"
}

func (s *FinalizeExecutionStep) Order() int {
	return 100
}

func (s *FinalizeExecutionStep) Execute(ctx context.Context, execution *pipeline.ExecutionContext) error {
	// Let the Reliability layer own all failure, metrics and retry logic.
	// It absorbs errors internally to prevent pipeline cleanup crashes.
	s.reliabilityManager.Handle(ctx, execution)
	return nil
}
