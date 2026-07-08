package steps

import (
	"context"
	"fmt"

	"github.com/OmkarLande/notification-worker/internal/pipeline"
	"github.com/OmkarLande/notification-worker/internal/services/deliverymanager"
)

// ChannelDeliveryStep invokes the DeliveryManager to dispatch payloads to
// configured channels for the given execution context.
type ChannelDeliveryStep struct {
	deliveryManager *deliverymanager.DeliveryManager
}

func NewChannelDeliveryStep(deliveryManager *deliverymanager.DeliveryManager) *ChannelDeliveryStep {
	return &ChannelDeliveryStep{
		deliveryManager: deliveryManager,
	}
}

func (s *ChannelDeliveryStep) Name() string {
	return "ChannelDeliveryStep"
}

func (s *ChannelDeliveryStep) Order() int {
	return 50 // Run after PayloadTransformationStep (Order=40)
}

func (s *ChannelDeliveryStep) Execute(ctx context.Context, execution *pipeline.ExecutionContext) error {
	if err := s.deliveryManager.DeliverAll(ctx, execution); err != nil {
		return fmt.Errorf("channel delivery step failed: %w", err)
	}

	return nil
}
