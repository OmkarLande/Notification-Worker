package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/OmkarLande/notification-worker/internal/contracts"
	"github.com/OmkarLande/notification-worker/internal/logger"
	"github.com/OmkarLande/notification-worker/internal/providers/dto"
)

// InsightService generates human-readable insights from provider execution outputs.
// In Phase 5A, this uses deterministic, rule-based logic. Future phases may integrate AI.
type InsightService struct {
	log logger.Logger
}

func NewInsightService(log logger.Logger) *InsightService {
	return &InsightService{log: log}
}

// Generate processes the raw ExecutionOutput and returns a structured InsightResult.
func (s *InsightService) Generate(ctx context.Context, output *dto.ExecutionOutput) (*contracts.InsightResult, error) {
	if output == nil {
		return nil, fmt.Errorf("insight service: execution output is nil")
	}

	// This is a naive, deterministic implementation.
	// We marshal the payload to inspect it loosely, or we could type-switch.
	// For demonstration, we'll return a generic success insight unless we detect "PendingTasks"
	
	bytes, err := json.Marshal(output.Payload)
	if err != nil {
		return nil, fmt.Errorf("insight service: failed to process payload: %w", err)
	}

	payloadStr := string(bytes)
	
	var severity contracts.InsightSeverity = contracts.SeverityInfo
	var title = "Execution Insight"
	var summary = "Your data has been processed successfully."

	// Basic rule-based analysis (mock logic)
	if len(payloadStr) > 0 {
		title = "Daily Activity Processed"
		summary = "We've compiled your latest data. Keep up the good momentum!"
		severity = contracts.SeveritySuccess
	}

	return &contracts.InsightResult{
		Title:    title,
		Summary:  summary,
		Severity: severity,
		Metadata: map[string]any{"generated_at": output.Duration},
	}, nil
}
