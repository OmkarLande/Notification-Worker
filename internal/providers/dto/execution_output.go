package dto

import "time"

// ExecutionOutput is the result of a single provider execution.
// It carries a generic Payload (the provider-specific data produced),
// Metadata for observability and future channel processors, and Duration
// for performance tracking in task logs.
type ExecutionOutput struct {
	// Payload is the provider-specific result (e.g. DailyDigestData).
	// Future channel processors (email, Discord, PDF) will consume this.
	Payload any

	// Metadata carries key-value pairs describing what happened during execution.
	// Examples: {"email_sent": true}, {"pdf_pages": 3}, {"discord_message_id": "…"}
	Metadata map[string]any

	// Duration is the time taken for the provider execution step.
	// Persisted to task_logs.performance_log for metrics and alerting.
	Duration time.Duration
}
