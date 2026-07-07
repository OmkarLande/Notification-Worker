package pipeline

import (
	"github.com/OmkarLande/notification-worker/internal/contracts"
	entities "github.com/OmkarLande/notification-worker/internal/entites"
	"github.com/OmkarLande/notification-worker/internal/interfaces"
	"github.com/OmkarLande/notification-worker/internal/providers/dto"
)

// ExecutionContext represents the state of a single Task execution as it flows
// through the transformation pipeline. Every step enriches the context with
// strongly-typed payloads.
type ExecutionContext struct {
	// Base context setup by Dispatcher/Worker
	Task     *entities.Task
	Job      *entities.Job
	App      *entities.App
	Provider interfaces.Provider

	// ExecutionOutput is populated by the ProviderExecutionStep
	ExecutionOutput *dto.ExecutionOutput

	// Insight is populated by the InsightGenerationStep
	Insight *contracts.InsightResult

	// Payloads are populated by the PayloadTransformationStep
	EmailPayload    *contracts.EmailPayload
	DiscordPayload  *contracts.DiscordPayload
	SlackPayload    *contracts.SlackPayload
	WhatsAppPayload *contracts.WhatsAppPayload

	// Metadata provides an extensibility point for temporary cross-step state.
	// It should NOT be used for primary execution state.
	Metadata map[string]any
}
