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

// Steps returns the registered steps in order.
func (p *Pipeline) Steps() []Step {
	return p.steps
}

// Run executes all steps sequentially. It logs start, completion, and duration
// for each step. If a step fails, the pipeline halts immediately, logs the error,
// and returns an ExecutionResult representing the failure.
func (p *Pipeline) Run(ctx context.Context, execution *ExecutionContext) (*ExecutionResult, error) {
	if execution.Metrics == nil {
		execution.Metrics = &ExecutionMetrics{
			PipelineStart: time.Now(),
			StepDurations: make(map[string]time.Duration),
		}
	} else if execution.Metrics.PipelineStart.IsZero() {
		execution.Metrics.PipelineStart = time.Now()
	}

	for _, step := range p.steps {
		stepName := step.Name()
		taskID := execution.Task.ID
		jobID := execution.Job.ID

		// If pipeline has failed, skip all remaining steps except FinalizeExecutionStep
		if execution.Error != nil && stepName != "FinalizeExecutionStep" {
			continue
		}

		p.logger.Info("Pipeline Step Started", "step", stepName, "task_id", taskID)
		stepStart := time.Now()

		err := step.Execute(ctx, execution)
		duration := time.Since(stepStart)
		execution.Metrics.StepDurations[stepName] = duration

		if err != nil {
			p.logger.Error("Pipeline Step Failed",
				"step", stepName,
				"task_id", taskID,
				"job_id", jobID,
				"error", err,
			)
			if execution.Error == nil {
				// Record the FIRST failure
				execution.Error = err
				execution.FailedStep = stepName
			}
			// Note: We don't return early anymore. We continue the loop
			// so that FinalizeExecutionStep can process the failure.
		} else {
			p.logger.Info("Pipeline Step Completed",
				"step", stepName,
				"task_id", taskID,
				"duration", duration,
			)
		}
	}

	success := execution.Error == nil
	res := &ExecutionResult{
		Success:  success,
		Duration: time.Since(execution.Metrics.PipelineStart),
		Metadata: make(map[string]any),
	}
	if !success {
		res.Metadata["failed_step"] = execution.FailedStep
	}
	return res, execution.Error
}
