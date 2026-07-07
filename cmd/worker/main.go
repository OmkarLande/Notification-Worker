package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	"github.com/OmkarLande/notification-worker/internal/app"
	"github.com/OmkarLande/notification-worker/internal/config"
)

func main() {
	// Load .env file if present. Non-fatal — production environments inject
	// variables directly and will not have a .env file.
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found; falling back to system environment variables")
	}

	// 1. Load and validate configuration. Fail immediately on any missing or
	//    invalid required setting so infrastructure issues surface early.
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Configuration error: %v", err)
	}

	// 2. Wire all infrastructure and build the application.
	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("❌ Startup failed: %v", err)
	}

	// 3. Log the startup banner.
	application.Start()

	// 4. Block until the OS sends an interrupt or termination signal.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	// 5. Graceful shutdown: give in-flight work time to complete.
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Worker.ShutdownTimeout)
	defer cancel()

	application.Shutdown(ctx)
}