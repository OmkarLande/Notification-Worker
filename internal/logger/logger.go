// Package logger provides a thin, replaceable logging abstraction for the
// Notification Worker. It wraps the standard library's log/slog package and
// exposes a Logger interface so the underlying implementation can be swapped
// for Zap, Zerolog, or any other library without touching application code.
package logger

import (
	"context"
	"log/slog"
	"os"
)

// Logger is the logging contract used throughout the application.
// All methods are safe for concurrent use.
type Logger interface {
	// Info logs a message at INFO level with optional key-value attributes.
	Info(msg string, args ...any)

	// Debug logs a message at DEBUG level with optional key-value attributes.
	Debug(msg string, args ...any)

	// Warn logs a message at WARN level with optional key-value attributes.
	Warn(msg string, args ...any)

	// Error logs a message at ERROR level with optional key-value attributes.
	Error(msg string, args ...any)

	// Fatal logs a message at ERROR level and then terminates the process.
	Fatal(msg string, args ...any)

	// With returns a new Logger that includes the given key-value pairs in
	// every subsequent log record. Useful for adding request-scoped context.
	With(args ...any) Logger
}

// slogLogger is the default Logger implementation backed by log/slog.
type slogLogger struct {
	inner *slog.Logger
}

// New creates a Logger configured for the given environment.
//   - "production": JSON output to stdout at INFO level.
//   - anything else: human-readable text output to stdout at DEBUG level.
func New(env string) Logger {
	var handler slog.Handler

	if env == "production" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	}

	return &slogLogger{inner: slog.New(handler)}
}

// Info logs a message at INFO level.
func (l *slogLogger) Info(msg string, args ...any) {
	l.inner.InfoContext(context.Background(), msg, args...)
}

// Debug logs a message at DEBUG level.
func (l *slogLogger) Debug(msg string, args ...any) {
	l.inner.DebugContext(context.Background(), msg, args...)
}

// Warn logs a message at WARN level.
func (l *slogLogger) Warn(msg string, args ...any) {
	l.inner.WarnContext(context.Background(), msg, args...)
}

// Error logs a message at ERROR level.
func (l *slogLogger) Error(msg string, args ...any) {
	l.inner.ErrorContext(context.Background(), msg, args...)
}

// Fatal logs a message at ERROR level and terminates the process with exit
// code 1. Use only for unrecoverable startup failures.
func (l *slogLogger) Fatal(msg string, args ...any) {
	l.inner.ErrorContext(context.Background(), msg, args...)
	os.Exit(1)
}

// With returns a new Logger that always includes the given attributes.
func (l *slogLogger) With(args ...any) Logger {
	return &slogLogger{inner: l.inner.With(args...)}
}
