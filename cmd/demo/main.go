// demo is a CLI entry point for manually triggering the Phase 3 execution
// pipeline. It seeds one full cycle — job trigger + task dispatch — and exits.
//
// Usage:
//
//	go run ./cmd/demo/
//
// Prerequisites:
//  1. Docker Postgres is running (see docker-compose.yml)
//  2. The seed script has been applied:
//     docker exec -i notification_postgres psql -U postgres -d notification_db < scripts/seed.sql
package main

import (
	"context"
	"log"
	"time"

	"github.com/joho/godotenv"

	"github.com/OmkarLande/notification-worker/internal/app"
	"github.com/OmkarLande/notification-worker/internal/config"
)

func main() {
	// Load .env — required for local development.
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found; using system environment variables")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Config error: %v", err)
	}

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("❌ Startup failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	c := application.Container()
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("🚀 Notification Worker — Phase 3 Demo")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Step 1: Trigger job ID 1 (seeded Daily Digest).
	const jobID = 1
	log.Printf("▶  Triggering job ID %d...\n", jobID)
	if err := c.JobExecutionService.TriggerJob(ctx, jobID); err != nil {
		log.Fatalf("❌ TriggerJob failed: %v", err)
	}
	log.Println("✅ Job triggered — tasks created with status NeedToPick")

	// Step 2: Load job to obtain MaxThreadCount for the dispatcher.
	job, err := c.Repos.Jobs.GetByID(ctx, jobID)
	if err != nil {
		log.Fatalf("❌ Failed to load job for dispatcher: %v", err)
	}

	// Step 3: Run one dispatch cycle.
	log.Printf("▶  Dispatching tasks (max_workers=%d)...\n", job.MaxThreadCount)
	if err := c.Dispatcher.Run(ctx, job.MaxThreadCount); err != nil {
		log.Fatalf("❌ Dispatch failed: %v", err)
	}

	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("✅ Demo complete — check task_logs for error details")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	application.Shutdown(ctx)
}
