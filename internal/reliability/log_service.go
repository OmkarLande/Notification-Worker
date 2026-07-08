package reliability

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	entities "github.com/OmkarLande/notification-worker/internal/entites"
	"github.com/OmkarLande/notification-worker/internal/interfaces"
	"github.com/OmkarLande/notification-worker/internal/pipeline"
)

// TaskLogService records failures into the database for audit and tracking.
type TaskLogService interface {
	LogFailure(ctx context.Context, execution *pipeline.ExecutionContext, failure FailureContext) error
}

type taskLogServiceImpl struct {
	repo interfaces.TaskLogRepository
}

// NewTaskLogService constructs a TaskLogService.
func NewTaskLogService(repo interfaces.TaskLogRepository) TaskLogService {
	return &taskLogServiceImpl{repo: repo}
}

// LogFailure constructs a TaskLog row with performance data and persists it.
func (s *taskLogServiceImpl) LogFailure(ctx context.Context, execution *pipeline.ExecutionContext, failure FailureContext) error {
	perfLog := make(map[string]any)

	if execution.Metrics != nil {
		perfLog["pipelineDuration"] = time.Since(execution.Metrics.PipelineStart).String()

		stepDurations := make([]map[string]string, 0, len(execution.Metrics.StepDurations))
		for step, dur := range execution.Metrics.StepDurations {
			stepDurations = append(stepDurations, map[string]string{
				"step":     step,
				"duration": dur.String(),
			})
		}
		perfLog["stepDurations"] = stepDurations
	}

	perfBytes, _ := json.Marshal(perfLog)

	errMap := map[string]any{
		"step":    failure.StepName,
		"message": failure.Error.Error(),
	}
	if failure.Channel != "" {
		errMap["channel"] = failure.Channel
	}
	errBytes, _ := json.Marshal(errMap)

	logEntry := &entities.TaskLog{
		TaskID:         execution.Task.ID,
		JobID:          execution.Job.ID,
		StepName:       failure.StepName,
		PerformanceLog: perfBytes,
		ErrorLog:       errBytes,
	}

	if err := s.repo.Create(ctx, logEntry); err != nil {
		return fmt.Errorf("task log service: %w", err)
	}

	return nil
}
