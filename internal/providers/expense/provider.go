// Package expense provides the notification provider for the Expense Tracker
// application. In Phase 3 it uses placeholder initialization (no real Firebase
// SDK). Real Firebase integration will be added in Phase 4.
package expense

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/OmkarLande/notification-worker/internal/config"
	entities "github.com/OmkarLande/notification-worker/internal/entites"
	"github.com/OmkarLande/notification-worker/internal/logger"
	"github.com/OmkarLande/notification-worker/internal/providers/dto"
)

// Provider is the Expense Tracker notification provider.
// Phase 3: uses placeholder Firebase init and mock data.
// Phase 4: will integrate with the Firebase Admin SDK.
type Provider struct {
	cfg config.FirebaseConfig
	log logger.Logger
}

// New constructs the Expense Tracker provider. Firebase initialization is
// deferred to Phase 4 — this phase only validates that config is available.
func New(cfg config.FirebaseConfig, log logger.Logger) *Provider {
	log.Info("Expense provider initialized (Firebase placeholder — Phase 3)")
	return &Provider{cfg: cfg, log: log}
}

// Name returns the provider's registry key — must match apps.name.
func (p *Provider) Name() string { return "expense" }

// GetNotificationEnabledUsers returns mock users.
// Phase 4 will query Firebase Firestore for real user data.
func (p *Provider) GetNotificationEnabledUsers(_ context.Context) ([]dto.User, error) {
	p.log.Debug("Expense: fetching notification-enabled users (mock)")
	return []dto.User{
		{ID: 201, Name: "David Lee", Email: "david@expense.com"},
		{ID: 202, Name: "Eva Green", Email: "eva@expense.com"},
	}, nil
}

// Execute dispatches the task to the appropriate private method.
func (p *Provider) Execute(ctx context.Context, ec entities.ExecutionContext) (*dto.ExecutionOutput, error) {
	start := time.Now()
	switch ec.Job.Name {
	case "Daily Digest":
		return p.executeDailyDigest(ctx, ec, start)
	case "Monthly Summary":
		return p.executeMonthlySummary(ctx, ec, start)
	default:
		return nil, fmt.Errorf("expense provider: unknown job %q", ec.Job.Name)
	}
}

func (p *Provider) executeDailyDigest(_ context.Context, ec entities.ExecutionContext, start time.Time) (*dto.ExecutionOutput, error) {
	var args struct {
		UserID int `json:"user_id"`
	}
	if err := json.Unmarshal(ec.Task.Arguments, &args); err != nil {
		return nil, fmt.Errorf("expense: daily digest: unmarshal args: %w", err)
	}

	p.log.Info("Executing Daily Digest",
		"user_id", args.UserID, "provider", "expense", "job", ec.Job.Name)

	digest := dto.DailyDigestData{
		UserID:      args.UserID,
		PendingTasks: []dto.PendingTask{{ID: 10, Title: "Submit expense report", DueAt: time.Now().Add(48 * time.Hour)}},
		Goals:       []dto.GoalProgress{{ID: 10, Title: "Monthly budget", Progress: 0.72}},
		GeneratedAt: time.Now(),
	}

	return &dto.ExecutionOutput{
		Payload:  digest,
		Metadata: map[string]any{"provider": "expense", "user_id": args.UserID},
		Duration: time.Since(start),
	}, nil
}

func (p *Provider) executeMonthlySummary(_ context.Context, ec entities.ExecutionContext, start time.Time) (*dto.ExecutionOutput, error) {
	var args struct {
		UserID int `json:"user_id"`
	}
	if err := json.Unmarshal(ec.Task.Arguments, &args); err != nil {
		return nil, fmt.Errorf("expense: monthly summary: unmarshal args: %w", err)
	}

	now := time.Now()
	p.log.Info("Executing Monthly Summary", "user_id", args.UserID, "provider", "expense")
	return &dto.ExecutionOutput{
		Payload:  dto.MonthlySummaryData{UserID: args.UserID, Month: now.Month(), Year: now.Year(), Summary: "Mock expense summary — Phase 4"},
		Metadata: map[string]any{"provider": "expense"},
		Duration: time.Since(start),
	}, nil
}

// Health returns nil — no real connection to check in Phase 3.
func (p *Provider) Health(_ context.Context) error { return nil }