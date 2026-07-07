package entities

import (
	"encoding/json"
	"time"
)

// Task represents a single unit of execution created by JobExecutionService.
// One Task is created per user per job trigger. Arguments is a JSON payload
// that the assigned provider deserializes at execution time.
type Task struct {
	ID                int
	JobID             int
	ParentTaskID      *int
	StatusID          int
	Arguments         json.RawMessage
	TaskTriggerTime   time.Time
	TaskStartTime     *time.Time
	TaskEndTime       *time.Time
	CurrentRetryCount int
	CreatedAt         time.Time
	UpdatedAt         time.Time
}