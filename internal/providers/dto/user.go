// Package dto contains data transfer objects used by provider implementations.
// These types flow between the provider and the worker engine and must never
// expose or reference database models directly.
package dto

// User represents a notification-enabled user returned by a provider.
type User struct {
	ID    int
	Name  string
	Email string
}
