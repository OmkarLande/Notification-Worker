package entities

import (
	"encoding/json"
	"time"
)

// TaskLog records a failure event during task execution. It is persisted to
// task_logs only when a task step fails, providing a full audit trail.
type TaskLog struct {
	ID             int
	TaskID         int
	JobID          int
	StepName       string
	PerformanceLog json.RawMessage
	ErrorLog       json.RawMessage
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
