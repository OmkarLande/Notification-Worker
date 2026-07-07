package entities

import "time"

// Channel represents a notification delivery channel (e.g. Email, Discord).
type Channel struct {
	ID        int
	Name      string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}