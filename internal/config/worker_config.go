package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// WorkerConfig holds general application and worker identity settings.
type WorkerConfig struct {
	// AppEnv is the deployment environment (e.g. "development", "production").
	AppEnv string

	// Port is the HTTP port the worker listens on.
	Port string

	// WorkerID uniquely identifies this worker instance. Useful when running
	// multiple workers in parallel so logs and metrics can be correlated.
	WorkerID string

	// ShutdownTimeout is the maximum duration the worker waits for in-flight
	// work to complete before forcefully terminating.
	ShutdownTimeout time.Duration
}

// loadWorkerConfig reads worker-level settings from environment variables.
func loadWorkerConfig() (WorkerConfig, error) {
	timeoutSec := getEnvInt("SHUTDOWN_TIMEOUT_SECONDS", 30)
	if timeoutSec <= 0 {
		return WorkerConfig{}, fmt.Errorf("SHUTDOWN_TIMEOUT_SECONDS must be a positive integer, got %d", timeoutSec)
	}

	return WorkerConfig{
		AppEnv:          getEnv("APP_ENV", "development"),
		Port:            getEnv("PORT", "8090"),
		WorkerID:        getEnv("WORKER_ID", "notification-worker-01"),
		ShutdownTimeout: time.Duration(timeoutSec) * time.Second,
	}, nil
}

// getEnv returns the value of the environment variable named by key, or
// defaultVal if the variable is not set or empty.
func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

// getEnvInt returns the integer value of the environment variable named by
// key, or defaultVal if the variable is not set or cannot be parsed.
func getEnvInt(key string, defaultVal int) int {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return defaultVal
	}
	return n
}
