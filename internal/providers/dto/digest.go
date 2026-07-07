package dto

import "time"

// DailyDigestData holds the aggregated data for a single user's daily digest.
// It is the payload returned by the Stackday provider's daily digest execution.
type DailyDigestData struct {
	UserID       int
	PendingTasks []PendingTask
	Goals        []GoalProgress
	GeneratedAt  time.Time
}

// PendingTask is a task that is due or overdue for the user.
type PendingTask struct {
	ID    int
	Title string
	DueAt time.Time
}

// GoalProgress represents a user's progress toward a goal.
type GoalProgress struct {
	ID       int
	Title    string
	Progress float64 // 0.0 – 1.0
}
