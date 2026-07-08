package reliability

import (
	"context"
	"fmt"

	"github.com/OmkarLande/notification-worker/internal/logger"
	"github.com/OmkarLande/notification-worker/internal/pipeline"
)

// Manager orchestrates all reliability concerns: metrics, logging, and retries.
type Manager struct {
	retryService   RetryService
	metricsService MetricsService
	taskLogService TaskLogService
	log            logger.Logger
}

// NewManager constructs a ReliabilityManager.
func NewManager(
	retryService RetryService,
	metricsService MetricsService,
	taskLogService TaskLogService,
	log logger.Logger,
) *Manager {
	return &Manager{
		retryService:   retryService,
		metricsService: metricsService,
		taskLogService: taskLogService,
		log:            log,
	}
}

// Handle inspects the completed ExecutionContext and orchestrates reliability tasks.
// It will never return an error to ensure cleanup tasks don't crash the pipeline.
func (m *Manager) Handle(ctx context.Context, execution *pipeline.ExecutionContext) {
	// Always record metrics
	m.metricsService.Record(ctx, execution)

	// Case 1: Pipeline-level Failure (e.g. Provider failed)
	if execution.Error != nil {
		failureCtx := FailureContext{
			StepName: execution.FailedStep,
			Error:    execution.Error,
			// Duration could be pulled from metrics if needed
		}

		if err := m.taskLogService.LogFailure(ctx, execution, failureCtx); err != nil {
			m.log.Error("ReliabilityManager: failed to log pipeline failure", "error", err)
		}

		if err := m.retryService.ScheduleRetry(ctx, execution, failureCtx); err != nil {
			m.log.Error("ReliabilityManager: failed to schedule retry for pipeline failure", "error", err)
		}

		return
	}

	// Case 2: Channel-level Failures (Pipeline succeeded, but some channels failed)
	if execution.Delivery != nil {
		for _, result := range execution.Delivery.Results {
			if !result.Success {
				failureCtx := FailureContext{
					StepName: "ChannelDeliveryStep",
					Channel:  result.Channel,
					Error:    fmt.Errorf("%s", result.ErrorMessage),
				}

				if err := m.taskLogService.LogFailure(ctx, execution, failureCtx); err != nil {
					m.log.Error("ReliabilityManager: failed to log channel failure", "channel", result.Channel, "error", err)
				}

				// We record the channel as a retry candidate.
				// In a future phase, RetryService will retry individual channels.
				// For now, we just pass the failure context down to the RetryService,
				// which will currently just retry the whole task because of how we built it.
				if err := m.retryService.ScheduleRetry(ctx, execution, failureCtx); err != nil {
					m.log.Error("ReliabilityManager: failed to schedule retry for channel failure", "channel", result.Channel, "error", err)
				}
			}
		}
	}
}
