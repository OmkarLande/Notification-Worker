package channels

import (
	"context"
	"time"

	"github.com/OmkarLande/notification-worker/internal/contracts"
)

// DeliveryResult captures the outcome of a channel delivery attempt.
type DeliveryResult struct {
	Success           bool
	Channel           string
	Duration          time.Duration
	ProviderMessageID string
	ErrorMessage      string
	Metadata          map[string]any
}

// Channel acts as a transport adapter. It should not be aware of execution
// pipeline specifics (Task, Job, App), but strictly handles payload delivery.
type Channel interface {
	Name() string
	Validate(ctx context.Context, payload contracts.Payload) error
	Deliver(ctx context.Context, payload contracts.Payload) (*DeliveryResult, error)
}
