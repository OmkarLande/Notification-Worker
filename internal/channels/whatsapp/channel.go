package whatsapp

import (
	"context"
	"fmt"
	"time"

	"github.com/OmkarLande/notification-worker/internal/channels"
	"github.com/OmkarLande/notification-worker/internal/contracts"
)

type WhatsAppChannel struct {
}

func NewWhatsAppChannel() *WhatsAppChannel {
	return &WhatsAppChannel{}
}

func (c *WhatsAppChannel) Name() string {
	return "whatsapp"
}

func (c *WhatsAppChannel) Validate(ctx context.Context, payload contracts.Payload) error {
	p, ok := payload.(*contracts.WhatsAppPayload)
	if !ok {
		return fmt.Errorf("whatsapp channel requires *contracts.WhatsAppPayload")
	}

	if p.Text == "" {
		return fmt.Errorf("whatsapp payload requires text content")
	}

	return nil
}

func (c *WhatsAppChannel) Deliver(ctx context.Context, payload contracts.Payload) (*channels.DeliveryResult, error) {
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
		ProviderMessageID: "mock-whatsapp-id",
		Metadata: map[string]any{
			"payload_size": 100, // mock size
		},
	}, nil
}
