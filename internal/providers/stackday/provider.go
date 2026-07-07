// Package stackday provides the notification provider for the Stackday application.
// The provider connects to Stackday's PostgreSQL database in its constructor so
// that factory.Get("stackday") always returns a live, ready-to-use provider.
package stackday

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	entities "github.com/OmkarLande/notification-worker/internal/entites"
	"github.com/OmkarLande/notification-worker/internal/logger"
	"github.com/OmkarLande/notification-worker/internal/providers/dto"
)

// Provider is the Stackday notification provider. It connects to the Stackday
// PostgreSQL database and returns mock data in Phase 3 (real queries in Phase 4).
type Provider struct {
	pool *pgxpool.Pool
	log  logger.Logger
}

// New constructs and initializes the Stackday provider. It connects to the
// database using connectionString and pings it to verify connectivity.
// Returns an error if the connection cannot be established.
//
// In production, connectionString comes from apps.connection_string in the DB.
// For Phase 3 development, it reuses the worker's own connection string.
func New(connectionString string, log logger.Logger) (*Provider, error) {
	if connectionString == "" {
		return nil, fmt.Errorf("stackday provider: connection string is required")
	}

	pool, err := pgxpool.New(context.Background(), connectionString)
	if err != nil {
		return nil, fmt.Errorf("stackday provider: failed to create pool: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("stackday provider: ping failed: %w", err)
	}

	log.Info("Stackday provider initialized")
	return &Provider{pool: pool, log: log}, nil
}

// Name returns the provider's registry key — must match apps.name.
func (p *Provider) Name() string { return "stackday" }

// GetNotificationEnabledUsers returns users who should receive notifications.
// Phase 3: returns mock users. Phase 4 will query the Stackday database.
func (p *Provider) GetNotificationEnabledUsers(ctx context.Context) ([]dto.User, error) {
	p.log.Debug("Stackday: fetching notification-enabled users (mock)")
	return []dto.User{
		{ID: 101, Name: "Alice Johnson", Email: "alice@stackday.com"},
		{ID: 102, Name: "Bob Smith", Email: "bob@stackday.com"},
		{ID: 103, Name: "Carol White", Email: "carol@stackday.com"},
	}, nil
}

// Execute dispatches the task to the appropriate private method based on the
// job name. The worker remains generic — all Stackday-specific routing lives here.
func (p *Provider) Execute(ctx context.Context, ec entities.ExecutionContext) (*dto.ExecutionOutput, error) {
	start := time.Now()
	switch ec.Job.Name {
	case "Daily Digest":
		return p.executeDailyDigest(ctx, ec, start)
	case "Monthly Summary":
		return p.executeMonthlySummary(ctx, ec, start)
	default:
		return nil, fmt.Errorf("stackday provider: unknown job %q", ec.Job.Name)
	}
}

func (p *Provider) executeDailyDigest(_ context.Context, ec entities.ExecutionContext, start time.Time) (*dto.ExecutionOutput, error) {
	var args struct {
		UserID int `json:"user_id"`
	}
	if err := json.Unmarshal(ec.Task.Arguments, &args); err != nil {
		return nil, fmt.Errorf("stackday: daily digest: unmarshal args: %w", err)
	}

	p.log.Info("Executing Daily Digest",
		"user_id", args.UserID, "provider", "stackday", "job", ec.Job.Name)

	digest := dto.DailyDigestData{
		UserID: args.UserID,
		PendingTasks: []dto.PendingTask{
			{ID: 1, Title: "Review PR #42", DueAt: time.Now().Add(2 * time.Hour)},
			{ID: 2, Title: "Update documentation", DueAt: time.Now().Add(24 * time.Hour)},
		},
		Goals:       []dto.GoalProgress{{ID: 1, Title: "Complete sprint", Progress: 0.65}},
		GeneratedAt: time.Now(),
	}

	return &dto.ExecutionOutput{
		Payload: digest,
		Metadata: map[string]any{
			"provider": "stackday", "job": ec.Job.Name,
			"user_id": args.UserID, "pending_tasks": len(digest.PendingTasks),
		},
		Duration: time.Since(start),
	}, nil
}

func (p *Provider) executeMonthlySummary(_ context.Context, ec entities.ExecutionContext, start time.Time) (*dto.ExecutionOutput, error) {
	var args struct {
		UserID int `json:"user_id"`
	}
	if err := json.Unmarshal(ec.Task.Arguments, &args); err != nil {
		return nil, fmt.Errorf("stackday: monthly summary: unmarshal args: %w", err)
	}

	now := time.Now()
	p.log.Info("Executing Monthly Summary", "user_id", args.UserID, "provider", "stackday")
	return &dto.ExecutionOutput{
		Payload:  dto.MonthlySummaryData{UserID: args.UserID, Month: now.Month(), Year: now.Year(), Summary: "Mock — Phase 4"},
		Metadata: map[string]any{"provider": "stackday", "month": now.Month().String()},
		Duration: time.Since(start),
	}, nil
}

// Health pings the Stackday database connection.
func (p *Provider) Health(ctx context.Context) error {
	if err := p.pool.Ping(ctx); err != nil {
		return fmt.Errorf("stackday provider: health check failed: %w", err)
	}
	return nil
}
