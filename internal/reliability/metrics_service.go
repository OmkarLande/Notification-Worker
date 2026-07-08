package reliability

import (
	"context"
	"time"

	"github.com/OmkarLande/notification-worker/internal/logger"
	"github.com/OmkarLande/notification-worker/internal/pipeline"
)

// MetricsService captures operational timing and success metrics.
type MetricsService interface {
	Record(ctx context.Context, execution *pipeline.ExecutionContext)
}

type metricsServiceImpl struct {
	log logger.Logger
}

// NewMetricsService constructs a MetricsService.
func NewMetricsService(log logger.Logger) MetricsService {
	return &metricsServiceImpl{log: log}
}

func (s *metricsServiceImpl) Record(ctx context.Context, execution *pipeline.ExecutionContext) {
	if execution.Metrics == nil {
		return
	}

	pipelineDuration := time.Since(execution.Metrics.PipelineStart)
	
	providerDur, _ := execution.Metrics.StepDurations["ProviderExecutionStep"]
	deliveryDur, _ := execution.Metrics.StepDurations["ChannelDeliveryStep"]

	s.log.Info("Metrics Recorded",
		"task_id", execution.Task.ID,
		"job_id", execution.Job.ID,
		"pipeline_duration", pipelineDuration,
		"provider_duration", providerDur,
		"delivery_duration", deliveryDur,
		"retry_count", execution.Task.CurrentRetryCount,
	)
}
