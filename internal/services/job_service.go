// Package services contains the application service layer. Services coordinate
// repositories and domain logic without directly touching the database.
package services

import (
	"context"
	"fmt"

	"github.com/OmkarLande/notification-worker/internal/cache"
	entities "github.com/OmkarLande/notification-worker/internal/entites"
	"github.com/OmkarLande/notification-worker/internal/interfaces"
	"github.com/OmkarLande/notification-worker/internal/logger"
)

// JobService is responsible for validating a job and its owning app before
// execution begins. It contains no provider knowledge — it only enforces
// pre-execution invariants.
type JobService struct {
	jobRepo     interfaces.JobRepository
	appRepo     interfaces.AppRepository
	statusCache *cache.StatusCache
	log         logger.Logger
}

// NewJobService constructs a JobService with its required dependencies.
func NewJobService(
	jobRepo interfaces.JobRepository,
	appRepo interfaces.AppRepository,
	statusCache *cache.StatusCache,
	log logger.Logger,
) *JobService {
	return &JobService{
		jobRepo:     jobRepo,
		appRepo:     appRepo,
		statusCache: statusCache,
		log:         log,
	}
}

// GetValidatedJob loads the job and its app, then verifies that execution is
// permitted. Returns descriptive errors so callers can surface them cleanly.
func (s *JobService) GetValidatedJob(ctx context.Context, jobID int) (*entities.Job, *entities.App, error) {
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return nil, nil, fmt.Errorf("job service: %w", err)
	}

	activeStatusID := s.statusCache.JobStatusID("Active")
	if job.StatusID != activeStatusID {
		return nil, nil, fmt.Errorf("job service: job %d is not active (status_id=%d)", jobID, job.StatusID)
	}

	app, err := s.appRepo.GetByID(ctx, job.AppID)
	if err != nil {
		return nil, nil, fmt.Errorf("job service: %w", err)
	}

	if app.MaintenanceMode {
		return nil, nil, fmt.Errorf("job service: app %q is in maintenance mode — skipping job %d", app.Name, jobID)
	}

	s.log.Info("Job validated", "job_id", job.ID, "job", job.Name, "app", app.Name)
	return job, app, nil
}