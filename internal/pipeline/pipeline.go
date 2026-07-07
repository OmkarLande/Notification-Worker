package pipeline

import (
	"context"
	"sort"
	"time"

	"github.com/OmkarLande/notification-worker/internal/logger"
)

// Pipeline manages the sequential execution of Steps.
// It fully owns all logging and timing for the steps it executes.
type Pipeline struct {
	logger logger.Logger
	steps  []Step
}

// NewPipeline constructs a new empty execution pipeline.
func NewPipeline(log logger.Logger) *Pipeline {
	return &Pipeline{
		logger: log,
		steps:  make([]Step, 0),
	}
}

// AddStep appends a step to the pipeline and re-sorts all steps by their Order().
func (p *Pipeline) AddStep(step Step) {
	p.steps = append(p.steps, step)
	sort.SliceStable(p.steps, func(i, j int) bool {
		return p.steps[i].Order() < p.steps[j].Order()
	})
}

// Run executes all steps sequentially. It logs start, completion, and duration
// for each step. If a step fails, the pipeline halts immediately, logs the error,
// and returns an ExecutionResult representing the failure.
func (p *Pipeline) Run(ctx context.Context, execution *ExecutionContext) (*ExecutionResult, error) {
	pipelineStart := time.Now()

	for _, step := range p.steps {
		stepName := step.Name()
		taskID := execution.Task.ID
		jobID := execution.Job.ID

		p.logger.Info("Pipeline Step Started", "step", stepName, "task_id", taskID)
		stepStart := time.Now()

		if err := step.Execute(ctx, execution); err != nil {
			p.logger.Error("Pipeline Step Failed",
				"step", stepName,
				"task_id", taskID,
				"job_id", jobID,
				"error", err,
			)
			return &ExecutionResult{
				Success:  false,
				Duration: time.Since(pipelineStart),
				Metadata: map[string]any{"failed_step": stepName},
			}, err
		}

		p.logger.Info("Pipeline Step Completed",
			"step", stepName,
			"task_id", taskID,
			"duration", time.Since(stepStart),
		)
	}

	return &ExecutionResult{
		Success:  true,
		Duration: time.Since(pipelineStart),
		Metadata: make(map[string]any),
	}, nil
}
