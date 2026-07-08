package payloadresolver

import (
	"fmt"

	"github.com/OmkarLande/notification-worker/internal/contracts"
	"github.com/OmkarLande/notification-worker/internal/pipeline"
)

// Resolver extracts the correct payload from the ExecutionContext based on the channel name.
type Resolver struct{}

// New returns a new Resolver.
func New() *Resolver {
	return &Resolver{}
}

// ResolvePayload extracts the appropriate payload from the ExecutionContext.
func (r *Resolver) ResolvePayload(execution *pipeline.ExecutionContext, channelName string) (contracts.Payload, error) {
	switch channelName {
	case "email":
		if execution.EmailPayload == nil {
			return nil, fmt.Errorf("email payload is nil in context")
		}
		return execution.EmailPayload, nil
	case "discord":
		if execution.DiscordPayload == nil {
			return nil, fmt.Errorf("discord payload is nil in context")
		}
		return execution.DiscordPayload, nil
	case "slack":
		if execution.SlackPayload == nil {
			return nil, fmt.Errorf("slack payload is nil in context")
		}
		return execution.SlackPayload, nil
	case "whatsapp":
		if execution.WhatsAppPayload == nil {
			return nil, fmt.Errorf("whatsapp payload is nil in context")
		}
		return execution.WhatsAppPayload, nil
	default:
		return nil, fmt.Errorf("no payload resolver found for channel %q", channelName)
	}
}
