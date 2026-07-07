// Package entities contains the core domain models for the Notification Worker.
// These types represent the worker's own database schema and are never exposed
// directly to external APIs or provider implementations.
package entities

import (
	"encoding/json"
	"time"
)

// Job represents a scheduled notification job stored in the jobs table.
type Job struct {
	ID             int
	AppID          int
	Name           string
	Description    string
	StatusID       int
	MaxThreadCount int
	Arguments      json.RawMessage
	CreatedAt      time.Time
	UpdatedAt      time.Time
}