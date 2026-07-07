package pipeline

import "context"

// Step defines a single, independent unit of work within the execution pipeline.
type Step interface {
	// Name returns the identifier for logging and tracing.
	Name() string

	// Order returns an integer used to sort the step within the pipeline.
	// Lower numbers execute earlier (e.g., Validation=10, Provider=20, Finalize=100).
	Order() int

	// Execute performs the step's logic, mutating the ExecutionContext if necessary.
	// Returning an error immediately halts the pipeline execution.
	Execute(ctx context.Context, execution *ExecutionContext) error
}
