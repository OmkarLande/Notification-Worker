package pipeline

import (
	entities "github.com/OmkarLande/notification-worker/internal/entites"
	"github.com/OmkarLande/notification-worker/internal/interfaces"
)

const (
	// ContextExecutionOutput is the key used in ExecutionContext.Data to store
	// the provider's dto.ExecutionOutput. Future steps (like templates or delivery)
	// will consume this payload.
	ContextExecutionOutput = "executionOutput"
)

// ExecutionContext represents the state of a single Task execution as it flows
// through the pipeline. Every step can read from or enrich the Data map.
type ExecutionContext struct {
	Task     *entities.Task
	Job      *entities.Job
	App      *entities.App
	Provider interfaces.Provider
	Data     map[string]any
}
