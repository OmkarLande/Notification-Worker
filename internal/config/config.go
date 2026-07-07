// Package config provides the application configuration system for the
// Notification Worker. It loads settings from environment variables, groups
// them into strongly typed structs, and validates required values before
// returning a single AppConfig that the rest of the application depends on.
//
// Usage:
//
//	cfg, err := config.Load()
//	if err != nil {
//	    log.Fatal(err)
//	}
package config

import "fmt"

// AppConfig is the single source of truth for all application settings.
// Every sub-config is loaded and validated before AppConfig is returned.
type AppConfig struct {
	// Worker holds general application and worker identity settings.
	Worker WorkerConfig

	// Database holds PostgreSQL connection and pool settings.
	Database DatabaseConfig

	// SMTP holds email delivery settings.
	SMTP SMTPConfig

	// Firebase holds Firebase Cloud Messaging settings (optional at startup).
	Firebase FirebaseConfig

	// S3 holds AWS S3 storage settings.
	S3 S3Config

	// N8N holds n8n workflow automation settings (optional at startup).
	N8N N8NConfig
}

// Load reads every configuration group from the environment, assembles a
// complete AppConfig, and validates it. It returns a descriptive error if any
// required setting is missing or invalid so that the application fails fast
// before attempting to connect to external services.
func Load() (*AppConfig, error) {
	worker, err := loadWorkerConfig()
	if err != nil {
		return nil, fmt.Errorf("config: worker: %w", err)
	}

	db, err := loadDatabaseConfig()
	if err != nil {
		return nil, fmt.Errorf("config: database: %w", err)
	}

	smtp, err := loadSMTPConfig()
	if err != nil {
		return nil, fmt.Errorf("config: smtp: %w", err)
	}

	firebase, err := loadFirebaseConfig()
	if err != nil {
		return nil, fmt.Errorf("config: firebase: %w", err)
	}

	s3, err := loadS3Config()
	if err != nil {
		return nil, fmt.Errorf("config: s3: %w", err)
	}

	n8n, err := loadN8NConfig()
	if err != nil {
		return nil, fmt.Errorf("config: n8n: %w", err)
	}

	cfg := &AppConfig{
		Worker:   worker,
		Database: db,
		SMTP:     smtp,
		Firebase: firebase,
		S3:       s3,
		N8N:      n8n,
	}

	if err := Validate(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
