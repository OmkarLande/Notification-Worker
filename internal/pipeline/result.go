package pipeline

import "time"

// ExecutionResult is the final outcome produced by the pipeline.
// It is returned by Pipeline.Run() to the TaskWorker.
type ExecutionResult struct {
	Success  bool
	Duration time.Duration
	Metadata map[string]any
}
