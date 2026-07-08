package reliability

import "time"

// FailureContext encapsulates details of a failure event.
// It could be an entire pipeline step failure or a specific channel failure.
type FailureContext struct {
	StepName string
	Channel  string // Empty if it's a generic pipeline step failure
	Error    error
	Duration time.Duration
}
