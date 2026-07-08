package deliverymanager

import (
	"context"
	"strings"

	"github.com/OmkarLande/notification-worker/internal/channels"
	"github.com/OmkarLande/notification-worker/internal/interfaces"
	"github.com/OmkarLande/notification-worker/internal/logger"
	"github.com/OmkarLande/notification-worker/internal/pipeline"
	"github.com/OmkarLande/notification-worker/internal/services/payloadresolver"
)

// DeliveryManager orchestrates the delivery of payloads to configured channels.
type DeliveryManager struct {
	channelTaskRepo interfaces.ChannelTaskRepository
	registry        *channels.Registry
	resolver        *payloadresolver.Resolver
	logger          logger.Logger
}

// New constructs a DeliveryManager.
func New(
	channelTaskRepo interfaces.ChannelTaskRepository,
	registry *channels.Registry,
	resolver *payloadresolver.Resolver,
	logger logger.Logger,
) *DeliveryManager {
	return &DeliveryManager{
		channelTaskRepo: channelTaskRepo,
		registry:        registry,
		resolver:        resolver,
		logger:          logger,
	}
}

// DeliverAll fetches enabled channels for the given task, resolves payloads,
// and delivers them sequentially, collecting results.
func (m *DeliveryManager) DeliverAll(ctx context.Context, execution *pipeline.ExecutionContext) error {
	if execution.Delivery == nil {
		execution.Delivery = &pipeline.DeliveryContext{
			Results: make([]channels.DeliveryResult, 0),
		}
	}

	taskChannels, err := m.channelTaskRepo.GetChannelsByTaskID(ctx, execution.Task.ID)
	if err != nil {
		m.logger.Error("DeliveryManager: failed to fetch channels for task", "task_id", execution.Task.ID, "error", err)
		return err
	}

	for _, ch := range taskChannels {
		// Normalize channel name to match registry (e.g. "Email" -> "email")
		name := strings.ToLower(ch.Name)

		channelImpl, err := m.registry.Get(name)
		if err != nil {
			m.logger.Warn("DeliveryManager: channel not found in registry", "channel", name)
			execution.Delivery.Results = append(execution.Delivery.Results, channels.DeliveryResult{
				Success:      false,
				Channel:      name,
				ErrorMessage: err.Error(),
			})
			continue
		}

		payload, err := m.resolver.ResolvePayload(execution, name)
		if err != nil {
			m.logger.Warn("DeliveryManager: payload resolution failed", "channel", name, "error", err)
			execution.Delivery.Results = append(execution.Delivery.Results, channels.DeliveryResult{
				Success:      false,
				Channel:      name,
				ErrorMessage: err.Error(),
			})
			continue
		}

		m.logger.Info("DeliveryManager: starting delivery", "channel", name, "task_id", execution.Task.ID)

		result, err := channelImpl.Deliver(ctx, payload)
		if err != nil {
			// Fatal channel errors that completely crash delivery for this channel
			m.logger.Error("DeliveryManager: channel delivery crashed", "channel", name, "error", err)
			execution.Delivery.Results = append(execution.Delivery.Results, channels.DeliveryResult{
				Success:      false,
				Channel:      name,
				ErrorMessage: err.Error(),
			})
			continue
		}

		if result != nil {
			execution.Delivery.Results = append(execution.Delivery.Results, *result)
			if result.Success {
				m.logger.Info("DeliveryManager: delivery successful", 
					"channel", name, 
					"duration", result.Duration, 
					"message_id", result.ProviderMessageID,
				)
			} else {
				m.logger.Warn("DeliveryManager: delivery failed", 
					"channel", name, 
					"duration", result.Duration, 
					"error", result.ErrorMessage,
				)
			}
		}
	}

	return nil
}
