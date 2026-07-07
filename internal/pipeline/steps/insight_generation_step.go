package steps

import (
	"context"
	"fmt"

	"github.com/OmkarLande/notification-worker/internal/pipeline"
	"github.com/OmkarLande/notification-worker/internal/services"
)

// InsightGenerationStep processes the raw ExecutionOutput into deterministic insights.
type InsightGenerationStep struct {
	insightService *services.InsightService
}

func NewInsightGenerationStep(insightService *services.InsightService) *InsightGenerationStep {
	return &InsightGenerationStep{
		insightService: insightService,
	}
}

func (s *InsightGenerationStep) Name() string {
	return "InsightGenerationStep"
}

func (s *InsightGenerationStep) Order() int {
	return 30
}

func (s *InsightGenerationStep) Execute(ctx context.Context, execution *pipeline.ExecutionContext) error {
	if execution.ExecutionOutput == nil {
		return fmt.Errorf("insight step: ExecutionOutput is nil")
	}

	insight, err := s.insightService.Generate(ctx, execution.ExecutionOutput)
	if err != nil {
		return fmt.Errorf("failed to generate insight: %w", err)
	}

	execution.Insight = insight
	return nil
}
