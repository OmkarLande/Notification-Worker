package slack

import (
	"context"
	"fmt"
	"time"

	"github.com/OmkarLande/notification-worker/internal/channels"
	"github.com/OmkarLande/notification-worker/internal/contracts"
)

type SlackChannel struct {
}

func NewSlackChannel() *SlackChannel {
	return &SlackChannel{}
}

func (c *SlackChannel) Name() string {
	return "slack"
}

func (c *SlackChannel) Validate(ctx context.Context, payload contracts.Payload) error {
	p, ok := payload.(*contracts.SlackPayload)
	if !ok {
		return fmt.Errorf("slack channel requires *contracts.SlackPayload")
	}

	if len(p.Blocks) == 0 {
		return fmt.Errorf("slack payload requires at least one block")
	}

	return nil
}

func (c *SlackChannel) Deliver(ctx context.Context, payload contracts.Payload) (*channels.DeliveryResult, error) {
	start := time.Now()

	if err := c.Validate(ctx, payload); err != nil {
		return &channels.DeliveryResult{
			Success:      false,
			Channel:      c.Name(),
			Duration:     time.Since(start),
			ErrorMessage: err.Error(),
		}, nil
	}

	// Mock delivery
	time.Sleep(50 * time.Millisecond)

	return &channels.DeliveryResult{
		Success:           true,
		Channel:           c.Name(),
		Duration:          time.Since(start),
		ProviderMessageID: "mock-slack-id",
		Metadata: map[string]any{
			"payload_size": 100, // mock size
		},
	}, nil
}
