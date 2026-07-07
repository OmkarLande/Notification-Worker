package config

import (
	"errors"
	"strings"
)

// Validate performs a final consistency check on the fully loaded AppConfig.
// It collects all validation errors rather than returning on the first failure,
// giving the operator a complete picture of what needs to be fixed.
func Validate(cfg *AppConfig) error {
	var errs []string

	// Worker
	if cfg.Worker.ShutdownTimeout <= 0 {
		errs = append(errs, "worker: SHUTDOWN_TIMEOUT_SECONDS must be a positive integer")
	}
	if cfg.Worker.WorkerID == "" {
		errs = append(errs, "worker: WORKER_ID must not be empty")
	}

	// Database
	if cfg.Database.URL == "" {
		errs = append(errs, "database: DB_URL is required")
	}
	if cfg.Database.MaxConns < 1 {
		errs = append(errs, "database: DB_MAX_CONNS must be >= 1")
	}

	// SMTP — host, user, password, from are required
	if cfg.SMTP.Host == "" {
		errs = append(errs, "smtp: SMTP_HOST is required")
	}
	if cfg.SMTP.User == "" {
		errs = append(errs, "smtp: SMTP_USER is required")
	}
	if cfg.SMTP.Password == "" {
		errs = append(errs, "smtp: SMTP_PASSWORD is required")
	}
	if cfg.SMTP.From == "" {
		errs = append(errs, "smtp: SMTP_FROM is required")
	}
	if cfg.SMTP.Port <= 0 || cfg.SMTP.Port > 65535 {
		errs = append(errs, "smtp: SMTP_PORT must be a valid port number (1–65535)")
	}

	// S3
	if cfg.S3.Region == "" {
		errs = append(errs, "s3: AWS_REGION is required")
	}
	if cfg.S3.AccessKeyID == "" {
		errs = append(errs, "s3: AWS_ACCESS_KEY_ID is required")
	}
	if cfg.S3.SecretAccessKey == "" {
		errs = append(errs, "s3: AWS_SECRET_ACCESS_KEY is required")
	}
	if cfg.S3.Bucket == "" {
		errs = append(errs, "s3: AWS_S3_BUCKET is required")
	}
	if cfg.S3.MaxUploadSizeMB <= 0 {
		errs = append(errs, "s3: PROFILE_IMAGE_MAX_SIZE_MB must be a positive integer")
	}

	if len(errs) == 0 {
		return nil
	}

	return errors.New("configuration validation failed:\n  - " + strings.Join(errs, "\n  - "))
}
