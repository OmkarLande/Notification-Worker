package steps

import (
	"context"
	"fmt"

	"github.com/OmkarLande/notification-worker/internal/pipeline"
	"github.com/OmkarLande/notification-worker/internal/services"
)

// PayloadTransformationStep uses the TemplateService to render HTML/text and
// directly constructs channel-specific payloads (Email, Discord, Slack, WhatsApp).
type PayloadTransformationStep struct {
	templateService *services.TemplateService
}

func NewPayloadTransformationStep(templateService *services.TemplateService) *PayloadTransformationStep {
	return &PayloadTransformationStep{
		templateService: templateService,
	}
}

func (s *PayloadTransformationStep) Name() string {
	return "PayloadTransformationStep"
}

func (s *PayloadTransformationStep) Order() int {
	return 40
}

func (s *PayloadTransformationStep) Execute(ctx context.Context, execution *pipeline.ExecutionContext) error {
	if execution.ExecutionOutput == nil {
		return fmt.Errorf("payload step: ExecutionOutput is nil")
	}

	key := services.TemplateKey{
		App: execution.App.Name,
		Job: execution.Job.Name,
	}
	data := execution.ExecutionOutput.Payload

	// 1. Email
	emailPayload, err := s.templateService.GenerateEmailPayload(ctx, key, data)
	if err != nil {
		return fmt.Errorf("failed to generate email payload: %w", err)
	}
	execution.EmailPayload = emailPayload

	// 2. Discord
	discordPayload, err := s.templateService.GenerateDiscordPayload(ctx, key, data)
	if err != nil {
		return fmt.Errorf("failed to generate discord payload: %w", err)
	}
	execution.DiscordPayload = discordPayload

	// 3. Slack
	slackPayload, err := s.templateService.GenerateSlackPayload(ctx, key, data)
	if err != nil {
		return fmt.Errorf("failed to generate slack payload: %w", err)
	}
	execution.SlackPayload = slackPayload

	// 4. WhatsApp
	whatsappPayload, err := s.templateService.GenerateWhatsAppPayload(ctx, key, data)
	if err != nil {
		return fmt.Errorf("failed to generate whatsapp payload: %w", err)
	}
	execution.WhatsAppPayload = whatsappPayload

	return nil
}
